package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorResize changes viewport size. 调整视口大小。
func (c *ControllerV1) BrowserExecutorResize(ctx context.Context, req *v1.BrowserExecutorResizeReq) (res *v1.BrowserExecutorResizeRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Resize(ctx, req.Width, req.Height)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorResizeRes{Result: result}, nil
}
