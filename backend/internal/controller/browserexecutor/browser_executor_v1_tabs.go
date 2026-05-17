package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorTabs manages tabs. BrowserExecutorTabs 缁狅紕鎮婇弽鍥╊劮妞ょ偣鈧?
func (c *ControllerV1) BrowserExecutorTabs(ctx context.Context, req *v1.BrowserExecutorTabsReq) (res *v1.BrowserExecutorTabsRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Tabs(ctx, req.Action, req.URL, req.Index)
	return &v1.BrowserExecutorTabsRes{Result: result}, opErr
}
