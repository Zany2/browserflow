package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorPageInfo gets active page info. BrowserExecutorPageInfo 閼惧嘲褰囪ぐ鎾冲妞ょ敻娼版穱鈩冧紖閵?
func (c *ControllerV1) BrowserExecutorPageInfo(ctx context.Context, req *v1.BrowserExecutorPageInfoReq) (res *v1.BrowserExecutorPageInfoRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.PageInfo(ctx)
	return &v1.BrowserExecutorPageInfoRes{Result: result}, opErr
}
