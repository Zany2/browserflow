// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package automa

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/automa/v1"
)

type IAutomaV1 interface {
	AutomaWorkflowList(ctx context.Context, req *v1.AutomaWorkflowListReq) (res *v1.AutomaWorkflowListRes, err error)
	AutomaWorkflowDetail(ctx context.Context, req *v1.AutomaWorkflowDetailReq) (res *v1.AutomaWorkflowDetailRes, err error)
	AutomaWorkflowCreate(ctx context.Context, req *v1.AutomaWorkflowCreateReq) (res *v1.AutomaWorkflowCreateRes, err error)
	AutomaWorkflowUpdate(ctx context.Context, req *v1.AutomaWorkflowUpdateReq) (res *v1.AutomaWorkflowUpdateRes, err error)
	AutomaWorkflowProtected(ctx context.Context, req *v1.AutomaWorkflowProtectedReq) (res *v1.AutomaWorkflowProtectedRes, err error)
	AutomaWorkflowBatchDelete(ctx context.Context, req *v1.AutomaWorkflowBatchDeleteReq) (res *v1.AutomaWorkflowBatchDeleteRes, err error)
	AutomaWorkflowImportFiles(ctx context.Context, req *v1.AutomaWorkflowImportFilesReq) (res *v1.AutomaWorkflowImportFilesRes, err error)
	AutomaWorkflowSyncCandidates(ctx context.Context, req *v1.AutomaWorkflowSyncCandidatesReq) (res *v1.AutomaWorkflowSyncCandidatesRes, err error)
	AutomaWorkflowSync(ctx context.Context, req *v1.AutomaWorkflowSyncReq) (res *v1.AutomaWorkflowSyncRes, err error)
	AutomaWorkflowCache(ctx context.Context, req *v1.AutomaWorkflowCacheReq) (res *v1.AutomaWorkflowCacheRes, err error)
	AutomaWorkflowRun(ctx context.Context, req *v1.AutomaWorkflowRunReq) (res *v1.AutomaWorkflowRunRes, err error)
	AutomaWorkflowOpen(ctx context.Context, req *v1.AutomaWorkflowOpenReq) (res *v1.AutomaWorkflowOpenRes, err error)
}
