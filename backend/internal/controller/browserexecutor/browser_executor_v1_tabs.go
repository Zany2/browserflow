package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorTabs manages tabs. 管理标签页。
func (c *ControllerV1) BrowserExecutorTabs(ctx context.Context, req *v1.BrowserExecutorTabsReq) (res *v1.BrowserExecutorTabsRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Tabs(ctx, req.Action, req.URL, req.Index)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorTabsRes{Result: result}, nil
}
