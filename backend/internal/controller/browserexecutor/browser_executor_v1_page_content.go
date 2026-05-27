package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorPageContent gets compact page HTML. BrowserExecutorPageContent 閼惧嘲褰囩槐褍鍣炬い鐢告桨 HTML閵?
func (c *ControllerV1) BrowserExecutorPageContent(ctx context.Context, req *v1.BrowserExecutorPageContentReq) (res *v1.BrowserExecutorPageContentRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.PageContent(ctx, req.Limit)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorPageContentRes{Result: result}, nil
}
