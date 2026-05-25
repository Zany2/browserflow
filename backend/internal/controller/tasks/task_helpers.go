package tasks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/cronexpr"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/Zany2/browserflow/backend/utility/tasklock"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

const taskCronSyncBatchSize = 500

func buildTaskMap(ctx context.Context, record gdb.Record) (*model.TaskResModel, error) {
	if record.IsEmpty() {
		return nil, nil
	}

	columns := dao.Tasks.Columns()
	workflowID := strings.TrimSpace(gconv.String(record[columns.AutomaId]))
	clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
	nodeID := strings.TrimSpace(gconv.String(record[columns.NodeId]))
	params, err := taskdata.DecodeJSONMap(strings.TrimSpace(gconv.String(record[columns.ParamsJson])))
	if err != nil {
		return nil, err
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
		}
	}

	return &model.TaskResModel{
		ID:                        gconv.Int64(record[columns.Id]),
		Name:                      strings.TrimSpace(gconv.String(record[columns.Name])),
		Description:               strings.TrimSpace(gconv.String(record[columns.Description])),
		AutomaID:                  workflowID,
		WorkflowID:                workflowID,
		WorkflowName:              workflowName,
		ClientID:                  clientID,
		ClientName:                clientName,
		ClientIP:                  clientIP,
		NodeID:                    nodeID,
		NodeName:                  nodeID,
		TargetGroupID:             gconv.Int64(record[columns.TargetGroupId]),
		DispatchMode:              strings.TrimSpace(gconv.String(record[columns.DispatchMode])),
		QueuePolicy:               strings.TrimSpace(gconv.String(record[columns.QueuePolicy])),
		MaxAttempts:               gconv.Int(record[columns.MaxAttempts]),
		TimeoutSeconds:            gconv.Int(record[columns.TimeoutSeconds]),
		QueueWaitSeconds:          gconv.Int(record[columns.QueueWaitSeconds]),
		QueueRetryIntervalSeconds: gconv.Int(record[columns.QueueRetryIntervalSeconds]),
		CronExpression:            strings.TrimSpace(gconv.String(record[columns.CronExpression])),
		Params:                    params,
		Enabled:                   gconv.Bool(record[columns.Enabled]),
		CreatedAt:                 taskdata.RecordTime(record[columns.CreatedAt]),
		UpdatedAt:                 taskdata.RecordTime(record[columns.UpdatedAt]),
		DeletedAt:                 taskdata.RecordTime(record[columns.DeletedAt]),
	}, nil
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

func resolveClientTarget(ctx context.Context, clientID string, clientIP string, nodeID string) (string, string, error) {
	clientID = strings.TrimSpace(clientID)
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	columns := dao.Clients.Columns()

	if nodeID != "" {
		clientModel := dao.Clients.Ctx(ctx)
		if clientIP != "" {
			clientModel = clientModel.Where(columns.ClientIp, clientIP).Where(columns.NodeId, nodeID)
		} else {
			clientModel = clientModel.Where(columns.NodeId, nodeID)
		}
		record, err := clientModel.One()
		if err != nil {
			return "", "", err
		}
		if record.IsEmpty() {
			return clientIP, nodeID, nil
		}
		return strings.TrimSpace(gconv.String(record[columns.ClientIp])), strings.TrimSpace(gconv.String(record[columns.NodeId])), nil
	}

	resolvedClientIP, err := resolveClientIP(ctx, clientID, clientIP)
	if err != nil || resolvedClientIP == "" || clientIP != "" {
		return resolvedClientIP, "", err
	}
	record, err := dao.Clients.Ctx(ctx).Where(columns.ClientIp, resolvedClientIP).One()
	if err != nil || record.IsEmpty() {
		return resolvedClientIP, "", err
	}
	return resolvedClientIP, strings.TrimSpace(gconv.String(record[columns.NodeId])), nil
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

// LoadCronTaskMap loads enabled cron task expressions. 加载启用的定时任务表达式
func LoadCronTaskMap(ctx context.Context) (map[string]string, error) {
	columns := dao.Tasks.Columns()
	result := make(map[string]string)
	lastID := int64(0)

	for {
		records, err := dao.Tasks.Ctx(ctx).
			Fields(columns.Id, columns.CronExpression).
			Where(columns.Enabled, true).
			Where(columns.CronExpression+" IS NOT NULL").
			Where(columns.CronExpression+" <> ?", "").
			WhereIn(columns.DispatchMode, []string{"auto", "ip", "node", "group"}).
			WhereIn(columns.QueuePolicy, []string{"queue", "fail", "skip"}).
			WhereGT(columns.Id, lastID).
			OrderAsc(columns.Id).
			Limit(taskCronSyncBatchSize).
			All()
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return result, nil
		}

		for _, record := range records {
			lastID = gconv.Int64(record[columns.Id])
			taskID := gconv.String(record[columns.Id])
			cronExpression := cronexpr.Normalize(gconv.String(record[columns.CronExpression]))
			if taskID == "" || cronExpression == "" {
				continue
			}
			result[TaskCronName(taskID)] = cronExpression
		}
		if len(records) < taskCronSyncBatchSize {
			return result, nil
		}
	}
}

// RecoverActiveTaskState resets task records and node busy flags left by a previous server process.
func RecoverActiveTaskState(ctx context.Context) {
	columns := dao.TaskRecords.Columns()
	if _, err := dao.TaskRecords.Ctx(ctx).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		Data(do.TaskRecords{
			Status:       "failed",
			ErrorMessage: "server restarted before task completion",
			FinishedAt:   gtime.Now(),
		}).
		Update(); err != nil {
		g.Log().Line().Warningf(ctx, "mark active task records failed on startup: %+v", err)
	}

	clientColumns := dao.Clients.Columns()
	if _, err := dao.Clients.Ctx(ctx).
		Data(do.Clients{
			BusyStatus:          "idle",
			CurrentExecutionId:  "",
			CurrentTaskRecordId: 0,
		}).
		Where(clientColumns.BusyStatus, "busy").
		Update(); err != nil {
		g.Log().Line().Warningf(ctx, "reset busy client nodes failed on startup: %+v", err)
	}

	locks, err := tasklock.List(ctx)
	if err != nil {
		g.Log().Line().Warningf(ctx, "scan task locks failed on startup: %+v", err)
		return
	}
	for _, lockInfo := range locks {
		if strings.TrimSpace(lockInfo.CommandID) == "" {
			continue
		}
		if err = tasklock.ReleaseNode(ctx, lockInfo.NodeID, lockInfo.ClientIP, lockInfo.CommandID); err != nil {
			g.Log().Line().Warningf(ctx, "release task lock failed on startup: node_id=%s client_ip=%s command_id=%s err=%+v", lockInfo.NodeID, lockInfo.ClientIP, lockInfo.CommandID, err)
		}
	}
}

// SweepStaleTaskRecords marks stale active task records as failed. 清理超时卡住的执行记录
func SweepStaleTaskRecords(ctx context.Context) {
	columns := dao.TaskRecords.Columns()
	cutoff := gtime.New(time.Now().Add(-tasklock.StaleAfter))
	records, err := dao.TaskRecords.Ctx(ctx).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		Where("COALESCE("+columns.StartedAt+", "+columns.CreatedAt+") < ?", cutoff).
		Limit(100).
		All()
	if err != nil {
		g.Log().Line().Warningf(ctx, "scan stale task records failed: %+v", err)
		return
	}

	for _, record := range records {
		recordID := gconv.Int64(record[columns.Id])
		clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
		nodeID := strings.TrimSpace(gconv.String(record[columns.NodeId]))
		if recordID <= 0 {
			continue
		}

		commandID := "task-record-" + gconv.String(recordID)
		lockInfo, hasLock, lockErr := tasklock.GetNode(ctx, nodeID, clientIP)
		if lockErr != nil {
			g.Log().Line().Warningf(ctx, "read client task lock failed: record_id=%d node_id=%s client_ip=%s err=%+v", recordID, nodeID, clientIP, lockErr)
			continue
		}
		if hasLock && lockInfo.CommandID == commandID {
			continue
		}

		_, err = dao.TaskRecords.Ctx(ctx).
			WherePri(recordID).
			WhereIn(columns.Status, []string{"pending", "queued", "running"}).
			Data(do.TaskRecords{
				Status:       "failed",
				ErrorMessage: "client task execution timed out and was automatically ended",
				FinishedAt:   gtime.Now(),
			}).
			Update()
		if err != nil {
			g.Log().Line().Warningf(ctx, "mark stale task record failed: record_id=%d err=%+v", recordID, err)
			continue
		}
		if err = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID); err != nil {
			g.Log().Line().Warningf(ctx, "release stale task lock failed: record_id=%d node_id=%s client_ip=%s err=%+v", recordID, nodeID, clientIP, err)
		}
		if err = markTaskNodeIdle(ctx, clientIP, nodeID, commandID, recordID); err != nil {
			g.Log().Line().Warningf(ctx, "reset stale task node status failed: record_id=%d node_id=%s client_ip=%s err=%+v", recordID, nodeID, clientIP, err)
		}
	}
}

func markTaskNodeBusy(ctx context.Context, clientIP string, nodeID string, commandID string, recordID int64) error {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" && nodeID == "" {
		return nil
	}
	columns := dao.Clients.Columns()
	updateModel := dao.Clients.Ctx(ctx)
	if clientIP != "" && nodeID != "" {
		updateModel = updateModel.Where(columns.ClientIp, clientIP).Where(columns.NodeId, nodeID)
	} else if nodeID != "" {
		updateModel = updateModel.Where(columns.NodeId, nodeID)
	} else {
		updateModel = updateModel.Where(columns.ClientIp, clientIP)
	}
	_, err := updateModel.Data(do.Clients{
		BusyStatus:          "busy",
		CurrentExecutionId:  strings.TrimSpace(commandID),
		CurrentTaskRecordId: recordID,
		LastCommandAt:       gtime.Now(),
		LastLockRenewedAt:   gtime.Now(),
	}).Update()
	return err
}

func markTaskNodeIdle(ctx context.Context, clientIP string, nodeID string, commandID string, recordID int64) error {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" && nodeID == "" {
		return nil
	}
	columns := dao.Clients.Columns()
	updateModel := dao.Clients.Ctx(ctx)
	if clientIP != "" && nodeID != "" {
		updateModel = updateModel.Where(columns.ClientIp, clientIP).Where(columns.NodeId, nodeID)
	} else if nodeID != "" {
		updateModel = updateModel.Where(columns.NodeId, nodeID)
	} else {
		updateModel = updateModel.Where(columns.ClientIp, clientIP)
	}
	if commandID != "" {
		updateModel = updateModel.Where(columns.CurrentExecutionId, commandID)
	} else if recordID > 0 {
		updateModel = updateModel.Where(columns.CurrentTaskRecordId, recordID)
	}
	_, err := updateModel.Data(do.Clients{
		BusyStatus:          "idle",
		CurrentExecutionId:  "",
		CurrentTaskRecordId: 0,
	}).Update()
	return err
}

// TaskCronName builds cron job name. 构建定时任务名称
func TaskCronName(taskID string) string {
	return "task-cron-" + taskID
}
