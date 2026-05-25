// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskAttempts is the golang structure of table task_attempts for DAO operations like Where/Data.
type TaskAttempts struct {
	g.Meta       `orm:"table:task_attempts, do:true"`
	Id           interface{} // 自增 ID
	QueueId      interface{} // 对应任务队列 ID
	TaskId       interface{} // 关联任务配置 ID
	RecordId     interface{} // 关联任务执行记录 ID
	AttemptNo    interface{} // 尝试序号，从 1 开始
	WorkflowId   interface{} // 执行工作流 ID
	ClientIp     interface{} // 执行客户端 IP 快照
	NodeId       interface{} // 执行节点 ID
	ExecutionId  interface{} // 后端本次执行标识
	CommandId    interface{} // 对应节点命令 ID
	Status       interface{} // 尝试状态：pending、running、success、failed、timeout、cancelled
	ErrorMessage interface{} // 本次尝试失败原因
	StartedAt    *gtime.Time // 尝试开始时间
	FinishedAt   *gtime.Time // 尝试结束时间
	CreatedAt    *gtime.Time // 记录创建时间
	UpdatedAt    *gtime.Time // 记录更新时间
	DeletedAt    *gtime.Time // 软删除时间
}
