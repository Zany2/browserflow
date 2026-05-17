package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorMouse runs coordinate mouse operations. BrowserExecutorMouse 执行鼠标操作。
func (c *ControllerV1) BrowserExecutorMouse(ctx context.Context, req *v1.BrowserExecutorMouseReq) (res *v1.BrowserExecutorMouseRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Mouse(ctx, model.BrowserExecutorMouseOptions{
		Action: req.Action,
		X:      req.X,
		Y:      req.Y,
		DeltaX: req.DeltaX,
		DeltaY: req.DeltaY,
		Steps:  req.Steps,
		Button: req.Button,
	})
	return &v1.BrowserExecutorMouseRes{Result: result}, opErr
}
