package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/errors/gerror"
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
		return nil, gerror.New("任务不存在")
	}
	if !gconv.Bool(taskRecord[taskColumns.Enabled]) {
		return nil, gerror.New("任务已停用")
	}

	clientIP, err := taskdata.ResolveClientIP(ctx, req.ClientID, req.ClientIP)
	if err != nil {
		return nil, err
	}
	if clientIP == "" {
		clientIP = strings.TrimSpace(gconv.String(taskRecord[taskColumns.ClientIp]))
	}
	if clientIP == "" {
		return nil, gerror.New("执行客户端不能为空")
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
	recordID, err := dao.TaskRecords.Ctx(ctx).Data(do.TaskRecords{
		TaskId:      taskID,
		WorkflowId:  strings.TrimSpace(gconv.String(taskRecord[taskColumns.AutomaId])),
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
		Payload: g.Map{
			"task_id":      taskID,
			"task_name":    strings.TrimSpace(gconv.String(taskRecord[taskColumns.Name])),
			"workflow_id":  strings.TrimSpace(gconv.String(taskRecord[taskColumns.AutomaId])),
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
