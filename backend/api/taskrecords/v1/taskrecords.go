package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskRecordListReq 任务记录列表请求
type TaskRecordListReq struct {
	g.Meta `path:"/" method:"get" tags:"任务记录" summary:"获取任务记录列表"`
	rr.CommonTimeReq
	TaskID       string `json:"task_id,omitempty" in:"query" dc:"任务ID"`
	WorkflowID   string `json:"workflow_id,omitempty" in:"query" dc:"工作流ID"`
	WorkflowName string `json:"workflow_name,omitempty" in:"query" dc:"工作流名称"`
	ClientID     string `json:"client_id,omitempty" in:"query" dc:"客户端ID"`
	ClientIP     string `json:"client_ip,omitempty" in:"query" dc:"客户端IP"`
	Status       string `json:"status,omitempty" in:"query" dc:"任务记录状态"`
	Keyword      string `json:"keyword,omitempty" in:"query" dc:"关键字"`
}

// TaskRecordListResModel task record list item 任务记录列表项
type TaskRecordListResModel = model.TaskRecordResModel

// TaskRecordListRes 任务记录列表响应
type TaskRecordListRes struct {
	List  []*TaskRecordListResModel `json:"list,omitempty" dc:"任务记录列表"`
	Total int                       `json:"total" dc:"任务记录总数"`
}

// TaskRecordDetailReq 任务记录详情请求
type TaskRecordDetailReq struct {
	g.Meta `path:"/{id}" method:"get" tags:"任务记录" summary:"获取任务记录详情"`
	ID     string `json:"id" in:"path" dc:"任务记录ID"`
}

// TaskRecordDetailRes 任务记录详情响应
type TaskRecordDetailRes struct {
	Record *TaskRecordListResModel `json:"record,omitempty" dc:"任务记录详情"`
}
