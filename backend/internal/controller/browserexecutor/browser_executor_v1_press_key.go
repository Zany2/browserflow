package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorPressKey presses keyboard shortcut. BrowserExecutorPressKey 閸欐垿鈧焦瀵滈柨顔衡偓?
func (c *ControllerV1) BrowserExecutorPressKey(ctx context.Context, req *v1.BrowserExecutorPressKeyReq) (res *v1.BrowserExecutorPressKeyRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.PressKey(ctx, req.Key, req.Ctrl, req.Shift, req.Alt, req.MetaKey)
	return &v1.BrowserExecutorPressKeyRes{Result: result}, opErr
}
