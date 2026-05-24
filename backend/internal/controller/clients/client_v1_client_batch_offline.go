package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/clientops"
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientBatchOffline forces clients offline in batch 批量强制客户端下线
func (c *ControllerV1) ClientBatchOffline(ctx context.Context, req *v1.ClientBatchOfflineReq) (res *v1.ClientBatchOfflineRes, err error) {
	// Stats 统计批量下线处理结果
	stats := v1.ClientBatchActionRes{
		Total: len(req.IDs),
	}
	for _, id := range req.IDs {
		// Query client 按选择的客户端标识查询记录
		record, queryErr := clientops.QueryRecord(ctx, id)
		if queryErr != nil {
			return nil, queryErr
		}
		if record.IsEmpty() {
			stats.NotFound++
			continue
		}

		// Mark offline 标记离线并关闭当前连接
		closed := clientops.CloseConnection(ctx, clientops.TargetConnectionID(record))
		if _, err = clientops.ScopedModel(ctx, record).
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

	return &v1.ClientBatchOfflineRes{
		ClientBatchActionRes: stats,
	}, nil
}
