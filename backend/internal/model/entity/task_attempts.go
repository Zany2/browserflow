// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskAttempts is the golang structure for table task_attempts.
type TaskAttempts struct {
	Id           int64       `json:"id"            orm:"id"            ` // 自增 ID
	QueueId      int64       `json:"queue_id"      orm:"queue_id"      ` // 对应任务队列 ID
	TaskId       int64       `json:"task_id"       orm:"task_id"       ` // 关联任务配置 ID
	RecordId     int64       `json:"record_id"     orm:"record_id"     ` // 关联任务执行记录 ID
	AttemptNo    int         `json:"attempt_no"    orm:"attempt_no"    ` // 尝试序号，从 1 开始
	WorkflowId   string      `json:"workflow_id"   orm:"workflow_id"   ` // 执行工作流 ID
	ClientIp     string      `json:"client_ip"     orm:"client_ip"     ` // 执行客户端 IP 快照
	MachineId    string      `json:"machine_id"    orm:"machine_id"    ` // 执行机器 ID
	NodeId       string      `json:"node_id"       orm:"node_id"       ` // 执行节点 ID
	ExecutionId  string      `json:"execution_id"  orm:"execution_id"  ` // 后端本次执行标识
	CommandId    string      `json:"command_id"    orm:"command_id"    ` // 对应节点命令 ID
	Status       string      `json:"status"        orm:"status"        ` // 尝试状态：pending、running、success、failed、timeout、cancelled
	ErrorMessage string      `json:"error_message" orm:"error_message" ` // 本次尝试失败原因
	StartedAt    *gtime.Time `json:"started_at"    orm:"started_at"    ` // 尝试开始时间
	FinishedAt   *gtime.Time `json:"finished_at"   orm:"finished_at"   ` // 尝试结束时间
	CreatedAt    *gtime.Time `json:"created_at"    orm:"created_at"    ` // 记录创建时间
	UpdatedAt    *gtime.Time `json:"updated_at"    orm:"updated_at"    ` // 记录更新时间
	DeletedAt    *gtime.Time `json:"deleted_at"    orm:"deleted_at"    ` // 软删除时间
}
