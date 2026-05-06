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
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// LlmConfigCreate creates model config 创建大模型配置
func (c *ControllerV1) LlmConfigCreate(ctx context.Context, req *v1.LlmConfigCreateReq) (res *v1.LlmConfigCreateRes, err error) {
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
	db := state.DB
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
	if config.ID == "" {
		config.ID = "llm_" + guid.S()
	}
	if config.Provider == "ollama" && config.APIKey == "" {
		config.APIKey = "ollama"
	}
	if err = llm.ValidateConfig(config); err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.BaseURL) == "" {
		return nil, gerror.New("Base URL 不能为空")
	}
	if err = db.SaveLLMConfig(config); err != nil {
		return nil, err
	}
	return &v1.LlmConfigCreateRes{Config: config}, nil
}
