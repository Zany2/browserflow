package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorType types text. 输入文本。
func (c *ControllerV1) BrowserExecutorType(ctx context.Context, req *v1.BrowserExecutorTypeReq) (res *v1.BrowserExecutorTypeRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	clear := true
	if req.Clear != nil {
		clear = *req.Clear
	}
	result, opErr := executor.Type(ctx, req.Identifier, req.Text, clear)
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorTypeRes{Result: result}, opErr
}
