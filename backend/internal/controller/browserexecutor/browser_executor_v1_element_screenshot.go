package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorElementScreenshot captures an element screenshot. BrowserExecutorElementScreenshot 截取元素图片。
func (c *ControllerV1) BrowserExecutorElementScreenshot(ctx context.Context, req *v1.BrowserExecutorElementScreenshotReq) (res *v1.BrowserExecutorElementScreenshotRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.ElementScreenshot(ctx, req.Identifier, req.Format, req.Quality)
	return &v1.BrowserExecutorElementScreenshotRes{Result: result}, opErr
}
