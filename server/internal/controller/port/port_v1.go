// ==========================================================================
// OmniWire - 端口管理控制器
// ==========================================================================

package port

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/port"
	svcPort "omniwire/internal/service/port"
)

// ControllerV1 端口管理控制器
type ControllerV1 struct{}

// NewV1 创建端口管理控制器实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Scan 扫描端口
func (c *ControllerV1) Scan(ctx context.Context, req *port.ScanReq) (res *port.ScanRes, err error) {
	ports, err := svcPort.Scan(ctx, req.StartPort, req.EndPort)
	if err != nil {
		return nil, err
	}
	res = &port.ScanRes{
		Ports: ports,
	}
	g.Log().Debugf(ctx, "扫描端口 %d-%d 完成", req.StartPort, req.EndPort)
	return
}

// Check 检查端口占用
func (c *ControllerV1) Check(ctx context.Context, req *port.CheckReq) (res *port.CheckRes, err error) {
	info, err := svcPort.Check(ctx, req.Port)
	if err != nil {
		return nil, err
	}
	res = &port.CheckRes{
		Port:    req.Port,
		InUse:   info.InUse,
		Process: info.Process,
		PID:     info.PID,
	}
	return
}

// Listen 获取监听端口列表
func (c *ControllerV1) Listen(ctx context.Context, req *port.ListenReq) (res *port.ListenRes, err error) {
	ports, err := svcPort.GetListeningPorts(ctx)
	if err != nil {
		return nil, err
	}
	res = &port.ListenRes{
		Ports: ports,
	}
	return
}

// Connections 获取端口连接信息
func (c *ControllerV1) Connections(ctx context.Context, req *port.ConnectionsReq) (res *port.ConnectionsRes, err error) {
	conns, err := svcPort.GetConnections(ctx, req.Port)
	if err != nil {
		return nil, err
	}
	res = &port.ConnectionsRes{
		Connections: conns,
	}
	return
}

// Firewall 防火墙状态
func (c *ControllerV1) Firewall(ctx context.Context, req *port.FirewallReq) (res *port.FirewallRes, err error) {
	status, err := svcPort.GetFirewallStatus(ctx)
	if err != nil {
		return nil, err
	}
	res = &port.FirewallRes{
		Enabled: status.Enabled,
		Rules:   status.Rules,
	}
	return
}

// FirewallOpen 开放端口
func (c *ControllerV1) FirewallOpen(ctx context.Context, req *port.FirewallOpenReq) (res *port.FirewallOpenRes, err error) {
	err = svcPort.OpenPort(ctx, req.Port, req.Protocol)
	if err != nil {
		return nil, err
	}
	res = &port.FirewallOpenRes{Success: true}
	g.Log().Infof(ctx, "端口 %d/%s 已开放", req.Port, req.Protocol)
	return
}

// FirewallClose 关闭端口
func (c *ControllerV1) FirewallClose(ctx context.Context, req *port.FirewallCloseReq) (res *port.FirewallCloseRes, err error) {
	err = svcPort.ClosePort(ctx, req.Port, req.Protocol)
	if err != nil {
		return nil, err
	}
	res = &port.FirewallCloseRes{Success: true}
	g.Log().Infof(ctx, "端口 %d/%s 已关闭", req.Port, req.Protocol)
	return
}
