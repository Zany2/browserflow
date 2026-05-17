package automa

import (
	"context"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/automa/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// AutomaWorkflowBatchDelete deletes local workflow records 批量删除本地工作流
func (c *ControllerV1) AutomaWorkflowBatchDelete(ctx context.Context, req *v1.AutomaWorkflowBatchDeleteReq) (res *v1.AutomaWorkflowBatchDeleteRes, err error) {
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

	for _, workflowID := range req.IDs {
		workflowID = strings.TrimSpace(workflowID)
		if workflowID == "" {
			continue
		}
		if err = db.DeleteAutomaWorkflowRecord(workflowID); err != nil {
			return nil, err
		}
	}
	return &v1.AutomaWorkflowBatchDeleteRes{}, nil
}
