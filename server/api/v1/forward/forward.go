// ==========================================================================
// OmniWire - 端口转发 API 定义
// ==========================================================================

package forward

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RuleInfo 转发规则信息
type RuleInfo struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol"` // tcp/udp
	ListenPort    int    `json:"listenPort"`
	TargetAddr    string `json:"targetAddr"`
	TargetPort    int    `json:"targetPort"`
	Enabled       bool   `json:"enabled"`
	Running       bool   `json:"running"`
	MaxConn       int    `json:"maxConn"`
	CurrentConn   int    `json:"currentConn"`
	UploadLimit   int64  `json:"uploadLimit"`   // bytes/s, 0=无限制
	DownloadLimit int64  `json:"downloadLimit"` // bytes/s, 0=无限制
	UploadSpeed   int64  `json:"uploadSpeed"`   // 当前上传速度 bytes/s
	DownloadSpeed int64  `json:"downloadSpeed"` // 当前下载速度 bytes/s
	TotalUpload   int64  `json:"totalUpload"`   // 历史总上传流量
	TotalDownload int64  `json:"totalDownload"` // 历史总下载流量
	Description   string `json:"description"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// RuleStats 转发规则统计
type RuleStats struct {
	Id            int    `json:"id"`
	TotalConn     int64  `json:"totalConn"`
	CurrentConn   int    `json:"currentConn"`
	BytesReceived int64  `json:"bytesReceived"`
	BytesSent     int64  `json:"bytesSent"`
	StartTime     string `json:"startTime"`
	Uptime        int64  `json:"uptime"` // 秒
}

// ===================== 规则管理 =====================

// ListReq 获取规则列表请求
type ListReq struct {
	g.Meta   `path:"/" method:"get" tags:"端口转发" summary:"获取转发规则列表"`
	Page     int `json:"page" d:"1" v:"min:1#页码最小为1"`
	PageSize int `json:"pageSize" d:"20" v:"min:1|max:100#每页数量最小1|每页数量最大100"`
}

// ListRes 获取规则列表响应
type ListRes struct {
	Rules []*RuleInfo `json:"rules"`
	Total int         `json:"total"`
}

// CreateReq 创建规则请求
type CreateReq struct {
	g.Meta        `path:"/" method:"post" tags:"端口转发" summary:"创建转发规则"`
	Name          string `json:"name" v:"required#规则名称必填"`
	Protocol      string `json:"protocol" v:"required|in:tcp,udp#协议必填|协议只能是tcp或udp"`
	ListenPort    int    `json:"listenPort" v:"required|min:1|max:65535#监听端口必填|端口范围错误|端口范围错误"`
	TargetAddr    string `json:"targetAddr" v:"required#目标地址必填"`
	TargetPort    int    `json:"targetPort" v:"required|min:1|max:65535#目标端口必填|端口范围错误|端口范围错误"`
	Enabled       bool   `json:"enabled" d:"true"`
	MaxConn       int    `json:"maxConn" d:"1000" v:"min:1|max:10000#最大连接数最小1|最大连接数最大10000"`
	UploadLimit   int64  `json:"uploadLimit" d:"0"`   // bytes/s, 0=无限制
	DownloadLimit int64  `json:"downloadLimit" d:"0"` // bytes/s, 0=无限制
	Description   string `json:"description"`
}

// CreateRes 创建规则响应
type CreateRes struct {
	Rule *RuleInfo `json:"rule"`
}

// UpdateReq 更新规则请求
type UpdateReq struct {
	g.Meta        `path:"/{id}" method:"put" tags:"端口转发" summary:"更新转发规则"`
	Id            int    `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol" v:"in:tcp,udp#协议只能是tcp或udp"`
	ListenPort    int    `json:"listenPort" v:"min:1|max:65535#端口范围错误|端口范围错误"`
	TargetAddr    string `json:"targetAddr"`
	TargetPort    int    `json:"targetPort" v:"min:1|max:65535#端口范围错误|端口范围错误"`
	Enabled       bool   `json:"enabled"`
	MaxConn       int    `json:"maxConn" v:"min:1|max:10000#最大连接数最小1|最大连接数最大10000"`
	UploadLimit   int64  `json:"uploadLimit"`   // bytes/s, 0=无限制
	DownloadLimit int64  `json:"downloadLimit"` // bytes/s, 0=无限制
	Description   string `json:"description"`
}

// UpdateRes 更新规则响应
type UpdateRes struct {
	Success bool `json:"success"`
}

// DeleteReq 删除规则请求
type DeleteReq struct {
	g.Meta `path:"/{id}" method:"delete" tags:"端口转发" summary:"删除转发规则"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// DeleteRes 删除规则响应
type DeleteRes struct {
	Success bool `json:"success"`
}

// ===================== 规则控制 =====================

// StartReq 启动规则请求
type StartReq struct {
	g.Meta `path:"/{id}/start" method:"post" tags:"端口转发" summary:"启动转发规则"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// StartRes 启动规则响应
type StartRes struct {
	Success bool `json:"success"`
}

// StopReq 停止规则请求
type StopReq struct {
	g.Meta `path:"/{id}/stop" method:"post" tags:"端口转发" summary:"停止转发规则"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// StopRes 停止规则响应
type StopRes struct {
	Success bool `json:"success"`
}

// ===================== 统计信息 =====================

// StatsReq 获取统计请求
type StatsReq struct {
	g.Meta `path:"/{id}/stats" method:"get" tags:"端口转发" summary:"获取转发规则统计"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// StatsRes 获取统计响应
type StatsRes struct {
	Stats *RuleStats `json:"stats"`
}
