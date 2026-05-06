// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package browser

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
)

type IBrowserV1 interface {
	BrowserStatus(ctx context.Context, req *v1.BrowserStatusReq) (res *v1.BrowserStatusRes, err error)
	BrowserStart(ctx context.Context, req *v1.BrowserStartReq) (res *v1.BrowserStartRes, err error)
	BrowserStop(ctx context.Context, req *v1.BrowserStopReq) (res *v1.BrowserStopRes, err error)
	BrowserInstanceList(ctx context.Context, req *v1.BrowserInstanceListReq) (res *v1.BrowserInstanceListRes, err error)
	BrowserInstanceCreate(ctx context.Context, req *v1.BrowserInstanceCreateReq) (res *v1.BrowserInstanceCreateRes, err error)
	BrowserInstanceCurrent(ctx context.Context, req *v1.BrowserInstanceCurrentReq) (res *v1.BrowserInstanceCurrentRes, err error)
	BrowserInstanceDetail(ctx context.Context, req *v1.BrowserInstanceDetailReq) (res *v1.BrowserInstanceDetailRes, err error)
	BrowserInstanceUpdate(ctx context.Context, req *v1.BrowserInstanceUpdateReq) (res *v1.BrowserInstanceUpdateRes, err error)
	BrowserInstanceDelete(ctx context.Context, req *v1.BrowserInstanceDeleteReq) (res *v1.BrowserInstanceDeleteRes, err error)
	BrowserInstanceStart(ctx context.Context, req *v1.BrowserInstanceStartReq) (res *v1.BrowserInstanceStartRes, err error)
	BrowserInstanceStop(ctx context.Context, req *v1.BrowserInstanceStopReq) (res *v1.BrowserInstanceStopRes, err error)
	BrowserInstanceSwitch(ctx context.Context, req *v1.BrowserInstanceSwitchReq) (res *v1.BrowserInstanceSwitchRes, err error)
}
