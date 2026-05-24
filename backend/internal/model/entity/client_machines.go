// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientMachines is the golang structure for table client_machines.
type ClientMachines struct {
	Id              int64       `json:"id"                orm:"id"                ` // 主键 ID，自增
	MachineId       string      `json:"machine_id"        orm:"machine_id"        ` // 物理机器稳定 ID，同一台 Windows 电脑保持一致
	MachineName     string      `json:"machine_name"      orm:"machine_name"      ` // Worker 上报的机器名称
	DisplayName     string      `json:"display_name"      orm:"display_name"      ` // 服务端自定义机器显示名称
	ClientIp        string      `json:"client_ip"         orm:"client_ip"         ` // 机器当前或最后一次连接 IP
	Hostname        string      `json:"hostname"          orm:"hostname"          ` // 系统主机名
	OsName          string      `json:"os_name"           orm:"os_name"           ` // 操作系统名称
	OsVersion       string      `json:"os_version"        orm:"os_version"        ` // 操作系统版本
	WorkerVersion   string      `json:"worker_version"    orm:"worker_version"    ` // BrowserFlow Worker 版本
	InstallDir      string      `json:"install_dir"       orm:"install_dir"       ` // Worker 安装目录
	DataDir         string      `json:"data_dir"          orm:"data_dir"          ` // Worker 数据目录
	Status          string      `json:"status"            orm:"status"            ` // 机器状态：online 在线，offline 离线
	IsBanned        bool        `json:"is_banned"         orm:"is_banned"         ` // 是否禁止该机器连接
	BanReason       string      `json:"ban_reason"        orm:"ban_reason"        ` // 机器被禁止连接的原因
	NodeCount       int         `json:"node_count"        orm:"node_count"        ` // 该机器配置的执行节点数量
	OnlineNodeCount int         `json:"online_node_count" orm:"online_node_count" ` // 该机器当前在线执行节点数量
	ConfigJson      string      `json:"config_json"       orm:"config_json"       ` // 机器级配置快照
	FirstSeenAt     *gtime.Time `json:"first_seen_at"     orm:"first_seen_at"     ` // 第一次连接到服务端的时间
	LastSeenAt      *gtime.Time `json:"last_seen_at"      orm:"last_seen_at"      ` // 最近一次心跳或交互时间
	CreatedAt       *gtime.Time `json:"created_at"        orm:"created_at"        ` // 记录创建时间
	UpdatedAt       *gtime.Time `json:"updated_at"        orm:"updated_at"        ` // 记录更新时间
	DeletedAt       *gtime.Time `json:"deleted_at"        orm:"deleted_at"        ` // 软删除时间
}
