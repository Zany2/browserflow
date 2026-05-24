// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AutomaPackagesDao is the data access object for the table automa_packages.
type AutomaPackagesDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  AutomaPackagesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// AutomaPackagesColumns defines and stores column names for the table automa_packages.
type AutomaPackagesColumns struct {
	Id           string // 自增 ID
	PackageName  string // 安装包名称，例如 Automa Chrome Extension
	Version      string // 安装包版本号
	FileName     string // 原始文件名
	FilePath     string // 服务端存储路径
	FileSize     string // 文件大小，单位字节
	Sha256       string // 安装包 sha256 校验值
	ManifestJson string // 插件 manifest 内容快照
	Status       string // 安装包状态：active 可用，disabled 禁用
	IsDefault    string // 是否为默认下发版本
	CreatedAt    string // 记录创建时间
	UpdatedAt    string // 记录更新时间
	DeletedAt    string // 软删除时间
}

// automaPackagesColumns holds the columns for the table automa_packages.
var automaPackagesColumns = AutomaPackagesColumns{
	Id:           "id",
	PackageName:  "package_name",
	Version:      "version",
	FileName:     "file_name",
	FilePath:     "file_path",
	FileSize:     "file_size",
	Sha256:       "sha256",
	ManifestJson: "manifest_json",
	Status:       "status",
	IsDefault:    "is_default",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewAutomaPackagesDao creates and returns a new DAO object for table data access.
func NewAutomaPackagesDao(handlers ...gdb.ModelHandler) *AutomaPackagesDao {
	return &AutomaPackagesDao{
		group:    "default",
		table:    "automa_packages",
		columns:  automaPackagesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AutomaPackagesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AutomaPackagesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AutomaPackagesDao) Columns() AutomaPackagesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AutomaPackagesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AutomaPackagesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AutomaPackagesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
