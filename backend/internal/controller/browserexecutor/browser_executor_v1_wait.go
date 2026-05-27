package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorWait waits for page or element. BrowserExecutorWait 缁涘绶熸い鐢告桨閹存牕鍘撶槐鐘偓?
func (c *ControllerV1) BrowserExecutorWait(ctx context.Context, req *v1.BrowserExecutorWaitReq) (res *v1.BrowserExecutorWaitRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Wait(ctx, req.Identifier, req.State, req.Timeout, req.Count)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorWaitRes{Result: result}, nil
}
