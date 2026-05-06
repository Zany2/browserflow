package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// LlmConfigPayload 大模型配置载荷
type LlmConfigPayload struct {
	ID        string `json:"id" dc:"配置ID"`
	Name      string `json:"name" dc:"配置名称"`
	Provider  string `json:"provider" dc:"模型厂商"`
	APIKey    string `json:"api_key" dc:"接口密钥"`
	Model     string `json:"model" dc:"模型名称"`
	BaseURL   string `json:"base_url" dc:"接口基础地址"`
	IsDefault bool   `json:"is_default" dc:"是否默认配置"`
	IsActive  bool   `json:"is_active" dc:"是否启用"`
}

// LlmConfigListResModel 大模型配置列表项
type LlmConfigListResModel = model.LLMConfig

// LlmConfigListReq 获取大模型配置列表
type LlmConfigListReq struct {
	g.Meta `path:"/configs" method:"get" tags:"大模型" summary:"获取大模型配置列表"`
}

// LlmConfigListRes 大模型配置列表响应
type LlmConfigListRes struct {
	Configs []*LlmConfigListResModel `json:"configs,omitempty" dc:"大模型配置列表"`
}

// LlmConfigDetailReq 获取大模型配置详情
type LlmConfigDetailReq struct {
	g.Meta `path:"/configs/{id}" method:"get" tags:"大模型" summary:"获取大模型配置详情"`
	ID     string `json:"id" in:"path" v:"required#配置ID不能为空" dc:"配置ID"`
}

// LlmConfigDetailRes 大模型配置详情响应
type LlmConfigDetailRes struct {
	Config *LlmConfigListResModel `json:"config,omitempty" dc:"大模型配置"`
}

// LlmConfigCreateReq 创建大模型配置
type LlmConfigCreateReq struct {
	g.Meta `path:"/configs" method:"post" tags:"大模型" summary:"创建大模型配置"`
	LlmConfigPayload
}

// LlmConfigCreateRes 大模型配置创建响应
type LlmConfigCreateRes struct {
	Config *LlmConfigListResModel `json:"config,omitempty" dc:"大模型配置"`
}

// LlmConfigUpdateReq 更新大模型配置
type LlmConfigUpdateReq struct {
	g.Meta `path:"/configs/{id}" method:"put" tags:"大模型" summary:"更新大模型配置"`
	ID     string `json:"id" in:"path" v:"required#配置ID不能为空" dc:"配置ID"`
	LlmConfigPayload
}

// LlmConfigUpdateRes 大模型配置更新响应
type LlmConfigUpdateRes struct {
	Config *LlmConfigListResModel `json:"config,omitempty" dc:"大模型配置"`
}

// LlmConfigDeleteReq 删除大模型配置
type LlmConfigDeleteReq struct {
	g.Meta `path:"/configs/{id}" method:"delete" tags:"大模型" summary:"删除大模型配置"`
	ID     string `json:"id" in:"path" v:"required#配置ID不能为空" dc:"配置ID"`
}

// LlmConfigDeleteRes 删除响应
type LlmConfigDeleteRes struct {
	Message string `json:"message,omitempty" dc:"删除结果消息"`
}

// LlmConfigDeleteManyReq 批量删除大模型配置
type LlmConfigDeleteManyReq struct {
	g.Meta `path:"/configs" method:"delete" tags:"大模型" summary:"批量删除大模型配置"`
	IDs    []string `json:"ids" v:"required|min-length:1#请选择要删除的大模型配置|请选择要删除的大模型配置" dc:"配置ID列表"`
}

// LlmConfigDeleteManyRes 批量删除响应
type LlmConfigDeleteManyRes struct {
	Message string `json:"message,omitempty" dc:"删除结果消息"`
}

// LlmConfigCheckReq 检查大模型配置
type LlmConfigCheckReq struct {
	g.Meta `path:"/configs/test" method:"post" tags:"大模型" summary:"测试大模型配置"`
	LlmConfigPayload
}

// LlmConfigCheckRes 检查响应
type LlmConfigCheckRes struct {
	Success bool   `json:"success" dc:"是否测试成功"`
	Message string `json:"message" dc:"测试结果消息"`
}
