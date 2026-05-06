package llm

import (
	"context"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/llm/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// LlmConfigCheck tests model config 测试大模型配置
func (c *ControllerV1) LlmConfigCheck(ctx context.Context, req *v1.LlmConfigCheckReq) (res *v1.LlmConfigCheckRes, err error) {
	state.DBMu.Lock()
	if state.DB == nil {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
		}
		state.DB, err = storage.NewBoltDB(dbPath)
		if err != nil {
			state.DBMu.Unlock()
			return nil, err
		}
	}
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}
	llmClient := state.LLMClient
	state.DBMu.Unlock()

	config := &model.LLMConfig{
		ID:        req.LlmConfigPayload.ID,
		Name:      strings.TrimSpace(req.LlmConfigPayload.Name),
		Provider:  llm.NormalizeProvider(req.LlmConfigPayload.Provider),
		APIKey:    req.LlmConfigPayload.APIKey,
		Model:     strings.TrimSpace(req.LlmConfigPayload.Model),
		BaseURL:   llm.BaseURL(req.LlmConfigPayload.Provider, strings.TrimSpace(req.LlmConfigPayload.BaseURL)),
		IsDefault: req.LlmConfigPayload.IsDefault,
		IsActive:  req.LlmConfigPayload.IsActive,
	}
	if config.Provider == "ollama" && config.APIKey == "" {
		config.APIKey = "ollama"
	}
	if err = llm.ValidateConfig(config); err != nil {
		return &v1.LlmConfigCheckRes{Success: false, Message: err.Error()}, nil
	}
	if strings.TrimSpace(config.BaseURL) == "" {
		return &v1.LlmConfigCheckRes{Success: false, Message: "Base URL 不能为空"}, nil
	}
	if err = llmClient.Test(ctx, config); err != nil {
		return &v1.LlmConfigCheckRes{Success: false, Message: err.Error()}, nil
	}
	return &v1.LlmConfigCheckRes{Success: true, Message: "连接成功"}, nil
}
