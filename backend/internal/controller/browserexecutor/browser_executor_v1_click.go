package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorClick clicks element. 点击元素。
func (c *ControllerV1) BrowserExecutorClick(ctx context.Context, req *v1.BrowserExecutorClickReq) (res *v1.BrowserExecutorClickRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Click(ctx, req.Identifier)
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorClickRes{Result: result}, opErr
}
