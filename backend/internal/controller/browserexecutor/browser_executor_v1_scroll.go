package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorScroll scrolls page or element. 滚动页面或元素。
func (c *ControllerV1) BrowserExecutorScroll(ctx context.Context, req *v1.BrowserExecutorScrollReq) (res *v1.BrowserExecutorScrollRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Scroll(ctx, req.Direction, req.Pixels, req.Identifier)
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorScrollRes{Result: result}, opErr
}
