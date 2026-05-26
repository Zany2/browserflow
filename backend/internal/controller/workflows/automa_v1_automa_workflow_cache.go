package workflows

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowCache returns cached workflow snapshot 返回本地工作流快照
func (c *ControllerV1) WorkflowCache(ctx context.Context, req *v1.WorkflowCacheReq) (res *v1.WorkflowCacheRes, err error) {
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer {
		columns := dao.AutomaWorkflows.Columns()
		records := []entity.AutomaWorkflows{}
		if err = dao.AutomaWorkflows.Ctx(ctx).OrderDesc(columns.CreatedAt).OrderDesc(columns.Id).Scan(&records); err != nil {
			return nil, err
		}
		workflows := make([]json.RawMessage, 0, len(records))
		for _, record := range records {
			if record.RawJson == "" {
				continue
			}
			workflows = append(workflows, json.RawMessage(record.RawJson))
		}
		payload, err := json.Marshal(workflows)
		if err != nil {
			return nil, err
		}
		return &v1.WorkflowCacheRes{Snapshot: &model.AutomaWorkflowSnapshot{ID: "latest", Workflows: payload}}, nil
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

	snapshot, err := db.GetAutomaWorkflowSnapshot("latest")
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowCacheRes{Snapshot: snapshot}, nil
}
