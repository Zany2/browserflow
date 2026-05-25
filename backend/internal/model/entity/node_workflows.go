// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeWorkflows is the golang structure for table node_workflows.
type NodeWorkflows struct {
	Id               int64       `json:"id"                orm:"id"                ` // 自增 ID
	NodeId           string      `json:"node_id"           orm:"node_id"           ` // 执行节点 ID
	ClientIp         string      `json:"client_ip"         orm:"client_ip"         ` // 客户端 IP 快照，用于区分同名执行节点
	AutomaId         string      `json:"automa_id"         orm:"automa_id"         ` // Automa 原始工作流 ID
	WorkflowId       int64       `json:"workflow_id"       orm:"workflow_id"       ` // 服务端工作流主表 ID
	ExpectedRevision int         `json:"expected_revision" orm:"expected_revision" ` // 服务端期望节点安装的工作流版本
	ActualRevision   int         `json:"actual_revision"   orm:"actual_revision"   ` // 节点实际上报的工作流版本
	ContentHash      string      `json:"content_hash"      orm:"content_hash"      ` // 节点当前工作流内容 hash
	DesiredPresent   bool        `json:"desired_present"   orm:"desired_present"   ` // 是否期望该工作流存在于节点中
	SyncStatus       string      `json:"sync_status"       orm:"sync_status"       ` // 同步状态：pending、syncing、synced、failed、deleting、deleted
	LastSyncAt       *gtime.Time `json:"last_sync_at"      orm:"last_sync_at"      ` // 最近一次安装、更新或删除完成时间
	LastCheckAt      *gtime.Time `json:"last_check_at"     orm:"last_check_at"     ` // 最近一次节点清单上报时间
	LastError        string      `json:"last_error"        orm:"last_error"        ` // 最近一次同步错误信息
	CreatedAt        *gtime.Time `json:"created_at"        orm:"created_at"        ` // 记录创建时间
	UpdatedAt        *gtime.Time `json:"updated_at"        orm:"updated_at"        ` // 记录更新时间
	DeletedAt        *gtime.Time `json:"deleted_at"        orm:"deleted_at"        ` // 软删除时间
}
