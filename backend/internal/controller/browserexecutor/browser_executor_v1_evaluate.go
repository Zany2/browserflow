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
		return nil, err
	}
	result, opErr := executor.Evaluate(ctx, req.Script)
	return &v1.BrowserExecutorEvaluateRes{Result: result}, opErr
}
