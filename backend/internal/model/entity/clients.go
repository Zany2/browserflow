// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Clients is the golang structure for table clients.
type Clients struct {
	Id             int64       `json:"id"              orm:"id"              ` // 主键 ID，自增
	ClientId       string      `json:"client_id"       orm:"client_id"       ` // 客户端唯一标识（建议前端生成 UUID 并存储在 localStorage，用于唯一识别一个浏览器实例）
	ClientName     string      `json:"client_name"     orm:"client_name"     ` // 客户端名称（可自定义，例如：办公室电脑、家用电脑）
	ClientIp       string      `json:"client_ip"       orm:"client_ip"       ` // 客户端当前或最后一次连接 IP
	UserAgent      string      `json:"user_agent"      orm:"user_agent"      ` // 客户端浏览器 User-Agent 字符串
	Status         string      `json:"status"          orm:"status"          ` // 客户端当前状态：online 在线，offline 离线
	PluginStatus   string      `json:"plugin_status"   orm:"plugin_status"   ` // Automa 插件状态：unknown 未知，installed 已安装，not_installed 未安装，disabled 已禁用，error 异常
	AutomaVersion  string      `json:"automa_version"  orm:"automa_version"  ` // Automa 插件版本号
	BrowserName    string      `json:"browser_name"    orm:"browser_name"    ` // 浏览器名称，例如 Chrome、Edge
	BrowserVersion string      `json:"browser_version" orm:"browser_version" ` // 浏览器版本号
	OsName         string      `json:"os_name"         orm:"os_name"         ` // 操作系统名称，例如 Windows、macOS、Linux
	OsVersion      string      `json:"os_version"      orm:"os_version"      ` // 操作系统版本号
	Hostname       string      `json:"hostname"        orm:"hostname"        ` // 客户端设备主机名
	IsBanned       bool        `json:"is_banned"       orm:"is_banned"       ` // 是否被拉黑（true 表示该客户端被禁止使用）
	BanReason      string      `json:"ban_reason"      orm:"ban_reason"      ` // 客户端被拉黑的原因
	FirstSeenAt    *gtime.Time `json:"first_seen_at"   orm:"first_seen_at"   ` // 客户端第一次连接到服务端的时间（注册时间）
	LastSeenAt     *gtime.Time `json:"last_seen_at"    orm:"last_seen_at"    ` // 客户端最近一次心跳或交互时间（用于判断是否在线）
	ConnectedAt    *gtime.Time `json:"connected_at"    orm:"connected_at"    ` // 最近一次建立 WebSocket 连接的时间
	DisconnectedAt *gtime.Time `json:"disconnected_at" orm:"disconnected_at" ` // 最近一次断开 WebSocket 连接的时间
	CreatedAt      *gtime.Time `json:"created_at"      orm:"created_at"      ` // 记录创建时间
	UpdatedAt      *gtime.Time `json:"updated_at"      orm:"updated_at"      ` // 记录更新时间
	DeletedAt      *gtime.Time `json:"deleted_at"      orm:"deleted_at"      ` // 软删除时间
}
