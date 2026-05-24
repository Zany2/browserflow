// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskRecords is the golang structure of table task_records for DAO operations like Where/Data.
type TaskRecords struct {
	g.Meta            `orm:"table:task_records, do:true"`
	Id                interface{} // 自增 ID
	TaskId            interface{} // 关联的任务配置 ID
	WorkflowId        interface{} // 执行时使用的 Automa 工作流 ID
	ClientIp          interface{} // 执行目标客户端 IP，仅用于展示和兼容旧逻辑
	MachineId         interface{} // 执行目标机器 ID
	NodeId            interface{} // 执行目标节点 ID
	NodeName          interface{} // 执行目标节点名称快照
	ExecutionId       interface{} // 后端本次执行标识，用于客户端回调和状态恢复，例如 task-record-{id}
	AutomaExecutionId interface{} // Automa 客户端侧本次执行实例 ID，例如 stateId 或 historyId
	CommandId         interface{} // 对应下发给节点的命令 ID
	QueueId           interface{} // 对应的任务队列记录 ID
	AttemptNo         interface{} // 第几次执行尝试
	TriggerType       interface{} // 触发类型：manual 手动触发，cron 定时触发，task_create 创建任务触发，skill Skill 触发，system 系统触发
	Status            interface{} // 执行状态：pending、queued、running、success、failed、cancelled、timeout
	ParamsJson        interface{} // 本次执行使用的参数快照
	ResultJson        interface{} // 本次执行结果内容
	ErrorMessage      interface{} // 执行失败时的错误信息
	StartedAt         *gtime.Time // 开始执行时间
	FinishedAt        *gtime.Time // 执行结束时间
	CreatedAt         *gtime.Time // 记录创建时间
	UpdatedAt         *gtime.Time // 记录更新时间
	DeletedAt         *gtime.Time // 软删除时间
}
