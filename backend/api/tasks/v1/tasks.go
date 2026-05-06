package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskPayload 任务配置载荷
type TaskPayload struct {
	Name               string        `json:"name" dc:"任务名称"`
	Description        string        `json:"description,omitempty" dc:"任务描述"`
	WorkflowID         string        `json:"workflow_id" dc:"工作流ID"`
	WorkflowName       string        `json:"workflow_name,omitempty" dc:"工作流名称"`
	ClientID           string        `json:"client_id,omitempty" dc:"客户端ID"`
	ClientName         string        `json:"client_name,omitempty" dc:"客户端名称"`
	ClientIP           string        `json:"client_ip,omitempty" dc:"客户端IP"`
	CronExpression     string        `json:"cron_expression,omitempty" dc:"Cron 表达式"`
	RunOnceAfterCreate *bool         `json:"run_once_after_create,omitempty" dc:"创建后是否立即运行一次"`
	Params             model.JSONMap `json:"params,omitempty" dc:"任务参数"`
	Enabled            *bool         `json:"enabled,omitempty" dc:"是否启用"`
}

// TaskListReq 任务列表请求
type TaskListReq struct {
	g.Meta `path:"/" method:"get" tags:"任务" summary:"获取任务列表"`
	rr.CommonTimeReq
	Keyword      string `json:"keyword,omitempty" in:"query" dc:"关键字"`
	WorkflowID   string `json:"workflow_id,omitempty" in:"query" dc:"工作流ID"`
	WorkflowName string `json:"workflow_name,omitempty" in:"query" dc:"工作流名称"`
	ClientID     string `json:"client_id,omitempty" in:"query" dc:"客户端ID"`
	Enabled      string `json:"enabled,omitempty" in:"query" dc:"启用状态"`
}

// TaskListResModel task list item 任务列表项
type TaskListResModel = model.TaskResModel

// TaskListRes 任务列表响应
type TaskListRes struct {
	List  []*TaskListResModel `json:"list,omitempty" dc:"任务列表"`
	Total int                 `json:"total" dc:"任务总数"`
}

// TaskDetailReq 任务详情请求
type TaskDetailReq struct {
	g.Meta `path:"/{id}" method:"get" tags:"任务" summary:"获取任务详情"`
	ID     string `json:"id" in:"path" dc:"任务ID"`
}

// TaskDetailRes 任务详情响应
type TaskDetailRes struct {
	Task *TaskListResModel `json:"task,omitempty" dc:"任务详情"`
}

// TaskCreateReq 新建任务请求
type TaskCreateReq struct {
	g.Meta `path:"/" method:"post" tags:"任务" summary:"新建任务"`
	TaskPayload
}

// TaskCreateRes 新建任务响应
type TaskCreateRes struct {
	Task *TaskListResModel `json:"task,omitempty" dc:"任务详情"`
}

// TaskUpdateReq 更新任务请求
type TaskUpdateReq struct {
	g.Meta `path:"/{id}" method:"put" tags:"任务" summary:"更新任务"`
	ID     string `json:"id" in:"path" dc:"任务ID"`
	TaskPayload
}

// TaskUpdateRes 更新任务响应
type TaskUpdateRes struct {
	Task *TaskListResModel `json:"task,omitempty" dc:"任务详情"`
}

// TaskDeleteReq 删除任务请求
type TaskDeleteReq struct {
	g.Meta `path:"/{id}" method:"delete" tags:"任务" summary:"删除任务"`
	ID     string `json:"id" in:"path" dc:"任务ID"`
}

// TaskDeleteRes 删除任务响应
type TaskDeleteRes struct {
	Message string `json:"message,omitempty" dc:"删除结果消息"`
}

// TaskExecuteReq 执行任务请求
type TaskExecuteReq struct {
	g.Meta      `path:"/{id}/execute" method:"post" tags:"任务" summary:"执行任务"`
	ID          string        `json:"id" in:"path" dc:"任务ID"`
	ClientID    string        `json:"client_id,omitempty" dc:"客户端ID"`
	ClientIP    string        `json:"client_ip,omitempty" dc:"客户端IP"`
	TriggerType string        `json:"trigger_type,omitempty" dc:"触发类型"`
	Params      model.JSONMap `json:"params,omitempty" dc:"执行参数"`
}

// TaskExecuteRes 执行任务响应
type TaskExecuteRes struct {
	Record *model.TaskRecordResModel `json:"record,omitempty" dc:"任务记录详情"`
	Result *model.AgentCommandResult `json:"result,omitempty" dc:"执行结果"`
}
