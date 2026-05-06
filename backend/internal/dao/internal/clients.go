// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ClientsDao is the data access object for the table clients.
type ClientsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ClientsColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ClientsColumns defines and stores column names for the table clients.
type ClientsColumns struct {
	Id             string // 主键 ID，自增
	ClientId       string // 客户端唯一标识（建议前端生成 UUID 并存储在 localStorage，用于唯一识别一个浏览器实例）
	ClientName     string // 客户端名称（可自定义，例如：办公室电脑、家用电脑）
	ClientIp       string // 客户端当前或最后一次连接 IP
	UserAgent      string // 客户端浏览器 User-Agent 字符串
	Status         string // 客户端当前状态：online 在线，offline 离线
	PluginStatus   string // Automa 插件状态：unknown 未知，installed 已安装，not_installed 未安装，disabled 已禁用，error 异常
	AutomaVersion  string // Automa 插件版本号
	BrowserName    string // 浏览器名称，例如 Chrome、Edge
	BrowserVersion string // 浏览器版本号
	OsName         string // 操作系统名称，例如 Windows、macOS、Linux
	OsVersion      string // 操作系统版本号
	Hostname       string // 客户端设备主机名
	IsBanned       string // 是否被拉黑（true 表示该客户端被禁止使用）
	BanReason      string // 客户端被拉黑的原因
	FirstSeenAt    string // 客户端第一次连接到服务端的时间（注册时间）
	LastSeenAt     string // 客户端最近一次心跳或交互时间（用于判断是否在线）
	ConnectedAt    string // 最近一次建立 WebSocket 连接的时间
	DisconnectedAt string // 最近一次断开 WebSocket 连接的时间
	CreatedAt      string // 记录创建时间
	UpdatedAt      string // 记录更新时间
	DeletedAt      string // 软删除时间
}

// clientsColumns holds the columns for the table clients.
var clientsColumns = ClientsColumns{
	Id:             "id",
	ClientId:       "client_id",
	ClientName:     "client_name",
	ClientIp:       "client_ip",
	UserAgent:      "user_agent",
	Status:         "status",
	PluginStatus:   "plugin_status",
	AutomaVersion:  "automa_version",
	BrowserName:    "browser_name",
	BrowserVersion: "browser_version",
	OsName:         "os_name",
	OsVersion:      "os_version",
	Hostname:       "hostname",
	IsBanned:       "is_banned",
	BanReason:      "ban_reason",
	FirstSeenAt:    "first_seen_at",
	LastSeenAt:     "last_seen_at",
	ConnectedAt:    "connected_at",
	DisconnectedAt: "disconnected_at",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewClientsDao creates and returns a new DAO object for table data access.
func NewClientsDao(handlers ...gdb.ModelHandler) *ClientsDao {
	return &ClientsDao{
		group:    "default",
		table:    "clients",
		columns:  clientsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ClientsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ClientsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ClientsDao) Columns() ClientsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ClientsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ClientsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ClientsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
