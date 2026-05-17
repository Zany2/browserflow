package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorDrag drags one element to another. 拖拽元素到目标元素。
func (c *ControllerV1) BrowserExecutorDrag(ctx context.Context, req *v1.BrowserExecutorDragReq) (res *v1.BrowserExecutorDragRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Drag(ctx, req.FromIdentifier, req.ToIdentifier)
	return &v1.BrowserExecutorDragRes{Result: result}, opErr
}
