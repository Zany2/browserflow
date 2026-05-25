// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Tasks is the golang structure for table tasks.
type Tasks struct {
	Id                        int64       `json:"id"                           orm:"id"                           ` // 自增 ID
	Name                      string      `json:"name"                         orm:"name"                         ` // 任务名称
	Description               string      `json:"description"                  orm:"description"                  ` // 任务说明
	AutomaId                  string      `json:"automa_id"                    orm:"automa_id"                    ` // Automa 原始工作流 ID
	ClientIp                  string      `json:"client_ip"                    orm:"client_ip"                    ` // 目标客户端 IP，为空表示由服务端自动选择
	NodeId                    string      `json:"node_id"                      orm:"node_id"                      ` // 目标执行节点 ID，为空表示由服务端选择可用节点
	TargetGroupId             int64       `json:"target_group_id"              orm:"target_group_id"              ` // 目标节点分组 ID，为 0 表示不限定分组
	DispatchMode              string      `json:"dispatch_mode"                orm:"dispatch_mode"                ` // 调度模式：auto 自动选择，group 指定分组，node 指定节点，ip 指定 IP
	QueuePolicy               string      `json:"queue_policy"                 orm:"queue_policy"                 ` // 繁忙策略：queue 排队，fail 直接失败，skip 跳过本次
	ConflictWindowSeconds     int         `json:"conflict_window_seconds"      orm:"conflict_window_seconds"      ` // 创建或编辑定时任务时用于冲突提醒的时间窗口秒数
	MaxAttempts               int         `json:"max_attempts"                 orm:"max_attempts"                 ` // 任务最多尝试次数
	TimeoutSeconds            int         `json:"timeout_seconds"              orm:"timeout_seconds"              ` // 任务执行超时时间，单位秒
	QueueWaitSeconds          int         `json:"queue_wait_seconds"           orm:"queue_wait_seconds"           ` // 等待可用最大等待时间，单位秒，仅 queue_policy=queue 时生效
	QueueRetryIntervalSeconds int         `json:"queue_retry_interval_seconds" orm:"queue_retry_interval_seconds" ` // 等待可用重试间隔，单位秒，仅 queue_policy=queue 时生效
	CronExpression            string      `json:"cron_expression"              orm:"cron_expression"              ` // Cron 表达式，为空表示立即执行
	ParamsJson                string      `json:"params_json"                  orm:"params_json"                  ` // 任务自定义参数快照，JSONB 存储
	Enabled                   bool        `json:"enabled"                      orm:"enabled"                      ` // 是否启用
	CreatedAt                 *gtime.Time `json:"created_at"                   orm:"created_at"                   ` // 创建时间
	UpdatedAt                 *gtime.Time `json:"updated_at"                   orm:"updated_at"                   ` // 更新时间
	DeletedAt                 *gtime.Time `json:"deleted_at"                   orm:"deleted_at"                   ` // 软删除时间
}
