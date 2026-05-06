package workflows

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowCache returns cached workflow snapshot 返回本地工作流快照
func (c *ControllerV1) WorkflowCache(ctx context.Context, req *v1.WorkflowCacheReq) (res *v1.WorkflowCacheRes, err error) {
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

	snapshot, err := db.GetAutomaWorkflowSnapshot("latest")
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowCacheRes{Snapshot: snapshot}, nil
}
