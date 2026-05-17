package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorResize changes viewport size. BrowserExecutorResize 璋冩暣瑙嗗彛澶у皬銆?
func (c *ControllerV1) BrowserExecutorResize(ctx context.Context, req *v1.BrowserExecutorResizeReq) (res *v1.BrowserExecutorResizeRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Resize(ctx, req.Width, req.Height)
	return &v1.BrowserExecutorResizeRes{Result: result}, opErr
}
