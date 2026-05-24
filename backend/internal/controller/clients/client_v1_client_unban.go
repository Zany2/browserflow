package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/clientops"
)

// ClientUnban removes client ban 解除客户端拉黑
func (c *ControllerV1) ClientUnban(ctx context.Context, req *v1.ClientUnbanReq) (res *v1.ClientUnbanRes, err error) {
	// Query target client 查询目标客户端
	record, err := clientops.QueryRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientUnbanRes{Message: "客户端不存在或未注册"}, nil
	}

	// Clear ban state 清除拉黑状态
	_, err = clientops.ScopedModel(ctx, record).
		Data(do.Clients{
			IsBanned:  false,
			BanReason: "",
		}).
		Update()
	if err != nil {
		return nil, err
	}

	client, err := clientops.RecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientUnbanRes{Client: client, Message: "客户端已解除拉黑"}, nil
}
