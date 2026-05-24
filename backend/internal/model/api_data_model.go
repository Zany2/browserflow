package model

import (
	"encoding/json"

	"github.com/gogf/gf/v2/os/gtime"
)

// JSONMap keeps dynamic JSON object values without using any.
type JSONMap map[string]json.RawMessage

// TaskResModel task response item.
type TaskResModel struct {
	ID             int64       `json:"id" dc:"任务ID"`
	Name           string      `json:"name" dc:"任务名称"`
	Description    string      `json:"description" dc:"任务描述"`
	AutomaID       string      `json:"automa_id" dc:"Automa工作流ID"`
	WorkflowID     string      `json:"workflow_id" dc:"工作流ID"`
	WorkflowName   string      `json:"workflow_name" dc:"工作流名称"`
	ClientID       string      `json:"client_id" dc:"客户端ID"`
	ClientName     string      `json:"client_name" dc:"客户端名称"`
	ClientIP       string      `json:"client_ip" dc:"客户端IP"`
	CronExpression string      `json:"cron_expression" dc:"Cron表达式"`
	Params         JSONMap     `json:"params" dc:"任务参数"`
	Enabled        bool        `json:"enabled" dc:"是否启用"`
	CreatedAt      *gtime.Time `json:"created_at" dc:"创建时间"`
	LastExecutedAt *gtime.Time `json:"last_executed_at" dc:"最近执行时间"`
	UpdatedAt      *gtime.Time `json:"updated_at" dc:"更新时间"`
	DeletedAt      *gtime.Time `json:"deleted_at" dc:"删除时间"`
}

// TaskRecordResModel task execution record response item.
type TaskRecordResModel struct {
	ID                int64       `json:"id" dc:"任务记录ID"`
	TaskID            int64       `json:"task_id" dc:"任务ID"`
	TaskName          string      `json:"task_name" dc:"任务名称"`
	WorkflowID        string      `json:"workflow_id" dc:"工作流ID"`
	WorkflowName      string      `json:"workflow_name" dc:"工作流名称"`
	ClientID          string      `json:"client_id" dc:"客户端ID"`
	ClientName        string      `json:"client_name" dc:"客户端名称"`
	ClientIP          string      `json:"client_ip" dc:"客户端IP"`
	ExecutionID       string      `json:"execution_id" dc:"后端本次执行标识"`
	AutomaExecutionID string      `json:"automa_execution_id" dc:"Automa客户端侧执行实例ID"`
	TriggerType       string      `json:"trigger_type" dc:"触发类型"`
	Status            string      `json:"status" dc:"执行状态"`
	Params            JSONMap     `json:"params" dc:"执行参数"`
	Result            JSONMap     `json:"result" dc:"执行结果"`
	ErrorMessage      string      `json:"error_message" dc:"错误信息"`
	DurationMs        int64       `json:"duration_ms" dc:"执行耗时毫秒数"`
	StartedAt         *gtime.Time `json:"started_at" dc:"开始时间"`
	FinishedAt        *gtime.Time `json:"finished_at" dc:"结束时间"`
	CreatedAt         *gtime.Time `json:"created_at" dc:"创建时间"`
	UpdatedAt         *gtime.Time `json:"updated_at" dc:"更新时间"`
	DeletedAt         *gtime.Time `json:"deleted_at" dc:"删除时间"`
}
