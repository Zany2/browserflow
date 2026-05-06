package workflows

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// WorkflowList returns local-file workflow records 返回本地文件中的工作流列表
func (c *ControllerV1) WorkflowList(ctx context.Context, req *v1.WorkflowListReq) (res *v1.WorkflowListRes, err error) {
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

	records, err := db.ListAutomaWorkflowRecords()
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
			text := strings.ToLower(strings.Join([]string{record.AutomaID, record.Name, record.Description, record.SourceIP}, " "))
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
		item := v1.WorkflowListResModel{Id: record.ID, AutomaId: record.AutomaID, Name: record.Name, Description: record.Description, Source: sourceText, SourceIp: record.SourceIP, CreatedAtAutoma: record.CreatedAtAutoma, UpdatedAtAutoma: record.UpdatedAtAutoma, IsDisabled: record.IsDisabled, IsProtected: record.IsProtected, NodeCount: record.NodeCount, EdgeCount: record.EdgeCount, ContentHash: record.ContentHash, Revision: record.Revision}
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
