package clients

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
)

// ClientDetail returns one registered client 返回单个已注册客户端
func (c *ControllerV1) ClientDetail(ctx context.Context, req *v1.ClientDetailReq) (res *v1.ClientDetailRes, err error) {
	// Query client by IP or primary id 按 IP 或主键查询客户端
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
