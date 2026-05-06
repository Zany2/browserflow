// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package agents

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/agents/v1"
)

type IAgentsV1 interface {
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
}
