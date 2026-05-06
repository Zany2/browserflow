// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
)

type IClientsV1 interface {
	ClientList(ctx context.Context, req *v1.ClientListReq) (res *v1.ClientListRes, err error)
	ClientDetail(ctx context.Context, req *v1.ClientDetailReq) (res *v1.ClientDetailRes, err error)
	ClientUpdate(ctx context.Context, req *v1.ClientUpdateReq) (res *v1.ClientUpdateRes, err error)
	ClientCheck(ctx context.Context, req *v1.ClientCheckReq) (res *v1.ClientCheckRes, err error)
	ClientOffline(ctx context.Context, req *v1.ClientOfflineReq) (res *v1.ClientOfflineRes, err error)
	ClientBatchOffline(ctx context.Context, req *v1.ClientBatchOfflineReq) (res *v1.ClientBatchOfflineRes, err error)
	ClientBan(ctx context.Context, req *v1.ClientBanReq) (res *v1.ClientBanRes, err error)
	ClientBatchBan(ctx context.Context, req *v1.ClientBatchBanReq) (res *v1.ClientBatchBanRes, err error)
	ClientUnban(ctx context.Context, req *v1.ClientUnbanReq) (res *v1.ClientUnbanRes, err error)
}
