package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorExtract extracts data from page. BrowserExecutorExtract 閹绘劕褰囨い鐢告桨閺佺増宓侀妴?
func (c *ControllerV1) BrowserExecutorExtract(ctx context.Context, req *v1.BrowserExecutorExtractReq) (res *v1.BrowserExecutorExtractRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Extract(ctx, req.Selector, req.Fields, req.Multiple)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorExtractRes{Result: result}, nil
}
