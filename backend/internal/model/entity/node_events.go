// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeEvents is the golang structure for table node_events.
type NodeEvents struct {
	Id          int64       `json:"id"           orm:"id"           ` // 自增 ID
	MachineId   string      `json:"machine_id"   orm:"machine_id"   ` // 机器 ID
	NodeId      string      `json:"node_id"      orm:"node_id"      ` // 执行节点 ID
	ClientIp    string      `json:"client_ip"    orm:"client_ip"    ` // 客户端 IP 快照
	EventType   string      `json:"event_type"   orm:"event_type"   ` // 事件类型：register、heartbeat、disconnect、command、workflow_sync、package_install、error 等
	Level       string      `json:"level"        orm:"level"        ` // 事件级别：debug、info、warn、error
	Message     string      `json:"message"      orm:"message"      ` // 可读事件消息
	PayloadJson string      `json:"payload_json" orm:"payload_json" ` // 事件扩展数据 JSON
	CreatedAt   *gtime.Time `json:"created_at"   orm:"created_at"   ` // 记录创建时间
	UpdatedAt   *gtime.Time `json:"updated_at"   orm:"updated_at"   ` // 记录更新时间
	DeletedAt   *gtime.Time `json:"deleted_at"   orm:"deleted_at"   ` // 软删除时间
}
