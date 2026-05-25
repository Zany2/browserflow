package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/Zany2/browserflow/backend/utility/tasklock"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskExecute executes task 鎵ц浠诲姟
func (c *ControllerV1) TaskExecute(ctx context.Context, req *v1.TaskExecuteReq) (res *v1.TaskExecuteRes, err error) {
	taskID := gconv.Int64(req.ID)
	taskColumns := dao.Tasks.Columns()
	taskRecord, err := dao.Tasks.Ctx(ctx).
		WherePri(taskID).
		One()
	if err != nil {
		return nil, err
	}
	if taskRecord.IsEmpty() {
		request := g.RequestFromCtx(ctx)
		if request != nil {
			rr.FailedJsonWithMessageExitAll(request, "任务不存在")
		}
		return nil, nil
	}
	if !gconv.Bool(taskRecord[taskColumns.Enabled]) {
		request := g.RequestFromCtx(ctx)
		if request != nil {
			rr.FailedJsonWithMessageExitAll(request, "任务不存在")
		}
		return nil, nil
	}

	workflowID := strings.TrimSpace(gconv.String(taskRecord[taskColumns.AutomaId]))
	clientIP, nodeID, err := resolveClientTarget(ctx, req.ClientID, req.ClientIP, req.NodeID)
	if err != nil {
		return nil, err
	}
	storedClientIP := strings.TrimSpace(gconv.String(taskRecord[taskColumns.ClientIp]))
	storedNodeID := strings.TrimSpace(gconv.String(taskRecord[taskColumns.NodeId]))
	dispatchMode := strings.TrimSpace(gconv.String(taskRecord[taskColumns.DispatchMode]))
	queuePolicy := normalizeQueuePolicy(gconv.String(taskRecord[taskColumns.QueuePolicy]))
	targetGroupID := gconv.Int64(taskRecord[taskColumns.TargetGroupId])
	maxAttempts := normalizeMaxAttempts(gconv.Int(taskRecord[taskColumns.MaxAttempts]))
	queueWaitSeconds := normalizeQueueWaitSeconds(gconv.Int(taskRecord[taskColumns.QueueWaitSeconds]))
	queueRetryIntervalSeconds := normalizeQueueRetryIntervalSeconds(gconv.Int(taskRecord[taskColumns.QueueRetryIntervalSeconds]))
	if clientIP == "" {
		clientIP = storedClientIP
	}
	if nodeID == "" {
		nodeID = storedNodeID
	}

	params := req.Params
	if params == nil {
		taskMap, mapErr := buildTaskMap(ctx, taskRecord)
		if mapErr != nil {
			return nil, mapErr
		}
		params = taskMap.Params
	}
	paramsJSON, err := taskdata.EncodeJSONMap(params)
	if err != nil {
		return nil, err
	}

	triggerType := taskdata.NormalizeTriggerType(req.TriggerType)
	serverMode := consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer
	failedResponse := func(recordID int64, message string) (*v1.TaskExecuteRes, error) {
		record, recordErr := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
		if recordErr != nil {
			return nil, recordErr
		}
		recordMap, recordErr := buildTaskRecordMap(ctx, record)
		if recordErr != nil {
			return nil, recordErr
		}
		if request := g.RequestFromCtx(ctx); request != nil {
			rr.FailedJsonWithMessageAndDataExitAll(request, message, &v1.TaskExecuteRes{Record: recordMap})
			return nil, nil
		}
		return &v1.TaskExecuteRes{Record: recordMap}, nil
	}
	createFailedRecord := func(message string) (*v1.TaskExecuteRes, error) {
		failedRecordID, recordErr := dao.TaskRecords.Ctx(ctx).Data(do.TaskRecords{
			TaskId:       taskID,
			WorkflowId:   workflowID,
			ClientIp:     clientIP,
			NodeId:       nodeID,
			TriggerType:  triggerType,
			Status:       "failed",
			ParamsJson:   paramsJSON,
			ErrorMessage: message,
			FinishedAt:   gtime.Now(),
		}).InsertAndGetId()
		if recordErr != nil {
			return nil, recordErr
		}
		failedCommandID := "task-record-" + gconv.String(failedRecordID)
		if _, recordErr = dao.TaskRecords.Ctx(ctx).
			WherePri(failedRecordID).
			Data(do.TaskRecords{ExecutionId: failedCommandID}).
			Update(); recordErr != nil {
			return nil, recordErr
		}
		return failedResponse(failedRecordID, message)
	}
	failRecord := func(failedRecordID int64, status string, message string) (*v1.TaskExecuteRes, error) {
		_, _ = dao.TaskRecords.Ctx(ctx).
			WherePri(failedRecordID).
			Data(do.TaskRecords{
				Status:       status,
				ErrorMessage: message,
				FinishedAt:   gtime.Now(),
			}).
			Update()
		return failedResponse(failedRecordID, message)
	}
	dispatchMessage := ""
	candidateTargets := make([]dispatchTarget, 0, 1)
	workflowID = strings.TrimSpace(workflowID)
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)

	if workflowID == "" {
		dispatchMessage = "工作流不能为空"
	} else if !serverMode {
		if clientIP == "" {
			dispatchMessage = "执行客户端不能为空"
		} else {
			candidateTargets = append(candidateTargets, dispatchTarget{ClientIP: clientIP, NodeID: nodeID})
		}
	} else {
		var targetErr error
		candidateTargets, dispatchMessage, targetErr = buildDispatchTargets(ctx, workflowID, dispatchMode, clientIP, nodeID, targetGroupID)
		if targetErr != nil {
			return nil, targetErr
		}
	}
	if dispatchMessage != "" {
		return createFailedRecord(dispatchMessage)
	}

	recordID, err := dao.TaskRecords.Ctx(ctx).Data(do.TaskRecords{
		TaskId:      taskID,
		WorkflowId:  workflowID,
		ClientIp:    clientIP,
		NodeId:      nodeID,
		TriggerType: triggerType,
		Status:      "pending",
		ParamsJson:  paramsJSON,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}

	commandID := "task-record-" + gconv.String(recordID)
	if _, err = dao.TaskRecords.Ctx(ctx).
		WherePri(recordID).
		Data(do.TaskRecords{ExecutionId: commandID}).
		Update(); err != nil {
		return nil, err
	}
	websockets.Init(ctx)
	timeout := workflowagent.NormalizeRunTimeout(req.Timeout)
	if req.Timeout <= 0 {
		timeout = normalizeTimeoutSeconds(gconv.Int(taskRecord[taskColumns.TimeoutSeconds]))
	}
	returnData := workflowexecution.NormalizeReturnData(req.ReturnData)
	if returnData == nil {
		returnData = &model.WorkflowExecutionReturnData{
			IncludeTable: true,
			TableLimit:   100,
		}
	}

	var resultCh chan model.AgentCommandResult
	locked := false
	sendFailed := false
	attemptCount := 0
	queueDeadline := time.Now()
	if queuePolicy == "queue" {
		queueDeadline = queueDeadline.Add(time.Duration(queueWaitSeconds) * time.Second)
	}
	for {
		attemptCount++
		for _, target := range candidateTargets {
			target.ClientIP = strings.TrimSpace(target.ClientIP)
			target.NodeID = strings.TrimSpace(target.NodeID)
			if target.ClientIP == "" && target.NodeID == "" {
				continue
			}
			ok, _, lockErr := tasklock.Acquire(ctx, tasklock.LockInfo{
				ClientIP:   target.ClientIP,
				NodeID:     target.NodeID,
				TaskID:     taskID,
				RecordID:   recordID,
				WorkflowID: workflowID,
				CommandID:  commandID,
			})
			if lockErr != nil {
				return nil, lockErr
			}
			if !ok {
				continue
			}

			clientIP = target.ClientIP
			nodeID = target.NodeID
			if _, err = dao.TaskRecords.Ctx(ctx).
				WherePri(recordID).
				Data(do.TaskRecords{
					ClientIp:  clientIP,
					NodeId:    nodeID,
					CommandId: commandID,
				}).
				Update(); err != nil {
				_ = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID)
				return nil, err
			}

			if req.WaitResult {
				resultCh = make(chan model.AgentCommandResult, 1)
				state.SetPendingCommand(commandID, resultCh)
			}
			if err = markTaskNodeBusy(ctx, clientIP, nodeID, commandID, recordID); err != nil {
				state.RemovePendingCommand(commandID)
				_ = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID)
				return nil, err
			}
			sentCount := websockets.SendNodeMessage(nodeID, clientIP, &model.WSResponse{
				Type:      model.WSMessageTypeAgentCommand,
				ClientIP:  clientIP,
				NodeID:    nodeID,
				CommandID: commandID,
				Command:   "task.execute",
				Payload: map[string]any{
					"task_id":      taskID,
					"task_name":    strings.TrimSpace(gconv.String(taskRecord[taskColumns.Name])),
					"workflow_id":  workflowID,
					"params":       params,
					"check_params": false,
					"execution_id": commandID,
					"wait_result":  req.WaitResult,
					"timeout":      timeout,
					"return_data":  returnData,
				},
			})
			if sentCount <= 0 {
				sendFailed = true
				state.RemovePendingCommand(commandID)
				_ = markTaskNodeIdle(ctx, clientIP, nodeID, commandID, recordID)
				_ = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID)
				if serverMode {
					_ = workflowcache.ClearNode(ctx, websockets.NodeConnectionID(clientIP, nodeID))
				}
				if len(candidateTargets) > 1 {
					continue
				}
				break
			}

			locked = true
			break
		}
		if locked || queuePolicy != "queue" || attemptCount >= maxAttempts || !time.Now().Before(queueDeadline) {
			break
		}

		sleepDuration := time.Duration(queueRetryIntervalSeconds) * time.Second
		remainingDuration := time.Until(queueDeadline)
		if remainingDuration < sleepDuration {
			sleepDuration = remainingDuration
		}
		if sleepDuration <= 0 {
			break
		}
		if _, err = dao.TaskRecords.Ctx(ctx).
			WherePri(recordID).
			Where(dao.TaskRecords.Columns().Status, "pending").
			Data(do.TaskRecords{Status: "queued"}).
			Update(); err != nil {
			return nil, err
		}
		select {
		case <-time.After(sleepDuration):
		case <-ctx.Done():
			state.RemovePendingCommand(commandID)
			return nil, ctx.Err()
		}
	}

	if !locked {
		status, message := busyPolicyMessage(queuePolicy)
		if len(candidateTargets) > 1 && sendFailed {
			message = "已遍历调度所有在线且拥有工作流的客户端，均已断开或 WebSocket 不可发送，任务执行失败"
		} else if sendFailed {
			message = "客户端不在线或 WebSocket 未连接"
		}
		return failRecord(recordID, status, message)
	}

	_, err = dao.TaskRecords.Ctx(ctx).
		WherePri(recordID).
		Data(do.TaskRecords{
			Status:    "queued",
			StartedAt: gtime.Now(),
		}).
		Update()
	if err != nil {
		return nil, err
	}

	if req.WaitResult {
		timer := time.NewTimer(time.Duration(timeout) * time.Second)
		defer timer.Stop()

		select {
		case result := <-resultCh:
			record, err := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
			if err != nil {
				return nil, err
			}
			recordMap, err := buildTaskRecordMap(ctx, record)
			if err != nil {
				return nil, err
			}
			return &v1.TaskExecuteRes{Record: recordMap, Result: &result}, nil
		case <-timer.C:
			state.RemovePendingCommand(commandID)
			_ = markTaskNodeIdle(ctx, clientIP, nodeID, commandID, recordID)
			_ = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID)
			_, _ = dao.TaskRecords.Ctx(ctx).
				WherePri(recordID).
				WhereIn(dao.TaskRecords.Columns().Status, []string{"pending", "queued", "running"}).
				Data(do.TaskRecords{
					Status:       "timeout",
					ErrorMessage: "task execution timed out while waiting for result",
					FinishedAt:   gtime.Now(),
				}).
				Update()
			record, recordErr := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
			if recordErr != nil {
				return nil, recordErr
			}
			recordMap, recordErr := buildTaskRecordMap(ctx, record)
			if recordErr != nil {
				return nil, recordErr
			}
			return &v1.TaskExecuteRes{Record: recordMap}, nil
		case <-ctx.Done():
			state.RemovePendingCommand(commandID)
			return nil, ctx.Err()
		}
	}

	record, err := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
	if err != nil {
		return nil, err
	}
	recordMap, err := buildTaskRecordMap(ctx, record)
	if err != nil {
		return nil, err
	}
	return &v1.TaskExecuteRes{Record: recordMap}, nil
}
