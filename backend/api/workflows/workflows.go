// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package workflows

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
)

type IWorkflowsV1 interface {
	WorkflowList(ctx context.Context, req *v1.WorkflowListReq) (res *v1.WorkflowListRes, err error)
	WorkflowDetail(ctx context.Context, req *v1.WorkflowDetailReq) (res *v1.WorkflowDetailRes, err error)
	WorkflowCreate(ctx context.Context, req *v1.WorkflowCreateReq) (res *v1.WorkflowCreateRes, err error)
	WorkflowUpdate(ctx context.Context, req *v1.WorkflowUpdateReq) (res *v1.WorkflowUpdateRes, err error)
	WorkflowProtected(ctx context.Context, req *v1.WorkflowProtectedReq) (res *v1.WorkflowProtectedRes, err error)
	WorkflowBatchDelete(ctx context.Context, req *v1.WorkflowBatchDeleteReq) (res *v1.WorkflowBatchDeleteRes, err error)
	WorkflowImportFiles(ctx context.Context, req *v1.WorkflowImportFilesReq) (res *v1.WorkflowImportFilesRes, err error)
	WorkflowSyncCandidates(ctx context.Context, req *v1.WorkflowSyncCandidatesReq) (res *v1.WorkflowSyncCandidatesRes, err error)
	WorkflowSync(ctx context.Context, req *v1.WorkflowSyncReq) (res *v1.WorkflowSyncRes, err error)
	WorkflowCache(ctx context.Context, req *v1.WorkflowCacheReq) (res *v1.WorkflowCacheRes, err error)
	WorkflowAgentList(ctx context.Context, req *v1.WorkflowAgentListReq) (res *v1.WorkflowAgentListRes, err error)
	WorkflowAgentExportSkill(ctx context.Context, req *v1.WorkflowAgentExportSkillReq) (res *v1.WorkflowAgentExportSkillRes, err error)
	WorkflowRun(ctx context.Context, req *v1.WorkflowRunReq) (res *v1.WorkflowRunRes, err error)
	WorkflowOpen(ctx context.Context, req *v1.WorkflowOpenReq) (res *v1.WorkflowOpenRes, err error)
}
