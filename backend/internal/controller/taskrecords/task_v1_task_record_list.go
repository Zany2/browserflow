package taskrecords

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskRecordList returns task records 获取任务记录列表
func (c *ControllerV1) TaskRecordList(ctx context.Context, req *v1.TaskRecordListReq) (res *v1.TaskRecordListRes, err error) {
	columns := dao.TaskRecords.Columns()
	gModel := dao.TaskRecords.Ctx(ctx).Where(columns.DeletedAt + " IS NULL")

	if taskID := strings.TrimSpace(req.TaskID); taskID != "" {
		gModel = gModel.Where(columns.TaskId, gconv.Int64(taskID))
	}
	if workflowID := strings.TrimSpace(req.WorkflowID); workflowID != "" {
		gModel = gModel.Where(columns.WorkflowId, workflowID)
	}
	// Workflow name filter 工作流名称模糊检索，转换为执行记录可匹配的工作流 ID
	if workflowName := strings.TrimSpace(req.WorkflowName); workflowName != "" {
		workflowIDs, workflowErr := findWorkflowIDsByName(ctx, workflowName)
		if workflowErr != nil {
			return nil, workflowErr
		}
		if len(workflowIDs) == 0 {
			return &v1.TaskRecordListRes{List: []*v1.TaskRecordListResModel{}, Total: 0}, nil
		}
		gModel = gModel.WhereIn(columns.WorkflowId, workflowIDs)
	}
	if clientIP, resolveErr := resolveClientIP(ctx, req.ClientID, req.ClientIP); resolveErr != nil {
		return nil, resolveErr
	} else if clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	} else if clientID := strings.TrimSpace(req.ClientID); clientID != "" {
		gModel = gModel.Where(columns.ClientIp, clientID)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		gModel = gModel.Where(columns.Status, status)
	}
	// Execute time range filter 按开始执行时间做起止范围筛选
	if startTime := strings.TrimSpace(req.StartTime); startTime != "" {
		gModel = gModel.WhereGTE(columns.StartedAt, startTime)
	}
	if endTime := strings.TrimSpace(req.EndTime); endTime != "" {
		gModel = gModel.WhereLTE(columns.StartedAt, endTime)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		gModel = gModel.Where("("+columns.WorkflowId+" LIKE ? OR "+columns.ClientIp+" LIKE ? OR "+columns.ErrorMessage+" LIKE ?)", likeKeyword, likeKeyword, likeKeyword)
	}

	records, err := gModel.OrderDesc(columns.CreatedAt).All()
	if err != nil {
		return nil, err
	}

	list := make([]*v1.TaskRecordListResModel, 0, len(records))
	for _, record := range records {
		item, mapErr := buildTaskRecordMap(ctx, record)
		if mapErr != nil {
			return nil, mapErr
		}
		list = append(list, item)
	}

	return &v1.TaskRecordListRes{List: list, Total: len(list)}, nil
}
