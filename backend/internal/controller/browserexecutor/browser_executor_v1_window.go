package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorWindow manages browser window bounds and state. BrowserExecutorWindow 执行窗口操作。
func (c *ControllerV1) BrowserExecutorWindow(ctx context.Context, req *v1.BrowserExecutorWindowReq) (res *v1.BrowserExecutorWindowRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Window(ctx, model.BrowserExecutorWindowOptions{
		Action: req.Action,
		Left:   req.Left,
		Top:    req.Top,
		Width:  req.Width,
		Height: req.Height,
	})
	return &v1.BrowserExecutorWindowRes{Result: result}, opErr
}
