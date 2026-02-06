// ==========================================================================
// OmniWire - 端口管理 API 定义
// ==========================================================================

package port

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PortInfo 端口信息
type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // tcp/udp
	State    string `json:"state"`    // listen/established/closed
	Process  string `json:"process"`
	PID      int    `json:"pid"`
	Address  string `json:"address"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	LocalAddr  string `json:"localAddr"`
	LocalPort  int    `json:"localPort"`
	RemoteAddr string `json:"remoteAddr"`
	RemotePort int    `json:"remotePort"`
	State      string `json:"state"`
	Protocol   string `json:"protocol"`
}

// FirewallRule 防火墙规则
type FirewallRule struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"` // allow/deny
	Source   string `json:"source"`
}

// ===================== 端口扫描 =====================

// ScanReq 扫描端口请求
type ScanReq struct {
	g.Meta    `path:"/scan" method:"post" tags:"端口管理" summary:"扫描端口"`
	StartPort int `json:"startPort" d:"1" v:"min:1|max:65535#起始端口最小1|起始端口最大65535"`
	EndPort   int `json:"endPort" d:"1024" v:"min:1|max:65535#结束端口最小1|结束端口最大65535"`
}

// ScanRes 扫描端口响应
type ScanRes struct {
	Ports []*PortInfo `json:"ports"`
}

// ===================== 端口检查 =====================

// CheckReq 检查端口请求
type CheckReq struct {
	g.Meta `path:"/check/{port}" method:"get" tags:"端口管理" summary:"检查端口占用"`
	Port   int `json:"port" in:"path" v:"required|min:1|max:65535#端口必填|端口范围错误|端口范围错误"`
}

// CheckRes 检查端口响应
type CheckRes struct {
	Port    int    `json:"port"`
	InUse   bool   `json:"inUse"`
	Process string `json:"process"`
	PID     int    `json:"pid"`
}

// ===================== 监听端口 =====================

// ListenReq 获取监听端口请求
type ListenReq struct {
	g.Meta `path:"/listen" method:"get" tags:"端口管理" summary:"获取所有监听端口"`
}

// ListenRes 获取监听端口响应
type ListenRes struct {
	Ports []*PortInfo `json:"ports"`
}

// ===================== 连接信息 =====================

// ConnectionsReq 获取端口连接请求
type ConnectionsReq struct {
	g.Meta `path:"/connections/{port}" method:"get" tags:"端口管理" summary:"获取端口连接信息"`
	Port   int `json:"port" in:"path" v:"required|min:1|max:65535#端口必填|端口范围错误|端口范围错误"`
}

// ConnectionsRes 获取端口连接响应
type ConnectionsRes struct {
	Connections []*ConnectionInfo `json:"connections"`
}

// ===================== 防火墙管理 =====================

// FirewallReq 防火墙状态请求
type FirewallReq struct {
	g.Meta `path:"/firewall" method:"get" tags:"端口管理" summary:"获取防火墙状态"`
}

// FirewallRes 防火墙状态响应
type FirewallRes struct {
	Enabled bool            `json:"enabled"`
	Rules   []*FirewallRule `json:"rules"`
}

// FirewallOpenReq 开放端口请求
type FirewallOpenReq struct {
	g.Meta   `path:"/firewall/open" method:"post" tags:"端口管理" summary:"开放端口"`
	Port     int    `json:"port" v:"required|min:1|max:65535#端口必填|端口范围错误|端口范围错误"`
	Protocol string `json:"protocol" d:"tcp" v:"in:tcp,udp,both#协议只能是tcp/udp/both"`
}

// FirewallOpenRes 开放端口响应
type FirewallOpenRes struct {
	Success bool `json:"success"`
}

// FirewallCloseReq 关闭端口请求
type FirewallCloseReq struct {
	g.Meta   `path:"/firewall/close" method:"post" tags:"端口管理" summary:"关闭端口"`
	Port     int    `json:"port" v:"required|min:1|max:65535#端口必填|端口范围错误|端口范围错误"`
	Protocol string `json:"protocol" d:"tcp" v:"in:tcp,udp,both#协议只能是tcp/udp/both"`
}

// FirewallCloseRes 关闭端口响应
type FirewallCloseRes struct {
	Success bool `json:"success"`
}
