// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AutomaWorkflowsDao is the data access object for the table automa_workflows.
type AutomaWorkflowsDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  AutomaWorkflowsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// AutomaWorkflowsColumns defines and stores column names for the table automa_workflows.
type AutomaWorkflowsColumns struct {
	Id                string // 后端自增主键
	AutomaId          string // Automa 原始工作流 ID，用于识别同一个工作流
	Name              string // 服务端自定义工作流名称
	Description       string // 服务端自定义工作流描述
	Source            string // 来源：1 页面导入，2 客户端同步，3 服务端生成
	SourceIp          string // 最后同步来源 IP，仅用于展示和兼容旧逻辑
	SourceMachineId   string // 最后同步来源机器 ID
	SourceNodeId      string // 最后同步来源执行节点 ID
	SourceUserAgent   string // 最后同步来源浏览器 User-Agent
	AutomaVersion     string // 工作流保存时的 Automa 插件版本
	ExtVersion        string // Automa 导出文件中的 extVersion
	CreatedAtAutoma   string // Automa 原始 createdAt，通常为毫秒时间戳
	UpdatedAtAutoma   string // Automa 原始 updatedAt，通常为毫秒时间戳
	IsDisabled        string // 是否被 Automa 禁用
	IsProtected       string // 是否受保护，受保护工作流不允许客户端覆盖
	NodeCount         string // 工作流节点数量
	EdgeCount         string // 工作流连线数量
	RawJson           string // Automa 导出的原始完整 JSON
	NormalizedJson    string // 规范化后的工作流 JSON，用于内容比对和下发
	ContentHash       string // 核心内容 hash，用于判断工作流是否变化
	Revision          string // 服务端工作流版本号，内容变化时递增
	FirstSyncedAt     string // 第一次同步到服务端的时间
	LastSyncedAt      string // 最近一次同步到服务端的时间
	CreatedAt         string // 记录创建时间
	UpdatedAt         string // 记录更新时间
	DeletedAt         string // 软删除时间
	AutomaName        string // Automa 工作流 JSON 中的原始名称
	AutomaDescription string // Automa 工作流 JSON 中的原始描述
}

// automaWorkflowsColumns holds the columns for the table automa_workflows.
var automaWorkflowsColumns = AutomaWorkflowsColumns{
	Id:                "id",
	AutomaId:          "automa_id",
	Name:              "name",
	Description:       "description",
	Source:            "source",
	SourceIp:          "source_ip",
	SourceMachineId:   "source_machine_id",
	SourceNodeId:      "source_node_id",
	SourceUserAgent:   "source_user_agent",
	AutomaVersion:     "automa_version",
	ExtVersion:        "ext_version",
	CreatedAtAutoma:   "created_at_automa",
	UpdatedAtAutoma:   "updated_at_automa",
	IsDisabled:        "is_disabled",
	IsProtected:       "is_protected",
	NodeCount:         "node_count",
	EdgeCount:         "edge_count",
	RawJson:           "raw_json",
	NormalizedJson:    "normalized_json",
	ContentHash:       "content_hash",
	Revision:          "revision",
	FirstSyncedAt:     "first_synced_at",
	LastSyncedAt:      "last_synced_at",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
	AutomaName:        "automa_name",
	AutomaDescription: "automa_description",
}

// NewAutomaWorkflowsDao creates and returns a new DAO object for table data access.
func NewAutomaWorkflowsDao(handlers ...gdb.ModelHandler) *AutomaWorkflowsDao {
	return &AutomaWorkflowsDao{
		group:    "default",
		table:    "automa_workflows",
		columns:  automaWorkflowsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AutomaWorkflowsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AutomaWorkflowsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AutomaWorkflowsDao) Columns() AutomaWorkflowsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AutomaWorkflowsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AutomaWorkflowsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AutomaWorkflowsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
