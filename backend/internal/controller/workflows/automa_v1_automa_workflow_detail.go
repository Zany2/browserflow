package workflows

import (
	"context"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// WorkflowDetail returns one local workflow detail 返回单个本地工作流详情
func (c *ControllerV1) WorkflowDetail(ctx context.Context, req *v1.WorkflowDetailReq) (res *v1.WorkflowDetailRes, err error) {
	var record *model.AutomaWorkflowRecord
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer {
		columns := dao.AutomaWorkflows.Columns()
		item := entity.AutomaWorkflows{}
		if primaryID := gconv.Int64(req.ID); primaryID > 0 {
			if err = dao.AutomaWorkflows.Ctx(ctx).WherePri(primaryID).Scan(&item); err != nil {
				return nil, err
			}
		}
		if item.Id <= 0 {
			if err = dao.AutomaWorkflows.Ctx(ctx).Where(columns.AutomaId, req.ID).Scan(&item); err != nil {
				return nil, err
			}
		}
		if item.Id <= 0 {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流不存在")
			return nil, nil
		}
		record = &model.AutomaWorkflowRecord{ID: item.Id, AutomaID: item.AutomaId, Name: item.Name, Description: item.Description, AutomaName: item.AutomaName, AutomaDescription: item.AutomaDescription, Source: item.Source, SourceIP: item.SourceIp, SourceNodeID: item.SourceNodeId, SourceUserAgent: item.SourceUserAgent, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, RawJSON: item.RawJson, NormalizedJSON: item.NormalizedJson, ContentHash: item.ContentHash, Revision: item.Revision}
		if item.FirstSyncedAt != nil && !item.FirstSyncedAt.IsZero() {
			record.FirstSyncedAt = item.FirstSyncedAt.Time
		}
		if item.LastSyncedAt != nil && !item.LastSyncedAt.IsZero() {
			record.LastSyncedAt = item.LastSyncedAt.Time
		}
		if item.CreatedAt != nil && !item.CreatedAt.IsZero() {
			record.CreatedAt = item.CreatedAt.Time
		}
		if item.UpdatedAt != nil && !item.UpdatedAt.IsZero() {
			record.UpdatedAt = item.UpdatedAt.Time
		}
	} else {
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

		record, err = db.GetAutomaWorkflowRecord(req.ID)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	source := ""
	if record.Source == 1 {
		source = "页面导入"
	}
	if record.Source == 2 {
		source = "客户端同步"
	}
	if !record.FirstSyncedAt.IsZero() {
		res = &v1.WorkflowDetailRes{FirstSyncedAt: gtime.NewFromTime(record.FirstSyncedAt)}
	} else {
		res = &v1.WorkflowDetailRes{}
	}
	if !record.LastSyncedAt.IsZero() {
		res.LastSyncedAt = gtime.NewFromTime(record.LastSyncedAt)
	}
	if !record.CreatedAt.IsZero() {
		res.CreatedAt = gtime.NewFromTime(record.CreatedAt)
	}
	if !record.UpdatedAt.IsZero() {
		res.UpdatedAt = gtime.NewFromTime(record.UpdatedAt)
	}
	res.Id = record.ID
	res.AutomaId = record.AutomaID
	res.Name = record.Name
	res.Description = record.Description
	res.AutomaName = strings.TrimSpace(record.AutomaName)
	if res.AutomaName == "" {
		res.AutomaName = record.Name
	}
	res.AutomaDescription = strings.TrimSpace(record.AutomaDescription)
	if res.AutomaDescription == "" {
		res.AutomaDescription = record.Description
	}
	res.Source = source
	res.SourceIp = record.SourceIP
	res.SourceNodeId = record.SourceNodeID
	res.SourceUserAgent = record.SourceUserAgent
	res.AutomaVersion = record.AutomaVersion
	res.ExtVersion = record.ExtVersion
	res.CreatedAtAutoma = record.CreatedAtAutoma
	res.UpdatedAtAutoma = record.UpdatedAtAutoma
	res.IsDisabled = record.IsDisabled
	res.IsProtected = record.IsProtected
	res.NodeCount = record.NodeCount
	res.EdgeCount = record.EdgeCount
	res.RawJson = record.RawJSON
	res.NormalizedJson = record.NormalizedJSON
	res.ContentHash = record.ContentHash
	res.Revision = record.Revision
	return res, nil
}
