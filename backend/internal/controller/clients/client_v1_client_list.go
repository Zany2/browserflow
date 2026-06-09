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

	switch strings.TrimSpace(req.BusyStatus) {
	case "idle", "busy", "unknown":
		gModel = gModel.Where(columns.BusyStatus, strings.TrimSpace(req.BusyStatus))
	}
	switch strings.TrimSpace(req.IsBanned) {
	case "true":
		gModel = gModel.Where(columns.IsBanned, true)
	case "false":
		gModel = gModel.Where(columns.IsBanned, false)
	}

	if clientIP := strings.TrimSpace(req.IP); clientIP != "" {
		gModel = gModel.Where(columns.ClientIp, clientIP)
	}
	if nodeID := strings.TrimSpace(req.NodeID); nodeID != "" {
		gModel = gModel.Where(columns.NodeId, nodeID)
	}
	if startTime := strings.TrimSpace(req.LastSeenStartTime); startTime != "" {
		gModel = gModel.WhereGTE(columns.LastSeenAt, startTime)
	}
	if endTime := strings.TrimSpace(req.LastSeenEndTime); endTime != "" {
		gModel = gModel.WhereLTE(columns.LastSeenAt, endTime)
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

	pageNum := req.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := req.PageSize
	if pageSize < 0 {
		pageSize = 0
	}

	filterByRuntimeStatus := strings.TrimSpace(req.Status) == "online" || strings.TrimSpace(req.Status) == "offline"
	total, err := gModel.Count()
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &v1.ClientListRes{List: []entity.Clients{}, Total: 0}, nil
	}

	clients := []entity.Clients{}
	queryModel := gModel.OrderDesc(columns.CreatedAt).OrderDesc(columns.Id)
	if pageSize > 0 && !filterByRuntimeStatus {
		queryModel = queryModel.Limit((pageNum-1)*pageSize, pageSize)
	}
	if err = queryModel.Scan(&clients); err != nil {
		return nil, err
	}
	markClientRuntimeStatus(ctx, clients)
	if filterByRuntimeStatus {
		filteredClients := make([]entity.Clients, 0, len(clients))
		for _, client := range clients {
			if client.Status == strings.TrimSpace(req.Status) {
				filteredClients = append(filteredClients, client)
			}
		}
		clients = filteredClients
		total = len(clients)
		if pageSize > 0 {
			start := (pageNum - 1) * pageSize
			if start >= len(clients) {
				clients = []entity.Clients{}
			} else {
				end := start + pageSize
				if end > len(clients) {
					end = len(clients)
				}
				clients = clients[start:end]
			}
		}
	}

	return &v1.ClientListRes{List: clients, Total: total}, nil
}

func markClientRuntimeStatus(ctx context.Context, clients []entity.Clients) {
	for i := range clients {
		if workflowcache.IsClientOnline(ctx, websockets.NodeConnectionID(clients[i].ClientIp, clients[i].NodeId)) {
			clients[i].Status = "online"
			continue
		}
		clients[i].Status = "offline"
	}
}
