// ==========================================================================
// OmniWire - WireGuard 服务层（使用纯 Go 实现）
// ==========================================================================

package wireguard

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/skip2/go-qrcode"

	"omniwire/api/v1/wireguard"
	"omniwire/internal/model/entity"
	"omniwire/internal/service/wgserver"
)

// StatusOutput WireGuard 状态输出
type StatusOutput struct {
	Running    bool
	Interface  string
	ListenPort int
	PublicKey  string
	PeerCount  int
}

// ConfigOutput WireGuard 配置输出
type ConfigOutput struct {
	Interface           string
	ListenPort          int
	PrivateKey          string
	PublicKey           string
	Address             string
	DNS                 string
	MTU                 int
	EndpointAddress     string
	PostUp              string
	PostDown            string
	EthDevice           string
	PersistentKeepalive int
	ClientAllowedIPs    string
	ProxyAddress        string
	LogLevel            string
}

// ConfigInput 配置输入
type ConfigInput struct {
	ListenPort          int
	Address             string
	DNS                 string
	MTU                 int
	EndpointAddress     string
	EthDevice           string
	PersistentKeepalive int
	ClientAllowedIPs    string
	ProxyAddress        string
	LogLevel            string
}

// PeerInput 客户端输入
type PeerInput struct {
	Name       string
	AllowedIPs string
	Enabled    bool
}

// Status 获取 WireGuard 服务状态
func Status(ctx context.Context) (*StatusOutput, error) {
	server := wgserver.GetServer()
	return &StatusOutput{
		Running:    server.IsRunning(),
		Interface:  server.GetInterfaceName(),
		ListenPort: server.GetListenPort(),
		PublicKey:  server.GetPublicKey(),
		PeerCount:  server.GetPeerCount(),
	}, nil
}

// Start 启动 WireGuard 服务
func Start(ctx context.Context) error {
	server := wgserver.GetServer()

	// 先初始化（确保密钥存在）
	if err := server.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化失败: %v", err)
	}

	config, err := GetConfig(ctx)
	if err != nil {
		return err
	}

	// 启动服务
	if err := server.Start(config.Interface, config.ListenPort, config.PrivateKey, config.Address, config.DNS); err != nil {
		return err
	}

	g.Log().Info(ctx, "[WireGuard] 服务已启动")
	return nil
}

// Stop 停止 WireGuard 服务
func Stop(ctx context.Context) error {
	server := wgserver.GetServer()
	if err := server.Stop(); err != nil {
		return err
	}

	g.Log().Info(ctx, "[WireGuard] 服务已停止")
	return nil
}

// Restart 重启 WireGuard 服务
func Restart(ctx context.Context) error {
	server := wgserver.GetServer()

	if server.IsRunning() {
		if err := server.Stop(); err != nil {
			g.Log().Warning(ctx, "停止服务时出错:", err)
		}
		time.Sleep(time.Second)
	}

	return Start(ctx)
}

// GetConfig 获取 WireGuard 配置
func GetConfig(ctx context.Context) (*ConfigOutput, error) {
	// 从数据库读取配置
	var config struct {
		InterfaceName       string
		ListenPort          int
		PrivateKey          string
		PublicKey           string
		Address             string
		Dns                 string
		Mtu                 int
		EndpointAddress     string
		EthDevice           string
		PersistentKeepalive int
		ClientAllowedIps    string
		ProxyAddress        string
		LogLevel            string
	}

	err := g.DB().Model("wireguard_config").Where("id", 1).Scan(&config)
	if err != nil {
		// 返回默认配置
		return &ConfigOutput{
			Interface:           "wg0",
			ListenPort:          g.Cfg().MustGet(ctx, "wireguard.listenPort", 51820).Int(),
			Address:             g.Cfg().MustGet(ctx, "wireguard.addressRange", "10.66.66.1/24").String(),
			DNS:                 g.Cfg().MustGet(ctx, "wireguard.dns", "1.1.1.1, 8.8.8.8").String(),
			MTU:                 g.Cfg().MustGet(ctx, "wireguard.mtu", 1420).Int(),
			EthDevice:           "eth0",
			PersistentKeepalive: 25,
			ClientAllowedIPs:    "0.0.0.0/0, ::/0",
			ProxyAddress:        ":50122",
			LogLevel:            "error",
		}, nil
	}

	// 如果 EndpointAddress 为空，尝试自动获取 (这里简单使用 placeholder 或配置值)
	endpoint := config.EndpointAddress
	if endpoint == "" { // 获取服务器公网 IP
		endpoint = g.Cfg().MustGet(ctx, "wireguard.endpoint", "YOUR_SERVER_IP").String()
	}

	return &ConfigOutput{
		Interface:           config.InterfaceName,
		ListenPort:          config.ListenPort,
		PrivateKey:          config.PrivateKey,
		PublicKey:           config.PublicKey,
		Address:             config.Address,
		DNS:                 config.Dns,
		MTU:                 config.Mtu,
		EndpointAddress:     endpoint,
		EthDevice:           config.EthDevice,
		PersistentKeepalive: config.PersistentKeepalive,
		ClientAllowedIPs:    config.ClientAllowedIps,
		ProxyAddress:        config.ProxyAddress,
		LogLevel:            config.LogLevel,
	}, nil
}

// UpdateConfig 更新 WireGuard 配置
func UpdateConfig(ctx context.Context, input *ConfigInput) error {
	g.Log().Infof(ctx, "[WireGuard] 更新配置请求: EndpointAddress='%s', Port=%d", input.EndpointAddress, input.ListenPort)
	// 更新数据库配置
	_, err := g.DB().Exec(ctx, `
		UPDATE wireguard_config SET 
			listen_port = ?, 
			address = ?, 
			dns = ?, 
			mtu = ?,
			endpoint_address = ?,
			eth_device = ?,
			persistent_keepalive = ?,
			client_allowed_ips = ?,
			proxy_address = ?,
			log_level = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, input.ListenPort, input.Address, input.DNS, input.MTU, input.EndpointAddress,
		input.EthDevice, input.PersistentKeepalive, input.ClientAllowedIPs, input.ProxyAddress, input.LogLevel)

	if err != nil {
		return fmt.Errorf("更新配置失败: %v", err)
	}

	// 如果服务正在运行，需要重启
	server := wgserver.GetServer()
	if server.IsRunning() {
		g.Log().Info(ctx, "[WireGuard] 配置已更新，需要重启服务生效")
	}

	return nil
}

// ... (CreatePeer, UpdatePeer, DeletePeer 等保持不变，这里省略以匹配替换范围)

// GetPeerConfig 获取客户端配置文件
func GetPeerConfig(ctx context.Context, id int) (string, error) {
	var peer entity.WireguardPeer
	err := g.DB().Model("wireguard_peer").Where("id", id).Scan(&peer)
	if err != nil {
		return "", fmt.Errorf("客户端不存在")
	}

	serverConfig, err := GetConfig(ctx)
	if err != nil {
		return "", err
	}

	endpoint := serverConfig.EndpointAddress
	g.Log().Infof(ctx, "[WireGuard] 生成客户端配置, EndpointAddress: '%s', DB原始值: '%s'", endpoint, serverConfig.EndpointAddress)

	listenPort := serverConfig.ListenPort
	allowedIPs := serverConfig.ClientAllowedIPs
	dns := serverConfig.DNS
	mtu := serverConfig.MTU
	persistentKeepalive := serverConfig.PersistentKeepalive

	// 构建客户端配置 (严格匹配用户要求的格式)
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s
MTU = %d

[Peer]
PublicKey = %s
AllowedIPs = %s
PersistentKeepalive = %d
Endpoint = %s:%d
`, peer.PrivateKey, peer.AllowedIps, dns, mtu, serverConfig.PublicKey, allowedIPs, persistentKeepalive, endpoint, listenPort)

	return config, nil
}

// GetPeers 获取客户端列表
func GetPeers(ctx context.Context) ([]*wireguard.PeerInfo, error) {
	peers := make([]*wireguard.PeerInfo, 0)

	// 从数据库获取客户端列表
	result, err := g.DB().Model("wireguard_peer").OrderDesc("id").All()
	if err != nil {
		return nil, err
	}

	// 获取运行时的客户端状态（包含实时流量和握手时间）
	server := wgserver.GetServer()
	runtimePeers := server.GetAllPeers()

	for _, row := range result {
		peer := &wireguard.PeerInfo{
			Id:            row["id"].Int(),
			Name:          row["name"].String(),
			PublicKey:     row["public_key"].String(),
			AllowedIPs:    row["allowed_ips"].String(),
			Enabled:       row["enabled"].Int() == 1,
			UploadLimit:   row["upload_limit"].Int64(),
			DownloadLimit: row["download_limit"].Int64(),
			TotalUpload:   row["total_upload"].Int64(),
			TotalDownload: row["total_download"].Int64(),
			CreatedAt:     row["created_at"].String(),
			UpdatedAt:     row["updated_at"].String(),
		}

		// 填充运行时状态（实时流量、握手时间、在线状态）
		if rp, ok := runtimePeers[peer.PublicKey]; ok {
			peer.Endpoint = rp.Endpoint
			if !rp.LastHandshake.IsZero() {
				peer.LatestHandshake = rp.LastHandshake.Format("2006-01-02 15:04:05")
				peer.Online = time.Since(rp.LastHandshake) < 3*time.Minute
			}
			peer.TransferRx = rp.TransferRx
			peer.TransferTx = rp.TransferTx
		}

		peers = append(peers, peer)
	}

	return peers, nil
}

// CreatePeer 创建客户端
func CreatePeer(ctx context.Context, input *PeerInput) (*entity.WireguardPeer, error) {
	// 生成密钥对 (Base64)
	privateKey, publicKey, err := wgserver.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥失败: %v", err)
	}

	// 分配 IP
	ip := input.AllowedIPs
	if ip == "" {
		ip, err = allocateIP(ctx)
		if err != nil {
			return nil, err
		}
	}

	peer := &entity.WireguardPeer{
		Name:       input.Name,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		AllowedIps: ip,
		Enabled:    1,
		CreatedAt:  gtime.Now(),
		UpdatedAt:  gtime.Now(),
	}

	// 保存到数据库 - 使用 OmitEmpty 跳过 Id=0，让 SQLite 自动生成 ID
	res, err := g.DB().Model("wireguard_peer").OmitEmpty().Insert(peer)
	if err != nil {
		return nil, fmt.Errorf("保存客户端失败: %v", err)
	}
	id, _ := res.LastInsertId()
	peer.Id = int(id)

	// 添加到运行时
	server := wgserver.GetServer()
	if server.IsRunning() {
		server.AddPeer(publicKey, ip)
	}

	g.Log().Infof(ctx, "[WireGuard] 创建客户端: %s (%s)", peer.Name, peer.AllowedIps)
	return peer, nil
}

// UpdatePeer 更新客户端
func UpdatePeer(ctx context.Context, id int, input *PeerInput) error {
	updateData := g.Map{
		"updated_at": time.Now(),
	}

	if input.Name != "" {
		updateData["name"] = input.Name
	}
	if input.AllowedIPs != "" {
		updateData["allowed_ips"] = input.AllowedIPs
	}
	updateData["enabled"] = func() int {
		if input.Enabled {
			return 1
		}
		return 0
	}()

	_, err := g.DB().Model("wireguard_peer").Where("id", id).Update(updateData)
	if err != nil {
		return fmt.Errorf("更新客户端失败: %v", err)
	}

	return nil
}

// DeletePeer 删除客户端
func DeletePeer(ctx context.Context, id int) error {
	// 获取客户端信息
	var peer entity.WireguardPeer
	err := g.DB().Model("wireguard_peer").Where("id", id).Scan(&peer)
	if err != nil {
		return fmt.Errorf("客户端不存在")
	}

	// 从数据库删除
	_, err = g.DB().Model("wireguard_peer").Where("id", id).Delete()
	if err != nil {
		return fmt.Errorf("删除客户端失败: %v", err)
	}

	// 从运行时移除
	server := wgserver.GetServer()
	server.RemovePeer(peer.PublicKey)

	g.Log().Infof(ctx, "[WireGuard] 删除客户端: %s", peer.Name)
	return nil
}

// GetPeerQRCode 获取客户端配置二维码
func GetPeerQRCode(ctx context.Context, id int) (string, error) {
	config, err := GetPeerConfig(ctx, id)
	if err != nil {
		return "", err
	}

	// 生成二维码
	qr, err := qrcode.Encode(config, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %v", err)
	}

	// 转换为 Base64
	base64Str := base64.StdEncoding.EncodeToString(qr)
	return "data:image/png;base64," + base64Str, nil
}

// ==================== 辅助函数 ====================

// allocateIP 分配 IP 地址
func allocateIP(ctx context.Context) (string, error) {
	addressRange := g.Cfg().MustGet(ctx, "wireguard.addressRange", "10.66.66.0/24").String()

	// 解析网段
	parts := splitCIDR(addressRange)
	if len(parts) != 2 {
		return "", fmt.Errorf("无效的地址范围: %s", addressRange)
	}

	baseIP := parts[0]
	mask := parts[1]
	ipParts := splitIP(baseIP)
	if len(ipParts) != 4 {
		return "", fmt.Errorf("无效的 IP 地址: %s", baseIP)
	}

	// 获取已使用的 IP
	usedIPsResult, err := g.DB().Model("wireguard_peer").Fields("allowed_ips").Array()
	usedSet := make(map[string]bool)
	if err == nil {
		for _, ip := range usedIPsResult {
			ipStr := ip.String()
			if idx := indexOf(ipStr, "/"); idx > 0 {
				usedSet[ipStr[:idx]] = true
			}
		}
	}

	// 分配新 IP (从 .2 开始, .1 是服务器)
	for i := 2; i < 255; i++ {
		newIP := fmt.Sprintf("%s.%s.%s.%d", ipParts[0], ipParts[1], ipParts[2], i)
		if !usedSet[newIP] {
			return fmt.Sprintf("%s/%s", newIP, mask), nil
		}
	}

	return "", fmt.Errorf("IP 地址已耗尽")
}

func splitCIDR(cidr string) []string {
	for i := 0; i < len(cidr); i++ {
		if cidr[i] == '/' {
			return []string{cidr[:i], cidr[i+1:]}
		}
	}
	return nil
}

func splitIP(ip string) []string {
	parts := make([]string, 0, 4)
	start := 0
	for i := 0; i < len(ip); i++ {
		if ip[i] == '.' {
			parts = append(parts, ip[start:i])
			start = i + 1
		}
	}
	parts = append(parts, ip[start:])
	return parts
}

func indexOf(s string, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "刚刚"
	} else if d < time.Hour {
		return fmt.Sprintf("%d 分钟前", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d 小时前", int(d.Hours()))
	}
	return fmt.Sprintf("%d 天前", int(d.Hours()/24))
}
