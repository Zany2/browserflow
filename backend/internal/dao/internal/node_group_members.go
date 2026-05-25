// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NodeGroupMembersDao is the data access object for the table node_group_members.
type NodeGroupMembersDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  NodeGroupMembersColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// NodeGroupMembersColumns defines and stores column names for the table node_group_members.
type NodeGroupMembersColumns struct {
	Id        string // 自增 ID
	GroupId   string // 节点分组 ID
	ClientIp  string // 客户端 IP 快照
	NodeId    string // 执行节点 ID
	CreatedAt string // 记录创建时间
	UpdatedAt string // 记录更新时间
	DeletedAt string // 软删除时间
}

// nodeGroupMembersColumns holds the columns for the table node_group_members.
var nodeGroupMembersColumns = NodeGroupMembersColumns{
	Id:        "id",
	GroupId:   "group_id",
	ClientIp:  "client_ip",
	NodeId:    "node_id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewNodeGroupMembersDao creates and returns a new DAO object for table data access.
func NewNodeGroupMembersDao(handlers ...gdb.ModelHandler) *NodeGroupMembersDao {
	return &NodeGroupMembersDao{
		group:    "default",
		table:    "node_group_members",
		columns:  nodeGroupMembersColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NodeGroupMembersDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NodeGroupMembersDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NodeGroupMembersDao) Columns() NodeGroupMembersColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NodeGroupMembersDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NodeGroupMembersDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NodeGroupMembersDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
