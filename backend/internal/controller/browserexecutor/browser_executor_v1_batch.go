package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorBatch executes operations sequentially. 顺序执行批量操作。
func (c *ControllerV1) BrowserExecutorBatch(ctx context.Context, req *v1.BrowserExecutorBatchReq) (res *v1.BrowserExecutorBatchRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	actions := make([]model.BrowserExecutorBatchAction, 0, len(req.Operations))
	for _, operation := range req.Operations {
		// Convert request item. 转换请求操作项。
		actions = append(actions, model.BrowserExecutorBatchAction{
			Type:        operation.Type,
			Params:      operation.Params,
			StopOnError: operation.StopOnError,
		})
	}
	result, err := executor.Batch(ctx, actions)
	return &v1.BrowserExecutorBatchRes{Result: result}, err
}
