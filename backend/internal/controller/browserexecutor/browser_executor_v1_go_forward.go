package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorGoForward navigates forward. BrowserExecutorGoForward 鍓嶈繘銆?
func (c *ControllerV1) BrowserExecutorGoForward(ctx context.Context, req *v1.BrowserExecutorGoForwardReq) (res *v1.BrowserExecutorGoForwardRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.GoForward(ctx)
	return &v1.BrowserExecutorGoForwardRes{Result: result}, opErr
}
