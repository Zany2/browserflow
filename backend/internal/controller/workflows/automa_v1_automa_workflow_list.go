package workflows

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// WorkflowList returns local-file workflow records 返回本地文件中的工作流列表
func (c *ControllerV1) WorkflowList(ctx context.Context, req *v1.WorkflowListReq) (res *v1.WorkflowListRes, err error) {
	var records []*model.AutomaWorkflowRecord
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer {
		columns := dao.AutomaWorkflows.Columns()
		items := []entity.AutomaWorkflows{}
		if err = dao.AutomaWorkflows.Ctx(ctx).OrderDesc(columns.UpdatedAt).Scan(&items); err != nil {
			return nil, err
		}
		records = make([]*model.AutomaWorkflowRecord, 0, len(items))
		for index := range items {
			item := items[index]
			record := &model.AutomaWorkflowRecord{ID: item.Id, AutomaID: item.AutomaId, Name: item.Name, Description: item.Description, AutomaName: item.AutomaName, AutomaDescription: item.AutomaDescription, Source: item.Source, SourceIP: item.SourceIp, SourceUserAgent: item.SourceUserAgent, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, RawJSON: item.RawJson, NormalizedJSON: item.NormalizedJson, ContentHash: item.ContentHash, Revision: item.Revision}
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
			records = append(records, record)
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

		records, err = db.ListAutomaWorkflowRecords()
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	sourceIP := strings.TrimSpace(req.SourceIP)
	filtered := make([]*model.AutomaWorkflowRecord, 0, len(records))
	for _, record := range records {
		if req.Source > 0 && record.Source != req.Source {
			continue
		}
		if sourceIP != "" && record.SourceIP != sourceIP {
			continue
		}
		if req.Syncable == 1 && record.IsProtected {
			continue
		}
		if req.Syncable == 2 && !record.IsProtected {
			continue
		}
		if keyword != "" {
			text := strings.ToLower(strings.Join([]string{record.AutomaID, record.Name, record.Description, record.AutomaName, record.AutomaDescription, record.SourceIP}, " "))
			if !strings.Contains(text, keyword) {
				continue
			}
		}
		filtered = append(filtered, record)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
	})
	pageNum := req.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}
	start := (pageNum - 1) * pageSize
	end := start + pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	list := make([]v1.WorkflowListResModel, 0, end-start)
	for _, record := range filtered[start:end] {
		sourceText := ""
		if record.Source == 1 {
			sourceText = "页面导入"
		}
		if record.Source == 2 {
			sourceText = "客户端同步"
		}
		automaName := strings.TrimSpace(record.AutomaName)
		if automaName == "" {
			automaName = record.Name
		}
		automaDescription := strings.TrimSpace(record.AutomaDescription)
		if automaDescription == "" {
			automaDescription = record.Description
		}
		item := v1.WorkflowListResModel{Id: record.ID, AutomaId: record.AutomaID, Name: record.Name, Description: record.Description, AutomaName: automaName, AutomaDescription: automaDescription, Source: sourceText, SourceIp: record.SourceIP, CreatedAtAutoma: record.CreatedAtAutoma, UpdatedAtAutoma: record.UpdatedAtAutoma, IsDisabled: record.IsDisabled, IsProtected: record.IsProtected, NodeCount: record.NodeCount, EdgeCount: record.EdgeCount, ContentHash: record.ContentHash, Revision: record.Revision}
		if !record.CreatedAt.IsZero() {
			item.CreatedAt = gtime.NewFromTime(record.CreatedAt)
		}
		if !record.UpdatedAt.IsZero() {
			item.UpdatedAt = gtime.NewFromTime(record.UpdatedAt)
		}
		list = append(list, item)
	}
	return &v1.WorkflowListRes{List: list, Total: len(filtered)}, nil
}
