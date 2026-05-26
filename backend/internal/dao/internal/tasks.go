// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TasksDao is the data access object for the table tasks.
type TasksDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TasksColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TasksColumns defines and stores column names for the table tasks.
type TasksColumns struct {
	Id                        string // 自增 ID
	Name                      string // 任务名称
	Description               string // 任务说明
	AutomaId                  string // Automa 原始工作流 ID
	ClientIp                  string // 目标客户端 IP，为空表示由服务端自动选择
	NodeId                    string // 目标执行节点 ID，为空表示由服务端选择可用节点
	TargetGroupId             string // 目标节点分组 ID，为 0 表示不限定分组
	DispatchMode              string // 调度模式：auto 自动选择，group 指定分组，node 指定节点，ip 指定 IP
	QueuePolicy               string // 繁忙策略：queue 排队，fail 直接失败，skip 跳过本次
	ConflictWindowSeconds     string // 创建或编辑定时任务时用于冲突提醒的时间窗口秒数
	MaxAttempts               string // 任务最多尝试次数
	TimeoutSeconds            string // 任务执行超时时间，单位秒
	QueueWaitSeconds          string // 等待可用最大等待时间，单位秒，仅 queue_policy=queue 时生效
	QueueRetryIntervalSeconds string // 等待可用重试间隔，单位秒，仅 queue_policy=queue 时生效
	CronExpression            string // Cron 表达式，为空表示立即执行
	ParamsJson                string // 任务自定义参数快照，JSONB 存储
	Enabled                   string // 是否启用
	CreatedAt                 string // 创建时间
	UpdatedAt                 string // 更新时间
	DeletedAt                 string // 软删除时间
}

// tasksColumns holds the columns for the table tasks.
var tasksColumns = TasksColumns{
	Id:                        "id",
	Name:                      "name",
	Description:               "description",
	AutomaId:                  "automa_id",
	ClientIp:                  "client_ip",
	NodeId:                    "node_id",
	TargetGroupId:             "target_group_id",
	DispatchMode:              "dispatch_mode",
	QueuePolicy:               "queue_policy",
	ConflictWindowSeconds:     "conflict_window_seconds",
	MaxAttempts:               "max_attempts",
	TimeoutSeconds:            "timeout_seconds",
	QueueWaitSeconds:          "queue_wait_seconds",
	QueueRetryIntervalSeconds: "queue_retry_interval_seconds",
	CronExpression:            "cron_expression",
	ParamsJson:                "params_json",
	Enabled:                   "enabled",
	CreatedAt:                 "created_at",
	UpdatedAt:                 "updated_at",
	DeletedAt:                 "deleted_at",
}

// NewTasksDao creates and returns a new DAO object for table data access.
func NewTasksDao(handlers ...gdb.ModelHandler) *TasksDao {
	return &TasksDao{
		group:    "default",
		table:    "tasks",
		columns:  tasksColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TasksDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TasksDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TasksDao) Columns() TasksColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TasksDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TasksDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TasksDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
