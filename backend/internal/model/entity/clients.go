// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Clients is the golang structure for table clients.
type Clients struct {
	Id                  int64       `json:"id"                     orm:"id"                     ` // 主键 ID，自增
	ClientIp            string      `json:"client_ip"              orm:"client_ip"              ` // 客户端 IP
	NodeId              string      `json:"node_id"                orm:"node_id"                ` // 执行节点 ID，同一 IP 下唯一，例如 node-1、node-2
	NodeIndex           int         `json:"node_index"             orm:"node_index"             ` // Worker 本机节点序号
	WorkerVersion       string      `json:"worker_version"         orm:"worker_version"         ` // BrowserFlow Worker 客户端版本
	ProfileDir          string      `json:"profile_dir"            orm:"profile_dir"            ` // 执行节点使用的 Chrome 用户数据目录
	ExtensionDir        string      `json:"extension_dir"          orm:"extension_dir"          ` // 执行节点加载的 Automa 插件目录
	UserAgent           string      `json:"user_agent"             orm:"user_agent"             ` // 客户端浏览器 User-Agent
	Status              string      `json:"status"                 orm:"status"                 ` // 客户端当前状态：online 在线，offline 离线
	BusyStatus          string      `json:"busy_status"            orm:"busy_status"            ` // 节点忙闲状态：idle 空闲，busy 执行中，unknown 未知
	CurrentExecutionId  string      `json:"current_execution_id"   orm:"current_execution_id"   ` // 节点当前执行中的后端执行 ID
	CurrentTaskRecordId int64       `json:"current_task_record_id" orm:"current_task_record_id" ` // 节点当前执行中的任务记录 ID
	PluginStatus        string      `json:"plugin_status"          orm:"plugin_status"          ` // Automa 插件状态：unknown、installed、not_installed、disabled、error
	AutomaVersion       string      `json:"automa_version"         orm:"automa_version"         ` // Automa 插件版本号
	BrowserName         string      `json:"browser_name"           orm:"browser_name"           ` // 浏览器名称，例如 Chrome、Edge
	BrowserVersion      string      `json:"browser_version"        orm:"browser_version"        ` // 浏览器版本号
	OsName              string      `json:"os_name"                orm:"os_name"                ` // 操作系统名称，例如 Windows、macOS、Linux
	OsVersion           string      `json:"os_version"             orm:"os_version"             ` // 操作系统版本号
	Hostname            string      `json:"hostname"               orm:"hostname"               ` // 客户端设备主机名
	CapabilitiesJson    string      `json:"capabilities_json"      orm:"capabilities_json"      ` // 节点能力快照，例如是否支持工作流安装、插件安装、文件回传
	LastCommandAt       *gtime.Time `json:"last_command_at"        orm:"last_command_at"        ` // 最近一次服务端命令下发时间
	LastWorkflowSyncAt  *gtime.Time `json:"last_workflow_sync_at"  orm:"last_workflow_sync_at"  ` // 最近一次工作流清单同步时间
	LastLockRenewedAt   *gtime.Time `json:"last_lock_renewed_at"   orm:"last_lock_renewed_at"   ` // 节点执行锁最近续期时间
	IsBanned            bool        `json:"is_banned"              orm:"is_banned"              ` // 是否被拉黑，true 表示禁止使用
	BanReason           string      `json:"ban_reason"             orm:"ban_reason"             ` // 客户端被拉黑的原因
	FirstSeenAt         *gtime.Time `json:"first_seen_at"          orm:"first_seen_at"          ` // 第一次连接到服务端的时间
	LastSeenAt          *gtime.Time `json:"last_seen_at"           orm:"last_seen_at"           ` // 最近一次心跳或交互时间
	ConnectedAt         *gtime.Time `json:"connected_at"           orm:"connected_at"           ` // 最近一次建立 WebSocket 连接的时间
	DisconnectedAt      *gtime.Time `json:"disconnected_at"        orm:"disconnected_at"        ` // 最近一次断开 WebSocket 连接的时间
	CreatedAt           *gtime.Time `json:"created_at"             orm:"created_at"             ` // 记录创建时间
	UpdatedAt           *gtime.Time `json:"updated_at"             orm:"updated_at"             ` // 记录更新时间
	DeletedAt           *gtime.Time `json:"deleted_at"             orm:"deleted_at"             ` // 软删除时间
	DisplayName         string      `json:"display_name"           orm:"display_name"           ` // 客户端自定义显示名称
}
