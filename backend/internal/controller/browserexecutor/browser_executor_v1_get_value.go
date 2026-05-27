package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorGetValue gets element value. BrowserExecutorGetValue 閼惧嘲褰囬崗鍐閸婄鈧?
func (c *ControllerV1) BrowserExecutorGetValue(ctx context.Context, req *v1.BrowserExecutorGetValueReq) (res *v1.BrowserExecutorGetValueRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.GetValue(ctx, req.Identifier)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorGetValueRes{Result: result}, nil
}
