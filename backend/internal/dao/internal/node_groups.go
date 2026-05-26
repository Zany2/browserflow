// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NodeGroupsDao is the data access object for the table node_groups.
type NodeGroupsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  NodeGroupsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// NodeGroupsColumns defines and stores column names for the table node_groups.
type NodeGroupsColumns struct {
	Id          string // 自增 ID
	GroupName   string // 节点分组名称，例如 财务组、客服组、闲置节点池
	Description string // 节点分组说明
	Enabled     string // 是否启用该分组
	CreatedAt   string // 记录创建时间
	UpdatedAt   string // 记录更新时间
	DeletedAt   string // 软删除时间
}

// nodeGroupsColumns holds the columns for the table node_groups.
var nodeGroupsColumns = NodeGroupsColumns{
	Id:          "id",
	GroupName:   "group_name",
	Description: "description",
	Enabled:     "enabled",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewNodeGroupsDao creates and returns a new DAO object for table data access.
func NewNodeGroupsDao(handlers ...gdb.ModelHandler) *NodeGroupsDao {
	return &NodeGroupsDao{
		group:    "default",
		table:    "node_groups",
		columns:  nodeGroupsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NodeGroupsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NodeGroupsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NodeGroupsDao) Columns() NodeGroupsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NodeGroupsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NodeGroupsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NodeGroupsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
