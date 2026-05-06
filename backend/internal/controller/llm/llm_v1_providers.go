package llm

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/llm/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
)

// LlmProviderList returns built-in model providers 获取内置大模型厂商列表
func (c *ControllerV1) LlmProviderList(ctx context.Context, req *v1.LlmProviderListReq) (res *v1.LlmProviderListRes, err error) {
	return &v1.LlmProviderListRes{Providers: llm.ModelCatalog()}, nil
}
