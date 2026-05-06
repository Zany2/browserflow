package llm

import "strings"

// NormalizeProvider normalizes provider id 统一厂商标识
func NormalizeProvider(provider string) string {
	value := strings.ToLower(strings.TrimSpace(provider))
	if value == "" {
		return "openai"
	}
	if value == "claude" {
		return "anthropic"
	}
	return value
}

// BaseURL returns provider default gateway 获取厂商默认网关
func BaseURL(provider, customBaseURL string) string {
	if customBaseURL != "" {
		return strings.TrimRight(customBaseURL, "/")
	}

	baseURLMap := map[string]string{
		"openai":      "https://api.openai.com/v1",
		"anthropic":   "https://api.anthropic.com",
		"gemini":      "https://generativelanguage.googleapis.com/v1beta/openai",
		"mistral":     "https://api.mistral.ai/v1",
		"deepseek":    "https://api.deepseek.com",
		"groq":        "https://api.groq.com/openai/v1",
		"cohere":      "https://api.cohere.ai/v1",
		"xai":         "https://api.x.ai/v1",
		"together":    "https://api.together.xyz/v1",
		"novita":      "https://api.novita.ai/v3/openai",
		"openrouter":  "https://openrouter.ai/api/v1",
		"qwen":        "https://dashscope.aliyuncs.com/compatible-mode/v1",
		"siliconflow": "https://api.siliconflow.cn/v1",
		"doubao":      "https://ark.cn-beijing.volces.com/api/v3",
		"ernie":       "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop",
		"spark":       "https://spark-api-open.xf-yun.com/v1",
		"chatglm":     "https://open.bigmodel.cn/api/paas/v4",
		"360":         "https://api.360.cn/v1",
		"hunyuan":     "https://hunyuan.tencentcloudapi.com",
		"moonshot":    "https://api.moonshot.cn/v1",
		"baichuan":    "https://api.baichuan-ai.com/v1",
		"minimax":     "https://api.minimax.chat/v1",
		"yi":          "https://api.lingyiwanwu.com/v1",
		"stepfun":     "https://api.stepfun.com/v1",
		"coze":        "https://api.coze.cn/open_api/v2",
		"ollama":      "http://localhost:11434/v1",
	}

	return baseURLMap[NormalizeProvider(provider)]
}
