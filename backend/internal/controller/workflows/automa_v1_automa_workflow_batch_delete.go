package workflows

import (
	"context"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowBatchDelete deletes local workflow records 批量删除本地工作流
func (c *ControllerV1) WorkflowBatchDelete(ctx context.Context, req *v1.WorkflowBatchDeleteReq) (res *v1.WorkflowBatchDeleteRes, err error) {
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
	return &v1.WorkflowBatchDeleteRes{}, nil
}
