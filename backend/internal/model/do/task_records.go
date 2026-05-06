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
	g.Meta       `orm:"table:task_records, do:true"`
	Id           interface{} // 自增ID
	TaskId       interface{} // 关联的任务配置ID
	WorkflowId   interface{} // 执行时使用的 Automa 工作流 ID
	ClientIp     interface{} // 执行目标客户端 IP
	TriggerType  interface{} // 触发类型：manual 手动触发，cron 定时触发，task_create 创建任务触发，skill Skill触发，system 系统触发
	Status       interface{} // 执行状态：pending、queued、running、success、failed、cancelled
	ParamsJson   interface{} // 本次执行使用的参数快照
	ResultJson   interface{} // 本次执行结果内容
	ErrorMessage interface{} // 执行失败时的错误信息
	StartedAt    *gtime.Time // 开始执行时间
	FinishedAt   *gtime.Time // 执行结束时间
	CreatedAt    *gtime.Time // 记录创建时间
	UpdatedAt    *gtime.Time // 记录更新时间
	DeletedAt    *gtime.Time // 软删除时间
}
