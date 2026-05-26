package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
)

// ClientDetail returns one registered client node. 返回单个客户端节点
func (c *ControllerV1) ClientDetail(ctx context.Context, req *v1.ClientDetailReq) (res *v1.ClientDetailRes, err error) {
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientDetailRes{}, nil
	}

	client, err := clientRecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientDetailRes{Client: client}, nil
}
