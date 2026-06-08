package app

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/app/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
)

// ServerDashboard returns server-mode homepage statistics.
func (c *ControllerV1) ServerDashboard(ctx context.Context, req *v1.ServerDashboardReq) (res *v1.ServerDashboardRes, err error) {
	res = &v1.ServerDashboardRes{}
	if consts.ResolveRuntimeMode(ctx) != consts.RuntimeModeServer {
		return res, nil
	}

	clientColumns := dao.Clients.Columns()
	clients := []entity.Clients{}
	if err = dao.Clients.Ctx(ctx).Fields(
		clientColumns.ClientIp,
		clientColumns.Status,
		clientColumns.BusyStatus,
		clientColumns.PluginStatus,
	).Scan(&clients); err != nil {
		return nil, err
	}

	clientIPs := make(map[string]struct{})
	onlineClientIPs := make(map[string]struct{})
	for _, client := range clients {
		clientIP := strings.TrimSpace(client.ClientIp)
		if clientIP != "" {
			clientIPs[clientIP] = struct{}{}
		}

		online := strings.TrimSpace(client.Status) == "online"
		if online && clientIP != "" {
			onlineClientIPs[clientIP] = struct{}{}
		}
		if !online {
			continue
		}

		res.Nodes.Value++
		switch strings.TrimSpace(client.BusyStatus) {
		case "idle":
			res.NodeBusy.Value++
		case "busy":
			res.NodeBusy.Total++
		default:
			res.NodeBusy.Extra++
		}
		if strings.TrimSpace(client.PluginStatus) == "installed" {
			res.Automa.Value++
		}
	}

	res.Clients.Value = len(onlineClientIPs)
	res.Clients.Total = len(clientIPs)
	if res.Clients.Total > res.Clients.Value {
		res.Clients.Extra = res.Clients.Total - res.Clients.Value
	}
	res.Nodes.Total = len(clients)
	if res.Nodes.Total > res.Nodes.Value {
		res.Nodes.Extra = res.Nodes.Total - res.Nodes.Value
	}
	res.Automa.Total = res.Nodes.Value
	if res.Automa.Total > res.Automa.Value {
		res.Automa.Extra = res.Automa.Total - res.Automa.Value
	}

	workflowColumns := dao.AutomaWorkflows.Columns()
	workflows := []entity.AutomaWorkflows{}
	if err = dao.AutomaWorkflows.Ctx(ctx).Fields(workflowColumns.IsProtected).Scan(&workflows); err != nil {
		return nil, err
	}
	for _, workflow := range workflows {
		if workflow.IsProtected {
			res.Workflows.Extra++
		} else {
			res.Workflows.Value++
		}
	}
	res.Workflows.Total = len(workflows)

	taskColumns := dao.Tasks.Columns()
	tasks := []entity.Tasks{}
	if err = dao.Tasks.Ctx(ctx).Fields(taskColumns.Enabled).Scan(&tasks); err != nil {
		return nil, err
	}
	for _, task := range tasks {
		if task.Enabled {
			res.Tasks.Value++
		} else {
			res.Tasks.Extra++
		}
	}
	res.Tasks.Total = len(tasks)

	return res, nil
}
