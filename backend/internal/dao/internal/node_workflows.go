// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NodeWorkflowsDao is the data access object for the table node_workflows.
type NodeWorkflowsDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  NodeWorkflowsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// NodeWorkflowsColumns defines and stores column names for the table node_workflows.
type NodeWorkflowsColumns struct {
	Id               string // 自增 ID
	NodeId           string // 执行节点 ID
	ClientIp         string // 客户端 IP 快照，用于区分同名执行节点
	AutomaId         string // Automa 原始工作流 ID
	WorkflowId       string // 服务端工作流主表 ID
	ExpectedRevision string // 服务端期望节点安装的工作流版本
	ActualRevision   string // 节点实际上报的工作流版本
	ContentHash      string // 节点当前工作流内容 hash
	DesiredPresent   string // 是否期望该工作流存在于节点中
	SyncStatus       string // 同步状态：pending、syncing、synced、failed、deleting、deleted
	LastSyncAt       string // 最近一次安装、更新或删除完成时间
	LastCheckAt      string // 最近一次节点清单上报时间
	LastError        string // 最近一次同步错误信息
	CreatedAt        string // 记录创建时间
	UpdatedAt        string // 记录更新时间
	DeletedAt        string // 软删除时间
}

// nodeWorkflowsColumns holds the columns for the table node_workflows.
var nodeWorkflowsColumns = NodeWorkflowsColumns{
	Id:               "id",
	NodeId:           "node_id",
	ClientIp:         "client_ip",
	AutomaId:         "automa_id",
	WorkflowId:       "workflow_id",
	ExpectedRevision: "expected_revision",
	ActualRevision:   "actual_revision",
	ContentHash:      "content_hash",
	DesiredPresent:   "desired_present",
	SyncStatus:       "sync_status",
	LastSyncAt:       "last_sync_at",
	LastCheckAt:      "last_check_at",
	LastError:        "last_error",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	DeletedAt:        "deleted_at",
}

// NewNodeWorkflowsDao creates and returns a new DAO object for table data access.
func NewNodeWorkflowsDao(handlers ...gdb.ModelHandler) *NodeWorkflowsDao {
	return &NodeWorkflowsDao{
		group:    "default",
		table:    "node_workflows",
		columns:  nodeWorkflowsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NodeWorkflowsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NodeWorkflowsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NodeWorkflowsDao) Columns() NodeWorkflowsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NodeWorkflowsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NodeWorkflowsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NodeWorkflowsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
