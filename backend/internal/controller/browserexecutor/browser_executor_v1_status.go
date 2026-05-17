package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorStatus gets current executor status. BrowserExecutorStatus 閼惧嘲褰囪ぐ鎾冲閹笛嗩攽閸ｃ劎濮搁幀浣碘偓?
func (c *ControllerV1) BrowserExecutorStatus(ctx context.Context, req *v1.BrowserExecutorStatusReq) (res *v1.BrowserExecutorStatusRes, err error) {
	status := browserexecutor.CurrentStatus(ctx)
	return &v1.BrowserExecutorStatusRes{Status: &status}, nil
}
