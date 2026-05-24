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

// TaskExecute executes task 执行任务
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
			rr.FailedJsonWithMessageExitAll(request, "任务已停用")
		}
		return nil, nil
	}

	workflowID := strings.TrimSpace(gconv.String(taskRecord[taskColumns.AutomaId]))
	clientIP, err := taskdata.ResolveClientIP(ctx, req.ClientID, req.ClientIP)
	if err != nil {
		return nil, err
	}
	storedClientIP := strings.TrimSpace(gconv.String(taskRecord[taskColumns.ClientIp]))
	if clientIP == "" {
		clientIP = storedClientIP
	}

	params := req.Params
	if params == nil {
		taskMap, mapErr := taskdata.BuildTaskMap(ctx, taskRecord)
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
	autoDispatch := serverMode && storedClientIP == "" && strings.TrimSpace(req.ClientIP) == "" && strings.TrimSpace(req.ClientID) == ""
	failedResponse := func(recordID int64, message string) (*v1.TaskExecuteRes, error) {
		record, recordErr := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
		if recordErr != nil {
			return nil, recordErr
		}
		recordMap, recordErr := taskdata.BuildTaskRecordMap(ctx, record)
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
	failRecord := func(failedRecordID int64, message string) (*v1.TaskExecuteRes, error) {
		_, _ = dao.TaskRecords.Ctx(ctx).
			WherePri(failedRecordID).
			Data(do.TaskRecords{
				Status:       "failed",
				ErrorMessage: message,
				FinishedAt:   gtime.Now(),
			}).
			Update()
		return failedResponse(failedRecordID, message)
	}
	dispatchMessage := ""
	candidateClientIPs := make([]string, 0, 1)
	workflowID = strings.TrimSpace(workflowID)
	clientIP = strings.TrimSpace(clientIP)

	if workflowID == "" {
		dispatchMessage = "工作流不能为空"
	} else if !serverMode {
		if clientIP == "" {
			dispatchMessage = "执行客户端不能为空"
		} else {
			candidateClientIPs = append(candidateClientIPs, clientIP)
		}
	} else if clientIP != "" {
		if !workflowcache.IsClientOnline(ctx, clientIP) {
			dispatchMessage = "客户端不在线或 WebSocket 未连接"
		} else if _, ok, cacheErr := workflowcache.GetClientWorkflow(ctx, clientIP, workflowID); cacheErr != nil {
			dispatchMessage = cacheErr.Error()
		} else if !ok {
			dispatchMessage = "客户端没有该工作流"
		} else {
			candidateClientIPs = append(candidateClientIPs, clientIP)
		}
	} else {
		items, listErr := workflowcache.ListWorkflowClients(ctx, workflowID)
		if listErr != nil {
			dispatchMessage = listErr.Error()
		} else if len(items) == 0 {
			dispatchMessage = "没有在线客户端拥有该工作流"
		} else {
			for _, item := range items {
				itemClientIP := strings.TrimSpace(item.SourceIp)
				if itemClientIP != "" {
					candidateClientIPs = append(candidateClientIPs, itemClientIP)
				}
			}
		}
	}
	if dispatchMessage != "" {
		return createFailedRecord(dispatchMessage)
	}

	recordID, err := dao.TaskRecords.Ctx(ctx).Data(do.TaskRecords{
		TaskId:      taskID,
		WorkflowId:  workflowID,
		ClientIp:    clientIP,
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
	for _, candidateClientIP := range candidateClientIPs {
		candidateClientIP = strings.TrimSpace(candidateClientIP)
		if candidateClientIP == "" {
			continue
		}
		ok, _, lockErr := tasklock.Acquire(ctx, tasklock.LockInfo{
			ClientIP:   candidateClientIP,
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

		clientIP = candidateClientIP
		if _, err = dao.TaskRecords.Ctx(ctx).
			WherePri(recordID).
			Data(do.TaskRecords{ClientIp: clientIP}).
			Update(); err != nil {
			_ = tasklock.Release(ctx, clientIP, commandID)
			return nil, err
		}

		if req.WaitResult {
			resultCh = make(chan model.AgentCommandResult, 1)
			state.SetPendingCommand(commandID, resultCh)
		}
		sentCount := websockets.SendClientMessage(clientIP, &model.WSResponse{
			Type:      model.WSMessageTypeAgentCommand,
			ClientIP:  clientIP,
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
			_ = tasklock.Release(ctx, clientIP, commandID)
			if serverMode {
				_ = workflowcache.ClearClient(ctx, clientIP)
			}
			if autoDispatch {
				continue
			}
			break
		}

		locked = true
		break
	}

	if !locked {
		message := "客户端正在执行其他任务，请稍后重试"
		if autoDispatch && sendFailed {
			message = "已遍历调度所有在线且拥有工作流的客户端，均已断开或 WebSocket 不可发送，任务执行失败"
		} else if autoDispatch {
			message = "已遍历调度所有在线且拥有工作流的客户端，均处于繁忙状态，任务执行失败"
		} else if sendFailed {
			message = "客户端不在线或 WebSocket 未连接"
		}
		return failRecord(recordID, message)
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
			recordMap, err := taskdata.BuildTaskRecordMap(ctx, record)
			if err != nil {
				return nil, err
			}
			return &v1.TaskExecuteRes{Record: recordMap, Result: &result}, nil
		case <-timer.C:
			state.RemovePendingCommand(commandID)
			record, recordErr := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
			if recordErr != nil {
				return nil, recordErr
			}
			recordMap, recordErr := taskdata.BuildTaskRecordMap(ctx, record)
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
	recordMap, err := taskdata.BuildTaskRecordMap(ctx, record)
	if err != nil {
		return nil, err
	}
	return &v1.TaskExecuteRes{Record: recordMap}, nil
}
