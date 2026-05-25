package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/model/do"
)

// ClientUnban removes client node ban. 解除客户端节点拉黑
func (c *ControllerV1) ClientUnban(ctx context.Context, req *v1.ClientUnbanReq) (res *v1.ClientUnbanRes, err error) {
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientUnbanRes{Message: "客户端不存在或未注册"}, nil
	}

	_, err = scopedClientModel(ctx, record).
		Data(do.Clients{
			IsBanned:  false,
			BanReason: "",
		}).
		Update()
	if err != nil {
		return nil, err
	}

	client, err := clientRecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientUnbanRes{Client: client, Message: "客户端已解除拉黑"}, nil
}
