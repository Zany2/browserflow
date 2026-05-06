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

// ClientUpdate updates editable client fields 更新客户端可编辑字段
func (c *ControllerV1) ClientUpdate(ctx context.Context, req *v1.ClientUpdateReq) (res *v1.ClientUpdateRes, err error) {
	// Query target client 查询目标客户端
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientUpdateRes{Message: "客户端不存在或未注册"}, nil
	}

	// Persist editable name 保存可编辑客户端名称
	columns := dao.Clients.Columns()
	_, err = dao.Clients.Ctx(ctx).
		WherePri(gconv.Int64(record[columns.Id])).
		Data(g.Map{
			columns.ClientName: strings.TrimSpace(req.ClientName),
			columns.UpdatedAt:  gtime.Now(),
		}).
		Update()
	if err != nil {
		return nil, err
	}

	updated, err := dao.Clients.Ctx(ctx).WherePri(gconv.Int64(record[columns.Id])).One()
	if err != nil {
		return nil, err
	}
	client, err := clientRecordToEntity(updated)
	if err != nil {
		return nil, err
	}
	return &v1.ClientUpdateRes{Client: client, Message: "客户端已更新"}, nil
}
