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

// LlmConfigDelete deletes model config 删除大模型配置
func (c *ControllerV1) LlmConfigDelete(ctx context.Context, req *v1.LlmConfigDeleteReq) (res *v1.LlmConfigDeleteRes, err error) {
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

	if err = db.DeleteLLMConfig(req.ID); err != nil {
		return nil, err
	}
	return &v1.LlmConfigDeleteRes{Message: "删除成功"}, nil
}
