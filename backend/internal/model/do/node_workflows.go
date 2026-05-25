// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeWorkflows is the golang structure of table node_workflows for DAO operations like Where/Data.
type NodeWorkflows struct {
	g.Meta           `orm:"table:node_workflows, do:true"`
	Id               interface{} // 自增 ID
	NodeId           interface{} // 执行节点 ID
	ClientIp         interface{} // 客户端 IP 快照，用于区分同名执行节点
	AutomaId         interface{} // Automa 原始工作流 ID
	WorkflowId       interface{} // 服务端工作流主表 ID
	ExpectedRevision interface{} // 服务端期望节点安装的工作流版本
	ActualRevision   interface{} // 节点实际上报的工作流版本
	ContentHash      interface{} // 节点当前工作流内容 hash
	DesiredPresent   interface{} // 是否期望该工作流存在于节点中
	SyncStatus       interface{} // 同步状态：pending、syncing、synced、failed、deleting、deleted
	LastSyncAt       *gtime.Time // 最近一次安装、更新或删除完成时间
	LastCheckAt      *gtime.Time // 最近一次节点清单上报时间
	LastError        interface{} // 最近一次同步错误信息
	CreatedAt        *gtime.Time // 记录创建时间
	UpdatedAt        *gtime.Time // 记录更新时间
	DeletedAt        *gtime.Time // 软删除时间
}
