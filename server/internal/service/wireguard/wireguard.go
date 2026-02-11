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
	AutoStart           bool
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
	AutoStart           bool
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
		AutoStart           int
	}

	err := g.DB().Model("wireguard_config").Where("id", 1).Scan(&config)
	if err != nil {
		// 返回默认配置
		return &ConfigOutput{
			Interface:           "omniwire",
			ListenPort:          g.Cfg().MustGet(ctx, "wireguard.listenPort", 51820).Int(),
			Address:             g.Cfg().MustGet(ctx, "wireguard.addressRange", "10.66.66.1/24").String(),
			DNS:                 g.Cfg().MustGet(ctx, "wireguard.dns", "223.5.5.5").String(),
			MTU:                 g.Cfg().MustGet(ctx, "wireguard.mtu", 1420).Int(),
			EthDevice:           "",
			PersistentKeepalive: 25,
			ClientAllowedIPs:    "0.0.0.0/0, ::/0",
			ProxyAddress:        ":50122",
			LogLevel:            "error",
			AutoStart:           false,
		}, nil
	}

	// 直接返回数据库中的 EndpointAddress，不做默认值替换
	// 这样前端可以正确显示和保存用户配置的值
	return &ConfigOutput{
		Interface:           config.InterfaceName,
		ListenPort:          config.ListenPort,
		PrivateKey:          config.PrivateKey,
		PublicKey:           config.PublicKey,
		Address:             config.Address,
		DNS:                 config.Dns,
		MTU:                 config.Mtu,
		EndpointAddress:     config.EndpointAddress,
		EthDevice:           config.EthDevice,
		PersistentKeepalive: config.PersistentKeepalive,
		ClientAllowedIPs:    config.ClientAllowedIps,
		ProxyAddress:        config.ProxyAddress,
		LogLevel:            config.LogLevel,
		AutoStart:           config.AutoStart == 1,
	}, nil
}

// UpdateConfig 更新 WireGuard 配置
func UpdateConfig(ctx context.Context, input *ConfigInput) error {
	g.Log().Infof(ctx, "[WireGuard] 更新配置请求: EndpointAddress='%s', Port=%d, AutoStart=%v, ClientAllowedIPs='%s'",
		input.EndpointAddress, input.ListenPort, input.AutoStart, input.ClientAllowedIPs)

	autoStart := 0
	if input.AutoStart {
		autoStart = 1
	}

	// 更新数据库配置
	result, err := g.DB().Exec(ctx, `
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
			auto_start = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, input.ListenPort, input.Address, input.DNS, input.MTU, input.EndpointAddress,
		input.EthDevice, input.PersistentKeepalive, input.ClientAllowedIPs, input.ProxyAddress, input.LogLevel, autoStart)

	if err != nil {
		g.Log().Errorf(ctx, "[WireGuard] 更新配置失败: %v", err)
		return fmt.Errorf("更新配置失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	g.Log().Infof(ctx, "[WireGuard] 配置更新完成, 影响行数: %d", rowsAffected)

	if rowsAffected == 0 {
		// 记录不存在，生成密钥并插入新记录
		privateKey, publicKey, keyErr := wgserver.GenerateKeyPair()
		if keyErr != nil {
			privateKey, publicKey = "", ""
		}
		_, err = g.DB().Exec(ctx, `
			INSERT INTO wireguard_config (id, private_key, public_key, listen_port, address, dns, mtu, endpoint_address, eth_device, persistent_keepalive, client_allowed_ips, proxy_address, log_level, auto_start)
			VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, privateKey, publicKey, input.ListenPort, input.Address, input.DNS, input.MTU, input.EndpointAddress,
			input.EthDevice, input.PersistentKeepalive, input.ClientAllowedIPs, input.ProxyAddress, input.LogLevel, autoStart)
		if err != nil {
			return fmt.Errorf("插入配置失败: %v", err)
		}
		g.Log().Infof(ctx, "[WireGuard] 配置记录不存在，已插入新记录")
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
	if err != nil || peer.Id == 0 {
		return "", fmt.Errorf("客户端不存在")
	}

	serverConfig, err := GetConfig(ctx)
	if err != nil {
		return "", err
	}

	// 检查公网地址是否已配置
	endpoint := serverConfig.EndpointAddress
	if endpoint == "" {
		return "", fmt.Errorf("请先在 WireGuard 配置中设置公网地址")
	}

	listenPort := serverConfig.ListenPort
	allowedIPs := serverConfig.ClientAllowedIPs
	dns := serverConfig.DNS
	mtu := serverConfig.MTU
	persistentKeepalive := serverConfig.PersistentKeepalive

	g.Log().Infof(ctx, "[WireGuard] 生成客户端配置, Endpoint: %s:%d, AllowedIPs: %s", endpoint, listenPort, allowedIPs)

	// 构建客户端配置
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
	// 先获取客户端信息（用于同步运行时状态）
	var peer entity.WireguardPeer
	err := g.DB().Model("wireguard_peer").Where("id", id).Scan(&peer)
	if err != nil {
		return fmt.Errorf("客户端不存在")
	}

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

	_, err = g.DB().Model("wireguard_peer").Where("id", id).Update(updateData)
	if err != nil {
		return fmt.Errorf("更新客户端失败: %v", err)
	}

	// 同步运行时状态
	server := wgserver.GetServer()
	allowedIPs := input.AllowedIPs
	if allowedIPs == "" {
		allowedIPs = peer.AllowedIps
	}

	if input.Enabled {
		// 启用：添加到 WireGuard 设备
		server.EnablePeer(peer.PublicKey, allowedIPs)
		g.Log().Infof(ctx, "[WireGuard] 启用客户端: %s", peer.Name)
	} else {
		// 禁用：从 WireGuard 设备移除
		server.DisablePeer(peer.PublicKey)
		g.Log().Infof(ctx, "[WireGuard] 禁用客户端: %s", peer.Name)
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

// GetConnectionLogs 获取连接日志
func GetConnectionLogs(ctx context.Context, peerId, page, pageSize int) ([]*wireguard.ConnectionLogInfo, int, error) {
	model := g.DB().Model("wireguard_connection_log")

	if peerId > 0 {
		model = model.Where("peer_id", peerId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("查询日志总数失败: %v", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []struct {
		Id         int
		PeerName   string
		Event      string
		Endpoint   string
		TransferRx int64
		TransferTx int64
		CreatedAt  string
	}

	err = g.DB().Model("wireguard_connection_log").
		Where(func() string {
			if peerId > 0 {
				return fmt.Sprintf("peer_id = %d", peerId)
			}
			return "1=1"
		}()).
		OrderDesc("id").
		Page(page, pageSize).
		Scan(&logs)
	if err != nil {
		return nil, 0, fmt.Errorf("查询日志失败: %v", err)
	}

	list := make([]*wireguard.ConnectionLogInfo, 0, len(logs))
	for _, row := range logs {
		list = append(list, &wireguard.ConnectionLogInfo{
			Id:         row.Id,
			PeerName:   row.PeerName,
			Event:      row.Event,
			Endpoint:   row.Endpoint,
			TransferRx: row.TransferRx,
			TransferTx: row.TransferTx,
			CreatedAt:  row.CreatedAt,
		})
	}

	return list, total, nil
}

// ==================== 辅助函数 ====================

// allocateIP 分配 IP 地址，从数据库读取服务端配置的网段
func allocateIP(ctx context.Context) (string, error) {
	// 从数据库读取服务端配置的 VPN 网段
	var addressRange string
	dbAddress, err := g.DB().Model("wireguard_config").Where("id", 1).Value("address")
	if err == nil && !dbAddress.IsEmpty() {
		addressRange = dbAddress.String()
	} else {
		// 回退到配置文件默认值
		addressRange = g.Cfg().MustGet(ctx, "wireguard.addressRange", "10.66.66.1/24").String()
	}

	// 解析网段（例如: "198.18.88.1/24"）
	parts := splitCIDR(addressRange)
	if len(parts) != 2 {
		return "", fmt.Errorf("无效的地址范围: %s", addressRange)
	}

	serverIP := parts[0]
	mask := parts[1]
	ipParts := splitIP(serverIP)
	if len(ipParts) != 4 {
		return "", fmt.Errorf("无效的 IP 地址: %s", serverIP)
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

	// 分配新 IP (从 .2 开始, .1 是服务器保留)
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

// InitWireGuard 初始化 WireGuard 服务（自动启动）
func InitWireGuard(ctx context.Context) {
	// 检查是否配置了自动启动
	var config struct {
		AutoStart int
	}
	err := g.DB().Model("wireguard_config").Where("id", 1).Scan(&config)
	if err != nil {
		g.Log().Debugf(ctx, "[WireGuard] 读取配置失败: %v", err)
		return
	}

	if config.AutoStart != 1 {
		g.Log().Info(ctx, "[WireGuard] 自动启动未开启，跳过")
		return
	}

	// 自动启动 WireGuard 服务
	g.Log().Info(ctx, "[WireGuard] 自动启动已开启，正在启动服务...")
	if err := Start(ctx); err != nil {
		g.Log().Errorf(ctx, "[WireGuard] 自动启动失败: %v", err)
	} else {
		g.Log().Info(ctx, "[WireGuard] 自动启动成功")
	}
}
