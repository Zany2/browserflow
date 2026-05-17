package model

import (
	"encoding/json"
	"time"
)

const (
	// WorkflowExecutionStatusQueued queued status 排队状态
	WorkflowExecutionStatusQueued = "queued"
	// WorkflowExecutionStatusRunning running status 运行中状态
	WorkflowExecutionStatusRunning = "running"
	// WorkflowExecutionStatusSuccess success status 成功状态
	WorkflowExecutionStatusSuccess = "success"
	// WorkflowExecutionStatusError error status 失败状态
	WorkflowExecutionStatusError = "error"
	// WorkflowExecutionStatusStopped stopped status 已停止状态
	WorkflowExecutionStatusStopped = "stopped"
	// WorkflowExecutionStatusTimeout timeout status 超时状态
	WorkflowExecutionStatusTimeout = "timeout"
)

// WorkflowExecutionReturnData requested return data config 回传数据配置
type WorkflowExecutionReturnData struct {
	// Variables variable names to return 需要回传的变量名
	Variables []string `json:"variables,omitempty"`
	// IncludeTable includes table data 是否回传表格数据
	IncludeTable bool `json:"include_table,omitempty"`
	// TableLimit table row limit 表格行数限制
	TableLimit int `json:"table_limit,omitempty"`
	// IncludeHistory includes workflow history 是否回传执行日志
	IncludeHistory bool `json:"include_history,omitempty"`
}

// WorkflowExecution workflow execution state 工作流执行状态
type WorkflowExecution struct {
	// ExecutionID execution id 执行标识
	ExecutionID string `json:"execution_id"`
	// WorkflowID workflow id 工作流标识
	WorkflowID string `json:"workflow_id"`
	// BrowserID browser agent id 浏览器执行端标识
	BrowserID string `json:"browser_id"`
	// CommandID websocket command id WebSocket 命令标识
	CommandID string `json:"command_id,omitempty"`
	// Status execution status 执行状态
	Status string `json:"status"`
	// WaitResult whether HTTP caller waited 是否同步等待结果
	WaitResult bool `json:"wait_result"`
	// ReturnData requested return data config 回传数据配置
	ReturnData *WorkflowExecutionReturnData `json:"return_data,omitempty"`
	// Result raw agent result 原始执行结果
	Result json.RawMessage `json:"result,omitempty"`
	// Message readable result message 可读结果信息
	Message string `json:"message,omitempty"`
	// Error error message 错误信息
	Error string `json:"error,omitempty"`
	// CreatedAt created time 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt updated time 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// StartedAt workflow started time 工作流开始时间
	StartedAt *time.Time `json:"started_at,omitempty"`
	// EndedAt workflow ended time 工作流结束时间
	EndedAt *time.Time `json:"ended_at,omitempty"`
}

// WorkflowExecutionCreateInput execution create input 执行创建参数
type WorkflowExecutionCreateInput struct {
	// ExecutionID execution id 执行标识
	ExecutionID string
	// WorkflowID workflow id 工作流标识
	WorkflowID string
	// BrowserID browser agent id 浏览器执行端标识
	BrowserID string
	// CommandID websocket command id WebSocket 命令标识
	CommandID string
	// WaitResult whether HTTP caller waits 是否同步等待
	WaitResult bool
	// ReturnData requested return data config 回传数据配置
	ReturnData *WorkflowExecutionReturnData
}
