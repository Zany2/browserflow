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
	g.Meta         `orm:"table:tasks, do:true"`
	Id             interface{} // 自增ID
	Name           interface{} // 任务名称
	Description    interface{} // 任务说明
	AutomaId       interface{} // Automa 原始工作流 ID
	ClientIp       interface{} // 目标客户端 IP，为空则随机匹配在线且拥有此工作流的客户端执行
	CronExpression interface{} // Cron 表达式，为空表示立即执行
	ParamsJson     interface{} // 任务自定义参数，key/value 格式，使用 JSONB 存储
	Enabled        interface{} // 是否启用
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
	DeletedAt      *gtime.Time // 软删除时间
}
