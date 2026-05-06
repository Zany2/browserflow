// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Tasks is the golang structure for table tasks.
type Tasks struct {
	Id             int64       `json:"id"              orm:"id"              ` // 自增ID
	Name           string      `json:"name"            orm:"name"            ` // 任务名称
	Description    string      `json:"description"     orm:"description"     ` // 任务说明
	AutomaId       string      `json:"automa_id"       orm:"automa_id"       ` // Automa 原始工作流 ID
	ClientIp       string      `json:"client_ip"       orm:"client_ip"       ` // 目标客户端 IP，为空则随机匹配在线且拥有此工作流的客户端执行
	CronExpression string      `json:"cron_expression" orm:"cron_expression" ` // Cron 表达式，为空表示立即执行
	ParamsJson     string      `json:"params_json"     orm:"params_json"     ` // 任务自定义参数，key/value 格式，使用 JSONB 存储
	Enabled        bool        `json:"enabled"         orm:"enabled"         ` // 是否启用
	CreatedAt      *gtime.Time `json:"created_at"      orm:"created_at"      ` // 创建时间
	UpdatedAt      *gtime.Time `json:"updated_at"      orm:"updated_at"      ` // 更新时间
	DeletedAt      *gtime.Time `json:"deleted_at"      orm:"deleted_at"      ` // 软删除时间
}
