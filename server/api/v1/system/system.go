// ==========================================================================
// OmniWire - 系统 API 定义
// ==========================================================================

package system

import (
	"github.com/gogf/gf/v2/frame/g"
)

// InfoReq 系统信息请求
type InfoReq struct {
	g.Meta `path:"/info" method:"get" tags:"系统管理" summary:"获取系统信息"`
}

// InfoRes 系统信息响应
type InfoRes struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// DashboardReq 仪表盘请求
type DashboardReq struct {
	g.Meta `path:"/dashboard" method:"get" tags:"系统管理" summary:"获取仪表盘数据"`
}

// DashboardRes 仪表盘响应
type DashboardRes struct {
	WireguardStatus   string `json:"wireguardStatus"`
	WireguardPeers    int    `json:"wireguardPeers"`
	ForwardRules      int    `json:"forwardRules"`
	ActiveConnections int    `json:"activeConnections"`
}

// LoginReq 登录请求
type LoginReq struct {
	g.Meta   `path:"/login" method:"post" tags:"系统管理" summary:"用户登录"`
	Username string `json:"username" v:"required"`
	Password string `json:"password" v:"required"`
}

// LoginRes 登录响应
type LoginRes struct {
	Token string `json:"token"`
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	g.Meta      `path:"/change-password" method:"post" tags:"系统管理" summary:"修改密码"`
	OldPassword string `json:"oldPassword" v:"required"`
	NewPassword string `json:"newPassword" v:"required|min-length:6"`
}

// ChangePasswordRes 修改密码响应
type ChangePasswordRes struct{}

// HealthReq 健康检查请求
type HealthReq struct {
	g.Meta `path:"/health" method:"get" tags:"系统管理" summary:"健康检查"`
}

// HealthRes 健康检查响应
type HealthRes struct {
	Status string `json:"status"`
}
