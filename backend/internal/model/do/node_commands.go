// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NodeCommands is the golang structure of table node_commands for DAO operations like Where/Data.
type NodeCommands struct {
	g.Meta         `orm:"table:node_commands, do:true"`
	Id             interface{} // 自增 ID
	CommandId      interface{} // 服务端命令唯一 ID，用于客户端回调和幂等处理
	ClientIp       interface{} // 目标客户端 IP
	NodeId         interface{} // 目标执行节点 ID
	CommandType    interface{} // 命令类型：install_workflow、delete_workflow、sync_inventory、install_package、restart_node、execute_workflow 等
	Status         interface{} // 命令状态：pending、sent、running、success、failed、cancelled、timeout
	PayloadJson    interface{} // 命令参数 JSON
	ResultJson     interface{} // 命令执行结果 JSON
	ErrorMessage   interface{} // 命令失败信息
	RetryCount     interface{} // 命令重试次数
	TimeoutSeconds interface{} // 命令超时时间，单位秒
	SentAt         *gtime.Time // 命令发送时间
	AckAt          *gtime.Time // 客户端确认收到时间
	FinishedAt     *gtime.Time // 命令完成时间
	CreatedAt      *gtime.Time // 记录创建时间
	UpdatedAt      *gtime.Time // 记录更新时间
	DeletedAt      *gtime.Time // 软删除时间
}
