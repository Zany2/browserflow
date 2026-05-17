package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorSnapshot returns accessibility snapshot. BrowserExecutorSnapshot 鏉╂柨娲栭崣顖濐問闂傤喗鈧冩彥閻撗佲偓?
func (c *ControllerV1) BrowserExecutorSnapshot(ctx context.Context, req *v1.BrowserExecutorSnapshotReq) (res *v1.BrowserExecutorSnapshotRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	result, opErr := executor.Snapshot(ctx)
	return &v1.BrowserExecutorSnapshotRes{Result: result}, opErr
}
