package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorStatus gets current executor status. 获取当前执行器状态。
func (c *ControllerV1) BrowserExecutorStatus(ctx context.Context, req *v1.BrowserExecutorStatusReq) (res *v1.BrowserExecutorStatusRes, err error) {
	status := browserexecutor.CurrentStatus(ctx)
	return &v1.BrowserExecutorStatusRes{Status: &status}, nil
}
