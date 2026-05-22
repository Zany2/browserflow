package taskrecords

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskRecordList returns task records.
func (c *ControllerV1) TaskRecordList(ctx context.Context, req *v1.TaskRecordListReq) (res *v1.TaskRecordListRes, err error) {
	columns := dao.TaskRecords.Columns()
	gModel := dao.TaskRecords.Ctx(ctx)

	if taskID := strings.TrimSpace(req.TaskID); taskID != "" {
		gModel = gModel.Where(columns.TaskId, gconv.Int64(taskID))
	}
	if taskName := strings.TrimSpace(req.TaskName); taskName != "" {
		taskIDs, taskErr := taskdata.FindTaskIDsByName(ctx, taskName)
		if taskErr != nil {
			return nil, taskErr
		}
		if len(taskIDs) == 0 {
			return &v1.TaskRecordListRes{List: []*v1.TaskRecordListResModel{}, Total: 0}, nil
		}
		gModel = gModel.WhereIn(columns.TaskId, taskIDs)
	}
	if workflowID := strings.TrimSpace(req.WorkflowID); workflowID != "" {
		gModel = gModel.Where(columns.WorkflowId, workflowID)
	}
	if workflowName := strings.TrimSpace(req.WorkflowName); workflowName != "" {
		workflowIDs, workflowErr := taskdata.FindWorkflowIDsByName(ctx, workflowName)
		if workflowErr != nil {
			return nil, workflowErr
		}
		if len(workflowIDs) == 0 {
			return &v1.TaskRecordListRes{List: []*v1.TaskRecordListResModel{}, Total: 0}, nil
		}
		gModel = gModel.WhereIn(columns.WorkflowId, workflowIDs)
	}
	if clientIP, resolveErr := taskdata.ResolveClientIP(ctx, req.ClientID, req.ClientIP); resolveErr != nil {
		return nil, resolveErr
	} else if clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	} else if clientID := strings.TrimSpace(req.ClientID); clientID != "" {
		gModel = gModel.Where(columns.ClientIp, clientID)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		gModel = gModel.Where(columns.Status, status)
	}
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

	total, err := gModel.Count()
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &v1.TaskRecordListRes{List: []*v1.TaskRecordListResModel{}, Total: 0}, nil
	}

	pageNum := req.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}
	start := (pageNum - 1) * pageSize

	records, err := gModel.OrderDesc(columns.CreatedAt).Limit(start, pageSize).All()
	if err != nil {
		return nil, err
	}

	list := make([]*v1.TaskRecordListResModel, 0, len(records))
	for _, record := range records {
		item, mapErr := taskdata.BuildTaskRecordMap(ctx, record)
		if mapErr != nil {
			return nil, mapErr
		}
		list = append(list, item)
	}

	return &v1.TaskRecordListRes{List: list, Total: total}, nil
}
