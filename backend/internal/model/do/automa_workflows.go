// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AutomaWorkflows is the golang structure of table automa_workflows for DAO operations like Where/Data.
type AutomaWorkflows struct {
	g.Meta          `orm:"table:automa_workflows, do:true"`
	Id              interface{} // 后端自增主键
	AutomaId        interface{} // Automa 原始工作流 ID，判断同一个工作流的唯一依据
	Name            interface{} // 工作流名称，默认从 Automa 导出的 JSON 中解析
	Description     interface{} // 工作流描述，默认从 Automa 导出的 JSON 中解析
	Source          interface{} // 来源：1页面导入 2客户端同步 默认1
	SourceIp        interface{} // 最后同步来源 IP
	SourceUserAgent interface{} // 最后同步来源浏览器 User-Agent
	AutomaVersion   interface{} // 工作流保存时的 Automa 插件版本
	ExtVersion      interface{} // Automa 导出文件中的 extVersion
	CreatedAtAutoma interface{} // Automa 原始 createdAt，通常为毫秒时间戳
	UpdatedAtAutoma interface{} // Automa 原始 updatedAt，通常为毫秒时间戳
	IsDisabled      interface{} // 是否被 Automa 禁用 默认false
	IsProtected     interface{} // 是否受保护 true客户端不能同步 false客户端可同步 默认false
	NodeCount       interface{} // 工作流节点数量
	EdgeCount       interface{} // 工作流连线数量
	RawJson         interface{} // Automa 导出的原始完整 JSON 数据
	NormalizedJson  interface{} // 规范化后的工作流 JSON，用于计算 content_hash 或内容对比
	ContentHash     interface{} // 核心内容 hash，用于判断工作流内容是否变化
	Revision        interface{} // 当前版本号，每次内容变化时加 1
	FirstSyncedAt   *gtime.Time // 第一次同步到后端的时间
	LastSyncedAt    *gtime.Time // 最近一次同步到后端的时间
	CreatedAt       *gtime.Time // 后端记录创建时间
	UpdatedAt       *gtime.Time // 后端记录更新时间
	DeletedAt       *gtime.Time // 软删除时间
}
