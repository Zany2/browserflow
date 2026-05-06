package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
)

// ClientList returns registered clients 返回已注册客户端列表
func (c *ControllerV1) ClientList(ctx context.Context, req *v1.ClientListReq) (res *v1.ClientListRes, err error) {
	// Build base query 构建基础查询
	columns := dao.Clients.Columns()
	gModel := dao.Clients.Ctx(ctx)

	// Apply status filter 应用状态筛选
	switch strings.TrimSpace(req.Status) {
	case "banned":
		gModel = gModel.Where(columns.IsBanned, true)
	case "online", "offline":
		gModel = gModel.Where(columns.Status, strings.TrimSpace(req.Status))
	}

	// Apply ip filter 应用 IP 筛选
	if clientIP := strings.TrimSpace(req.IP); clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	}

	// Apply keyword filter 应用关键词筛选
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		gModel = gModel.Where(
			"("+columns.ClientName+" LIKE ? OR "+
				columns.ClientIp+" LIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	// Query clients 查询客户端列表
	clients := []entity.Clients{}
	if err = gModel.OrderDesc(columns.UpdatedAt).Scan(&clients); err != nil {
		return nil, err
	}

	return &v1.ClientListRes{
		List:  clients,
		Total: len(clients),
	}, nil
}
