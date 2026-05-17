package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorSelect selects a dropdown option. 选择下拉选项。
func (c *ControllerV1) BrowserExecutorSelect(ctx context.Context, req *v1.BrowserExecutorSelectReq) (res *v1.BrowserExecutorSelectRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Select(ctx, req.Identifier, req.Value)
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorSelectRes{Result: result}, opErr
}
