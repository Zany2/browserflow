package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// StatusReq 获取在线执行端状态
type StatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"执行端" summary:"获取在线执行端状态"`
}

// StatusRes 在线执行端响应
type StatusRes struct {
	Agents []model.AgentStatus `json:"agents,omitempty" dc:"在线执行端列表"`
}
