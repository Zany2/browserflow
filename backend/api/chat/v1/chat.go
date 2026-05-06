package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// ChatSessionListResModel 对话会话列表项
type ChatSessionListResModel = model.ChatSession

// ChatSessionListReq 获取对话会话列表
type ChatSessionListReq struct {
	g.Meta `path:"/sessions" method:"get" tags:"对话" summary:"获取对话会话列表"`
}

// ChatSessionListRes 对话会话列表响应
type ChatSessionListRes struct {
	Sessions []*ChatSessionListResModel `json:"sessions,omitempty" dc:"对话会话列表"`
}

// ChatSessionCreateReq 创建对话会话
type ChatSessionCreateReq struct {
	g.Meta      `path:"/sessions" method:"post" tags:"对话" summary:"创建对话会话"`
	LLMConfigID string `json:"llm_config_id" dc:"大模型配置ID"`
}

// ChatSessionCreateRes 对话会话创建响应
type ChatSessionCreateRes struct {
	Session *ChatSessionListResModel `json:"session,omitempty" dc:"对话会话"`
}

// ChatSessionDeleteManyReq 批量删除对话会话
type ChatSessionDeleteManyReq struct {
	g.Meta `path:"/sessions" method:"delete" tags:"对话" summary:"批量删除对话会话"`
	IDs    []string `json:"ids" v:"required|min-length:1#请选择要删除的会话|请选择要删除的会话" dc:"会话ID列表"`
}

// ChatSessionDeleteManyRes 批量删除响应
type ChatSessionDeleteManyRes struct {
	Message string `json:"message,omitempty" dc:"删除结果消息"`
}

// ChatSessionDetailReq 获取对话会话详情
type ChatSessionDetailReq struct {
	g.Meta `path:"/sessions/{id}" method:"get" tags:"对话" summary:"获取对话会话详情"`
	ID     string `json:"id" in:"path" v:"required#会话ID不能为空" dc:"会话ID"`
}

// ChatSessionDetailRes 对话会话详情响应
type ChatSessionDetailRes struct {
	Session *ChatSessionListResModel `json:"session,omitempty" dc:"对话会话"`
}

// ChatSessionDeleteReq 删除对话会话
type ChatSessionDeleteReq struct {
	g.Meta `path:"/sessions/{id}" method:"delete" tags:"对话" summary:"删除对话会话"`
	ID     string `json:"id" in:"path" v:"required#会话ID不能为空" dc:"会话ID"`
}

// ChatSessionDeleteRes 删除响应
type ChatSessionDeleteRes struct {
	Message string `json:"message,omitempty" dc:"删除结果消息"`
}

// ChatMessageSendReq 发送对话消息
type ChatMessageSendReq struct {
	g.Meta      `path:"/sessions/{id}/messages" method:"post" tags:"对话" summary:"发送对话消息"`
	ID          string `json:"id" in:"path" v:"required#会话ID不能为空" dc:"会话ID"`
	Message     string `json:"message" v:"required#消息不能为空" dc:"消息内容"`
	LLMConfigID string `json:"llm_config_id" dc:"大模型配置ID"`
}

// ChatMessageSendRes 对话消息响应
type ChatMessageSendRes struct {
	Result *model.StreamChunk `json:"result,omitempty" dc:"发送结果"`
}
