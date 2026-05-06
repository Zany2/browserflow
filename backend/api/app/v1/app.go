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
}
