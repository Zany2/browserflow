package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAICompatibleClient OpenAI compatible client OpenAI 兼容客户端
type OpenAICompatibleClient struct{}

// NewOpenAICompatibleClient creates OpenAI compatible client 创建 OpenAI 兼容客户端
func NewOpenAICompatibleClient() *OpenAICompatibleClient {
	return &OpenAICompatibleClient{}
}

// StreamChat streams chat completion 流式对话
func (c *OpenAICompatibleClient) StreamChat(ctx context.Context, config *model.LLMConfig, messages []model.ChatMessage, emit func(string) error) error {
	if err := ValidateConfig(config); err != nil {
		return err
	}

	stream, err := c.newClient(config).CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    config.Model,
		Messages: buildMessages(messages),
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("create chat stream failed: %w", err)
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("receive chat stream failed: %w", err)
		}
		for _, choice := range resp.Choices {
			if choice.Delta.Content == "" {
				continue
			}
			if err := emit(choice.Delta.Content); err != nil {
				return err
			}
		}
	}
}

// Test tests model config 测试连接
func (c *OpenAICompatibleClient) Test(ctx context.Context, config *model.LLMConfig) error {
	return c.StreamChat(ctx, config, []model.ChatMessage{{
		Role:      "user",
		Content:   "Reply with OK.",
		Timestamp: time.Now(),
	}}, func(string) error { return nil })
}

// newClient creates SDK client from config 根据配置创建 SDK 客户端
func (c *OpenAICompatibleClient) newClient(config *model.LLMConfig) *openai.Client {
	apiKey := strings.TrimSpace(config.APIKey)
	if NormalizeProvider(config.Provider) == "ollama" && apiKey == "" {
		apiKey = "ollama"
	}

	clientConfig := openai.DefaultConfig(apiKey)
	clientConfig.BaseURL = BaseURL(config.Provider, config.BaseURL)
	return openai.NewClientWithConfig(clientConfig)
}

// buildMessages converts local messages to OpenAI messages 转换本地消息为 OpenAI 消息
func buildMessages(messages []model.ChatMessage) []openai.ChatCompletionMessage {
	result := []openai.ChatCompletionMessage{{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful assistant. Reply in the same language as the user.",
	}}
	for _, message := range messages {
		if message.Role == "" || message.Content == "" {
			continue
		}
		result = append(result, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return result
}
