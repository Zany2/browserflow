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

// LlmConfigDetail returns model config detail 获取大模型配置详情
func (c *ControllerV1) LlmConfigDetail(ctx context.Context, req *v1.LlmConfigDetailReq) (res *v1.LlmConfigDetailRes, err error) {
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

	config, err := db.GetLLMConfig(req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.LlmConfigDetailRes{Config: config}, nil
}
