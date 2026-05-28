package taskrecords

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/gogf/gf/v2/database/gdb"
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
		taskIDs, taskErr := findTaskIDsByName(ctx, taskName)
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
		workflowIDs, workflowErr := findWorkflowIDsByName(ctx, workflowName)
		if workflowErr != nil {
			return nil, workflowErr
		}
		if len(workflowIDs) == 0 {
			return &v1.TaskRecordListRes{List: []*v1.TaskRecordListResModel{}, Total: 0}, nil
		}
		gModel = gModel.WhereIn(columns.WorkflowId, workflowIDs)
	}
	if clientIPs := splitListFilter(req.ClientIPs); len(clientIPs) > 0 {
		gModel = gModel.WhereIn(columns.ClientIp, clientIPs)
	} else if clientIP, resolveErr := resolveClientIP(ctx, req.ClientID, req.ClientIP); resolveErr != nil {
		return nil, resolveErr
	} else if clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	} else if clientID := strings.TrimSpace(req.ClientID); clientID != "" {
		gModel = gModel.Where(columns.ClientIp, clientID)
	}
	if nodeIDs := splitListFilter(req.NodeIDs); len(nodeIDs) > 0 {
		gModel = gModel.WhereIn(columns.NodeId, nodeIDs)
	} else if nodeID := strings.TrimSpace(req.NodeID); nodeID != "" {
		gModel = gModel.Where(columns.NodeId, nodeID)
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

	records, err := gModel.OrderDesc(columns.CreatedAt).OrderDesc(columns.Id).Limit(start, pageSize).All()
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

	return &v1.TaskRecordListRes{List: list, Total: total}, nil
}

func splitListFilter(value string) []string {
	parts := strings.Split(value, ",")
	list := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		list = append(list, item)
	}
	return list
}

func buildTaskRecordMap(ctx context.Context, record gdb.Record) (*model.TaskRecordResModel, error) {
	if record.IsEmpty() {
		return nil, nil
	}

	columns := dao.TaskRecords.Columns()
	taskID := gconv.Int64(record[columns.TaskId])
	workflowID := strings.TrimSpace(gconv.String(record[columns.WorkflowId]))
	clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
	nodeID := strings.TrimSpace(gconv.String(record[columns.NodeId]))
	params, err := taskdata.DecodeJSONMap(strings.TrimSpace(gconv.String(record[columns.ParamsJson])))
	if err != nil {
		return nil, err
	}
	result := model.JSONMap{}
	if resultJSON := strings.TrimSpace(gconv.String(record[columns.ResultJson])); resultJSON != "" {
		if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
			rawResult, _ := json.Marshal(resultJSON)
			result = model.JSONMap{"raw": rawResult}
		}
	}

	taskName := ""
	if taskID > 0 {
		taskColumns := dao.Tasks.Columns()
		taskRecord, err := dao.Tasks.Ctx(ctx).WherePri(taskID).One()
		if err != nil {
			return nil, err
		}
		if !taskRecord.IsEmpty() {
			taskName = strings.TrimSpace(gconv.String(taskRecord[taskColumns.Name]))
		}
	}

	workflowName := ""
	if workflowID != "" {
		workflowColumns := dao.AutomaWorkflows.Columns()
		workflowRecord, err := dao.AutomaWorkflows.Ctx(ctx).Where(workflowColumns.AutomaId, workflowID).One()
		if err != nil {
			return nil, err
		}
		if !workflowRecord.IsEmpty() {
			workflowName = strings.TrimSpace(gconv.String(workflowRecord[workflowColumns.Name]))
		}
	}

	clientID := ""
	clientName := ""
	if nodeID != "" || clientIP != "" {
		clientColumns := dao.Clients.Columns()
		clientModel := dao.Clients.Ctx(ctx)
		if nodeID != "" && clientIP != "" {
			clientModel = clientModel.Where(clientColumns.ClientIp, clientIP).Where(clientColumns.NodeId, nodeID)
		} else if nodeID != "" {
			clientModel = clientModel.Where(clientColumns.NodeId, nodeID)
		} else {
			clientModel = clientModel.Where(clientColumns.ClientIp, clientIP)
		}
		clientRecord, err := clientModel.One()
		if err != nil {
			return nil, err
		}
		if !clientRecord.IsEmpty() {
			clientID = strings.TrimSpace(gconv.String(clientRecord[clientColumns.Id]))
			clientName = strings.TrimSpace(gconv.String(clientRecord[clientColumns.ClientIp]))
			if nodeID == "" {
				nodeID = strings.TrimSpace(gconv.String(clientRecord[clientColumns.NodeId]))
			}
		}
	}

	startedAt := taskdata.RecordTime(record[columns.StartedAt])
	finishedAt := taskdata.RecordTime(record[columns.FinishedAt])
	durationMs := int64(0)
	if startedAt != nil && finishedAt != nil {
		durationMs = finishedAt.Time.Sub(startedAt.Time).Milliseconds()
	}

	return &model.TaskRecordResModel{
		ID:                gconv.Int64(record[columns.Id]),
		TaskID:            taskID,
		TaskName:          taskName,
		WorkflowID:        workflowID,
		WorkflowName:      workflowName,
		ClientID:          clientID,
		ClientName:        clientName,
		ClientIP:          clientIP,
		NodeID:            nodeID,
		NodeName:          nodeID,
		ExecutionID:       strings.TrimSpace(gconv.String(record[columns.ExecutionId])),
		AutomaExecutionID: strings.TrimSpace(gconv.String(record[columns.AutomaExecutionId])),
		TriggerType:       strings.TrimSpace(gconv.String(record[columns.TriggerType])),
		Status:            strings.TrimSpace(gconv.String(record[columns.Status])),
		Params:            params,
		Result:            result,
		ErrorMessage:      strings.TrimSpace(gconv.String(record[columns.ErrorMessage])),
		DurationMs:        durationMs,
		StartedAt:         startedAt,
		FinishedAt:        finishedAt,
		CreatedAt:         taskdata.RecordTime(record[columns.CreatedAt]),
		UpdatedAt:         taskdata.RecordTime(record[columns.UpdatedAt]),
		DeletedAt:         taskdata.RecordTime(record[columns.DeletedAt]),
	}, nil
}

func resolveClientIP(ctx context.Context, clientID string, clientIP string) (string, error) {
	clientID = strings.TrimSpace(clientID)
	clientIP = strings.TrimSpace(clientIP)
	if clientIP != "" {
		return clientIP, nil
	}
	if clientID == "" {
		return "", nil
	}

	columns := dao.Clients.Columns()
	record, err := dao.Clients.Ctx(ctx).WherePri(gconv.Int64(clientID)).One()
	if err != nil || record.IsEmpty() {
		return "", err
	}
	return strings.TrimSpace(gconv.String(record[columns.ClientIp])), nil
}

func findWorkflowIDsByName(ctx context.Context, workflowName string) ([]string, error) {
	columns := dao.AutomaWorkflows.Columns()
	records, err := dao.AutomaWorkflows.Ctx(ctx).
		Fields(columns.AutomaId).
		Where(columns.Name+" LIKE ?", "%"+workflowName+"%").
		All()
	if err != nil {
		return nil, err
	}

	workflowIDs := make([]string, 0, len(records))
	for _, record := range records {
		workflowID := strings.TrimSpace(gconv.String(record[columns.AutomaId]))
		if workflowID != "" {
			workflowIDs = append(workflowIDs, workflowID)
		}
	}
	return workflowIDs, nil
}

func findTaskIDsByName(ctx context.Context, taskName string) ([]int64, error) {
	columns := dao.Tasks.Columns()
	records, err := dao.Tasks.Ctx(ctx).
		Fields(columns.Id).
		Where(columns.Name+" LIKE ?", "%"+taskName+"%").
		All()
	if err != nil {
		return nil, err
	}

	taskIDs := make([]int64, 0, len(records))
	for _, record := range records {
		taskID := gconv.Int64(record[columns.Id])
		if taskID > 0 {
			taskIDs = append(taskIDs, taskID)
		}
	}
	return taskIDs, nil
}
