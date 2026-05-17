package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorScreenshot captures screenshot. BrowserExecutorScreenshot 閹搭亜娴橀妴?
func (c *ControllerV1) BrowserExecutorScreenshot(ctx context.Context, req *v1.BrowserExecutorScreenshotReq) (res *v1.BrowserExecutorScreenshotRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Screenshot(ctx, req.FullPage, req.Format, req.Quality)
	return &v1.BrowserExecutorScreenshotRes{Result: result}, opErr
}
