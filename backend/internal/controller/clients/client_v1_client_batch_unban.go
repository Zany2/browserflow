package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
)

// ClientBatchUnban removes client node bans in batch. 批量解除客户端节点拉黑
func (c *ControllerV1) ClientBatchUnban(ctx context.Context, req *v1.ClientBatchUnbanReq) (res *v1.ClientBatchUnbanRes, err error) {
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

		if _, err = scopedClientModel(ctx, record).
			Data(do.Clients{
				IsBanned:  false,
				BanReason: "",
			}).
			Update(); err != nil {
			return nil, err
		}

		stats.Success++
	}

	return &v1.ClientBatchUnbanRes{ClientBatchActionRes: stats}, nil
}
