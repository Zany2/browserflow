package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorStorage manages localStorage or sessionStorage. 管理 localStorage 或 sessionStorage。
func (c *ControllerV1) BrowserExecutorStorage(ctx context.Context, req *v1.BrowserExecutorStorageReq) (res *v1.BrowserExecutorStorageRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Storage(ctx, model.BrowserExecutorStorageOptions{
		Action: req.Action,
		Type:   req.Type,
		Key:    req.Key,
		Value:  req.Value,
	})
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorStorageRes{Result: result}, nil
}
