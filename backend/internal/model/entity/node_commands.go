// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeCommands is the golang structure for table node_commands.
type NodeCommands struct {
	Id             int64       `json:"id"              orm:"id"              ` // 自增 ID
	CommandId      string      `json:"command_id"      orm:"command_id"      ` // 服务端命令唯一 ID，用于客户端回调和幂等处理
	NodeId         string      `json:"node_id"         orm:"node_id"         ` // 目标执行节点 ID
	MachineId      string      `json:"machine_id"      orm:"machine_id"      ` // 目标机器 ID
	CommandType    string      `json:"command_type"    orm:"command_type"    ` // 命令类型：install_workflow、delete_workflow、sync_inventory、install_package、restart_node、execute_workflow 等
	Status         string      `json:"status"          orm:"status"          ` // 命令状态：pending、sent、running、success、failed、cancelled、timeout
	PayloadJson    string      `json:"payload_json"    orm:"payload_json"    ` // 命令参数 JSON
	ResultJson     string      `json:"result_json"     orm:"result_json"     ` // 命令执行结果 JSON
	ErrorMessage   string      `json:"error_message"   orm:"error_message"   ` // 命令失败信息
	RetryCount     int         `json:"retry_count"     orm:"retry_count"     ` // 命令重试次数
	TimeoutSeconds int         `json:"timeout_seconds" orm:"timeout_seconds" ` // 命令超时时间，单位秒
	SentAt         *gtime.Time `json:"sent_at"         orm:"sent_at"         ` // 命令发送时间
	AckAt          *gtime.Time `json:"ack_at"          orm:"ack_at"          ` // 客户端确认收到时间
	FinishedAt     *gtime.Time `json:"finished_at"     orm:"finished_at"     ` // 命令完成时间
	CreatedAt      *gtime.Time `json:"created_at"      orm:"created_at"      ` // 记录创建时间
	UpdatedAt      *gtime.Time `json:"updated_at"      orm:"updated_at"      ` // 记录更新时间
	DeletedAt      *gtime.Time `json:"deleted_at"      orm:"deleted_at"      ` // 软删除时间
}
