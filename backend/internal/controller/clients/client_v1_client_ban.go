package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientBan bans one client node. 拉黑单个客户端节点
func (c *ControllerV1) ClientBan(ctx context.Context, req *v1.ClientBanReq) (res *v1.ClientBanRes, err error) {
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientBanRes{Message: "客户端不存在或未注册"}, nil
	}

	clientID := clientPrimaryIDFromRecord(record)
	clientIP := clientIPFromRecord(record)
	nodeID := clientNodeIDFromRecord(record)
	reason := strings.TrimSpace(req.Reason)
	_, err = scopedClientModel(ctx, record).
		Data(do.Clients{
			IsBanned:       true,
			BanReason:      reason,
			Status:         "offline",
			DisconnectedAt: gtime.Now(),
		}).
		Update()
	if err != nil {
		return nil, err
	}

	sent := clientops.NotifyBanned(ctx, clientIP, nodeID, clientID, reason)
	message := "客户端已拉黑"
	if sent > 0 {
		message = "客户端已拉黑，并已通知客户端暂停当前连接"
	}
	client, err := clientRecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientBanRes{Client: client, Message: message}, nil
}
