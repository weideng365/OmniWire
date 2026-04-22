// ==========================================================================
// OmniWire - OpenVPN API 定义
// ==========================================================================

package openvpn

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ===================== 服务状态 =====================

// StatusReq OpenVPN 状态请求
type StatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"OpenVPN" summary:"获取OpenVPN服务状态"`
}

// StatusRes OpenVPN 状态响应
type StatusRes struct {
	Running     bool   `json:"running"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	ClientCount int    `json:"clientCount"`
	RxBytes     int64  `json:"rxBytes"`
	TxBytes     int64  `json:"txBytes"`
}

// StartReq 启动服务请求
type StartReq struct {
	g.Meta `path:"/start" method:"post" tags:"OpenVPN" summary:"启动OpenVPN服务"`
}

// StartRes 启动服务响应
type StartRes struct {
	Success bool `json:"success"`
}

// StopReq 停止服务请求
type StopReq struct {
	g.Meta `path:"/stop" method:"post" tags:"OpenVPN" summary:"停止OpenVPN服务"`
}

// StopRes 停止服务响应
type StopRes struct {
	Success bool `json:"success"`
}

// RestartReq 重启服务请求
type RestartReq struct {
	g.Meta `path:"/restart" method:"post" tags:"OpenVPN" summary:"重启OpenVPN服务"`
}

// RestartRes 重启服务响应
type RestartRes struct {
	Success bool `json:"success"`
}

// ===================== 配置管理 =====================

// ConfigReq 获取配置请求
type ConfigReq struct {
	g.Meta `path:"/config" method:"get" tags:"OpenVPN" summary:"获取OpenVPN配置"`
}

// ConfigRes 获取配置响应
type ConfigRes struct {
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	Endpoint    string `json:"endpoint"`
	Subnet      string `json:"subnet"`
	DNS         string `json:"dns"`
	AutoStart   bool   `json:"autoStart"`
	RouteMode   string `json:"routeMode"`   // full 或 split
	SplitRoutes string `json:"splitRoutes"` // 逗号分隔的 CIDR，仅 split 模式有效
}

// UpdateConfigReq 更新配置请求
type UpdateConfigReq struct {
	g.Meta      `path:"/config" method:"put" tags:"OpenVPN" summary:"更新OpenVPN配置"`
	Protocol    string `json:"protocol" v:"required|in:udp,tcp#协议必填|协议只能是udp或tcp"`
	Port        int    `json:"port" v:"required|min:1|max:65535#端口必填|端口范围错误|端口范围错误"`
	Endpoint    string `json:"endpoint"`
	Subnet      string `json:"subnet" v:"required#子网必填"`
	DNS         string `json:"dns"`
	AutoStart   bool   `json:"autoStart"`
	RouteMode   string `json:"routeMode"`
	SplitRoutes string `json:"splitRoutes"`
}

// UpdateConfigRes 更新配置响应
type UpdateConfigRes struct {
	Success bool `json:"success"`
}

// ===================== 用户管理 =====================

// UserInfo 用户信息
type UserInfo struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	Enabled     bool   `json:"enabled"`
	Online      bool   `json:"online"`
	IP          string `json:"ip"`
	ConnectedAt string `json:"connectedAt"`
	CreatedAt   string `json:"createdAt"`
	RxBytes     int64  `json:"rxBytes"`
	TxBytes     int64  `json:"txBytes"`
}

// UserListReq 获取用户列表请求
type UserListReq struct {
	g.Meta `path:"/users" method:"get" tags:"OpenVPN" summary:"获取用户列表"`
}

// UserListRes 获取用户列表响应
type UserListRes struct {
	Users []*UserInfo `json:"users"`
}

// UserCreateReq 创建用户请求
type UserCreateReq struct {
	g.Meta   `path:"/users" method:"post" tags:"OpenVPN" summary:"创建用户"`
	Username string `json:"username" v:"required#用户名必填"`
	Password string `json:"password" v:"required|length:6,32#密码必填|密码长度6-32位"`
}

// UserCreateRes 创建用户响应
type UserCreateRes struct {
	User *UserInfo `json:"user"`
}

// UserUpdateReq 更新用户请求
type UserUpdateReq struct {
	g.Meta   `path:"/users/{id}" method:"put" tags:"OpenVPN" summary:"更新用户"`
	Id       int    `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

// UserUpdateRes 更新用户响应
type UserUpdateRes struct {
	Success bool `json:"success"`
}

// UserDeleteReq 删除用户请求
type UserDeleteReq struct {
	g.Meta `path:"/users/{id}" method:"delete" tags:"OpenVPN" summary:"删除用户"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// UserDeleteRes 删除用户响应
type UserDeleteRes struct {
	Success bool `json:"success"`
}

// UserConfigReq 获取用户配置请求
type UserConfigReq struct {
	g.Meta `path:"/users/{id}/config" method:"get" tags:"OpenVPN" summary:"获取用户配置文件"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// UserConfigRes 获取用户配置响应
type UserConfigRes struct {
	Config string `json:"config"` // .ovpn 文件内容
}

// ===================== 认证回调 =====================

// AuthReq 用户名密码认证请求（供 openvpn auth 脚本回调）
type AuthReq struct {
	g.Meta   `path:"/auth" method:"post" tags:"OpenVPN" summary:"验证用户名密码"`
	Username string `json:"username" v:"required#用户名必填"`
	Password string `json:"password" v:"required#密码必填"`
}

// AuthRes 认证响应
type AuthRes struct {
	Success bool `json:"success"`
}

// ConnectReq 客户端连接回调
type ConnectReq struct {
	g.Meta   `path:"/connect" method:"post" tags:"OpenVPN" summary:"客户端连接回调"`
	Username string `json:"username"`
	IP       string `json:"ip"`
}

type ConnectRes struct{}

// DisconnectReq 客户端断开回调
type DisconnectReq struct {
	g.Meta   `path:"/disconnect" method:"post" tags:"OpenVPN" summary:"客户端断开回调"`
	Username string `json:"username"`
}

type DisconnectRes struct{}
