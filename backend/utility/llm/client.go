package llm

import (
	"context"
	"errors"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
)

// Client large model client 大模型统一客户端
type Client struct {
	openAICompatible *OpenAICompatibleClient
}

// NewClient creates large model client 创建大模型客户端
func NewClient() *Client {
	return &Client{openAICompatible: NewOpenAICompatibleClient()}
}

// Providers returns supported providers 获取支持的大模型厂商
func (c *Client) Providers() []model.LLMProvider {
	return ModelCatalog()
}

// StreamChat streams chat completion 流式对话
func (c *Client) StreamChat(ctx context.Context, config *model.LLMConfig, messages []model.ChatMessage, emit func(string) error) error {
	return c.openAICompatible.StreamChat(ctx, config, messages, emit)
}

// Test tests model config 测试模型配置
func (c *Client) Test(ctx context.Context, config *model.LLMConfig) error {
	return c.openAICompatible.Test(ctx, config)
}

// ValidateConfig validates large model config 校验大模型配置
func ValidateConfig(config *model.LLMConfig) error {
	if strings.TrimSpace(config.Name) == "" {
		return errors.New("配置名称不能为空")
	}
	if strings.TrimSpace(config.Provider) == "" {
		return errors.New("提供商不能为空")
	}
	if strings.TrimSpace(config.Model) == "" {
		return errors.New("模型名称不能为空")
	}
	if NormalizeProvider(config.Provider) != "ollama" && strings.TrimSpace(config.APIKey) == "" {
		return errors.New("API Key 不能为空")
	}
	return nil
}
