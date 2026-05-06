// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package taskrecords

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
)

type ITaskrecordsV1 interface {
	TaskRecordList(ctx context.Context, req *v1.TaskRecordListReq) (res *v1.TaskRecordListRes, err error)
	TaskRecordDetail(ctx context.Context, req *v1.TaskRecordDetailReq) (res *v1.TaskRecordDetailRes, err error)
}
