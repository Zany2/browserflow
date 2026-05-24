package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
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
	if machineID := strings.TrimSpace(req.MachineID); machineID != "" {
		gModel = gModel.Where(columns.MachineId, machineID)
	}
	if nodeID := strings.TrimSpace(req.NodeID); nodeID != "" {
		gModel = gModel.Where(columns.NodeId, nodeID)
	}

	// Apply keyword filter 应用关键词筛选
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		gModel = gModel.Where(
			"("+columns.ClientName+" LIKE ? OR "+
				columns.ClientIp+" LIKE ? OR "+
				columns.MachineId+" LIKE ? OR "+
				columns.NodeId+" LIKE ? OR "+
				columns.NodeName+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	// Query clients 查询客户端列表
	clients := []entity.Clients{}
	if err = gModel.OrderDesc(columns.CreatedAt).OrderDesc(columns.Id).Scan(&clients); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Status) == "online" {
		onlineClients := make([]entity.Clients, 0, len(clients))
		for _, client := range clients {
			if workflowcache.IsClientOnline(ctx, firstNonEmpty(client.NodeId, client.ClientIp)) {
				onlineClients = append(onlineClients, client)
			}
		}
		clients = onlineClients
	}

	return &v1.ClientListRes{
		List:  clients,
		Total: len(clients),
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
