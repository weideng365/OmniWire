// ==========================================================================
// OmniWire - 数据库实体定义
// ==========================================================================

package entity

import "github.com/gogf/gf/v2/os/gtime"

// WireguardPeer WireGuard 客户端
type WireguardPeer struct {
	Id            int         `json:"id" orm:"id"`
	Name          string      `json:"name" orm:"name"`
	PublicKey     string      `json:"publicKey" orm:"public_key"`
	PrivateKey    string      `json:"privateKey" orm:"private_key"`
	AllowedIps    string      `json:"allowedIps" orm:"allowed_ips"`
	Enabled       int         `json:"enabled" orm:"enabled"`
	UploadLimit   int64       `json:"uploadLimit" orm:"upload_limit"`     // 上传速率限制 (bytes/s), 0=无限制
	DownloadLimit int64       `json:"downloadLimit" orm:"download_limit"` // 下载速率限制 (bytes/s), 0=无限制
	TotalUpload   int64       `json:"totalUpload" orm:"total_upload"`     // 历史总上传流量
	TotalDownload int64       `json:"totalDownload" orm:"total_download"` // 历史总下载流量
	CreatedAt     *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt     *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

// ForwardRule 端口转发规则
type ForwardRule struct {
	Id            int         `json:"id"`
	Name          string      `json:"name"`
	Protocol      string      `json:"protocol"`
	ListenPort    int         `json:"listenPort"`
	TargetAddr    string      `json:"targetAddr"`
	TargetPort    int         `json:"targetPort"`
	Enabled       int         `json:"enabled"`
	MaxConn       int         `json:"maxConn"`
	UploadLimit   int64       `json:"uploadLimit"`   // 上传速率限制 (bytes/s), 0=无限制
	DownloadLimit int64       `json:"downloadLimit"` // 下载速率限制 (bytes/s), 0=无限制
	TotalUpload   int64       `json:"totalUpload"`   // 历史总上传流量
	TotalDownload int64       `json:"totalDownload"` // 历史总下载流量
	Description   string      `json:"description"`
	CreatedAt     *gtime.Time `json:"createdAt"`
	UpdatedAt     *gtime.Time `json:"updatedAt"`
}

// User 用户
type User struct {
	Id        int         `json:"id"`
	Username  string      `json:"username"`
	Password  string      `json:"password"`
	Role      string      `json:"role"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// OperationLog 操作日志
type OperationLog struct {
	Id        int         `json:"id"`
	UserId    int         `json:"userId"`
	Action    string      `json:"action"`
	Target    string      `json:"target"`
	Detail    string      `json:"detail"`
	Ip        string      `json:"ip"`
	CreatedAt *gtime.Time `json:"createdAt"`
}
