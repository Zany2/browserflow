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

// ClientUnban removes client ban 解除客户端拉黑
func (c *ControllerV1) ClientUnban(ctx context.Context, req *v1.ClientUnbanReq) (res *v1.ClientUnbanRes, err error) {
	// Query target client 查询目标客户端
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientUnbanRes{Message: "客户端不存在或未注册"}, nil
	}

	// Clear ban state 清除拉黑状态
	columns := dao.Clients.Columns()
	clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
	_, err = dao.Clients.Ctx(ctx).
		Where(columns.ClientIp, clientIP).
		Data(g.Map{
			columns.IsBanned:  false,
			columns.BanReason: "",
			columns.UpdatedAt: gtime.Now(),
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
