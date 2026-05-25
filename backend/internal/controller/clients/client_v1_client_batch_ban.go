package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientBatchBan bans client nodes in batch. 批量拉黑客户端节点
func (c *ControllerV1) ClientBatchBan(ctx context.Context, req *v1.ClientBatchBanReq) (res *v1.ClientBatchBanRes, err error) {
	stats := v1.ClientBatchActionRes{Total: len(req.IDs)}
	reason := strings.TrimSpace(req.Reason)

	for _, id := range req.IDs {
		record, queryErr := queryClientRecord(ctx, id)
		if queryErr != nil {
			return nil, queryErr
		}
		if record.IsEmpty() {
			stats.NotFound++
			continue
		}

		clientID := clientPrimaryIDFromRecord(record)
		clientIP := clientIPFromRecord(record)
		nodeID := clientNodeIDFromRecord(record)
		if _, err = scopedClientModel(ctx, record).
			Data(do.Clients{
				IsBanned:       true,
				BanReason:      reason,
				Status:         "offline",
				DisconnectedAt: gtime.Now(),
			}).
			Update(); err != nil {
			return nil, err
		}

		stats.Success++
		stats.Notified += clientops.NotifyBanned(ctx, clientIP, nodeID, clientID, reason)
	}

	return &v1.ClientBatchBanRes{ClientBatchActionRes: stats}, nil
}
