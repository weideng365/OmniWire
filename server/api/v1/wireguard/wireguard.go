// ==========================================================================
// OmniWire - WireGuard API 定义
// ==========================================================================

package wireguard

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ===================== 服务状态 =====================

// StatusReq WireGuard 状态请求
type StatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"WireGuard" summary:"获取WireGuard服务状态"`
}

// StatusRes WireGuard 状态响应
type StatusRes struct {
	Running    bool   `json:"running"`
	Interface  string `json:"interface"`
	ListenPort int    `json:"listenPort"`
	PublicKey  string `json:"publicKey"`
	PeerCount  int    `json:"peerCount"`
}

// StartReq 启动服务请求
type StartReq struct {
	g.Meta `path:"/start" method:"post" tags:"WireGuard" summary:"启动WireGuard服务"`
}

// StartRes 启动服务响应
type StartRes struct {
	Success bool `json:"success"`
}

// StopReq 停止服务请求
type StopReq struct {
	g.Meta `path:"/stop" method:"post" tags:"WireGuard" summary:"停止WireGuard服务"`
}

// StopRes 停止服务响应
type StopRes struct {
	Success bool `json:"success"`
}

// RestartReq 重启服务请求
type RestartReq struct {
	g.Meta `path:"/restart" method:"post" tags:"WireGuard" summary:"重启WireGuard服务"`
}

// RestartRes 重启服务响应
type RestartRes struct {
	Success bool `json:"success"`
}

// ===================== 配置管理 =====================

// ConfigReq 获取配置请求
type ConfigReq struct {
	g.Meta `path:"/config" method:"get" tags:"WireGuard" summary:"获取WireGuard配置"`
}

// ConfigRes 获取配置响应
type ConfigRes struct {
	Interface           string `json:"interface"`
	ListenPort          int    `json:"listenPort"`
	PrivateKey          string `json:"privateKey"`
	PublicKey           string `json:"publicKey"`
	Address             string `json:"address"`
	DNS                 string `json:"dns"`
	MTU                 int    `json:"mtu"`
	EndpointAddress     string `json:"endpointAddress"`
	PostUp              string `json:"postUp"`
	PostDown            string `json:"postDown"`
	EthDevice           string `json:"ethDevice"`
	PersistentKeepalive int    `json:"persistentKeepalive"`
	ClientAllowedIPs    string `json:"clientAllowedIPs"`
	ProxyAddress        string `json:"proxyAddress"`
	LogLevel            string `json:"logLevel"`
	AutoStart           bool   `json:"autoStart"`
}

// UpdateConfigReq 更新配置请求
type UpdateConfigReq struct {
	g.Meta              `path:"/config" method:"put" tags:"WireGuard" summary:"更新WireGuard配置"`
	ListenPort          int    `json:"listenPort" v:"required|min:1|max:65535#监听端口必填|端口范围错误|端口范围错误"`
	EndpointAddress     string `json:"endpointAddress"`
	Address             string `json:"address" v:"required#地址必填"`
	DNS                 string `json:"dns"`
	MTU                 int    `json:"mtu" v:"min:1280|max:1500#MTU最小1280|MTU最大1500"`
	EthDevice           string `json:"ethDevice"`
	PersistentKeepalive int    `json:"persistentKeepalive"`
	ClientAllowedIPs    string `json:"clientAllowedIPs"`
	ProxyAddress        string `json:"proxyAddress"`
	LogLevel            string `json:"logLevel"`
	AutoStart           bool   `json:"autoStart"`
}

// UpdateConfigRes 更新配置响应
type UpdateConfigRes struct {
	Success bool `json:"success"`
}

// ===================== 客户端管理 =====================

// PeerInfo 客户端信息
type PeerInfo struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	PublicKey       string `json:"publicKey"`
	AllowedIPs      string `json:"allowedIPs"`
	Endpoint        string `json:"endpoint"`
	LatestHandshake string `json:"latestHandshake"`
	TransferRx      int64  `json:"transferRx"`
	TransferTx      int64  `json:"transferTx"`
	UploadLimit     int64  `json:"uploadLimit"`   // bytes/s, 0=无限制
	DownloadLimit   int64  `json:"downloadLimit"` // bytes/s, 0=无限制
	TotalUpload     int64  `json:"totalUpload"`   // 历史总上传流量
	TotalDownload   int64  `json:"totalDownload"` // 历史总下载流量
	Enabled         bool   `json:"enabled"`
	Online          bool   `json:"online"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// PeerListReq 获取客户端列表请求
type PeerListReq struct {
	g.Meta `path:"/peers" method:"get" tags:"WireGuard" summary:"获取客户端列表"`
}

// PeerListRes 获取客户端列表响应
type PeerListRes struct {
	Peers []*PeerInfo `json:"peers"`
}

// PeerCreateReq 创建客户端请求
type PeerCreateReq struct {
	g.Meta        `path:"/peers" method:"post" tags:"WireGuard" summary:"创建客户端"`
	Name          string `json:"name" v:"required#客户端名称必填"`
	AllowedIPs    string `json:"allowedIPs"`
	UploadLimit   int64  `json:"uploadLimit" d:"0"`   // bytes/s, 0=无限制
	DownloadLimit int64  `json:"downloadLimit" d:"0"` // bytes/s, 0=无限制
}

// PeerCreateRes 创建客户端响应
type PeerCreateRes struct {
	Peer *PeerInfo `json:"peer"`
}

// PeerUpdateReq 更新客户端请求
type PeerUpdateReq struct {
	g.Meta        `path:"/peers/{id}" method:"put" tags:"WireGuard" summary:"更新客户端"`
	Id            int    `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
	Name          string `json:"name"`
	AllowedIPs    string `json:"allowedIPs"`
	UploadLimit   int64  `json:"uploadLimit"`   // bytes/s, 0=无限制
	DownloadLimit int64  `json:"downloadLimit"` // bytes/s, 0=无限制
	Enabled       bool   `json:"enabled"`
}

// PeerUpdateRes 更新客户端响应
type PeerUpdateRes struct {
	Success bool `json:"success"`
}

// PeerDeleteReq 删除客户端请求
type PeerDeleteReq struct {
	g.Meta `path:"/peers/{id}" method:"delete" tags:"WireGuard" summary:"删除客户端"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// PeerDeleteRes 删除客户端响应
type PeerDeleteRes struct {
	Success bool `json:"success"`
}

// PeerConfigReq 获取客户端配置请求
type PeerConfigReq struct {
	g.Meta `path:"/peers/{id}/config" method:"get" tags:"WireGuard" summary:"获取客户端配置文件"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// PeerConfigRes 获取客户端配置响应
type PeerConfigRes struct {
	Config string `json:"config"`
}

// PeerQRCodeReq 获取客户端二维码请求
type PeerQRCodeReq struct {
	g.Meta `path:"/peers/{id}/qrcode" method:"get" tags:"WireGuard" summary:"获取客户端配置二维码"`
	Id     int `json:"id" in:"path" v:"required|min:1#ID必填|ID无效"`
}

// PeerQRCodeRes 获取客户端二维码响应
type PeerQRCodeRes struct {
	QRCode string `json:"qrcode"` // Base64 编码的 PNG 图片
}

// ===================== 连接日志 =====================

// ConnectionLogsReq 获取连接日志请求
type ConnectionLogsReq struct {
	g.Meta   `path:"/connection-logs" method:"get" tags:"WireGuard" summary:"获取客户端连接日志"`
	PeerId   int `json:"peerId" in:"query"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

// ConnectionLogInfo 连接日志信息
type ConnectionLogInfo struct {
	Id         int    `json:"id"`
	PeerName   string `json:"peerName"`
	Event      string `json:"event"`
	Endpoint   string `json:"endpoint"`
	TransferRx int64  `json:"transferRx"`
	TransferTx int64  `json:"transferTx"`
	CreatedAt  string `json:"createdAt"`
}

// ConnectionLogsRes 获取连接日志响应
type ConnectionLogsRes struct {
	List  []*ConnectionLogInfo `json:"list"`
	Total int                  `json:"total"`
	Page  int                  `json:"page"`
}
