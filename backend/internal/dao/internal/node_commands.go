// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NodeCommandsDao is the data access object for the table node_commands.
type NodeCommandsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  NodeCommandsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// NodeCommandsColumns defines and stores column names for the table node_commands.
type NodeCommandsColumns struct {
	Id             string // 自增 ID
	CommandId      string // 服务端命令唯一 ID，用于客户端回调和幂等处理
	NodeId         string // 目标执行节点 ID
	MachineId      string // 目标机器 ID
	CommandType    string // 命令类型：install_workflow、delete_workflow、sync_inventory、install_package、restart_node、execute_workflow 等
	Status         string // 命令状态：pending、sent、running、success、failed、cancelled、timeout
	PayloadJson    string // 命令参数 JSON
	ResultJson     string // 命令执行结果 JSON
	ErrorMessage   string // 命令失败信息
	RetryCount     string // 命令重试次数
	TimeoutSeconds string // 命令超时时间，单位秒
	SentAt         string // 命令发送时间
	AckAt          string // 客户端确认收到时间
	FinishedAt     string // 命令完成时间
	CreatedAt      string // 记录创建时间
	UpdatedAt      string // 记录更新时间
	DeletedAt      string // 软删除时间
}

// nodeCommandsColumns holds the columns for the table node_commands.
var nodeCommandsColumns = NodeCommandsColumns{
	Id:             "id",
	CommandId:      "command_id",
	NodeId:         "node_id",
	MachineId:      "machine_id",
	CommandType:    "command_type",
	Status:         "status",
	PayloadJson:    "payload_json",
	ResultJson:     "result_json",
	ErrorMessage:   "error_message",
	RetryCount:     "retry_count",
	TimeoutSeconds: "timeout_seconds",
	SentAt:         "sent_at",
	AckAt:          "ack_at",
	FinishedAt:     "finished_at",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewNodeCommandsDao creates and returns a new DAO object for table data access.
func NewNodeCommandsDao(handlers ...gdb.ModelHandler) *NodeCommandsDao {
	return &NodeCommandsDao{
		group:    "default",
		table:    "node_commands",
		columns:  nodeCommandsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NodeCommandsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NodeCommandsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NodeCommandsDao) Columns() NodeCommandsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NodeCommandsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NodeCommandsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NodeCommandsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
