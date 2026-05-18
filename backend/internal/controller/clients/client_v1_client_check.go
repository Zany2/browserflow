package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/util/gconv"
)

// ClientCheck checks whether client can connect 检查客户端是否允许连接
func (c *ControllerV1) ClientCheck(ctx context.Context, req *v1.ClientCheckReq) (res *v1.ClientCheckRes, err error) {
	// Resolve request ip 解析请求 IP
	clientIP := clientops.RequestIP(ctx)
	if clientIP == "" {
		return &v1.ClientCheckRes{Allowed: true}, nil
	}

	// Query ban record 查询拉黑记录
	record, err := clientops.QueryBannedRecord(ctx, clientIP)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientCheckRes{Allowed: true}, nil
	}

	// Return banned result 返回拉黑状态
	reason := strings.TrimSpace(gconv.String(record["ban_reason"]))
	client, err := clientops.RecordToEntity(record)
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
