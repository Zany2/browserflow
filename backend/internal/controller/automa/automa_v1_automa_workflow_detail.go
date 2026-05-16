package automa

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/automa/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AutomaWorkflowDetail returns one local workflow detail 返回单个本地工作流详情
func (c *ControllerV1) AutomaWorkflowDetail(ctx context.Context, req *v1.AutomaWorkflowDetailReq) (res *v1.AutomaWorkflowDetailRes, err error) {
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

	record, err := db.GetAutomaWorkflowRecord(req.ID)
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
		res = &v1.AutomaWorkflowDetailRes{FirstSyncedAt: gtime.NewFromTime(record.FirstSyncedAt)}
	} else {
		res = &v1.AutomaWorkflowDetailRes{}
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
	res.Source = source
	res.SourceIp = record.SourceIP
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
