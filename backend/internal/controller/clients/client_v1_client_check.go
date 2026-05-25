package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/util/gconv"
)

// ClientCheck checks whether a client can connect. 检查客户端是否允许连接
func (c *ControllerV1) ClientCheck(ctx context.Context, req *v1.ClientCheckReq) (res *v1.ClientCheckRes, err error) {
	clientIP := clientops.RequestIP(ctx)
	if clientIP == "" {
		return &v1.ClientCheckRes{Allowed: true}, nil
	}

	columns := dao.Clients.Columns()
	record, err := dao.Clients.Ctx(ctx).
		Where(columns.ClientIp, clientIP).
		Where(columns.IsBanned, true).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientCheckRes{Allowed: true}, nil
	}

	reason := strings.TrimSpace(gconv.String(record[columns.BanReason]))
	client, err := clientRecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientCheckRes{
		Allowed:  false,
		IsBanned: true,
		Reason:   reason,
		Client:   client,
	}, nil
}
