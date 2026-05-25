// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NodePackagesDao is the data access object for the table node_packages.
type NodePackagesDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  NodePackagesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// NodePackagesColumns defines and stores column names for the table node_packages.
type NodePackagesColumns struct {
	Id              string // 自增 ID
	NodeId          string // 执行节点 ID
	ClientIp        string // 客户端 IP 快照，用于区分同名执行节点
	PackageId       string // 关联的 Automa 插件安装包 ID
	ExpectedVersion string // 服务端期望安装的插件版本
	ActualVersion   string // 节点实际上报的插件版本
	ExpectedSha256  string // 服务端期望安装包 sha256
	ActualSha256    string // 节点实际上报的安装内容 sha256
	DesiredStatus   string // 期望状态：installed 安装，removed 移除
	InstallStatus   string // 安装状态：unknown、pending、installing、installed、failed、removed
	LastInstallAt   string // 最近一次安装或更新完成时间
	LastCheckAt     string // 最近一次节点上报检查时间
	LastError       string // 最近一次安装或检查错误
	CreatedAt       string // 记录创建时间
	UpdatedAt       string // 记录更新时间
	DeletedAt       string // 软删除时间
}

// nodePackagesColumns holds the columns for the table node_packages.
var nodePackagesColumns = NodePackagesColumns{
	Id:              "id",
	NodeId:          "node_id",
	ClientIp:        "client_ip",
	PackageId:       "package_id",
	ExpectedVersion: "expected_version",
	ActualVersion:   "actual_version",
	ExpectedSha256:  "expected_sha256",
	ActualSha256:    "actual_sha256",
	DesiredStatus:   "desired_status",
	InstallStatus:   "install_status",
	LastInstallAt:   "last_install_at",
	LastCheckAt:     "last_check_at",
	LastError:       "last_error",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewNodePackagesDao creates and returns a new DAO object for table data access.
func NewNodePackagesDao(handlers ...gdb.ModelHandler) *NodePackagesDao {
	return &NodePackagesDao{
		group:    "default",
		table:    "node_packages",
		columns:  nodePackagesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NodePackagesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NodePackagesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NodePackagesDao) Columns() NodePackagesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NodePackagesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NodePackagesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NodePackagesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
