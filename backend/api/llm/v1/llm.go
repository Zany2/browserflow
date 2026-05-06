package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// LlmProviderListResModel 模型厂商列表项
type LlmProviderListResModel = model.LLMProvider

// LlmProviderListReq 获取内置模型厂商
type LlmProviderListReq struct {
	g.Meta `path:"/providers" method:"get" tags:"大模型" summary:"获取大模型厂商列表"`
}

// LlmProviderListRes 模型厂商列表响应
type LlmProviderListRes struct {
	Providers []LlmProviderListResModel `json:"providers,omitempty" dc:"大模型厂商列表"`
}
