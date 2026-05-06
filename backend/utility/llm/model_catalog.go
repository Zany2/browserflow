package llm

import "github.com/Zany2/browserflow/backend/internal/model"

// ModelCatalog returns built-in model providers 返回内置模型厂商
func ModelCatalog() []model.LLMProvider {
	return []model.LLMProvider{
		provider("openai", "OpenAI", []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"}),
		provider("anthropic", "Anthropic Claude", []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229"}),
		provider("gemini", "Google Gemini", []string{"gemini-2.0-flash-exp", "gemini-1.5-pro", "gemini-1.5-flash"}),
		provider("mistral", "Mistral AI", []string{"mistral-large-latest", "mistral-medium-latest", "mistral-small-latest"}),
		provider("deepseek", "DeepSeek", []string{"deepseek-chat", "deepseek-coder"}),
		provider("qwen", "通义千问", []string{"qwen-max", "qwen-plus", "qwen-turbo", "qwen-long"}),
		provider("siliconflow", "SiliconFlow", []string{"deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"}),
		provider("doubao", "豆包", []string{"doubao-pro-32k", "doubao-lite-32k"}),
		provider("moonshot", "Moonshot", []string{"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"}),
		provider("ollama", "Ollama", []string{"qwen2.5:latest", "llama3.3:latest", "deepseek-r1:latest", "mistral:latest"}),
		{ID: "custom", Name: "自定义 OpenAI 兼容服务", BaseURL: "", Compatible: true, Models: []string{}},
	}
}

// provider builds provider metadata 构建厂商元数据
func provider(id, name string, models []string) model.LLMProvider {
	return model.LLMProvider{
		ID:         id,
		Name:       name,
		BaseURL:    BaseURL(id, ""),
		Compatible: true,
		Models:     models,
	}
}
