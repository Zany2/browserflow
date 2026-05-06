// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Clients is the golang structure of table clients for DAO operations like Where/Data.
type Clients struct {
	g.Meta         `orm:"table:clients, do:true"`
	Id             interface{} // 主键 ID，自增
	ClientId       interface{} // 客户端唯一标识（建议前端生成 UUID 并存储在 localStorage，用于唯一识别一个浏览器实例）
	ClientName     interface{} // 客户端名称（可自定义，例如：办公室电脑、家用电脑）
	ClientIp       interface{} // 客户端当前或最后一次连接 IP
	UserAgent      interface{} // 客户端浏览器 User-Agent 字符串
	Status         interface{} // 客户端当前状态：online 在线，offline 离线
	PluginStatus   interface{} // Automa 插件状态：unknown 未知，installed 已安装，not_installed 未安装，disabled 已禁用，error 异常
	AutomaVersion  interface{} // Automa 插件版本号
	BrowserName    interface{} // 浏览器名称，例如 Chrome、Edge
	BrowserVersion interface{} // 浏览器版本号
	OsName         interface{} // 操作系统名称，例如 Windows、macOS、Linux
	OsVersion      interface{} // 操作系统版本号
	Hostname       interface{} // 客户端设备主机名
	IsBanned       interface{} // 是否被拉黑（true 表示该客户端被禁止使用）
	BanReason      interface{} // 客户端被拉黑的原因
	FirstSeenAt    *gtime.Time // 客户端第一次连接到服务端的时间（注册时间）
	LastSeenAt     *gtime.Time // 客户端最近一次心跳或交互时间（用于判断是否在线）
	ConnectedAt    *gtime.Time // 最近一次建立 WebSocket 连接的时间
	DisconnectedAt *gtime.Time // 最近一次断开 WebSocket 连接的时间
	CreatedAt      *gtime.Time // 记录创建时间
	UpdatedAt      *gtime.Time // 记录更新时间
	DeletedAt      *gtime.Time // 软删除时间
}
