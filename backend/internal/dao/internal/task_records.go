// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskRecordsDao is the data access object for the table task_records.
type TaskRecordsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TaskRecordsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TaskRecordsColumns defines and stores column names for the table task_records.
type TaskRecordsColumns struct {
	Id           string // 自增ID
	TaskId       string // 关联的任务配置ID
	WorkflowId   string // 执行时使用的 Automa 工作流 ID
	ClientIp     string // 执行目标客户端 IP
	TriggerType  string // 触发类型：manual 手动触发，cron 定时触发，task_create 创建任务触发，skill Skill触发，system 系统触发
	Status       string // 执行状态：pending、queued、running、success、failed、cancelled
	ParamsJson   string // 本次执行使用的参数快照
	ResultJson   string // 本次执行结果内容
	ErrorMessage string // 执行失败时的错误信息
	StartedAt    string // 开始执行时间
	FinishedAt   string // 执行结束时间
	CreatedAt    string // 记录创建时间
	UpdatedAt    string // 记录更新时间
	DeletedAt    string // 软删除时间
}

// taskRecordsColumns holds the columns for the table task_records.
var taskRecordsColumns = TaskRecordsColumns{
	Id:           "id",
	TaskId:       "task_id",
	WorkflowId:   "workflow_id",
	ClientIp:     "client_ip",
	TriggerType:  "trigger_type",
	Status:       "status",
	ParamsJson:   "params_json",
	ResultJson:   "result_json",
	ErrorMessage: "error_message",
	StartedAt:    "started_at",
	FinishedAt:   "finished_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewTaskRecordsDao creates and returns a new DAO object for table data access.
func NewTaskRecordsDao(handlers ...gdb.ModelHandler) *TaskRecordsDao {
	return &TaskRecordsDao{
		group:    "default",
		table:    "task_records",
		columns:  taskRecordsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TaskRecordsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TaskRecordsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TaskRecordsDao) Columns() TaskRecordsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TaskRecordsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TaskRecordsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TaskRecordsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
