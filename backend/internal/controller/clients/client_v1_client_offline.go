package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// ClientOffline forces one client offline 强制单个客户端下线
func (c *ControllerV1) ClientOffline(ctx context.Context, req *v1.ClientOfflineReq) (res *v1.ClientOfflineRes, err error) {
	// Query target client 查询目标客户端
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientOfflineRes{Message: "客户端不存在或未注册"}, nil
	}

	// Close websocket and mark offline 关闭 WebSocket 并标记离线
	columns := dao.Clients.Columns()
	clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
	closed := closeClientConnection(ctx, clientIP)
	now := gtime.Now()
	_, err = dao.Clients.Ctx(ctx).
		Where(columns.ClientIp, clientIP).
		Data(g.Map{
			columns.Status:         "offline",
			columns.DisconnectedAt: now,
			columns.UpdatedAt:      now,
		}).
		Update()
	if err != nil {
		return nil, err
	}

	message := "客户端未在线，已更新为离线"
	if closed > 0 {
		message = "客户端连接已断开，将由客户端自动重连"
	}
	client, err := clientRecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientOfflineRes{Client: client, Message: message}, nil
}
