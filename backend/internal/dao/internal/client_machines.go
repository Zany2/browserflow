// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ClientMachinesDao is the data access object for the table client_machines.
type ClientMachinesDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  ClientMachinesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// ClientMachinesColumns defines and stores column names for the table client_machines.
type ClientMachinesColumns struct {
	Id              string // 主键 ID，自增
	MachineId       string // 物理机器稳定 ID，同一台 Windows 电脑保持一致
	MachineName     string // Worker 上报的机器名称
	DisplayName     string // 服务端自定义机器显示名称
	ClientIp        string // 机器当前或最后一次连接 IP
	Hostname        string // 系统主机名
	OsName          string // 操作系统名称
	OsVersion       string // 操作系统版本
	WorkerVersion   string // BrowserFlow Worker 版本
	InstallDir      string // Worker 安装目录
	DataDir         string // Worker 数据目录
	Status          string // 机器状态：online 在线，offline 离线
	IsBanned        string // 是否禁止该机器连接
	BanReason       string // 机器被禁止连接的原因
	NodeCount       string // 该机器配置的执行节点数量
	OnlineNodeCount string // 该机器当前在线执行节点数量
	ConfigJson      string // 机器级配置快照
	FirstSeenAt     string // 第一次连接到服务端的时间
	LastSeenAt      string // 最近一次心跳或交互时间
	CreatedAt       string // 记录创建时间
	UpdatedAt       string // 记录更新时间
	DeletedAt       string // 软删除时间
}

// clientMachinesColumns holds the columns for the table client_machines.
var clientMachinesColumns = ClientMachinesColumns{
	Id:              "id",
	MachineId:       "machine_id",
	MachineName:     "machine_name",
	DisplayName:     "display_name",
	ClientIp:        "client_ip",
	Hostname:        "hostname",
	OsName:          "os_name",
	OsVersion:       "os_version",
	WorkerVersion:   "worker_version",
	InstallDir:      "install_dir",
	DataDir:         "data_dir",
	Status:          "status",
	IsBanned:        "is_banned",
	BanReason:       "ban_reason",
	NodeCount:       "node_count",
	OnlineNodeCount: "online_node_count",
	ConfigJson:      "config_json",
	FirstSeenAt:     "first_seen_at",
	LastSeenAt:      "last_seen_at",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewClientMachinesDao creates and returns a new DAO object for table data access.
func NewClientMachinesDao(handlers ...gdb.ModelHandler) *ClientMachinesDao {
	return &ClientMachinesDao{
		group:    "default",
		table:    "client_machines",
		columns:  clientMachinesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ClientMachinesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ClientMachinesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ClientMachinesDao) Columns() ClientMachinesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ClientMachinesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ClientMachinesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ClientMachinesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
