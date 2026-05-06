package llm

import (
	"context"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/llm/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// LlmConfigDeleteMany deletes model configs in batch 批量删除大模型配置
func (c *ControllerV1) LlmConfigDeleteMany(ctx context.Context, req *v1.LlmConfigDeleteManyReq) (res *v1.LlmConfigDeleteManyRes, err error) {
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

	for _, id := range req.IDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if err = db.DeleteLLMConfig(id); err != nil {
			return nil, err
		}
	}
	return &v1.LlmConfigDeleteManyRes{Message: "批量删除成功"}, nil
}
