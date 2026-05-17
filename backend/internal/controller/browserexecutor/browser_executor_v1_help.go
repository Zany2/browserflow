package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorHelp returns available executor commands. BrowserExecutorHelp 鏉╂柨娲栭崣顖滄暏閸涙垝鎶ら妴?
func (c *ControllerV1) BrowserExecutorHelp(ctx context.Context, req *v1.BrowserExecutorHelpReq) (res *v1.BrowserExecutorHelpRes, err error) {
	return &v1.BrowserExecutorHelpRes{
		BasePath: browserexecutor.BasePath,
		Commands: browserexecutor.Help(req.Command),
	}, nil
}
