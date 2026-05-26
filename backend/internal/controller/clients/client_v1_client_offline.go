package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientOffline forces one client node offline. 强制单个客户端节点下线
func (c *ControllerV1) ClientOffline(ctx context.Context, req *v1.ClientOfflineReq) (res *v1.ClientOfflineRes, err error) {
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientOfflineRes{Message: "客户端不存在或未注册"}, nil
	}

	clientIP := clientIPFromRecord(record)
	nodeID := clientNodeIDFromRecord(record)
	closed := clientops.CloseConnection(ctx, clientIP, nodeID)
	_, err = scopedClientModel(ctx, record).
		Data(do.Clients{
			Status:         "offline",
			DisconnectedAt: gtime.Now(),
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
