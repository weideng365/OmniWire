// ==========================================================================
// OmniWire - 系统管理控制器
// ==========================================================================

package system

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/system"
)

// ControllerV1 系统管理控制器
type ControllerV1 struct{}

// NewV1 创建系统管理控制器实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Info 系统信息
func (c *ControllerV1) Info(ctx context.Context, req *system.InfoReq) (res *system.InfoRes, err error) {
	res = &system.InfoRes{
		Name:    "OmniWire",
		Version: "1.0.0",
		Status:  "running",
	}
	return
}

// Dashboard 仪表盘数据
func (c *ControllerV1) Dashboard(ctx context.Context, req *system.DashboardReq) (res *system.DashboardRes, err error) {
	res = &system.DashboardRes{
		WireguardStatus:   "stopped",
		WireguardPeers:    0,
		ForwardRules:      0,
		ActiveConnections: 0,
	}

	// TODO: 获取实际的系统状态
	g.Log().Debug(ctx, "获取仪表盘数据")

	return
}

// Health 健康检查
func (c *ControllerV1) Health(ctx context.Context, req *system.HealthReq) (res *system.HealthRes, err error) {
	res = &system.HealthRes{
		Status: "healthy",
	}
	return
}
