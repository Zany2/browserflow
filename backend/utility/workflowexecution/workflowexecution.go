package workflowexecution

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/util/guid"
)

const (
	// defaultTableLimit default table row limit 默认表格行数限制
	defaultTableLimit = 20
	// maxExecutionCount max retained execution count 最大保留执行数
	maxExecutionCount = 500
)

var (
	// mu protects execution store 保护执行状态存储
	mu sync.RWMutex
	// executions stores execution states 执行状态存储
	executions = map[string]*model.WorkflowExecution{}
	// commandIndex maps command id to execution id 命令到执行标识索引
	commandIndex = map[string]string{}
)

// NewID creates execution id 创建执行标识
func NewID() string {
	return "exec_" + guid.S()
}

// NormalizeReturnData normalizes return data request 规范化回传数据配置
func NormalizeReturnData(in *model.WorkflowExecutionReturnData) *model.WorkflowExecutionReturnData {
	if in == nil {
		return nil
	}

	variables := make([]string, 0, len(in.Variables))
	for _, variable := range in.Variables {
		variable = strings.TrimSpace(variable)
		if variable != "" {
			variables = append(variables, variable)
		}
	}

	tableLimit := in.TableLimit
	if tableLimit <= 0 {
		tableLimit = defaultTableLimit
	}

	return &model.WorkflowExecutionReturnData{
		Variables:      variables,
		IncludeTable:   in.IncludeTable,
		TableLimit:     tableLimit,
		IncludeHistory: in.IncludeHistory,
	}
}

// Create stores initial execution state 创建执行状态
func Create(in model.WorkflowExecutionCreateInput) *model.WorkflowExecution {
	now := time.Now()
	executionID := strings.TrimSpace(in.ExecutionID)
	if executionID == "" {
		executionID = NewID()
	}

	execution := &model.WorkflowExecution{
		ExecutionID: executionID,
		WorkflowID:  strings.TrimSpace(in.WorkflowID),
		BrowserID:   strings.TrimSpace(in.BrowserID),
		CommandID:   strings.TrimSpace(in.CommandID),
		Status:      model.WorkflowExecutionStatusQueued,
		WaitResult:  in.WaitResult,
		ReturnData:  NormalizeReturnData(in.ReturnData),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mu.Lock()
	executions[execution.ExecutionID] = execution
	if execution.CommandID != "" {
		commandIndex[execution.CommandID] = execution.ExecutionID
	}
	pruneLocked()
	mu.Unlock()

	return cloneExecution(execution)
}

// MarkRunning marks execution as running 标记执行中
func MarkRunning(executionID string) {
	update(executionID, func(execution *model.WorkflowExecution) {
		execution.Status = model.WorkflowExecutionStatusRunning
		now := time.Now()
		execution.StartedAt = &now
	})
}

// CompleteByCommand completes execution by command result 根据命令结果完成执行状态
func CompleteByCommand(commandID string, result *model.AgentCommandResult) {
	commandID = strings.TrimSpace(commandID)
	if commandID == "" || result == nil {
		return
	}

	mu.RLock()
	executionID := commandIndex[commandID]
	mu.RUnlock()
	if executionID == "" {
		return
	}

	Complete(executionID, result)
}

// Complete completes execution from agent result 根据执行端结果完成执行状态
func Complete(executionID string, result *model.AgentCommandResult) {
	if result == nil {
		return
	}

	update(executionID, func(execution *model.WorkflowExecution) {
		status := resolveResultStatus(result)
		now := time.Now()
		execution.Status = status
		execution.Result = cloneRaw(result.Data)
		execution.Error = strings.TrimSpace(result.Error)
		execution.Message = resolveResultMessage(result)
		if execution.StartedAt == nil {
			execution.StartedAt = resolveResultTime(result.Data, "started_at", "startedAt", "startedTimestamp")
		}
		if endedAt := resolveResultTime(result.Data, "ended_at", "endedAt", "endedTimestamp"); endedAt != nil {
			execution.EndedAt = endedAt
		} else {
			execution.EndedAt = &now
		}
	})
}

// MarkTimeout marks execution timeout 标记执行超时
func MarkTimeout(executionID string, message string) {
	update(executionID, func(execution *model.WorkflowExecution) {
		now := time.Now()
		execution.Status = model.WorkflowExecutionStatusTimeout
		execution.Message = strings.TrimSpace(message)
		execution.Error = strings.TrimSpace(message)
		execution.EndedAt = &now
	})
}

// Get returns execution state 获取执行状态
func Get(executionID string) (*model.WorkflowExecution, bool) {
	mu.RLock()
	execution, ok := executions[strings.TrimSpace(executionID)]
	mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneExecution(execution), true
}

// update updates one execution 更新单个执行状态
func update(executionID string, fn func(execution *model.WorkflowExecution)) {
	executionID = strings.TrimSpace(executionID)
	if executionID == "" || fn == nil {
		return
	}

	mu.Lock()
	execution := executions[executionID]
	if execution != nil {
		fn(execution)
		execution.UpdatedAt = time.Now()
	}
	mu.Unlock()
}

// resolveResultStatus resolves terminal status 解析终态状态
func resolveResultStatus(result *model.AgentCommandResult) string {
	if !result.Success {
		return model.WorkflowExecutionStatusError
	}

	var data map[string]any
	if len(result.Data) > 0 && json.Unmarshal(result.Data, &data) == nil {
		if status := strings.TrimSpace(stringValue(data["status"])); status != "" {
			if status == model.WorkflowExecutionStatusQueued && result.Success {
				return model.WorkflowExecutionStatusRunning
			}
			return status
		}
		if okValue, ok := data["ok"].(bool); ok && !okValue {
			return model.WorkflowExecutionStatusError
		}
	}

	return model.WorkflowExecutionStatusSuccess
}

// resolveResultMessage resolves readable result message 解析可读结果信息
func resolveResultMessage(result *model.AgentCommandResult) string {
	if strings.TrimSpace(result.Error) != "" {
		return strings.TrimSpace(result.Error)
	}

	var data map[string]any
	if len(result.Data) > 0 && json.Unmarshal(result.Data, &data) == nil {
		for _, key := range []string{"message", "error", "error_message", "errorMessage"} {
			if value := strings.TrimSpace(stringValue(data[key])); value != "" {
				return value
			}
		}
	}

	return ""
}

// resolveResultTime resolves timestamp field 解析时间字段
func resolveResultTime(raw json.RawMessage, keys ...string) *time.Time {
	var data map[string]any
	if len(raw) == 0 || json.Unmarshal(raw, &data) != nil {
		return nil
	}

	for _, key := range keys {
		if parsed := parseAnyTime(data[key]); parsed != nil {
			return parsed
		}
	}
	return nil
}

// parseAnyTime parses common timestamp values 解析常见时间值
func parseAnyTime(value any) *time.Time {
	switch typed := value.(type) {
	case float64:
		return parseUnixLike(int64(typed))
	case int64:
		return parseUnixLike(typed)
	case int:
		return parseUnixLike(int64(typed))
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return nil
		}
		if parsed, err := time.Parse(time.RFC3339, text); err == nil {
			return &parsed
		}
	}
	return nil
}

// parseUnixLike parses seconds or milliseconds timestamp 解析秒或毫秒时间戳
func parseUnixLike(value int64) *time.Time {
	if value <= 0 {
		return nil
	}
	if value < 10000000000 {
		value *= 1000
	}
	parsed := time.UnixMilli(value)
	return &parsed
}

// stringValue converts value to string 转换字符串
func stringValue(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return strings.TrimSpace(strings.Trim(string(mustJSON(typed)), `"`))
	}
}

// mustJSON marshals value to json bytes 序列化 JSON
func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

// cloneRaw clones raw message 复制原始 JSON
func cloneRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	next := make([]byte, len(raw))
	copy(next, raw)
	return next
}

// cloneExecution clones execution state 复制执行状态
func cloneExecution(execution *model.WorkflowExecution) *model.WorkflowExecution {
	if execution == nil {
		return nil
	}
	next := *execution
	next.Result = cloneRaw(execution.Result)
	if execution.ReturnData != nil {
		returnData := *execution.ReturnData
		returnData.Variables = append([]string(nil), execution.ReturnData.Variables...)
		next.ReturnData = &returnData
	}
	if execution.StartedAt != nil {
		startedAt := *execution.StartedAt
		next.StartedAt = &startedAt
	}
	if execution.EndedAt != nil {
		endedAt := *execution.EndedAt
		next.EndedAt = &endedAt
	}
	return &next
}

// pruneLocked keeps recent executions only 保留最近执行状态
func pruneLocked() {
	if len(executions) <= maxExecutionCount {
		return
	}

	var oldestID string
	var oldestTime time.Time
	for id, execution := range executions {
		if oldestID == "" || execution.UpdatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = execution.UpdatedAt
		}
	}
	if oldestID != "" {
		if commandID := executions[oldestID].CommandID; commandID != "" {
			delete(commandIndex, commandID)
		}
		delete(executions, oldestID)
	}
}
