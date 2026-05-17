package browserexecutor

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorNavigate navigates current page. 导航当前页面。
func (c *ControllerV1) BrowserExecutorNavigate(ctx context.Context, req *v1.BrowserExecutorNavigateReq) (res *v1.BrowserExecutorNavigateRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Navigate(ctx, strings.TrimSpace(req.URL), req.Timeout)
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorNavigateRes{Result: result}, opErr
}
