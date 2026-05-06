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

// ClientBatchBan bans clients in batch 批量拉黑客户端
func (c *ControllerV1) ClientBatchBan(ctx context.Context, req *v1.ClientBatchBanReq) (res *v1.ClientBatchBanRes, err error) {
	// Stats 统计批量拉黑处理结果
	stats := v1.ClientBatchActionRes{
		Total: len(req.IDs),
	}
	columns := dao.Clients.Columns()
	now := gtime.Now()
	reason := strings.TrimSpace(req.Reason)

	for _, id := range req.IDs {
		// Query client 按选择的客户端标识查询记录
		record, queryErr := queryClientRecord(ctx, id)
		if queryErr != nil {
			return nil, queryErr
		}
		if record.IsEmpty() {
			stats.NotFound++
			continue
		}

		// Persist ban state 保存拉黑状态
		clientID := strings.TrimSpace(gconv.String(record[columns.ClientId]))
		clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
		if _, err = dao.Clients.Ctx(ctx).
			Where(columns.ClientIp, clientIP).
			Data(g.Map{
				columns.IsBanned:       true,
				columns.BanReason:      reason,
				columns.Status:         "offline",
				columns.DisconnectedAt: now,
				columns.UpdatedAt:      now,
			}).
			Update(); err != nil {
			return nil, err
		}

		stats.Success++
		stats.Notified += notifyClientBanned(ctx, clientIP, clientID, reason)
	}

	return &v1.ClientBatchBanRes{
		ClientBatchActionRes: stats,
	}, nil
}
