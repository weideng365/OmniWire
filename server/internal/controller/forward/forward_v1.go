// ==========================================================================
// OmniWire - 端口转发控制器
// ==========================================================================

package forward

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/forward"
	svcForward "omniwire/internal/service/forward"
)

// ControllerV1 端口转发控制器
type ControllerV1 struct{}

// NewV1 创建端口转发控制器实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// List 获取转发规则列表
func (c *ControllerV1) List(ctx context.Context, req *forward.ListReq) (res *forward.ListRes, err error) {
	rules, total, err := svcForward.GetList(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	res = &forward.ListRes{
		Rules: rules,
		Total: total,
	}
	return
}

// Create 创建转发规则
func (c *ControllerV1) Create(ctx context.Context, req *forward.CreateReq) (res *forward.CreateRes, err error) {
	rule, err := svcForward.Create(ctx, &svcForward.RuleInput{
		Name:        req.Name,
		Protocol:    req.Protocol,
		ListenPort:  req.ListenPort,
		TargetAddr:  req.TargetAddr,
		TargetPort:  req.TargetPort,
		Enabled:     req.Enabled,
		MaxConn:     req.MaxConn,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	res = &forward.CreateRes{
		Rule: rule,
	}
	g.Log().Infof(ctx, "转发规则 %s 已创建", req.Name)
	return
}

// Update 更新转发规则
func (c *ControllerV1) Update(ctx context.Context, req *forward.UpdateReq) (res *forward.UpdateRes, err error) {
	err = svcForward.Update(ctx, req.Id, &svcForward.RuleInput{
		Name:        req.Name,
		Protocol:    req.Protocol,
		ListenPort:  req.ListenPort,
		TargetAddr:  req.TargetAddr,
		TargetPort:  req.TargetPort,
		Enabled:     req.Enabled,
		MaxConn:     req.MaxConn,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	res = &forward.UpdateRes{Success: true}
	g.Log().Infof(ctx, "转发规则 %d 已更新", req.Id)
	return
}

// Delete 删除转发规则
func (c *ControllerV1) Delete(ctx context.Context, req *forward.DeleteReq) (res *forward.DeleteRes, err error) {
	err = svcForward.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &forward.DeleteRes{Success: true}
	g.Log().Infof(ctx, "转发规则 %d 已删除", req.Id)
	return
}

// Start 启动转发规则
func (c *ControllerV1) Start(ctx context.Context, req *forward.StartReq) (res *forward.StartRes, err error) {
	err = svcForward.Start(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &forward.StartRes{Success: true}
	g.Log().Infof(ctx, "转发规则 %d 已启动", req.Id)
	return
}

// Stop 停止转发规则
func (c *ControllerV1) Stop(ctx context.Context, req *forward.StopReq) (res *forward.StopRes, err error) {
	err = svcForward.Stop(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &forward.StopRes{Success: true}
	g.Log().Infof(ctx, "转发规则 %d 已停止", req.Id)
	return
}

// Stats 获取转发统计
func (c *ControllerV1) Stats(ctx context.Context, req *forward.StatsReq) (res *forward.StatsRes, err error) {
	stats, err := svcForward.GetStats(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &forward.StatsRes{
		Stats: stats,
	}
	return
}
