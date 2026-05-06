// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package tasks

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
)

type ITasksV1 interface {
	TaskList(ctx context.Context, req *v1.TaskListReq) (res *v1.TaskListRes, err error)
	TaskDetail(ctx context.Context, req *v1.TaskDetailReq) (res *v1.TaskDetailRes, err error)
	TaskCreate(ctx context.Context, req *v1.TaskCreateReq) (res *v1.TaskCreateRes, err error)
	TaskUpdate(ctx context.Context, req *v1.TaskUpdateReq) (res *v1.TaskUpdateRes, err error)
	TaskDelete(ctx context.Context, req *v1.TaskDeleteReq) (res *v1.TaskDeleteRes, err error)
	TaskExecute(ctx context.Context, req *v1.TaskExecuteReq) (res *v1.TaskExecuteRes, err error)
}
