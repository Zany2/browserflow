package automa

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/automa/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// AutomaWorkflowCache returns cached workflow snapshot 返回本地工作流快照
func (c *ControllerV1) AutomaWorkflowCache(ctx context.Context, req *v1.AutomaWorkflowCacheReq) (res *v1.AutomaWorkflowCacheRes, err error) {
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
	return &v1.AutomaWorkflowCacheRes{Snapshot: snapshot}, nil
}
