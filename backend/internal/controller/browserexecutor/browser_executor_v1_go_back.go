package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorGoBack navigates backward. BrowserExecutorGoBack 鍚庨€€銆?
func (c *ControllerV1) BrowserExecutorGoBack(ctx context.Context, req *v1.BrowserExecutorGoBackReq) (res *v1.BrowserExecutorGoBackRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.GoBack(ctx)
	return &v1.BrowserExecutorGoBackRes{Result: result}, opErr
}
