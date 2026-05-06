package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowProtected updates local workflow protected status 修改本地工作流保护状态
func (c *ControllerV1) WorkflowProtected(ctx context.Context, req *v1.WorkflowProtectedReq) (res *v1.WorkflowProtectedRes, err error) {
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

	record, err := db.GetAutomaWorkflowRecord(fmt.Sprintf("%d", req.ID))
	if err != nil {
		return nil, err
	}
	if req.Revision > 0 && record.Revision != req.Revision {
		return nil, fmt.Errorf("版本已变化，请刷新后重试")
	}
	record.IsProtected = req.IsProtected
	record.Revision++
	if err = db.SaveAutomaWorkflowRecord(record); err != nil {
		return nil, err
	}
	return &v1.WorkflowProtectedRes{}, nil
}
