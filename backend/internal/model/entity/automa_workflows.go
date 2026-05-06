// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AutomaWorkflows is the golang structure for table automa_workflows.
type AutomaWorkflows struct {
	Id              int64       `json:"id"                orm:"id"                ` // 后端自增主键
	AutomaId        string      `json:"automa_id"         orm:"automa_id"         ` // Automa 原始工作流 ID，判断同一个工作流的唯一依据
	Name            string      `json:"name"              orm:"name"              ` // 工作流名称，默认从 Automa 导出的 JSON 中解析
	Description     string      `json:"description"       orm:"description"       ` // 工作流描述，默认从 Automa 导出的 JSON 中解析
	Source          int         `json:"source"            orm:"source"            ` // 来源：1页面导入 2客户端同步 默认1
	SourceIp        string      `json:"source_ip"         orm:"source_ip"         ` // 最后同步来源 IP
	SourceUserAgent string      `json:"source_user_agent" orm:"source_user_agent" ` // 最后同步来源浏览器 User-Agent
	AutomaVersion   string      `json:"automa_version"    orm:"automa_version"    ` // 工作流保存时的 Automa 插件版本
	ExtVersion      string      `json:"ext_version"       orm:"ext_version"       ` // Automa 导出文件中的 extVersion
	CreatedAtAutoma int64       `json:"created_at_automa" orm:"created_at_automa" ` // Automa 原始 createdAt，通常为毫秒时间戳
	UpdatedAtAutoma int64       `json:"updated_at_automa" orm:"updated_at_automa" ` // Automa 原始 updatedAt，通常为毫秒时间戳
	IsDisabled      bool        `json:"is_disabled"       orm:"is_disabled"       ` // 是否被 Automa 禁用 默认false
	IsProtected     bool        `json:"is_protected"      orm:"is_protected"      ` // 是否受保护 true客户端不能同步 false客户端可同步 默认false
	NodeCount       int         `json:"node_count"        orm:"node_count"        ` // 工作流节点数量
	EdgeCount       int         `json:"edge_count"        orm:"edge_count"        ` // 工作流连线数量
	RawJson         string      `json:"raw_json"          orm:"raw_json"          ` // Automa 导出的原始完整 JSON 数据
	NormalizedJson  string      `json:"normalized_json"   orm:"normalized_json"   ` // 规范化后的工作流 JSON，用于计算 content_hash 或内容对比
	ContentHash     string      `json:"content_hash"      orm:"content_hash"      ` // 核心内容 hash，用于判断工作流内容是否变化
	Revision        int         `json:"revision"          orm:"revision"          ` // 当前版本号，每次内容变化时加 1
	FirstSyncedAt   *gtime.Time `json:"first_synced_at"   orm:"first_synced_at"   ` // 第一次同步到后端的时间
	LastSyncedAt    *gtime.Time `json:"last_synced_at"    orm:"last_synced_at"    ` // 最近一次同步到后端的时间
	CreatedAt       *gtime.Time `json:"created_at"        orm:"created_at"        ` // 后端记录创建时间
	UpdatedAt       *gtime.Time `json:"updated_at"        orm:"updated_at"        ` // 后端记录更新时间
	DeletedAt       *gtime.Time `json:"deleted_at"        orm:"deleted_at"        ` // 软删除时间
}
