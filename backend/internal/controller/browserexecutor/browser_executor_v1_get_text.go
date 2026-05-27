package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorGetText gets element text. BrowserExecutorGetText 閼惧嘲褰囬崗鍐閺傚洦婀伴妴?
func (c *ControllerV1) BrowserExecutorGetText(ctx context.Context, req *v1.BrowserExecutorGetTextReq) (res *v1.BrowserExecutorGetTextRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.GetText(ctx, req.Identifier)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorGetTextRes{Result: result}, nil
}
