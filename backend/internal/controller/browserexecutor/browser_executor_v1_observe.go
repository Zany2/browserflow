package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorObserve returns compact page context. BrowserExecutorObserve 鏉╂柨娲栫槐褍鍣炬い鐢告桨娑撳﹣绗呴弬鍥モ偓?
func (c *ControllerV1) BrowserExecutorObserve(ctx context.Context, req *v1.BrowserExecutorObserveReq) (res *v1.BrowserExecutorObserveRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Observe(ctx, req.IncludeText, req.TextLimit)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorObserveRes{Result: result}, nil
}
