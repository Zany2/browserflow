package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorInputElements returns compact input element refs. BrowserExecutorInputElements 返回输入元素引用。
func (c *ControllerV1) BrowserExecutorInputElements(ctx context.Context, req *v1.BrowserExecutorInputElementsReq) (res *v1.BrowserExecutorInputElementsRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.ElementRefs(ctx, "input", req.Limit)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorInputElementsRes{Result: result}, nil
}
