package taskrecords

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskRecordBatchDelete deletes selected task records.
func (c *ControllerV1) TaskRecordBatchDelete(ctx context.Context, req *v1.TaskRecordBatchDeleteReq) (res *v1.TaskRecordBatchDeleteRes, err error) {
	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "请选择需要删除的任务记录")
		return nil, nil
	}

	columns := dao.TaskRecords.Columns()
	if _, err = dao.TaskRecords.Ctx(ctx).WhereIn(columns.Id, ids).Delete(); err != nil {
		return nil, err
	}

	fileColumns := dao.TaskRecordFiles.Columns()
	now := gtime.Now()
	if _, err = dao.TaskRecordFiles.Ctx(ctx).
		WhereIn(fileColumns.RecordId, ids).
		Where(fileColumns.DeletedAt + " IS NULL").
		Data(do.TaskRecordFiles{
			UpdatedAt: now,
			DeletedAt: now,
		}).
		Update(); err != nil {
		return nil, err
	}

	return &v1.TaskRecordBatchDeleteRes{}, nil
}
