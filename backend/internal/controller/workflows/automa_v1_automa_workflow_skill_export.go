package workflows

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/workflowskill"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowExportSkill exports stored workflows as SKILL.md 导出已入库工作流为 SKILL.md
func (c *ServerControllerV1) WorkflowExportSkill(ctx context.Context, req *v1.WorkflowExportSkillReq) (res *v1.WorkflowExportSkillRes, err error) {
	if consts.ResolveRuntimeMode(ctx) != consts.RuntimeModeServer {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "当前接口仅支持 Server 模式")
		return nil, nil
	}

	records, err := loadServerWorkflowRecordsForSkill(ctx)
	if err != nil {
		return nil, err
	}

	workflows := make([]map[string]any, 0, len(records))
	for _, record := range records {
		workflow := storedWorkflowRecordToSkillMap(record)
		if len(workflow) > 0 {
			workflows = append(workflows, workflow)
		}
	}

	workflows = workflowskill.FilterServerWorkflows(workflows, req.Scope, req.WorkflowIDs)
	if len(workflows) == 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "没有可导出的工作流")
		return nil, nil
	}

	request := g.RequestFromCtx(ctx)
	baseURL := workflowskill.BaseURLFromFrontendURL(g.Cfg().MustGet(ctx, "frontend.url", "").String())
	if baseURL == "" {
		baseURL = workflowskill.BaseURLFromFrontendURL(os.Getenv("FRONTEND_URL"))
	}
	if baseURL == "" {
		baseURL = workflowskill.BaseURL(request.Host, request.TLS != nil)
	}
	request.Response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	request.Response.Header().Set("Content-Disposition", workflowskill.ContentDisposition(workflowskill.FileName))
	request.Response.Write(workflowskill.GenerateServerMarkdown(workflows, baseURL))

	request.ExitAll()
	return nil, nil
}

func loadServerWorkflowRecordsForSkill(ctx context.Context) ([]*model.AutomaWorkflowRecord, error) {
	columns := dao.AutomaWorkflows.Columns()
	items := []entity.AutomaWorkflows{}
	if err := dao.AutomaWorkflows.Ctx(ctx).OrderDesc(columns.CreatedAt).OrderDesc(columns.Id).Scan(&items); err != nil {
		return nil, err
	}

	records := make([]*model.AutomaWorkflowRecord, 0, len(items))
	for index := range items {
		item := items[index]
		records = append(records, &model.AutomaWorkflowRecord{
			ID:                item.Id,
			AutomaID:          item.AutomaId,
			Name:              item.Name,
			Description:       item.Description,
			AutomaName:        item.AutomaName,
			AutomaDescription: item.AutomaDescription,
			Source:            item.Source,
			SourceIP:          item.SourceIp,
			AutomaVersion:     item.AutomaVersion,
			ExtVersion:        item.ExtVersion,
			CreatedAtAutoma:   item.CreatedAtAutoma,
			UpdatedAtAutoma:   item.UpdatedAtAutoma,
			IsDisabled:        item.IsDisabled,
			NodeCount:         item.NodeCount,
			EdgeCount:         item.EdgeCount,
			RawJSON:           item.RawJson,
			NormalizedJSON:    item.NormalizedJson,
		})
	}
	return records, nil
}

func storedWorkflowRecordToSkillMap(record *model.AutomaWorkflowRecord) map[string]any {
	if record == nil {
		return nil
	}

	workflow := parseStoredWorkflowJSON(record.NormalizedJSON)
	if len(workflow) == 0 {
		workflow = parseStoredWorkflowJSON(record.RawJSON)
	}
	if workflow == nil {
		workflow = map[string]any{}
	}

	serverID := record.ID
	if serverID > 0 {
		workflow["id"] = serverID
		workflow["server_id"] = serverID
	}
	if strings.TrimSpace(record.AutomaID) != "" {
		workflow["automa_id"] = record.AutomaID
	}
	if strings.TrimSpace(record.Name) != "" {
		workflow["name"] = record.Name
	}
	if strings.TrimSpace(record.Description) != "" {
		workflow["description"] = record.Description
	}
	if strings.TrimSpace(record.AutomaName) != "" {
		workflow["automa_name"] = record.AutomaName
	}
	if strings.TrimSpace(record.AutomaDescription) != "" {
		workflow["automa_description"] = record.AutomaDescription
	}
	workflow["is_disabled"] = record.IsDisabled
	workflow["node_count"] = record.NodeCount
	workflow["edge_count"] = record.EdgeCount
	if record.CreatedAtAutoma > 0 {
		workflow["created_at"] = record.CreatedAtAutoma
	}
	if record.UpdatedAtAutoma > 0 {
		workflow["updated_at"] = record.UpdatedAtAutoma
	}
	return workflow
}

func parseStoredWorkflowJSON(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var workflow map[string]any
	if err := json.Unmarshal([]byte(raw), &workflow); err != nil {
		return nil
	}
	return workflow
}
