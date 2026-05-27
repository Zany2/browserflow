package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorReload reloads current page. BrowserExecutorReload 鍒锋柊褰撳墠椤甸潰銆?
func (c *ControllerV1) BrowserExecutorReload(ctx context.Context, req *v1.BrowserExecutorReloadReq) (res *v1.BrowserExecutorReloadRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Reload(ctx)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorReloadRes{Result: result}, nil
}
