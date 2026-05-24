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
	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	sourceIP := strings.TrimSpace(req.SourceIP)
	pageNum := req.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}
	start := (pageNum - 1) * pageSize

	var records []*model.AutomaWorkflowRecord
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer {
		serverKeyword := keyword
		customKeyword := strings.ToLower(strings.TrimSpace(req.CustomKeyword))
		if customKeyword == "" && req.Syncable == 1 {
			customKeyword = serverKeyword
			serverKeyword = ""
		}
		columns := dao.AutomaWorkflows.Columns()
		dbModel := dao.AutomaWorkflows.Ctx(ctx)
		if req.Source > 0 {
			dbModel = dbModel.Where(columns.Source, req.Source)
		}
		if sourceIP != "" {
			dbModel = dbModel.Where(columns.SourceIp, sourceIP)
		}
		if req.Syncable == 1 {
			dbModel = dbModel.Where(columns.IsProtected, false)
		}
		if req.Syncable == 2 {
			dbModel = dbModel.Where(columns.IsProtected, true)
		}
		if customKeyword != "" {
			likeKeyword := "%" + customKeyword + "%"
			dbModel = dbModel.Where("(LOWER("+columns.Name+") LIKE ? OR LOWER("+columns.Description+") LIKE ?)", likeKeyword, likeKeyword)
		}
		if serverKeyword != "" {
			likeKeyword := "%" + serverKeyword + "%"
			dbModel = dbModel.Where(
				"(LOWER("+columns.AutomaId+") LIKE ? OR "+
					"LOWER("+columns.Name+") LIKE ? OR "+
					"LOWER("+columns.Description+") LIKE ? OR "+
					"LOWER("+columns.AutomaName+") LIKE ? OR "+
					"LOWER("+columns.AutomaDescription+") LIKE ? OR "+
					"LOWER("+columns.SourceIp+") LIKE ?)",
				likeKeyword,
				likeKeyword,
				likeKeyword,
				likeKeyword,
				likeKeyword,
				likeKeyword,
			)
		}

		total, countErr := dbModel.Count()
		if countErr != nil {
			return nil, countErr
		}

		items := []entity.AutomaWorkflows{}
		if err = dbModel.OrderDesc(columns.CreatedAt).OrderDesc(columns.Id).Limit(start, pageSize).Scan(&items); err != nil {
			return nil, err
		}

		list := make([]v1.WorkflowListResModel, 0, len(items))
		for index := range items {
			item := items[index]
			sourceText := ""
			if item.Source == 1 {
				sourceText = "页面导入"
			}
			if item.Source == 2 {
				sourceText = "客户端同步"
			}
			automaName := strings.TrimSpace(item.AutomaName)
			if automaName == "" {
				automaName = item.Name
			}
			automaDescription := strings.TrimSpace(item.AutomaDescription)
			if automaDescription == "" {
				automaDescription = item.Description
			}
			listItem := v1.WorkflowListResModel{Id: item.Id, AutomaId: item.AutomaId, Name: item.Name, Description: item.Description, AutomaName: automaName, AutomaDescription: automaDescription, Source: sourceText, SourceIp: item.SourceIp, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, ContentHash: item.ContentHash, Revision: item.Revision}
			if item.CreatedAt != nil && !item.CreatedAt.IsZero() {
				listItem.CreatedAt = item.CreatedAt
			}
			if item.UpdatedAt != nil && !item.UpdatedAt.IsZero() {
				listItem.UpdatedAt = item.UpdatedAt
			}
			list = append(list, listItem)
		}
		return &v1.WorkflowListRes{List: list, Total: total}, nil
	}

	{
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
