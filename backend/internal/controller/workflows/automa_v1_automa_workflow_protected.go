package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowProtected updates local workflow protected status 修改本地工作流保护状态
func (c *ControllerV1) WorkflowProtected(ctx context.Context, req *v1.WorkflowProtectedReq) (res *v1.WorkflowProtectedRes, err error) {
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer {
		item := entity.AutomaWorkflows{}
		if err = dao.AutomaWorkflows.Ctx(ctx).WherePri(req.ID).Scan(&item); err != nil {
			return nil, err
		}
		if item.Id <= 0 {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流不存在")
			return nil, nil
		}
		if req.Revision > 0 && item.Revision != req.Revision {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "版本已变化，请刷新后重试")
			return nil, nil
		}
		if _, err = dao.AutomaWorkflows.Ctx(ctx).WherePri(req.ID).Data(do.AutomaWorkflows{IsProtected: req.IsProtected, Revision: item.Revision + 1}).Update(); err != nil {
			return nil, err
		}
		return &v1.WorkflowProtectedRes{}, nil
	}

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
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "版本已变化，请刷新后重试")
		return nil, nil
	}
	record.IsProtected = req.IsProtected
	record.Revision++
	if err = db.SaveAutomaWorkflowRecord(record); err != nil {
		return nil, err
	}
	return &v1.WorkflowProtectedRes{}, nil
}
