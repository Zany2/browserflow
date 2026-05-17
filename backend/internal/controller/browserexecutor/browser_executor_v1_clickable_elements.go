package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorClickableElements returns compact clickable element refs. BrowserExecutorClickableElements 返回可点击元素引用。
func (c *ControllerV1) BrowserExecutorClickableElements(ctx context.Context, req *v1.BrowserExecutorClickableElementsReq) (res *v1.BrowserExecutorClickableElementsRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.ElementRefs(ctx, "clickable", req.Limit)
	return &v1.BrowserExecutorClickableElementsRes{Result: result}, opErr
}
