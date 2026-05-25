// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Tasks is the golang structure of table tasks for DAO operations like Where/Data.
type Tasks struct {
	g.Meta                `orm:"table:tasks, do:true"`
	Id                    interface{} // 自增 ID
	Name                  interface{} // 任务名称
	Description           interface{} // 任务说明
	AutomaId              interface{} // Automa 原始工作流 ID
	ClientIp              interface{} // 目标客户端 IP，为空表示由服务端自动选择
	NodeId                interface{} // 目标执行节点 ID，为空表示由服务端选择可用节点
	TargetGroupId         interface{} // 目标节点分组 ID，为 0 表示不限定分组
	DispatchMode          interface{} // 调度模式：auto 自动选择，group 指定分组，node 指定节点，ip 指定 IP
	QueuePolicy           interface{} // 繁忙策略：queue 排队，fail 直接失败，skip 跳过本次
	ConflictWindowSeconds interface{} // 创建或编辑定时任务时用于冲突提醒的时间窗口秒数
	MaxAttempts           interface{} // 任务最多尝试次数
	TimeoutSeconds        interface{} // 任务执行超时时间，单位秒
	CronExpression        interface{} // Cron 表达式，为空表示立即执行
	ParamsJson            interface{} // 任务自定义参数快照，JSONB 存储
	Enabled               interface{} // 是否启用
	CreatedAt             *gtime.Time // 创建时间
	UpdatedAt             *gtime.Time // 更新时间
	DeletedAt             *gtime.Time // 软删除时间
}
