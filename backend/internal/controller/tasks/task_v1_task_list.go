package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
)

// TaskList returns tasks 获取任务列表
func (c *ControllerV1) TaskList(ctx context.Context, req *v1.TaskListReq) (res *v1.TaskListRes, err error) {
	columns := dao.Tasks.Columns()
	gModel := dao.Tasks.Ctx(ctx)

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		gModel = gModel.Where("("+columns.Name+" LIKE ? OR "+columns.Description+" LIKE ?)", likeKeyword, likeKeyword)
	}
	if workflowID := strings.TrimSpace(req.WorkflowID); workflowID != "" {
		gModel = gModel.Where(columns.AutomaId, workflowID)
	}
	// Workflow name filter 工作流名称模糊检索，转换为任务表可匹配的工作流 ID
	if workflowName := strings.TrimSpace(req.WorkflowName); workflowName != "" {
		workflowIDs, workflowErr := taskdata.FindWorkflowIDsByName(ctx, workflowName)
		if workflowErr != nil {
			return nil, workflowErr
		}
		if len(workflowIDs) == 0 {
			return &v1.TaskListRes{List: []*v1.TaskListResModel{}, Total: 0}, nil
		}
		gModel = gModel.WhereIn(columns.AutomaId, workflowIDs)
	}
	if clientIP, resolveErr := taskdata.ResolveClientIP(ctx, req.ClientID, ""); resolveErr != nil {
		return nil, resolveErr
	} else if clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	} else if clientID := strings.TrimSpace(req.ClientID); clientID != "" {
		gModel = gModel.Where(columns.ClientIp, clientID)
	}
	if enabled := strings.TrimSpace(req.Enabled); enabled == "true" || enabled == "false" {
		gModel = gModel.Where(columns.Enabled, enabled == "true")
	}
	// Created time range filter 创建时间起止范围筛选
	if startTime := strings.TrimSpace(req.StartTime); startTime != "" {
		gModel = gModel.WhereGTE(columns.CreatedAt, startTime)
	}
	if endTime := strings.TrimSpace(req.EndTime); endTime != "" {
		gModel = gModel.WhereLTE(columns.CreatedAt, endTime)
	}

	records, err := gModel.OrderDesc(columns.UpdatedAt).All()
	if err != nil {
		return nil, err
	}

	list := make([]*v1.TaskListResModel, 0, len(records))
	for _, record := range records {
		item, mapErr := taskdata.BuildTaskMap(ctx, record)
		if mapErr != nil {
			return nil, mapErr
		}
		list = append(list, item)
	}

	return &v1.TaskListRes{List: list, Total: len(list)}, nil
}
