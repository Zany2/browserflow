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
	g.Meta              `orm:"table:clients, do:true"`
	Id                  interface{} // 主键 ID，自增
	ClientIp            interface{} // 客户端 IP
	NodeId              interface{} // 执行节点 ID，同一 IP 下唯一，例如 node-1、node-2
	NodeIndex           interface{} // Worker 本机节点序号
	WorkerVersion       interface{} // BrowserFlow Worker 客户端版本
	ProfileDir          interface{} // 执行节点使用的 Chrome 用户数据目录
	ExtensionDir        interface{} // 执行节点加载的 Automa 插件目录
	UserAgent           interface{} // 客户端浏览器 User-Agent
	Status              interface{} // 客户端当前状态：online 在线，offline 离线
	BusyStatus          interface{} // 节点忙闲状态：idle 空闲，busy 执行中，unknown 未知
	CurrentExecutionId  interface{} // 节点当前执行中的后端执行 ID
	CurrentTaskRecordId interface{} // 节点当前执行中的任务记录 ID
	PluginStatus        interface{} // Automa 插件状态：unknown、installed、not_installed、disabled、error
	AutomaVersion       interface{} // Automa 插件版本号
	BrowserName         interface{} // 浏览器名称，例如 Chrome、Edge
	BrowserVersion      interface{} // 浏览器版本号
	OsName              interface{} // 操作系统名称，例如 Windows、macOS、Linux
	OsVersion           interface{} // 操作系统版本号
	Hostname            interface{} // 客户端设备主机名
	CapabilitiesJson    interface{} // 节点能力快照，例如是否支持工作流安装、插件安装、文件回传
	LastCommandAt       *gtime.Time // 最近一次服务端命令下发时间
	LastWorkflowSyncAt  *gtime.Time // 最近一次工作流清单同步时间
	LastLockRenewedAt   *gtime.Time // 节点执行锁最近续期时间
	IsBanned            interface{} // 是否被拉黑，true 表示禁止使用
	BanReason           interface{} // 客户端被拉黑的原因
	FirstSeenAt         *gtime.Time // 第一次连接到服务端的时间
	LastSeenAt          *gtime.Time // 最近一次心跳或交互时间
	ConnectedAt         *gtime.Time // 最近一次建立 WebSocket 连接的时间
	DisconnectedAt      *gtime.Time // 最近一次断开 WebSocket 连接的时间
	CreatedAt           *gtime.Time // 记录创建时间
	UpdatedAt           *gtime.Time // 记录更新时间
	DeletedAt           *gtime.Time // 软删除时间
	DisplayName         interface{} // 客户端自定义显示名称
}
