// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskQueueDao is the data access object for the table task_queue.
type TaskQueueDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TaskQueueColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TaskQueueColumns defines and stores column names for the table task_queue.
type TaskQueueColumns struct {
	Id            string // 自增 ID
	TaskId        string // 关联任务配置 ID
	WorkflowId    string // 待执行 Automa 工作流 ID
	DispatchMode  string // 调度模式快照：auto、group、machine、node、ip
	ClientIp      string // 目标客户端 IP 快照，仅用于兼容旧逻辑
	MachineId     string // 目标机器 ID
	NodeId        string // 目标节点 ID
	TargetGroupId string // 目标节点分组 ID
	TriggerType   string // 触发类型
	Priority      string // 优先级，数字越大优先级越高
	Status        string // 队列状态：pending、reserved、running、success、failed、cancelled、skipped
	AttemptCount  string // 已尝试次数
	MaxAttempts   string // 最大尝试次数
	ParamsJson    string // 本次排队任务参数快照
	LastError     string // 最近一次失败原因
	ScheduledAt   string // 计划执行时间
	ReservedAt    string // 被调度器占用时间
	StartedAt     string // 开始执行时间
	FinishedAt    string // 执行完成时间
	CreatedAt     string // 记录创建时间
	UpdatedAt     string // 记录更新时间
	DeletedAt     string // 软删除时间
}

// taskQueueColumns holds the columns for the table task_queue.
var taskQueueColumns = TaskQueueColumns{
	Id:            "id",
	TaskId:        "task_id",
	WorkflowId:    "workflow_id",
	DispatchMode:  "dispatch_mode",
	ClientIp:      "client_ip",
	MachineId:     "machine_id",
	NodeId:        "node_id",
	TargetGroupId: "target_group_id",
	TriggerType:   "trigger_type",
	Priority:      "priority",
	Status:        "status",
	AttemptCount:  "attempt_count",
	MaxAttempts:   "max_attempts",
	ParamsJson:    "params_json",
	LastError:     "last_error",
	ScheduledAt:   "scheduled_at",
	ReservedAt:    "reserved_at",
	StartedAt:     "started_at",
	FinishedAt:    "finished_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
}

// NewTaskQueueDao creates and returns a new DAO object for table data access.
func NewTaskQueueDao(handlers ...gdb.ModelHandler) *TaskQueueDao {
	return &TaskQueueDao{
		group:    "default",
		table:    "task_queue",
		columns:  taskQueueColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TaskQueueDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TaskQueueDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TaskQueueDao) Columns() TaskQueueColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TaskQueueDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TaskQueueDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TaskQueueDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
