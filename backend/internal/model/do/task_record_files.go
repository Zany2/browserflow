// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskRecordFiles is the golang structure of table task_record_files for DAO operations like Where/Data.
type TaskRecordFiles struct {
	g.Meta      `orm:"table:task_record_files, do:true"`
	Id          interface{} // 自增ID
	RecordId    interface{} // 关联任务执行记录ID
	TaskId      interface{} // 关联任务配置ID
	WorkflowId  interface{} // 执行时使用的 Automa 工作流 ID
	ClientIp    interface{} // 执行客户端 IP
	FileType    interface{} // 文件类型：table_json、table_csv、variables_json、log、screenshot 等
	FileName    interface{} // 原始/展示文件名
	FilePath    interface{} // 服务端保存路径或对象存储 Key
	MimeType    interface{} // 文件 MIME 类型
	FileSize    interface{} // 文件大小，单位字节
	RowCount    interface{} // 表格结果行数
	PreviewJson interface{} // 预览数据，例如前几行表格
	Remark      interface{} // 备注
	CreatedAt   *gtime.Time // 记录创建时间
	UpdatedAt   *gtime.Time // 记录更新时间
	DeletedAt   *gtime.Time // 软删除时间
}
