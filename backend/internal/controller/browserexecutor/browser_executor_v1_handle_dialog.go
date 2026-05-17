package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorHandleDialog handles next JavaScript dialog. 处理下一个 JavaScript 弹窗。
func (c *ControllerV1) BrowserExecutorHandleDialog(ctx context.Context, req *v1.BrowserExecutorHandleDialogReq) (res *v1.BrowserExecutorHandleDialogRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.HandleDialog(ctx, req.Accept, req.Text, req.Timeout)
	return &v1.BrowserExecutorHandleDialogRes{Result: result}, opErr
}
