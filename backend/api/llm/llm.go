// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package llm

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/llm/v1"
)

type ILlmV1 interface {
	LlmConfigList(ctx context.Context, req *v1.LlmConfigListReq) (res *v1.LlmConfigListRes, err error)
	LlmConfigDetail(ctx context.Context, req *v1.LlmConfigDetailReq) (res *v1.LlmConfigDetailRes, err error)
	LlmConfigCreate(ctx context.Context, req *v1.LlmConfigCreateReq) (res *v1.LlmConfigCreateRes, err error)
	LlmConfigUpdate(ctx context.Context, req *v1.LlmConfigUpdateReq) (res *v1.LlmConfigUpdateRes, err error)
	LlmConfigDelete(ctx context.Context, req *v1.LlmConfigDeleteReq) (res *v1.LlmConfigDeleteRes, err error)
	LlmConfigDeleteMany(ctx context.Context, req *v1.LlmConfigDeleteManyReq) (res *v1.LlmConfigDeleteManyRes, err error)
	LlmConfigCheck(ctx context.Context, req *v1.LlmConfigCheckReq) (res *v1.LlmConfigCheckRes, err error)
	LlmProviderList(ctx context.Context, req *v1.LlmProviderListReq) (res *v1.LlmProviderListRes, err error)
}
