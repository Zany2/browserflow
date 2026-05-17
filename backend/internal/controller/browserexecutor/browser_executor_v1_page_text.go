package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorPageText gets compact page text. BrowserExecutorPageText 閼惧嘲褰囩槐褍鍣炬い鐢告桨閺傚洦婀伴妴?
func (c *ControllerV1) BrowserExecutorPageText(ctx context.Context, req *v1.BrowserExecutorPageTextReq) (res *v1.BrowserExecutorPageTextRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.PageText(ctx, req.Limit)
	return &v1.BrowserExecutorPageTextRes{Result: result}, opErr
}
