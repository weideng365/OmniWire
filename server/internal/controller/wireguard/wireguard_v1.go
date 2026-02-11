// ==========================================================================
// OmniWire - WireGuard 管理控制器
// ==========================================================================

package wireguard

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/wireguard"
	svcWireguard "omniwire/internal/service/wireguard"
)

// ControllerV1 WireGuard 管理控制器
type ControllerV1 struct{}

// NewV1 创建 WireGuard 管理控制器实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Status WireGuard 服务状态
func (c *ControllerV1) Status(ctx context.Context, req *wireguard.StatusReq) (res *wireguard.StatusRes, err error) {
	status, err := svcWireguard.Status(ctx)
	if err != nil {
		return nil, err
	}
	res = &wireguard.StatusRes{
		Running:    status.Running,
		Interface:  status.Interface,
		ListenPort: status.ListenPort,
		PublicKey:  status.PublicKey,
		PeerCount:  status.PeerCount,
	}
	return
}

// Start 启动 WireGuard 服务
func (c *ControllerV1) Start(ctx context.Context, req *wireguard.StartReq) (res *wireguard.StartRes, err error) {
	err = svcWireguard.Start(ctx)
	if err != nil {
		return nil, err
	}
	res = &wireguard.StartRes{Success: true}
	g.Log().Info(ctx, "WireGuard 服务已启动")
	return
}

// Stop 停止 WireGuard 服务
func (c *ControllerV1) Stop(ctx context.Context, req *wireguard.StopReq) (res *wireguard.StopRes, err error) {
	err = svcWireguard.Stop(ctx)
	if err != nil {
		return nil, err
	}
	res = &wireguard.StopRes{Success: true}
	g.Log().Info(ctx, "WireGuard 服务已停止")
	return
}

// Restart 重启 WireGuard 服务
func (c *ControllerV1) Restart(ctx context.Context, req *wireguard.RestartReq) (res *wireguard.RestartRes, err error) {
	err = svcWireguard.Restart(ctx)
	if err != nil {
		return nil, err
	}
	res = &wireguard.RestartRes{Success: true}
	g.Log().Info(ctx, "WireGuard 服务已重启")
	return
}

// Config 获取 WireGuard 配置
func (c *ControllerV1) Config(ctx context.Context, req *wireguard.ConfigReq) (res *wireguard.ConfigRes, err error) {
	config, err := svcWireguard.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	res = &wireguard.ConfigRes{
		Interface:           config.Interface,
		ListenPort:          config.ListenPort,
		PrivateKey:          config.PrivateKey,
		PublicKey:           config.PublicKey,
		Address:             config.Address,
		DNS:                 config.DNS,
		MTU:                 config.MTU,
		EndpointAddress:     config.EndpointAddress,
		PostUp:              config.PostUp,
		PostDown:            config.PostDown,
		EthDevice:           config.EthDevice,
		PersistentKeepalive: config.PersistentKeepalive,
		ClientAllowedIPs:    config.ClientAllowedIPs,
		ProxyAddress:        config.ProxyAddress,
		LogLevel:            config.LogLevel,
		AutoStart:           config.AutoStart,
	}
	return
}

// UpdateConfig 更新 WireGuard 配置
func (c *ControllerV1) UpdateConfig(ctx context.Context, req *wireguard.UpdateConfigReq) (res *wireguard.UpdateConfigRes, err error) {
	err = svcWireguard.UpdateConfig(ctx, &svcWireguard.ConfigInput{
		ListenPort:          req.ListenPort,
		EndpointAddress:     req.EndpointAddress,
		Address:             req.Address,
		DNS:                 req.DNS,
		MTU:                 req.MTU,
		EthDevice:           req.EthDevice,
		PersistentKeepalive: req.PersistentKeepalive,
		ClientAllowedIPs:    req.ClientAllowedIPs,
		ProxyAddress:        req.ProxyAddress,
		LogLevel:            req.LogLevel,
		AutoStart:           req.AutoStart,
	})
	if err != nil {
		return nil, err
	}
	res = &wireguard.UpdateConfigRes{Success: true}
	g.Log().Info(ctx, "WireGuard 配置已更新")
	return
}

// PeerList 获取客户端列表
func (c *ControllerV1) PeerList(ctx context.Context, req *wireguard.PeerListReq) (res *wireguard.PeerListRes, err error) {
	peers, err := svcWireguard.GetPeers(ctx)
	if err != nil {
		return nil, err
	}
	res = &wireguard.PeerListRes{
		Peers: peers,
	}
	return
}

// PeerCreate 创建客户端
func (c *ControllerV1) PeerCreate(ctx context.Context, req *wireguard.PeerCreateReq) (res *wireguard.PeerCreateRes, err error) {
	peer, err := svcWireguard.CreatePeer(ctx, &svcWireguard.PeerInput{
		Name:       req.Name,
		AllowedIPs: req.AllowedIPs,
	})
	if err != nil {
		return nil, err
	}
	res = &wireguard.PeerCreateRes{
		Peer: &wireguard.PeerInfo{
			Id:         peer.Id,
			Name:       peer.Name,
			PublicKey:  peer.PublicKey,
			AllowedIPs: peer.AllowedIps,
			Enabled:    peer.Enabled == 1,
			CreatedAt:  peer.CreatedAt.String(),
			UpdatedAt:  peer.UpdatedAt.String(),
		},
	}
	g.Log().Infof(ctx, "客户端 %s 已创建", req.Name)
	return
}

// PeerUpdate 更新客户端
func (c *ControllerV1) PeerUpdate(ctx context.Context, req *wireguard.PeerUpdateReq) (res *wireguard.PeerUpdateRes, err error) {
	err = svcWireguard.UpdatePeer(ctx, req.Id, &svcWireguard.PeerInput{
		Name:       req.Name,
		AllowedIPs: req.AllowedIPs,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	res = &wireguard.PeerUpdateRes{Success: true}
	g.Log().Infof(ctx, "客户端 %d 已更新", req.Id)
	return
}

// PeerDelete 删除客户端
func (c *ControllerV1) PeerDelete(ctx context.Context, req *wireguard.PeerDeleteReq) (res *wireguard.PeerDeleteRes, err error) {
	err = svcWireguard.DeletePeer(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &wireguard.PeerDeleteRes{Success: true}
	g.Log().Infof(ctx, "客户端 %d 已删除", req.Id)
	return
}

// PeerConfig 获取客户端配置文件
func (c *ControllerV1) PeerConfig(ctx context.Context, req *wireguard.PeerConfigReq) (res *wireguard.PeerConfigRes, err error) {
	config, err := svcWireguard.GetPeerConfig(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &wireguard.PeerConfigRes{
		Config: config,
	}
	return
}

// PeerQRCode 获取客户端配置二维码
func (c *ControllerV1) PeerQRCode(ctx context.Context, req *wireguard.PeerQRCodeReq) (res *wireguard.PeerQRCodeRes, err error) {
	qrcode, err := svcWireguard.GetPeerQRCode(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &wireguard.PeerQRCodeRes{
		QRCode: qrcode,
	}
	return
}

// ConnectionLogs 获取连接日志
func (c *ControllerV1) ConnectionLogs(ctx context.Context, req *wireguard.ConnectionLogsReq) (res *wireguard.ConnectionLogsRes, err error) {
	list, total, err := svcWireguard.GetConnectionLogs(ctx, req.PeerId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	res = &wireguard.ConnectionLogsRes{
		List:  list,
		Total: total,
		Page:  req.Page,
	}
	return
}
