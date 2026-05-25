package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
)

// ClientList returns registered client nodes. 返回已注册客户端节点列表
func (c *ControllerV1) ClientList(ctx context.Context, req *v1.ClientListReq) (res *v1.ClientListRes, err error) {
	columns := dao.Clients.Columns()
	gModel := dao.Clients.Ctx(ctx)

	switch strings.TrimSpace(req.Status) {
	case "banned":
		gModel = gModel.Where(columns.IsBanned, true)
	case "online", "offline":
		gModel = gModel.Where(columns.Status, strings.TrimSpace(req.Status))
	}

	if clientIP := strings.TrimSpace(req.IP); clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	}
	if nodeID := strings.TrimSpace(req.NodeID); nodeID != "" {
		gModel = gModel.Where(columns.NodeId, nodeID)
	}

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		gModel = gModel.Where(
			"("+columns.ClientIp+" LIKE ? OR "+
				columns.NodeId+" LIKE ? OR "+
				columns.DisplayName+" LIKE ? OR "+
				columns.Hostname+" LIKE ? OR "+
				columns.BrowserName+" LIKE ? OR "+
				columns.BrowserVersion+" LIKE ? OR "+
				columns.WorkerVersion+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	clients := []entity.Clients{}
	if err = gModel.OrderDesc(columns.CreatedAt).OrderDesc(columns.Id).Scan(&clients); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Status) == "online" {
		onlineClients := make([]entity.Clients, 0, len(clients))
		for _, client := range clients {
			if workflowcache.IsClientOnline(ctx, websockets.NodeConnectionID(client.ClientIp, client.NodeId)) {
				onlineClients = append(onlineClients, client)
			}
		}
		clients = onlineClients
	}

	return &v1.ClientListRes{List: clients, Total: len(clients)}, nil
}
