package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorEvaluate executes JS. BrowserExecutorEvaluate 閹笛嗩攽 JS閵?
func (c *ControllerV1) BrowserExecutorEvaluate(ctx context.Context, req *v1.BrowserExecutorEvaluateReq) (res *v1.BrowserExecutorEvaluateRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Evaluate(ctx, req.Script)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorEvaluateRes{Result: result}, nil
}
