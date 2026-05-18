package tasks

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskDelete deletes task 删除任务
func (c *ControllerV1) TaskDelete(ctx context.Context, req *v1.TaskDeleteReq) (res *v1.TaskDeleteRes, err error) {
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

	_, err = dao.Tasks.Ctx(ctx).
		WherePri(taskID).
		Data(do.Tasks{Enabled: false}).
		Update()
	if err != nil {
		return nil, err
	}
	if _, err = dao.Tasks.Ctx(ctx).WherePri(taskID).Delete(); err != nil {
		return nil, err
	}

	return &v1.TaskDeleteRes{}, nil
}
