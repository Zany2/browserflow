// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskRecordFilesDao is the data access object for the table task_record_files.
type TaskRecordFilesDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  TaskRecordFilesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// TaskRecordFilesColumns defines and stores column names for the table task_record_files.
type TaskRecordFilesColumns struct {
	Id          string // 自增 ID
	RecordId    string // 关联任务执行记录 ID
	TaskId      string // 关联任务配置 ID
	WorkflowId  string // 执行时使用的 Automa 工作流 ID
	ClientIp    string // 执行客户端 IP
	NodeId      string // 执行节点 ID
	ExecutionId string // 后端本次执行标识
	FileType    string // 文件类型：table_json、table_csv、variables_json、log、screenshot 等
	FileName    string // 原始或展示文件名
	FilePath    string // 服务端保存路径或对象存储 Key
	MimeType    string // 文件 MIME 类型
	FileSize    string // 文件大小，单位字节
	RowCount    string // 表格结果行数
	PreviewJson string // 预览数据，例如前几行表格
	Remark      string // 备注
	CreatedAt   string // 记录创建时间
	UpdatedAt   string // 记录更新时间
	DeletedAt   string // 软删除时间
}

// taskRecordFilesColumns holds the columns for the table task_record_files.
var taskRecordFilesColumns = TaskRecordFilesColumns{
	Id:          "id",
	RecordId:    "record_id",
	TaskId:      "task_id",
	WorkflowId:  "workflow_id",
	ClientIp:    "client_ip",
	NodeId:      "node_id",
	ExecutionId: "execution_id",
	FileType:    "file_type",
	FileName:    "file_name",
	FilePath:    "file_path",
	MimeType:    "mime_type",
	FileSize:    "file_size",
	RowCount:    "row_count",
	PreviewJson: "preview_json",
	Remark:      "remark",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewTaskRecordFilesDao creates and returns a new DAO object for table data access.
func NewTaskRecordFilesDao(handlers ...gdb.ModelHandler) *TaskRecordFilesDao {
	return &TaskRecordFilesDao{
		group:    "default",
		table:    "task_record_files",
		columns:  taskRecordFilesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TaskRecordFilesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TaskRecordFilesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TaskRecordFilesDao) Columns() TaskRecordFilesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TaskRecordFilesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TaskRecordFilesDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *TaskRecordFilesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
