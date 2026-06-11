package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/database/gdb"
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

	runtimeStatus := strings.TrimSpace(req.Status)
	filterByRuntimeStatus := runtimeStatus == "online" || runtimeStatus == "offline"
	var onlineSet map[string]struct{}
	if filterByRuntimeStatus {
		onlineClients, listErr := workflowcache.ListOnlineClients(ctx)
		if listErr != nil {
			return nil, listErr
		}
		onlineSet = buildOnlineClientSet(onlineClients)
		if runtimeStatus == "online" && len(onlineSet) == 0 {
			return &v1.ClientListRes{List: []entity.Clients{}, Total: 0}, nil
		}
		gModel = applyRuntimeStatusFilter(gModel, columns.ClientIp, columns.NodeId, onlineSet, runtimeStatus)
	}

	total, err := gModel.Count()
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &v1.ClientListRes{List: []entity.Clients{}, Total: 0}, nil
	}

	clients := []entity.Clients{}
	queryModel := gModel.OrderDesc(columns.CreatedAt).OrderDesc(columns.Id)
	if pageSize > 0 {
		queryModel = queryModel.Limit((pageNum-1)*pageSize, pageSize)
	}
	if err = queryModel.Scan(&clients); err != nil {
		return nil, err
	}
	if onlineSet != nil {
		markClientRuntimeStatusWithSet(clients, onlineSet)
	} else {
		if onlineSet, err = onlineSetForClients(ctx, clients); err != nil {
			return nil, err
		}
		markClientRuntimeStatusWithSet(clients, onlineSet)
	}

	return &v1.ClientListRes{List: clients, Total: total}, nil
}

func onlineSetForClients(ctx context.Context, clients []entity.Clients) (map[string]struct{}, error) {
	identities := make([]string, 0, len(clients))
	for _, client := range clients {
		identities = append(identities, clientRuntimeIdentity(client))
	}
	return workflowcache.GetOnlineClientSet(ctx, identities)
}

func buildOnlineClientSet(onlineClients []string) map[string]struct{} {
	onlineSet := make(map[string]struct{}, len(onlineClients))
	for _, client := range onlineClients {
		if client = strings.TrimSpace(client); client != "" {
			onlineSet[client] = struct{}{}
		}
	}
	return onlineSet
}

func markClientRuntimeStatusWithSet(clients []entity.Clients, onlineSet map[string]struct{}) {
	for i := range clients {
		if _, ok := onlineSet[clientRuntimeIdentity(clients[i])]; ok {
			clients[i].Status = "online"
			continue
		}
		clients[i].Status = "offline"
	}
}

func clientRuntimeIdentity(client entity.Clients) string {
	return websockets.NodeConnectionID(client.ClientIp, client.NodeId)
}

func applyRuntimeStatusFilter(model *gdb.Model, clientIPColumn string, nodeIDColumn string, onlineSet map[string]struct{}, status string) *gdb.Model {
	if len(onlineSet) == 0 {
		return model
	}

	conditions := make([]string, 0, len(onlineSet))
	args := make([]any, 0, len(onlineSet)*2)
	for identity := range onlineSet {
		clientIP, nodeID := splitRuntimeIdentity(identity)
		if clientIP == "" && nodeID == "" {
			continue
		}
		if clientIP != "" && nodeID != "" {
			conditions = append(conditions, "("+clientIPColumn+" = ? AND "+nodeIDColumn+" = ?)")
			args = append(args, clientIP, nodeID)
			continue
		}
		if clientIP != "" {
			conditions = append(conditions, "("+clientIPColumn+" = ? AND ("+nodeIDColumn+" IS NULL OR "+nodeIDColumn+" = '' OR "+nodeIDColumn+" = ?))")
			args = append(args, clientIP, clientIP)
			continue
		}
		conditions = append(conditions, "("+nodeIDColumn+" = ?)")
		args = append(args, nodeID)
	}
	if len(conditions) == 0 {
		return model
	}

	condition := "(" + strings.Join(conditions, " OR ") + ")"
	if status == "offline" {
		condition = "NOT " + condition
	}
	return model.Where(condition, args...)
}

func splitRuntimeIdentity(identity string) (string, string) {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return "", ""
	}
	if strings.Contains(identity, "|") {
		parts := strings.SplitN(identity, "|", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return identity, ""
}
