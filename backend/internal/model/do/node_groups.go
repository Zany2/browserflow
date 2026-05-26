// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeGroups is the golang structure of table node_groups for DAO operations like Where/Data.
type NodeGroups struct {
	g.Meta      `orm:"table:node_groups, do:true"`
	Id          interface{} // 自增 ID
	GroupName   interface{} // 节点分组名称，例如 财务组、客服组、闲置节点池
	Description interface{} // 节点分组说明
	Enabled     interface{} // 是否启用该分组
	CreatedAt   *gtime.Time // 记录创建时间
	UpdatedAt   *gtime.Time // 记录更新时间
	DeletedAt   *gtime.Time // 软删除时间
}
