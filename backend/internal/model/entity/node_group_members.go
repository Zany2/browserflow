// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeGroupMembers is the golang structure for table node_group_members.
type NodeGroupMembers struct {
	Id        int64       `json:"id"         orm:"id"         ` // 自增 ID
	GroupId   int64       `json:"group_id"   orm:"group_id"   ` // 节点分组 ID
	ClientIp  string      `json:"client_ip"  orm:"client_ip"  ` // 客户端 IP 快照
	NodeId    string      `json:"node_id"    orm:"node_id"    ` // 执行节点 ID
	CreatedAt *gtime.Time `json:"created_at" orm:"created_at" ` // 记录创建时间
	UpdatedAt *gtime.Time `json:"updated_at" orm:"updated_at" ` // 记录更新时间
	DeletedAt *gtime.Time `json:"deleted_at" orm:"deleted_at" ` // 软删除时间
}
