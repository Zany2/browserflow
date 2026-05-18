package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskUpdate updates task 更新任务
func (c *ControllerV1) TaskUpdate(ctx context.Context, req *v1.TaskUpdateReq) (res *v1.TaskUpdateRes, err error) {
	name := strings.TrimSpace(req.Name)
	workflowID := strings.TrimSpace(req.WorkflowID)
	if name == "" {
		return nil, gerror.New("任务名称不能为空")
	}
	if workflowID == "" {
		return nil, gerror.New("工作流不能为空")
	}

	taskID := gconv.Int64(req.ID)
	record, err := dao.Tasks.Ctx(ctx).
		WherePri(taskID).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("任务不存在")
	}

	clientIP, err := taskdata.ResolveClientIP(ctx, req.ClientID, req.ClientIP)
	if err != nil {
		return nil, err
	}
	if clientIP == "" {
		return nil, gerror.New("执行客户端不能为空")
	}

	paramsJSON, err := taskdata.EncodeJSONMap(req.Params)
	if err != nil {
		return nil, err
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	_, err = dao.Tasks.Ctx(ctx).
		WherePri(taskID).
		Data(do.Tasks{
			Name:           name,
			Description:    strings.TrimSpace(req.Description),
			AutomaId:       workflowID,
			ClientIp:       clientIP,
			CronExpression: strings.TrimSpace(req.CronExpression),
			ParamsJson:     paramsJSON,
			Enabled:        enabled,
		}).
		Update()
	if err != nil {
		return nil, err
	}

	updated, err := dao.Tasks.Ctx(ctx).WherePri(taskID).One()
	if err != nil {
		return nil, err
	}
	task, err := taskdata.BuildTaskMap(ctx, updated)
	if err != nil {
		return nil, err
	}
	return &v1.TaskUpdateRes{Task: task}, nil
}
