package taskrecords

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskRecordDetail returns task record detail 获取任务记录详情
func (c *ControllerV1) TaskRecordDetail(ctx context.Context, req *v1.TaskRecordDetailReq) (res *v1.TaskRecordDetailRes, err error) {
	columns := dao.TaskRecords.Columns()
	record, err := dao.TaskRecords.Ctx(ctx).
		WherePri(gconv.Int64(req.ID)).
		Where(columns.DeletedAt + " IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("执行记录不存在")
	}

	recordMap, err := buildTaskRecordMap(ctx, record)
	if err != nil {
		return nil, err
	}
	return &v1.TaskRecordDetailRes{Record: recordMap}, nil
}
