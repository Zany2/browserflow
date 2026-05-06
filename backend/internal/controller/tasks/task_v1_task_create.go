package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskCreate creates task 创建任务
func (c *ControllerV1) TaskCreate(ctx context.Context, req *v1.TaskCreateReq) (res *v1.TaskCreateRes, err error) {
	name := strings.TrimSpace(req.Name)
	workflowID := strings.TrimSpace(req.WorkflowID)
	if name == "" {
		return nil, gerror.New("任务名称不能为空")
	}
	if workflowID == "" {
		return nil, gerror.New("工作流不能为空")
	}

	clientIP, err := resolveClientIP(ctx, req.ClientID, req.ClientIP)
	if err != nil {
		return nil, err
	}
	if clientIP == "" {
		return nil, gerror.New("执行客户端不能为空")
	}

	paramsJSON, err := encodeJSONMap(req.Params)
	if err != nil {
		return nil, err
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	columns := dao.Tasks.Columns()
	taskID, err := dao.Tasks.Ctx(ctx).Data(g.Map{
		columns.Name:           name,
		columns.Description:    strings.TrimSpace(req.Description),
		columns.AutomaId:       workflowID,
		columns.ClientIp:       clientIP,
		columns.CronExpression: strings.TrimSpace(req.CronExpression),
		columns.ParamsJson:     paramsJSON,
		columns.Enabled:        enabled,
		columns.CreatedAt:      gtime.Now(),
		columns.UpdatedAt:      gtime.Now(),
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
			TriggerType: "task_create",
			Params:      req.Params,
		})
	}

	return &v1.TaskCreateRes{Task: task}, nil
}
