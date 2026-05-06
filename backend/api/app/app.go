// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package app

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/app/v1"
)

type IAppV1 interface {
	Runtime(ctx context.Context, req *v1.RuntimeReq) (res *v1.RuntimeRes, err error)
}
