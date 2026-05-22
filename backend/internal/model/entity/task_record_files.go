// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskRecordFiles is the golang structure for table task_record_files.
type TaskRecordFiles struct {
	Id          int64       `json:"id"           orm:"id"           ` // 自增ID
	RecordId    int64       `json:"record_id"    orm:"record_id"    ` // 关联任务执行记录ID
	TaskId      int64       `json:"task_id"      orm:"task_id"      ` // 关联任务配置ID
	WorkflowId  string      `json:"workflow_id"  orm:"workflow_id"  ` // 执行时使用的 Automa 工作流 ID
	ClientIp    string      `json:"client_ip"    orm:"client_ip"    ` // 执行客户端 IP
	FileType    string      `json:"file_type"    orm:"file_type"    ` // 文件类型：table_json、table_csv、variables_json、log、screenshot 等
	FileName    string      `json:"file_name"    orm:"file_name"    ` // 原始/展示文件名
	FilePath    string      `json:"file_path"    orm:"file_path"    ` // 服务端保存路径或对象存储 Key
	MimeType    string      `json:"mime_type"    orm:"mime_type"    ` // 文件 MIME 类型
	FileSize    int64       `json:"file_size"    orm:"file_size"    ` // 文件大小，单位字节
	RowCount    int         `json:"row_count"    orm:"row_count"    ` // 表格结果行数
	PreviewJson string      `json:"preview_json" orm:"preview_json" ` // 预览数据，例如前几行表格
	Remark      string      `json:"remark"       orm:"remark"       ` // 备注
	CreatedAt   *gtime.Time `json:"created_at"   orm:"created_at"   ` // 记录创建时间
	UpdatedAt   *gtime.Time `json:"updated_at"   orm:"updated_at"   ` // 记录更新时间
	DeletedAt   *gtime.Time `json:"deleted_at"   orm:"deleted_at"   ` // 软删除时间
}
