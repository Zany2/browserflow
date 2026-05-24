// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskQueue is the golang structure of table task_queue for DAO operations like Where/Data.
type TaskQueue struct {
	g.Meta        `orm:"table:task_queue, do:true"`
	Id            interface{} // 自增 ID
	TaskId        interface{} // 关联任务配置 ID
	WorkflowId    interface{} // 待执行 Automa 工作流 ID
	DispatchMode  interface{} // 调度模式快照：auto、group、machine、node、ip
	ClientIp      interface{} // 目标客户端 IP 快照，仅用于兼容旧逻辑
	MachineId     interface{} // 目标机器 ID
	NodeId        interface{} // 目标节点 ID
	TargetGroupId interface{} // 目标节点分组 ID
	TriggerType   interface{} // 触发类型
	Priority      interface{} // 优先级，数字越大优先级越高
	Status        interface{} // 队列状态：pending、reserved、running、success、failed、cancelled、skipped
	AttemptCount  interface{} // 已尝试次数
	MaxAttempts   interface{} // 最大尝试次数
	ParamsJson    interface{} // 本次排队任务参数快照
	LastError     interface{} // 最近一次失败原因
	ScheduledAt   *gtime.Time // 计划执行时间
	ReservedAt    *gtime.Time // 被调度器占用时间
	StartedAt     *gtime.Time // 开始执行时间
	FinishedAt    *gtime.Time // 执行完成时间
	CreatedAt     *gtime.Time // 记录创建时间
	UpdatedAt     *gtime.Time // 记录更新时间
	DeletedAt     *gtime.Time // 软删除时间
}
