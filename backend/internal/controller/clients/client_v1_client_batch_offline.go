package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientBatchOffline forces client nodes offline in batch. 批量强制客户端节点下线
func (c *ControllerV1) ClientBatchOffline(ctx context.Context, req *v1.ClientBatchOfflineReq) (res *v1.ClientBatchOfflineRes, err error) {
	stats := v1.ClientBatchActionRes{Total: len(req.IDs)}
	for _, id := range req.IDs {
		record, queryErr := queryClientRecord(ctx, id)
		if queryErr != nil {
			return nil, queryErr
		}
		if record.IsEmpty() {
			stats.NotFound++
			continue
		}

		clientIP := clientIPFromRecord(record)
		nodeID := clientNodeIDFromRecord(record)
		closed := clientops.CloseConnection(ctx, clientIP, nodeID)
		if _, err = scopedClientModel(ctx, record).
			Data(do.Clients{
				Status:         "offline",
				DisconnectedAt: gtime.Now(),
			}).
			Update(); err != nil {
			return nil, err
		}

		stats.Success++
		stats.Notified += closed
	}

	return &v1.ClientBatchOfflineRes{ClientBatchActionRes: stats}, nil
}
