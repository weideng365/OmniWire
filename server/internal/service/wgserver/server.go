// ==========================================================================
// OmniWire - WireGuard 真实服务端 (基于 wireguard-go)
// ==========================================================================

package wgserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/curve25519"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

// WireGuardServer WireGuard 服务端
type WireGuardServer struct {
	mu         sync.RWMutex
	running    bool
	listenPort int
	privateKey string // Base64
	publicKey  string // Base64
	address    string
	mtu        int

	// WireGuard 核心组件
	dev *device.Device
	tun tun.Device

	peers  map[string]*Peer // 以公钥(Base64)为键
	ctx    context.Context
	cancel context.CancelFunc
	stats  *ServerStats

	// 连接日志监控
	lastHandshakes map[string]time.Time // 上次已知握手时间
	lastOnline     map[string]bool      // 上次在线状态
}

// Peer 客户端状态
type Peer struct {
	Name          string
	PublicKey     string
	PresharedKey  string
	AllowedIPs    string
	Endpoint      string
	LastHandshake time.Time
	TransferRx    int64
	TransferTx    int64
	Enabled       bool
}

// ServerStats 统计信息
type ServerStats struct {
	StartTime   time.Time
	TotalRx     int64
	TotalTx     int64
	Connections int64
}

var (
	instance *WireGuardServer
	once     sync.Once
)

// GetServer 获取单例
func GetServer() *WireGuardServer {
	once.Do(func() {
		instance = &WireGuardServer{
			peers: make(map[string]*Peer),
			stats: &ServerStats{},
		}
	})
	return instance
}

// Initialize 初始化
func (s *WireGuardServer) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 读取配置
	s.listenPort = g.Cfg().MustGet(ctx, "wireguard.listenPort", 51820).Int()
	s.address = g.Cfg().MustGet(ctx, "wireguard.addressRange", "10.66.66.1/24").String()
	s.mtu = g.Cfg().MustGet(ctx, "wireguard.mtu", 1420).Int()

	// 2. 加载或生成密钥
	var config struct {
		PrivateKey string
		PublicKey  string
	}
	err := g.DB().Model("wireguard_config").Where("id", 1).Scan(&config)
	if err != nil || config.PrivateKey == "" {
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			return err
		}
		s.privateKey = privateKey
		s.publicKey = publicKey

		// 保存到数据库
		_, _ = g.DB().Exec(ctx, `UPDATE wireguard_config SET private_key = ?, public_key = ? WHERE id = 1`, s.privateKey, s.publicKey)
	} else {
		s.privateKey = config.PrivateKey
		s.publicKey = config.PublicKey
	}

	// 3. 预加载 Peers
	// 实际启动时会再次加载以应用到 Device
	_ = s.loadPeersFromDB(ctx)

	g.Log().Infof(ctx, "[WireGuard] 初始化完成, 端口: %d, VPN地址: %s", s.listenPort, s.address)
	return nil
}

// Start 启动服务 (创建网卡 & 运行协议)
func (s *WireGuardServer) Start(interfaceName string, listenPort int, privateKey, address, dns string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("服务已在运行")
	}

	s.listenPort = listenPort
	s.privateKey = privateKey
	s.address = address
	s.peers = make(map[string]*Peer) // 清空缓存，重新加载

	// 1. 创建 TUN 设备 (跨平台)
	// Windows会自动创建 Wintun 适配器
	fmt.Printf("[WireGuard] Creating TUN device: %s\n", interfaceName)
	tunDevice, err := tun.CreateTUN(interfaceName, s.mtu)
	if err != nil {
		return fmt.Errorf("创建 TUN 设备失败: %v", err)
	}
	s.tun = tunDevice

	// 2. 获取真正的接口名称 (Windows上名字可能被重命名或GUID)
	realName, _ := tunDevice.Name()
	g.Log().Infof(context.Background(), "[WireGuard] TUN 设备已创建: %s", realName)

	// 3. 创建 WireGuard 实例
	fmt.Println("[DEBUG] Creating WireGuard device...")
	logger := device.NewLogger(device.LogLevelError, fmt.Sprintf("(%s) ", interfaceName))
	s.dev = device.NewDevice(tunDevice, conn.NewStdNetBind(), logger)
	fmt.Println("[DEBUG] WireGuard device created")

	// 4. 启动设备
	fmt.Println("[DEBUG] Calling dev.Up()...")
	if err := s.dev.Up(); err != nil {
		tunDevice.Close()
		return fmt.Errorf("启动 Device 失败: %v", err)
	}
	fmt.Println("[DEBUG] dev.Up() completed")

	// 5. 配置设备 (Private Key & Port)
	// wireguard-go 使用 IPC 文本协议配置
	// 密钥需要转换为 Hex
	fmt.Println("[DEBUG] Converting private key...")
	fmt.Printf("[DEBUG] privateKey length: %d\n", len(s.privateKey))
	hexPrivKey, err := base64ToHex(s.privateKey)
	if err != nil {
		fmt.Printf("[DEBUG] base64ToHex failed: %v\n", err)
		s.Stop()
		return fmt.Errorf("密钥格式错误: %v", err)
	}
	fmt.Printf("[DEBUG] hexPrivKey length: %d\n", len(hexPrivKey))

	fmt.Println("[DEBUG] Calling IpcSet()...")
	ipcConfig := fmt.Sprintf("private_key=%s\nlisten_port=%d\n", hexPrivKey, s.listenPort)
	if err := s.dev.IpcSet(ipcConfig); err != nil {
		fmt.Printf("[DEBUG] IpcSet failed: %v\n", err)
		s.Stop()
		return fmt.Errorf("配置 Device 失败: %v", err)
	}
	fmt.Println("[DEBUG] IpcSet() completed")

	// 6. 配置操作系统 IP 地址
	fmt.Println("[DEBUG] About to configure IP...")
	if runtime.GOOS == "windows" {
		// Windows: 使用 LUID API 配置 IP（更可靠）
		if err := configureIPWithLUID(tunDevice, s.address); err != nil {
			g.Log().Errorf(context.Background(), "[WireGuard] LUID 配置 IP 失败: %v，尝试备用方案...", err)
			// 备用方案：使用命令行
			if err := s.configureInterfaceIP(realName, s.address); err != nil {
				g.Log().Errorf(context.Background(), "[WireGuard] 备用方案也失败: %v", err)
			}
		}
	} else {
		// Linux/macOS: 使用命令行配置 IP
		if err := s.configureInterfaceIP(realName, s.address); err != nil {
			g.Log().Errorf(context.Background(), "配置 IP 失败: %v", err)
		}
	}

	// 7. 加载并应用所有 Clients
	if err := s.loadAndApplyPeers(context.Background()); err != nil {
		g.Log().Errorf(context.Background(), "加载 Clients 失败: %v", err)
	}

	s.running = true
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.stats.StartTime = time.Now()

	// 初始化连接日志监控状态
	s.lastHandshakes = make(map[string]time.Time)
	s.lastOnline = make(map[string]bool)

	// 启动连接监控 goroutine
	go s.monitorConnections()

	g.Log().Info(context.Background(), "[WireGuard] 服务启动成功!")
	return nil
}

// Stop 停止服务
func (s *WireGuardServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	// 关闭 Device
	if s.dev != nil {
		s.dev.Close()
		s.dev = nil
	}
	// 显式关闭 TUN 设备（Windows 上必须显式关闭，否则会残留）
	if s.tun != nil {
		s.tun.Close()
		s.tun = nil
	}

	s.running = false
	g.Log().Info(context.Background(), "[WireGuard] 服务已停止")
	return nil
}

// IsRunning 状态查询
func (s *WireGuardServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *WireGuardServer) GetInterfaceName() string {
	if s.tun != nil {
		name, _ := s.tun.Name()
		return name
	}
	return "omniwire" // default
}

func (s *WireGuardServer) GetListenPort() int {
	return s.listenPort
}

func (s *WireGuardServer) GetPublicKey() string {
	return s.publicKey
}

func (s *WireGuardServer) GetPeerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.peers)
}
func (s *WireGuardServer) IsPeerConnected(pubKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.peers[pubKey]; ok {
		// 3分钟内有握手算在线
		return time.Since(p.LastHandshake) < 3*time.Minute
	}
	return false
}

// GetAllPeers 获取所有客户端状态
func (s *WireGuardServer) GetAllPeers() map[string]*Peer {
	// 先刷新实时统计
	s.RefreshPeerStats()

	s.mu.RLock()
	defer s.mu.RUnlock()

	copyPeers := make(map[string]*Peer)
	for k, v := range s.peers {
		// 浅拷贝
		p := *v
		copyPeers[k] = &p
	}
	return copyPeers
}

// RefreshPeerStats 从 WireGuard 设备获取实时统计信息
func (s *WireGuardServer) RefreshPeerStats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.dev == nil {
		return
	}

	// 获取设备状态
	ipcData, err := s.dev.IpcGet()
	if err != nil {
		return
	}

	// 解析 IPC 输出
	// 格式: key=value\n，每个 peer 以 public_key= 开头
	var currentPubKey string
	var lastHandshakeSec, lastHandshakeNsec int64
	var rxBytes, txBytes int64

	for _, line := range strings.Split(ipcData, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		switch key {
		case "public_key":
			// 保存上一个 peer 的数据
			if currentPubKey != "" {
				s.updatePeerStats(currentPubKey, lastHandshakeSec, lastHandshakeNsec, rxBytes, txBytes)
			}
			// 开始新的 peer，将 hex 转为 base64
			if b64Key, err := hexToBase64(value); err == nil {
				currentPubKey = b64Key
			} else {
				currentPubKey = ""
			}
			lastHandshakeSec, lastHandshakeNsec, rxBytes, txBytes = 0, 0, 0, 0

		case "last_handshake_time_sec":
			lastHandshakeSec, _ = strconv.ParseInt(value, 10, 64)

		case "last_handshake_time_nsec":
			lastHandshakeNsec, _ = strconv.ParseInt(value, 10, 64)

		case "rx_bytes":
			rxBytes, _ = strconv.ParseInt(value, 10, 64)

		case "tx_bytes":
			txBytes, _ = strconv.ParseInt(value, 10, 64)
		}
	}

	// 保存最后一个 peer 的数据
	if currentPubKey != "" {
		s.updatePeerStats(currentPubKey, lastHandshakeSec, lastHandshakeNsec, rxBytes, txBytes)
	}
}

func (s *WireGuardServer) updatePeerStats(pubKey string, handshakeSec, handshakeNsec, rx, tx int64) {
	if peer, ok := s.peers[pubKey]; ok {
		if handshakeSec > 0 {
			peer.LastHandshake = time.Unix(handshakeSec, handshakeNsec)
		}
		peer.TransferRx = rx
		peer.TransferTx = tx
	}
}

// ==================== 内部逻辑 ====================

// configureInterfaceIP 配置网卡 IP (跨平台由 cmd 调用实现)
func (s *WireGuardServer) configureInterfaceIP(ifaceName, cidr string) error {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	// 如果 IP 等于网络地址（主机位全 0），自动修正为 .1
	if ip4 := ip.To4(); ip4 != nil {
		netIP := ipNet.IP.To4()
		// 检查主机位是否全 0
		mask := ipNet.Mask
		isNetworkAddr := true
		for i := 0; i < 4; i++ {
			if ip4[i]&^mask[i] != 0 {
				isNetworkAddr = false
				break
			}
		}
		if isNetworkAddr {
			copy(ip4, netIP)
			ip4[3] = 1
			ip = ip4
			g.Log().Infof(context.Background(), "[WireGuard] 检测到网络地址，自动修正服务端 IP 为: %s", ip.String())
		}
	}

	ipStr := ip.String()
	ones, _ := ipNet.Mask.Size()

	// Windows implementation
	if runtime.GOOS == "windows" {
		// 1. 使用 PowerShell 等待并查找 Wintun 网卡 (Go 的 net 包在 Windows 上可能无法正确识别 Wintun)
		g.Log().Infof(context.Background(), "[WireGuard] 等待网卡 %s 就绪...", ifaceName)

		// PowerShell 脚本：查找 WireGuard 接口的 InterfaceIndex
		// 优先通过名称匹配，其次通过 InterfaceDescription 包含 "WireGuard"
		findIfaceScript := fmt.Sprintf(`
			$ErrorActionPreference = "SilentlyContinue"
			$adapter = Get-NetAdapter | Where-Object { $_.Name -eq "%s" -or $_.InterfaceDescription -like "*WireGuard*" } | Select-Object -First 1
			if ($adapter) {
				Write-Output $adapter.ifIndex
			} else {
				Write-Output ""
			}
		`, ifaceName)

		var idxStr string
		for i := 0; i < 20; i++ { // 最多等待 10 秒
			time.Sleep(500 * time.Millisecond)

			cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", findIfaceScript)
			out, err := cmd.Output()
			if err == nil {
				idxStr = strings.TrimSpace(string(out))
				if idxStr != "" {
					g.Log().Infof(context.Background(), "[WireGuard] 网卡就绪: %s (Index: %s)", ifaceName, idxStr)
					break
				}
			}
			g.Log().Debugf(context.Background(), "[WireGuard] 等待网卡就绪 (%d/20)...", i+1)
		}

		if idxStr == "" {
			g.Log().Warningf(context.Background(), "[WireGuard] 无法获取网卡索引，将尝试使用名字: %s", ifaceName)
		}

		// 2. PowerShell 脚本: 配置 IP + 开启 NAT (动态识别网段)
		// ipNet.String() 会返回标准的 CIDR 网段，例如 "198.18.30.0/24"
		subnet := ipNet.String()
		natName := fmt.Sprintf("WG_NAT_%s", ifaceName) // 使用接口名防止冲突

		var psScript string
		if idxStr != "" {
			// 方案 A: 使用 PowerShell (更加智能的检测与配置)
			// 关键改进:
			// 1. 先禁用 DHCP 防止 APIPA (169.254.x.x) 分配
			// 2. 清理旧 IP 配置
			// 3. 设置新 IP
			// 4. 验证 IP 是否正确设置（非 APIPA）
			psScript = fmt.Sprintf(`
				$ErrorActionPreference = "Stop"
				$Idx = %s
				$TargetIP = "%s"
				$PrefixLen = %d

				try {
					# 步骤1: 禁用 DHCP 防止 APIPA 自动分配
					Write-Host "Disabling DHCP on interface $Idx..."
					Set-NetIPInterface -InterfaceIndex $Idx -Dhcp Disabled -ErrorAction SilentlyContinue

					# 步骤2: 检查并清理现有 IP 配置
					$cur = Get-NetIPAddress -InterfaceIndex $Idx -AddressFamily IPv4 -ErrorAction SilentlyContinue
					if ($cur) {
						# 检查是否已经是目标 IP
						$hasTargetIP = $cur | Where-Object { $_.IPAddress -eq $TargetIP -and $_.PrefixLength -eq $PrefixLen }
						if ($hasTargetIP) {
							Write-Host "IP already configured correctly: $TargetIP"
							exit 0
						}
						# 清理所有旧 IP (包括 APIPA)
						Write-Host "Removing old IP configuration..."
						$cur | Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue
					}

					# 步骤3: 等待系统处理完毕
					Start-Sleep -Milliseconds 800

					# 步骤4: 设置新 IP
					Write-Host "Setting new IP $TargetIP/$PrefixLen..."
					New-NetIPAddress -InterfaceIndex $Idx -IPAddress $TargetIP -PrefixLength $PrefixLen -Confirm:$false | Out-Null

					# 步骤5: 验证 IP 配置结果
					Start-Sleep -Milliseconds 500
					$verify = Get-NetIPAddress -InterfaceIndex $Idx -AddressFamily IPv4 -ErrorAction SilentlyContinue
					if (-not $verify) {
						throw "IP verification failed: no IP address found"
					}

					# 检查是否为 APIPA (169.254.x.x)
					$apipa = $verify | Where-Object { $_.IPAddress -like "169.254.*" }
					if ($apipa) {
						# 尝试再次删除 APIPA 并重新设置
						Write-Host "APIPA detected, removing and retrying..."
						$apipa | Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue
						Start-Sleep -Milliseconds 500
						New-NetIPAddress -InterfaceIndex $Idx -IPAddress $TargetIP -PrefixLength $PrefixLen -Confirm:$false | Out-Null
						Start-Sleep -Milliseconds 500
						$verify = Get-NetIPAddress -InterfaceIndex $Idx -AddressFamily IPv4 -ErrorAction SilentlyContinue
					}

					# 最终验证
					$finalIP = $verify | Where-Object { $_.IPAddress -eq $TargetIP }
					if (-not $finalIP) {
						$actualIPs = ($verify | ForEach-Object { $_.IPAddress }) -join ", "
						throw "IP mismatch: expected $TargetIP, got: $actualIPs"
					}

					Write-Host "IP configured successfully: $TargetIP"
				} catch {
					Write-Error $_.Exception.Message
					exit 1
				}

				# 自动配置 NAT (如果不存在)
				$subnet = "%s"
				if (Get-Command New-NetNat -ErrorAction SilentlyContinue) {
					$existingNat = Get-NetNat | Where-Object { $_.InternalIPInterfaceAddressPrefix -eq $subnet }
					if (-not $existingNat) {
						$natName = "%s"
						Write-Host "Creating NAT $natName..."
						try {
							New-NetNat -Name $natName -InternalIPInterfaceAddressPrefix $subnet -Confirm:$false -ErrorAction Stop
						} catch {
							Write-Host "NAT creation skipped: $_"
						}
					}
				}
			`, idxStr, ipStr, ones, subnet, natName)
		} else {
			// Fallback 使用 Name (逻辑类似，同样增加 DHCP 禁用和验证)
			psScript = fmt.Sprintf(`
				$ErrorActionPreference = "Stop"
				$Alias = "%s"
				$TargetIP = "%s"
				$PrefixLen = %d

				try {
					# 禁用 DHCP
					Set-NetIPInterface -InterfaceAlias $Alias -Dhcp Disabled -ErrorAction SilentlyContinue

					$cur = Get-NetIPAddress -InterfaceAlias $Alias -AddressFamily IPv4 -ErrorAction SilentlyContinue
					if ($cur) {
						$hasTargetIP = $cur | Where-Object { $_.IPAddress -eq $TargetIP -and $_.PrefixLength -eq $PrefixLen }
						if ($hasTargetIP) { exit 0 }
						$cur | Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue
					}
					Start-Sleep -Milliseconds 800
					New-NetIPAddress -InterfaceAlias $Alias -IPAddress $TargetIP -PrefixLength $PrefixLen -Confirm:$false | Out-Null

					# 验证
					Start-Sleep -Milliseconds 500
					$verify = Get-NetIPAddress -InterfaceAlias $Alias -AddressFamily IPv4 | Where-Object { $_.IPAddress -eq $TargetIP }
					if (-not $verify) {
						throw "IP verification failed"
					}
					Write-Host "IP configured: $TargetIP"
				} catch {
					Write-Error $_
					exit 1
				}
			`, ifaceName, ipStr, ones)
		}

		// 循环重试机制 (增加重试间隔)
		for i := 0; i < 5; i++ {
			// 首次等待 1 秒，之后每次增加 1.5 秒
			waitTime := time.Duration(1000+i*1500) * time.Millisecond
			time.Sleep(waitTime)

			cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
			g.Log().Debugf(context.Background(), "[WireGuard] PowerShell 配置 IP (Idx:%s) (%d/5)...", idxStr, i+1)

			// 捕获标准输出和错误输出
			out, err := cmd.CombinedOutput()
			outputStr := strings.TrimSpace(string(out))

			if err == nil {
				g.Log().Infof(context.Background(), "[WireGuard] Windows IP 配置成功: %s. Output: %s", ipStr, outputStr)
				return nil
			} else {
				// 记录详细错误以便调试
				g.Log().Warningf(context.Background(), "[WireGuard] PowerShell 尝试 %d/5 失败: %v. Output: %s", i+1, err, outputStr)
			}
		}

		// PowerShell 失败后，尝试使用 netsh 作为备选方案
		g.Log().Infof(context.Background(), "[WireGuard] PowerShell 配置失败，尝试使用 netsh...")

		// 计算子网掩码
		mask := net.CIDRMask(ones, 32)
		maskStr := fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])

		for i := 0; i < 3; i++ {
			time.Sleep(time.Duration(1000+i*1000) * time.Millisecond)

			// netsh interface ip set address name="接口名" static IP地址 子网掩码
			netshCmd := exec.Command("netsh", "interface", "ip", "set", "address",
				fmt.Sprintf("name=%s", ifaceName), "static", ipStr, maskStr)
			g.Log().Debugf(context.Background(), "[WireGuard] netsh 配置 IP (%d/3): %s", i+1, netshCmd.String())

			out, err := netshCmd.CombinedOutput()
			outputStr := strings.TrimSpace(string(out))

			if err == nil || strings.Contains(outputStr, "确定") || strings.Contains(outputStr, "Ok") {
				g.Log().Infof(context.Background(), "[WireGuard] netsh IP 配置成功: %s. Output: %s", ipStr, outputStr)
				return nil
			}
			g.Log().Warningf(context.Background(), "[WireGuard] netsh 尝试 %d/3 失败: %v. Output: %s", i+1, err, outputStr)
		}

		return fmt.Errorf("Windows IP 配置最终失败，请检查: 1) OmniWire 是否以管理员权限运行 2) 网卡 %s 是否存在 3) 手动配置 IP: %s/%d", ifaceName, ipStr, ones)

	} else if runtime.GOOS == "linux" {
		// ip address add 10.66.66.1/24 dev wg0
		addr := fmt.Sprintf("%s/%d", ipStr, ones)
		cmd := exec.Command("ip", "address", "add", addr, "dev", ifaceName)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ip addr error: %v, output: %s", err, string(out))
		}
		// ip link set up dev wg0
		cmdUp := exec.Command("ip", "link", "set", "up", "dev", ifaceName)
		if out, err := cmdUp.CombinedOutput(); err != nil {
			return fmt.Errorf("ip link error: %v, output: %s", err, string(out))
		}
		return nil
	} else if runtime.GOOS == "darwin" {
		// macOS: ifconfig utun0 10.0.0.1/24 up
		// 或者: ifconfig utun0 inet 10.0.0.1 10.0.0.1 up
		// 尝试使用 CIDR 格式
		cmd := exec.Command("ifconfig", ifaceName, "inet", fmt.Sprintf("%s/%d", ipStr, ones), "up")
		g.Log().Debugf(context.Background(), "Exec: %s", cmd.String())
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ifconfig error: %v, output: %s", err, string(out))
		}

		// macOS 可能需要显式添加路由，如果需要作为网关
		// 由于 wireguard-go/tun 会自动处理路由表，这里只需确保接口 UP 且有 IP
		return nil
	}

	return fmt.Errorf("unsupported os: %s", runtime.GOOS)
}

// loadPeersFromDB 读取数据库
func (s *WireGuardServer) loadPeersFromDB(ctx context.Context) error {
	type PeerRecord struct {
		Name       string
		PublicKey  string
		AllowedIps string
		Enabled    int
	}
	var records []PeerRecord
	if err := g.DB().Model("wireguard_peer").Scan(&records); err != nil {
		return err
	}
	for _, r := range records {
		s.peers[r.PublicKey] = &Peer{
			Name:       r.Name,
			PublicKey:  r.PublicKey,
			AllowedIPs: r.AllowedIps,
			Enabled:    r.Enabled == 1,
		}
	}
	return nil
}

// loadAndApplyPeers 加载并下发配置给 Device
func (s *WireGuardServer) loadAndApplyPeers(ctx context.Context) error {
	if err := s.loadPeersFromDB(ctx); err != nil {
		return err
	}

	// 构建 IPC 字符串
	// public_key=...
	// allowed_ip=...
	// endpoint=... (可选)
	var ipcBuilder strings.Builder
	for _, p := range s.peers {
		if !p.Enabled {
			continue // 跳过禁用的Peer
		}

		hexKey, err := base64ToHex(p.PublicKey)
		if err != nil {
			continue
		}

		ipcBuilder.WriteString(fmt.Sprintf("public_key=%s\n", hexKey))

		// Allowed IPs
		for _, cidr := range strings.Split(p.AllowedIPs, ",") {
			cidr = strings.TrimSpace(cidr)
			if cidr != "" {
				ipcBuilder.WriteString(fmt.Sprintf("allowed_ip=%s\n", cidr))
			}
		}
		// persistent_keepalive_interval=25
	}

	if s.dev != nil {
		return s.dev.IpcSet(ipcBuilder.String())
	}
	return nil
}

// AddPeer 动态添加 Peer
func (s *WireGuardServer) AddPeer(publicKey, allowedIPs string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 内存记录
	s.peers[publicKey] = &Peer{
		PublicKey:  publicKey,
		AllowedIPs: allowedIPs,
		Enabled:    true,
	}

	if !s.running || s.dev == nil {
		return nil
	}

	// 下发配置
	hexKey, _ := base64ToHex(publicKey)
	ipc := fmt.Sprintf("public_key=%s\n", hexKey)
	for _, cidr := range strings.Split(allowedIPs, ",") {
		ipc += fmt.Sprintf("allowed_ip=%s\n", strings.TrimSpace(cidr))
	}
	return s.dev.IpcSet(ipc)
}

// RemovePeer 移除 Peer
func (s *WireGuardServer) RemovePeer(publicKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.peers, publicKey)

	if !s.running || s.dev == nil {
		return nil
	}

	// wireguard-go remove peer: public_key=... remove=true
	hexKey, _ := base64ToHex(publicKey)
	ipc := fmt.Sprintf("public_key=%s\nremove=true\n", hexKey)
	return s.dev.IpcSet(ipc)
}

// DisablePeer 禁用 Peer（从设备移除但保留内存记录）
func (s *WireGuardServer) DisablePeer(publicKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新内存状态
	if peer, ok := s.peers[publicKey]; ok {
		peer.Enabled = false
	}

	if !s.running || s.dev == nil {
		return nil
	}

	// 从设备移除
	hexKey, _ := base64ToHex(publicKey)
	ipc := fmt.Sprintf("public_key=%s\nremove=true\n", hexKey)
	return s.dev.IpcSet(ipc)
}

// EnablePeer 启用 Peer（添加到设备）
func (s *WireGuardServer) EnablePeer(publicKey, allowedIPs string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新内存状态
	if peer, ok := s.peers[publicKey]; ok {
		peer.Enabled = true
	} else {
		s.peers[publicKey] = &Peer{
			PublicKey:  publicKey,
			AllowedIPs: allowedIPs,
			Enabled:    true,
		}
	}

	if !s.running || s.dev == nil {
		return nil
	}

	// 添加到设备
	hexKey, _ := base64ToHex(publicKey)
	ipc := fmt.Sprintf("public_key=%s\n", hexKey)
	for _, cidr := range strings.Split(allowedIPs, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr != "" {
			ipc += fmt.Sprintf("allowed_ip=%s\n", cidr)
		}
	}
	return s.dev.IpcSet(ipc)
}

// ==================== 辅助工具 ====================

// GenerateKeyPair 生成 Base64 密钥对
func GenerateKeyPair() (string, string, error) {
	var privateKey [32]byte
	var publicKey [32]byte

	_, err := rand.Read(privateKey[:])
	if err != nil {
		return "", "", err
	}

	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return base64.StdEncoding.EncodeToString(privateKey[:]),
		base64.StdEncoding.EncodeToString(publicKey[:]), nil
}

func base64ToHex(b64 string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hexToBase64(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// monitorConnections 后台监控 Peer 连接状态并记录日志
func (s *WireGuardServer) monitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkConnectionChanges()
		}
	}
}

// checkConnectionChanges 检测连接状态变化并写入日志
func (s *WireGuardServer) checkConnectionChanges() {
	// 先刷新实时统计
	s.RefreshPeerStats()

	s.mu.RLock()
	// 复制当前 Peer 状态，避免长时间持锁
	type peerSnapshot struct {
		PublicKey     string
		Name          string
		Endpoint      string
		LastHandshake time.Time
		TransferRx    int64
		TransferTx    int64
	}
	snapshots := make([]peerSnapshot, 0, len(s.peers))
	for _, p := range s.peers {
		snapshots = append(snapshots, peerSnapshot{
			PublicKey:     p.PublicKey,
			Name:          p.Name,
			Endpoint:      p.Endpoint,
			LastHandshake: p.LastHandshake,
			TransferRx:    p.TransferRx,
			TransferTx:    p.TransferTx,
		})
	}
	s.mu.RUnlock()

	ctx := context.Background()

	// 查询 peer_id 映射（public_key -> id）
	type peerIdRow struct {
		Id        int
		PublicKey string
		Name      string
	}
	var peerIdRows []peerIdRow
	_ = g.DB().Model("wireguard_peer").Fields("id, public_key, name").Scan(&peerIdRows)
	peerIdMap := make(map[string]peerIdRow)
	for _, row := range peerIdRows {
		peerIdMap[row.PublicKey] = row
	}

	for _, snap := range snapshots {
		nowOnline := !snap.LastHandshake.IsZero() && time.Since(snap.LastHandshake) < 3*time.Minute
		prevHandshake, hasPrev := s.lastHandshakes[snap.PublicKey]
		wasOnline := s.lastOnline[snap.PublicKey]

		// 获取 peer_id 和 name（优先用数据库的，因为 runtime 可能没有 name）
		peerID := 0
		peerName := snap.Name
		if row, ok := peerIdMap[snap.PublicKey]; ok {
			peerID = row.Id
			if peerName == "" {
				peerName = row.Name
			}
		}

		if !snap.LastHandshake.IsZero() {
			// 握手时间发生变化 → 记录 handshake 事件
			if !hasPrev || !snap.LastHandshake.Equal(prevHandshake) {
				s.insertConnectionLog(ctx, peerID, peerName, snap.PublicKey, "handshake", snap.Endpoint, snap.TransferRx, snap.TransferTx)
			}

			// 从离线变为在线
			if nowOnline && !wasOnline {
				s.insertConnectionLog(ctx, peerID, peerName, snap.PublicKey, "online", snap.Endpoint, snap.TransferRx, snap.TransferTx)
			}
		}

		// 从在线变为离线
		if !nowOnline && wasOnline {
			s.insertConnectionLog(ctx, peerID, peerName, snap.PublicKey, "offline", snap.Endpoint, snap.TransferRx, snap.TransferTx)
		}

		// 更新追踪状态
		if !snap.LastHandshake.IsZero() {
			s.lastHandshakes[snap.PublicKey] = snap.LastHandshake
		}
		s.lastOnline[snap.PublicKey] = nowOnline
	}
}

// insertConnectionLog 插入连接日志记录
func (s *WireGuardServer) insertConnectionLog(ctx context.Context, peerID int, peerName, publicKey, event, endpoint string, rx, tx int64) {
	_, err := g.DB().Exec(ctx,
		`INSERT INTO wireguard_connection_log (peer_id, peer_name, public_key, event, endpoint, transfer_rx, transfer_tx) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		peerID, peerName, publicKey, event, endpoint, rx, tx,
	)
	if err != nil {
		g.Log().Warningf(ctx, "[WireGuard] 写入连接日志失败: %v", err)
	} else {
		g.Log().Debugf(ctx, "[WireGuard] 连接日志: %s [%s] %s endpoint=%s", peerName, event, publicKey[:8]+"...", endpoint)
	}
}
