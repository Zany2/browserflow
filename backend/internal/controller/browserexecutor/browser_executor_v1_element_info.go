package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorElementInfo reads element diagnostics. BrowserExecutorElementInfo 获取元素详情。
func (c *ControllerV1) BrowserExecutorElementInfo(ctx context.Context, req *v1.BrowserExecutorElementInfoReq) (res *v1.BrowserExecutorElementInfoRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.ElementInfo(ctx, req.Identifier, req.Attributes)
	return &v1.BrowserExecutorElementInfoRes{Result: result}, opErr
}
