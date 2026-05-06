// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package ws

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/ws/v1"
)

type IWsV1 interface {
	Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error)
}
