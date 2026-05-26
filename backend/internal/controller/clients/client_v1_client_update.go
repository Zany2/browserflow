package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/gogf/gf/v2/util/gconv"
)

// ClientUpdate saves the client display name. 保存客户端自定义显示名称
func (c *ControllerV1) ClientUpdate(ctx context.Context, req *v1.ClientUpdateReq) (res *v1.ClientUpdateRes, err error) {
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientUpdateRes{Message: "客户端不存在或未注册"}, nil
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if _, err = scopedClientModel(ctx, record).Data(do.Clients{
		DisplayName: displayName,
	}).Update(); err != nil {
		return nil, err
	}

	columns := dao.Clients.Columns()
	updated, err := dao.Clients.Ctx(ctx).WherePri(gconv.Int64(record[columns.Id])).One()
	if err != nil {
		return nil, err
	}
	client, err := clientRecordToEntity(updated)
	if err != nil {
		return nil, err
	}
	return &v1.ClientUpdateRes{Client: client, Message: "客户端已保存"}, nil
}
