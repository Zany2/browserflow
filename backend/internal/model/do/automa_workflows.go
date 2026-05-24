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
	g.Meta            `orm:"table:automa_workflows, do:true"`
	Id                interface{} // 后端自增主键
	AutomaId          interface{} // Automa 原始工作流 ID，用于识别同一个工作流
	Name              interface{} // 服务端自定义工作流名称
	Description       interface{} // 服务端自定义工作流描述
	Source            interface{} // 来源：1 页面导入，2 客户端同步，3 服务端生成
	SourceIp          interface{} // 最后同步来源 IP，仅用于展示和兼容旧逻辑
	SourceMachineId   interface{} // 最后同步来源机器 ID
	SourceNodeId      interface{} // 最后同步来源执行节点 ID
	SourceUserAgent   interface{} // 最后同步来源浏览器 User-Agent
	AutomaVersion     interface{} // 工作流保存时的 Automa 插件版本
	ExtVersion        interface{} // Automa 导出文件中的 extVersion
	CreatedAtAutoma   interface{} // Automa 原始 createdAt，通常为毫秒时间戳
	UpdatedAtAutoma   interface{} // Automa 原始 updatedAt，通常为毫秒时间戳
	IsDisabled        interface{} // 是否被 Automa 禁用
	IsProtected       interface{} // 是否受保护，受保护工作流不允许客户端覆盖
	NodeCount         interface{} // 工作流节点数量
	EdgeCount         interface{} // 工作流连线数量
	RawJson           interface{} // Automa 导出的原始完整 JSON
	NormalizedJson    interface{} // 规范化后的工作流 JSON，用于内容比对和下发
	ContentHash       interface{} // 核心内容 hash，用于判断工作流是否变化
	Revision          interface{} // 服务端工作流版本号，内容变化时递增
	FirstSyncedAt     *gtime.Time // 第一次同步到服务端的时间
	LastSyncedAt      *gtime.Time // 最近一次同步到服务端的时间
	CreatedAt         *gtime.Time // 记录创建时间
	UpdatedAt         *gtime.Time // 记录更新时间
	DeletedAt         *gtime.Time // 软删除时间
	AutomaName        interface{} // Automa 工作流 JSON 中的原始名称
	AutomaDescription interface{} // Automa 工作流 JSON 中的原始描述
}
