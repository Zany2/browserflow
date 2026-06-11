package workflows

import (
	"context"
	"fmt"
	"os"
	"strings"

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

// WorkflowUpdate updates local workflow metadata 修改本地工作流元信息
func (c *ControllerV1) WorkflowUpdate(ctx context.Context, req *v1.WorkflowUpdateReq) (res *v1.WorkflowUpdateRes, err error) {
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
		updateData := do.AutomaWorkflows{Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description), IsProtected: req.IsProtected, Revision: item.Revision + 1}
		if req.Source == 1 || req.Source == 2 {
			updateData.Source = req.Source
		}
		if _, err = dao.AutomaWorkflows.Ctx(ctx).WherePri(req.ID).Data(updateData).Update(); err != nil {
			return nil, err
		}
		return &v1.WorkflowUpdateRes{}, nil
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

	record, err := db.GetAutomaWorkflowRecord(fmt.Sprintf("%v", req.ID))
	if err != nil {
		return nil, err
	}
	if req.Revision > 0 && record.Revision != req.Revision {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "版本已变化，请刷新后重试")
		return nil, nil
	}
	record.Name = strings.TrimSpace(req.Name)
	record.Description = strings.TrimSpace(req.Description)
	if req.Source == 1 || req.Source == 2 {
		record.Source = req.Source
	}
	record.IsProtected = req.IsProtected
	record.Revision++
	if err = db.SaveAutomaWorkflowRecord(record); err != nil {
		return nil, err
	}
	return &v1.WorkflowUpdateRes{}, nil
}
