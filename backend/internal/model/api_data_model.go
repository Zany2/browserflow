package model

import (
	"encoding/json"

	"github.com/gogf/gf/v2/os/gtime"
)

// JSONMap keeps dynamic JSON object values without using any JSONMap 用 RawMessage 承载动态 JSON 对象
type JSONMap map[string]json.RawMessage

// TaskResModel task response item 任务响应项
type TaskResModel struct {
	// ID task id 任务 ID
	ID int64 `json:"id" dc:"任务ID"`
	// Name task name 任务名称
	Name string `json:"name" dc:"任务名称"`
	// Description task description 任务描述
	Description string `json:"description" dc:"任务描述"`
	// AutomaID Automa workflow id Automa 工作流 ID
	AutomaID string `json:"automa_id" dc:"Automa工作流ID"`
	// WorkflowID compatible workflow id 兼容工作流 ID
	WorkflowID string `json:"workflow_id" dc:"工作流ID"`
	// WorkflowName workflow display name 工作流名称
	WorkflowName string `json:"workflow_name" dc:"工作流名称"`
	// ClientID client unique id 客户端 ID
	ClientID string `json:"client_id" dc:"客户端ID"`
	// ClientName client display name 客户端名称
	ClientName string `json:"client_name" dc:"客户端名称"`
	// ClientIP client ip 客户端 IP
	ClientIP string `json:"client_ip" dc:"客户端IP"`
	// CronExpression cron expression Cron 表达式
	CronExpression string `json:"cron_expression" dc:"Cron表达式"`
	// Params task parameters 任务参数
	Params JSONMap `json:"params" dc:"任务参数"`
	// Enabled enabled flag 是否启用
	Enabled bool `json:"enabled" dc:"是否启用"`
	// CreatedAt created time 创建时间
	CreatedAt *gtime.Time `json:"created_at" dc:"创建时间"`
	// UpdatedAt updated time 更新时间
	UpdatedAt *gtime.Time `json:"updated_at" dc:"更新时间"`
	// DeletedAt deleted time 删除时间
	DeletedAt *gtime.Time `json:"deleted_at" dc:"删除时间"`
}

// TaskRecordResModel task execution record response item 任务执行记录响应项
type TaskRecordResModel struct {
	// ID record id 记录 ID
	ID int64 `json:"id" dc:"任务记录ID"`
	// TaskID task id 任务 ID
	TaskID int64 `json:"task_id" dc:"任务ID"`
	// TaskName task name 任务名称
	TaskName string `json:"task_name" dc:"任务名称"`
	// WorkflowID workflow id 工作流 ID
	WorkflowID string `json:"workflow_id" dc:"工作流ID"`
	// WorkflowName workflow display name 工作流名称
	WorkflowName string `json:"workflow_name" dc:"工作流名称"`
	// ClientID client unique id 客户端 ID
	ClientID string `json:"client_id" dc:"客户端ID"`
	// ClientName client display name 客户端名称
	ClientName string `json:"client_name" dc:"客户端名称"`
	// ClientIP client ip 客户端 IP
	ClientIP string `json:"client_ip" dc:"客户端IP"`
	// TriggerType execution trigger type 执行触发类型
	TriggerType string `json:"trigger_type" dc:"触发类型"`
	// Status execution status 执行状态
	Status string `json:"status" dc:"执行状态"`
	// Params execution parameters 执行参数
	Params JSONMap `json:"params" dc:"执行参数"`
	// Result execution result 执行结果
	Result JSONMap `json:"result" dc:"执行结果"`
	// ErrorMessage execution error message 执行错误信息
	ErrorMessage string `json:"error_message" dc:"错误信息"`
	// StartedAt started time 开始时间
	StartedAt *gtime.Time `json:"started_at" dc:"开始时间"`
	// FinishedAt finished time 结束时间
	FinishedAt *gtime.Time `json:"finished_at" dc:"结束时间"`
	// CreatedAt created time 创建时间
	CreatedAt *gtime.Time `json:"created_at" dc:"创建时间"`
	// UpdatedAt updated time 更新时间
	UpdatedAt *gtime.Time `json:"updated_at" dc:"更新时间"`
	// DeletedAt deleted time 删除时间
	DeletedAt *gtime.Time `json:"deleted_at" dc:"删除时间"`
}
