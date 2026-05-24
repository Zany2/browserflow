// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeGroupMembers is the golang structure of table node_group_members for DAO operations like Where/Data.
type NodeGroupMembers struct {
	g.Meta    `orm:"table:node_group_members, do:true"`
	Id        interface{} // 自增 ID
	GroupId   interface{} // 节点分组 ID
	MachineId interface{} // 机器 ID 快照
	NodeId    interface{} // 执行节点 ID
	CreatedAt *gtime.Time // 记录创建时间
	UpdatedAt *gtime.Time // 记录更新时间
	DeletedAt *gtime.Time // 软删除时间
}
