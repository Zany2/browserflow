package llm

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/llm/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// LlmConfigList returns model configs 获取大模型配置列表
func (c *ControllerV1) LlmConfigList(ctx context.Context, req *v1.LlmConfigListReq) (res *v1.LlmConfigListRes, err error) {
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

	configs, err := db.ListLLMConfigs()
	if err != nil {
		return nil, err
	}
	return &v1.LlmConfigListRes{Configs: configs}, nil
}
