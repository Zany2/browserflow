// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskQueue is the golang structure for table task_queue.
type TaskQueue struct {
	Id            int64       `json:"id"              orm:"id"              ` // 自增 ID
	TaskId        int64       `json:"task_id"         orm:"task_id"         ` // 关联任务配置 ID
	WorkflowId    string      `json:"workflow_id"     orm:"workflow_id"     ` // 待执行 Automa 工作流 ID
	DispatchMode  string      `json:"dispatch_mode"   orm:"dispatch_mode"   ` // 调度模式快照：auto、group、node、ip
	ClientIp      string      `json:"client_ip"       orm:"client_ip"       ` // 目标客户端 IP 快照
	NodeId        string      `json:"node_id"         orm:"node_id"         ` // 目标节点 ID
	TargetGroupId int64       `json:"target_group_id" orm:"target_group_id" ` // 目标节点分组 ID
	TriggerType   string      `json:"trigger_type"    orm:"trigger_type"    ` // 触发类型
	Priority      int         `json:"priority"        orm:"priority"        ` // 优先级，数字越大优先级越高
	Status        string      `json:"status"          orm:"status"          ` // 队列状态：pending、reserved、running、success、failed、cancelled、skipped
	AttemptCount  int         `json:"attempt_count"   orm:"attempt_count"   ` // 已尝试次数
	MaxAttempts   int         `json:"max_attempts"    orm:"max_attempts"    ` // 最大尝试次数
	ParamsJson    string      `json:"params_json"     orm:"params_json"     ` // 本次排队任务参数快照
	LastError     string      `json:"last_error"      orm:"last_error"      ` // 最近一次失败原因
	ScheduledAt   *gtime.Time `json:"scheduled_at"    orm:"scheduled_at"    ` // 计划执行时间
	ReservedAt    *gtime.Time `json:"reserved_at"     orm:"reserved_at"     ` // 被调度器占用时间
	StartedAt     *gtime.Time `json:"started_at"      orm:"started_at"      ` // 开始执行时间
	FinishedAt    *gtime.Time `json:"finished_at"     orm:"finished_at"     ` // 执行完成时间
	CreatedAt     *gtime.Time `json:"created_at"      orm:"created_at"      ` // 记录创建时间
	UpdatedAt     *gtime.Time `json:"updated_at"      orm:"updated_at"      ` // 记录更新时间
	DeletedAt     *gtime.Time `json:"deleted_at"      orm:"deleted_at"      ` // 软删除时间
}
