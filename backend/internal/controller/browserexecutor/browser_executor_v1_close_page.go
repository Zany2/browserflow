package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorClosePage closes current page. BrowserExecutorClosePage 鍏抽棴褰撳墠椤甸潰銆?
func (c *ControllerV1) BrowserExecutorClosePage(ctx context.Context, req *v1.BrowserExecutorClosePageReq) (res *v1.BrowserExecutorClosePageRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.ClosePage(ctx)
	return &v1.BrowserExecutorClosePageRes{Result: result}, opErr
}
