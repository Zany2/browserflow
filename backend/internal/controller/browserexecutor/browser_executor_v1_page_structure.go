package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorPageStructure gets compact page structure. BrowserExecutorPageStructure 获取页面结构。
func (c *ControllerV1) BrowserExecutorPageStructure(ctx context.Context, req *v1.BrowserExecutorPageStructureReq) (res *v1.BrowserExecutorPageStructureRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.PageStructure(ctx, model.BrowserExecutorPageStructureOptions{
		IncludeLinks:   req.IncludeLinks,
		IncludeForms:   req.IncludeForms,
		IncludeTables:  req.IncludeTables,
		IncludeImages:  req.IncludeImages,
		IncludeButtons: req.IncludeButtons,
		Limit:          req.Limit,
	})
	return &v1.BrowserExecutorPageStructureRes{Result: result}, opErr
}
