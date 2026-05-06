package tasks

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskDetail returns task detail 获取任务详情
func (c *ControllerV1) TaskDetail(ctx context.Context, req *v1.TaskDetailReq) (res *v1.TaskDetailRes, err error) {
	columns := dao.Tasks.Columns()
	record, err := dao.Tasks.Ctx(ctx).
		WherePri(gconv.Int64(req.ID)).
		Where(columns.DeletedAt + " IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("任务不存在")
	}

	task, err := buildTaskMap(ctx, record)
	if err != nil {
		return nil, err
	}
	return &v1.TaskDetailRes{Task: task}, nil
}
