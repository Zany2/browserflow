// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NodeEventsDao is the data access object for the table node_events.
type NodeEventsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  NodeEventsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// NodeEventsColumns defines and stores column names for the table node_events.
type NodeEventsColumns struct {
	Id          string // 自增 ID
	MachineId   string // 机器 ID
	NodeId      string // 执行节点 ID
	ClientIp    string // 客户端 IP 快照
	EventType   string // 事件类型：register、heartbeat、disconnect、command、workflow_sync、package_install、error 等
	Level       string // 事件级别：debug、info、warn、error
	Message     string // 可读事件消息
	PayloadJson string // 事件扩展数据 JSON
	CreatedAt   string // 记录创建时间
	UpdatedAt   string // 记录更新时间
	DeletedAt   string // 软删除时间
}

// nodeEventsColumns holds the columns for the table node_events.
var nodeEventsColumns = NodeEventsColumns{
	Id:          "id",
	MachineId:   "machine_id",
	NodeId:      "node_id",
	ClientIp:    "client_ip",
	EventType:   "event_type",
	Level:       "level",
	Message:     "message",
	PayloadJson: "payload_json",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewNodeEventsDao creates and returns a new DAO object for table data access.
func NewNodeEventsDao(handlers ...gdb.ModelHandler) *NodeEventsDao {
	return &NodeEventsDao{
		group:    "default",
		table:    "node_events",
		columns:  nodeEventsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NodeEventsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NodeEventsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NodeEventsDao) Columns() NodeEventsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NodeEventsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NodeEventsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NodeEventsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
