// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeGroups is the golang structure for table node_groups.
type NodeGroups struct {
	Id          int64       `json:"id"          orm:"id"          ` // 自增 ID
	GroupName   string      `json:"group_name"  orm:"group_name"  ` // 节点分组名称，例如 财务组、客服组、闲置节点池
	Description string      `json:"description" orm:"description" ` // 节点分组说明
	Enabled     bool        `json:"enabled"     orm:"enabled"     ` // 是否启用该分组
	CreatedAt   *gtime.Time `json:"created_at"  orm:"created_at"  ` // 记录创建时间
	UpdatedAt   *gtime.Time `json:"updated_at"  orm:"updated_at"  ` // 记录更新时间
	DeletedAt   *gtime.Time `json:"deleted_at"  orm:"deleted_at"  ` // 软删除时间
}
