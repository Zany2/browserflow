package model

import "github.com/gogf/gf/v2/os/gtime"

// TaskRecordFileResModel task execution record file response item.
type TaskRecordFileResModel struct {
	ID          int64       `json:"id" dc:"File ID"`
	RecordID    int64       `json:"record_id" dc:"Task record ID"`
	TaskID      int64       `json:"task_id" dc:"Task ID"`
	WorkflowID  string      `json:"workflow_id" dc:"Workflow ID"`
	ClientIP    string      `json:"client_ip" dc:"Client IP"`
	MachineID   string      `json:"machine_id" dc:"Machine ID"`
	NodeID      string      `json:"node_id" dc:"Node ID"`
	ExecutionID string      `json:"execution_id" dc:"Execution ID"`
	FileType    string      `json:"file_type" dc:"File type"`
	FileName    string      `json:"file_name" dc:"File name"`
	FilePath    string      `json:"file_path" dc:"File path"`
	MimeType    string      `json:"mime_type" dc:"MIME type"`
	FileSize    int64       `json:"file_size" dc:"File size"`
	RowCount    int         `json:"row_count" dc:"Row count"`
	Preview     JSONMap     `json:"preview" dc:"Preview data"`
	Remark      string      `json:"remark" dc:"Remark"`
	CreatedAt   *gtime.Time `json:"created_at" dc:"Created time"`
	UpdatedAt   *gtime.Time `json:"updated_at" dc:"Updated time"`
	DeletedAt   *gtime.Time `json:"deleted_at" dc:"Deleted time"`
}
