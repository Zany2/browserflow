// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskRecords is the golang structure for table task_records.
type TaskRecords struct {
	Id           int64       `json:"id"            orm:"id"            ` // 自增ID
	TaskId       int64       `json:"task_id"       orm:"task_id"       ` // 关联的任务配置ID
	WorkflowId   string      `json:"workflow_id"   orm:"workflow_id"   ` // 执行时使用的 Automa 工作流 ID
	ClientIp     string      `json:"client_ip"     orm:"client_ip"     ` // 执行目标客户端 IP
	TriggerType  string      `json:"trigger_type"  orm:"trigger_type"  ` // 触发类型：manual 手动触发，cron 定时触发，task_create 创建任务触发，skill Skill触发，system 系统触发
	Status       string      `json:"status"        orm:"status"        ` // 执行状态：pending、queued、running、success、failed、cancelled
	ParamsJson   string      `json:"params_json"   orm:"params_json"   ` // 本次执行使用的参数快照
	ResultJson   string      `json:"result_json"   orm:"result_json"   ` // 本次执行结果内容
	ErrorMessage string      `json:"error_message" orm:"error_message" ` // 执行失败时的错误信息
	StartedAt    *gtime.Time `json:"started_at"    orm:"started_at"    ` // 开始执行时间
	FinishedAt   *gtime.Time `json:"finished_at"   orm:"finished_at"   ` // 执行结束时间
	CreatedAt    *gtime.Time `json:"created_at"    orm:"created_at"    ` // 记录创建时间
	UpdatedAt    *gtime.Time `json:"updated_at"    orm:"updated_at"    ` // 记录更新时间
	DeletedAt    *gtime.Time `json:"deleted_at"    orm:"deleted_at"    ` // 软删除时间
}
