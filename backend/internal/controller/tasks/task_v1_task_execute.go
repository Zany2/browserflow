package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/errors/gerror"
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
		return nil, gerror.New("任务不存在")
	}
	if !gconv.Bool(taskRecord[taskColumns.Enabled]) {
		return nil, gerror.New("任务已停用")
	}

	workflowID := strings.TrimSpace(gconv.String(taskRecord[taskColumns.AutomaId]))
	clientIP, err := taskdata.ResolveClientIP(ctx, req.ClientID, req.ClientIP)
	if err != nil {
		return nil, err
	}
	if clientIP == "" {
		clientIP = strings.TrimSpace(gconv.String(taskRecord[taskColumns.ClientIp]))
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
	clientIP, err = resolveTaskExecutionClient(ctx, workflowID, clientIP, consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer)
	if err != nil {
		recordMap, recordErr := createFailedTaskRecord(ctx, taskID, workflowID, clientIP, triggerType, paramsJSON, err.Error())
		if recordErr != nil {
			return nil, recordErr
		}
		return &v1.TaskExecuteRes{Record: recordMap}, err
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

	websockets.Init(ctx)
	sentCount := websockets.SendClientMessage(clientIP, &model.WSResponse{
		Type:      model.WSMessageTypeAgentCommand,
		ClientIP:  clientIP,
		CommandID: "task-record-" + gconv.String(recordID),
		Command:   "task.execute",
		Payload: map[string]any{
			"task_id":      taskID,
			"task_name":    strings.TrimSpace(gconv.String(taskRecord[taskColumns.Name])),
			"workflow_id":  workflowID,
			"params":       params,
			"check_params": false,
		},
	})
	if sentCount <= 0 {
		_, _ = dao.TaskRecords.Ctx(ctx).
			WherePri(recordID).
			Data(do.TaskRecords{
				Status:       "failed",
				ErrorMessage: "客户端不在线或 WebSocket 未连接",
				FinishedAt:   gtime.Now(),
			}).
			Update()
		return nil, gerror.New("客户端不在线或 WebSocket 未连接")
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

// resolveTaskExecutionClient picks executable client by workflow cache 解析任务执行客户端
func resolveTaskExecutionClient(ctx context.Context, workflowID string, clientIP string, serverMode bool) (string, error) {
	workflowID = strings.TrimSpace(workflowID)
	clientIP = strings.TrimSpace(clientIP)
	if workflowID == "" {
		return "", gerror.New("工作流不能为空")
	}
	if !serverMode {
		if clientIP == "" {
			return "", gerror.New("执行客户端不能为空")
		}
		return clientIP, nil
	}
	if clientIP != "" {
		if !workflowcache.IsClientOnline(ctx, clientIP) {
			return clientIP, gerror.New("客户端不在线或 WebSocket 未连接")
		}
		if _, ok, err := workflowcache.GetClientWorkflow(ctx, clientIP, workflowID); err != nil {
			return clientIP, err
		} else if !ok {
			return clientIP, gerror.New("客户端没有该工作流")
		}
		return clientIP, nil
	}

	items, err := workflowcache.ListWorkflowClients(ctx, workflowID)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", gerror.New("没有在线客户端拥有该工作流")
	}
	return strings.TrimSpace(items[0].SourceIp), nil
}

// createFailedTaskRecord records dispatch failure 创建失败执行记录
func createFailedTaskRecord(ctx context.Context, taskID int64, workflowID string, clientIP string, triggerType string, paramsJSON string, message string) (*model.TaskRecordResModel, error) {
	recordID, err := dao.TaskRecords.Ctx(ctx).Data(do.TaskRecords{
		TaskId:       taskID,
		WorkflowId:   strings.TrimSpace(workflowID),
		ClientIp:     strings.TrimSpace(clientIP),
		TriggerType:  triggerType,
		Status:       "failed",
		ParamsJson:   paramsJSON,
		ErrorMessage: strings.TrimSpace(message),
		FinishedAt:   gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	record, err := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
	if err != nil {
		return nil, err
	}
	recordMap, err := taskdata.BuildTaskRecordMap(ctx, record)
	if err != nil {
		return nil, err
	}
	return recordMap, nil
}
