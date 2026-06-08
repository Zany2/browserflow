package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeReq 获取当前应用运行模式
type RuntimeReq struct {
	g.Meta `path:"/runtime" method:"get" tags:"应用" summary:"获取应用运行模式"`
}

// RuntimeRes 应用运行模式响应
type RuntimeRes struct {
	Mode           string   `json:"mode"`
	DisabledRoutes []string `json:"disabled_routes"`
	// FrontendURL configured frontend access base URL 前端访问基础地址
	FrontendURL string `json:"frontend_url"`
}

// ServerDashboardReq 获取服务端首页统计
type ServerDashboardReq struct {
	g.Meta `path:"/server-dashboard" method:"get" tags:"应用" summary:"获取服务端首页统计"`
}

// ServerDashboardStat 服务端首页统计项
type ServerDashboardStat struct {
	Value int `json:"value" dc:"主统计数字"`
	Total int `json:"total" dc:"总数"`
	Extra int `json:"extra" dc:"辅助统计数字"`
}

// ServerDashboardRes 服务端首页统计响应
type ServerDashboardRes struct {
	Clients   ServerDashboardStat `json:"clients" dc:"客户端统计，value=在线客户端，total=总客户端，extra=离线客户端"`
	Nodes     ServerDashboardStat `json:"nodes" dc:"执行节点统计，value=在线节点，total=总节点，extra=离线节点"`
	NodeBusy  ServerDashboardStat `json:"node_busy" dc:"节点忙闲统计，value=空闲节点，total=忙碌节点，extra=未知节点"`
	Automa    ServerDashboardStat `json:"automa" dc:"Automa 统计，value=可用节点，total=在线节点，extra=异常节点"`
	Workflows ServerDashboardStat `json:"workflows" dc:"工作流统计，value=可同步工作流，total=总工作流，extra=受保护工作流"`
	Tasks     ServerDashboardStat `json:"tasks" dc:"任务统计，value=启用任务，total=总任务，extra=禁用任务"`
}
