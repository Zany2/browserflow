// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskAttemptsDao is the data access object for the table task_attempts.
type TaskAttemptsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  TaskAttemptsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// TaskAttemptsColumns defines and stores column names for the table task_attempts.
type TaskAttemptsColumns struct {
	Id           string // 自增 ID
	QueueId      string // 对应任务队列 ID
	TaskId       string // 关联任务配置 ID
	RecordId     string // 关联任务执行记录 ID
	AttemptNo    string // 尝试序号，从 1 开始
	WorkflowId   string // 执行工作流 ID
	ClientIp     string // 执行客户端 IP 快照
	MachineId    string // 执行机器 ID
	NodeId       string // 执行节点 ID
	ExecutionId  string // 后端本次执行标识
	CommandId    string // 对应节点命令 ID
	Status       string // 尝试状态：pending、running、success、failed、timeout、cancelled
	ErrorMessage string // 本次尝试失败原因
	StartedAt    string // 尝试开始时间
	FinishedAt   string // 尝试结束时间
	CreatedAt    string // 记录创建时间
	UpdatedAt    string // 记录更新时间
	DeletedAt    string // 软删除时间
}

// taskAttemptsColumns holds the columns for the table task_attempts.
var taskAttemptsColumns = TaskAttemptsColumns{
	Id:           "id",
	QueueId:      "queue_id",
	TaskId:       "task_id",
	RecordId:     "record_id",
	AttemptNo:    "attempt_no",
	WorkflowId:   "workflow_id",
	ClientIp:     "client_ip",
	MachineId:    "machine_id",
	NodeId:       "node_id",
	ExecutionId:  "execution_id",
	CommandId:    "command_id",
	Status:       "status",
	ErrorMessage: "error_message",
	StartedAt:    "started_at",
	FinishedAt:   "finished_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewTaskAttemptsDao creates and returns a new DAO object for table data access.
func NewTaskAttemptsDao(handlers ...gdb.ModelHandler) *TaskAttemptsDao {
	return &TaskAttemptsDao{
		group:    "default",
		table:    "task_attempts",
		columns:  taskAttemptsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TaskAttemptsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TaskAttemptsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TaskAttemptsDao) Columns() TaskAttemptsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TaskAttemptsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TaskAttemptsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TaskAttemptsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
