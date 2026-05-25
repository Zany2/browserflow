package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/cronexpr"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"
)

// TaskCreate creates task.
func (c *ControllerV1) TaskCreate(ctx context.Context, req *v1.TaskCreateReq) (res *v1.TaskCreateRes, err error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	workflowID := strings.TrimSpace(req.WorkflowID)
	cronExpression := strings.TrimSpace(req.CronExpression)
	if name == "" {
		name = "任务-" + grand.S(8)
	}
	if description == "" {
		description = "任务说明-" + grand.S(8)
	}
	if workflowID == "" {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流不能为空")
		return nil, nil
	}
	if err := cronexpr.Validate(cronExpression); err != nil {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), err.Error())
		return nil, nil
	}

	workflowColumns := dao.AutomaWorkflows.Columns()
	workflowRecord, err := dao.AutomaWorkflows.Ctx(ctx).
		Where(workflowColumns.AutomaId, workflowID).
		One()
	if err != nil {
		return nil, err
	}
	if workflowRecord.IsEmpty() {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流不存在，请先同步或导入工作流")
		return nil, nil
	}

	clientIP, nodeID, err := resolveClientTarget(ctx, req.ClientID, req.ClientIP, req.NodeID)
	if err != nil {
		return nil, err
	}
	if consts.ResolveRuntimeMode(ctx) != consts.RuntimeModeServer && clientIP == "" {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "执行客户端不能为空")
		return nil, nil
	}
	if clientIP != "" || nodeID != "" {
		clientColumns := dao.Clients.Columns()
		clientModel := dao.Clients.Ctx(ctx)
		if clientIP != "" && nodeID != "" {
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
		if clientRecord.IsEmpty() {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "执行客户端不存在，请先确认客户端已接入")
			return nil, nil
		}
	}

	paramsJSON, err := taskdata.EncodeJSONMap(req.Params)
	if err != nil {
		return nil, err
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	taskID, err := dao.Tasks.Ctx(ctx).Data(do.Tasks{
		Name:           name,
		Description:    description,
		AutomaId:       workflowID,
		ClientIp:       clientIP,
		NodeId:         nodeID,
		TargetGroupId:  req.TargetGroupID,
		DispatchMode:   normalizeDispatchMode(req.DispatchMode, nodeID, req.TargetGroupID, clientIP),
		QueuePolicy:    normalizeQueuePolicy(req.QueuePolicy),
		CronExpression: cronExpression,
		ParamsJson:     paramsJSON,
		Enabled:        enabled,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	record, err := dao.Tasks.Ctx(ctx).WherePri(taskID).One()
	if err != nil {
		return nil, err
	}
	task, err := buildTaskMap(ctx, record)
	if err != nil {
		return nil, err
	}

	if req.RunOnceAfterCreate != nil && *req.RunOnceAfterCreate && enabled {
		_, _ = c.TaskExecute(ctx, &v1.TaskExecuteReq{
			ID:          gconv.String(taskID),
			ClientIP:    clientIP,
			NodeID:      nodeID,
			TriggerType: "task_create",
			Params:      req.Params,
		})
	}

	return &v1.TaskCreateRes{Task: task}, nil
}
