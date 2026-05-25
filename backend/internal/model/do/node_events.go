// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeEvents is the golang structure of table node_events for DAO operations like Where/Data.
type NodeEvents struct {
	g.Meta      `orm:"table:node_events, do:true"`
	Id          interface{} // 自增 ID
	NodeId      interface{} // 执行节点 ID
	ClientIp    interface{} // 客户端 IP 快照
	EventType   interface{} // 事件类型：register、heartbeat、disconnect、command、workflow_sync、package_install、error 等
	Level       interface{} // 事件级别：debug、info、warn、error
	Message     interface{} // 可读事件消息
	PayloadJson interface{} // 事件扩展数据 JSON
	CreatedAt   *gtime.Time // 记录创建时间
	UpdatedAt   *gtime.Time // 记录更新时间
	DeletedAt   *gtime.Time // 软删除时间
}
