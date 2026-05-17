package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorHover hovers an element. BrowserExecutorHover 鎮仠鍏冪礌銆?
func (c *ControllerV1) BrowserExecutorHover(ctx context.Context, req *v1.BrowserExecutorHoverReq) (res *v1.BrowserExecutorHoverRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Hover(ctx, req.Identifier)
	return &v1.BrowserExecutorHoverRes{Result: result}, opErr
}
