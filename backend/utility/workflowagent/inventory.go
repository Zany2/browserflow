package workflowagent

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"
)

const (
	// workflowInventoryRefreshCommand asks client-agent to report current Automa workflows. 请求客户端刷新工作流清单
	workflowInventoryRefreshCommand = "automa.workflow.inventory.refresh"
)

// RefreshClientInventory refreshes one online client inventory. 刷新单个在线客户端工作流清单
func RefreshClientInventory(ctx context.Context, clientIP string, timeout time.Duration) error {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return nil
	}
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	if !workflowcache.IsClientOnline(ctx, clientIP) {
		return gerror.Newf("客户端 %s 不在线或 WebSocket 未连接", clientIP)
	}

	commandID := "cmd_" + guid.S()
	resultCh := make(chan model.AgentCommandResult, 1)
	state.SetPendingCommand(commandID, resultCh)

	// Send forced inventory refresh command. 下发强制刷新清单命令
	node, ok, nodeErr := workflowcache.GetOnlineNode(ctx, clientIP)
	if nodeErr != nil {
		state.RemovePendingCommand(commandID)
		return nodeErr
	}
	if !ok {
		node.NodeID = clientIP
		node.ClientIP = clientIP
	}
	if sent := websockets.SendNodeMessage(node.NodeID, node.ClientIP, &model.WSResponse{
		Type:      model.WSMessageTypeAgentCommand,
		ClientIP:  node.ClientIP,
		MachineID: node.MachineID,
		NodeID:    node.NodeID,
		NodeName:  node.NodeName,
		CommandID: commandID,
		Command:   workflowInventoryRefreshCommand,
		Payload: map[string]any{
			"force": true,
		},
	}); sent <= 0 {
		state.RemovePendingCommand(commandID)
		return gerror.Newf("客户端 %s 不在线或 WebSocket 未连接", clientIP)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case result := <-resultCh:
		if result.Success {
			return nil
		}
		if message := strings.TrimSpace(result.Error); message != "" {
			return gerror.New(message)
		}
		return gerror.Newf("客户端 %s 刷新工作流清单失败", clientIP)
	case <-timer.C:
		state.RemovePendingCommand(commandID)
		return gerror.Newf("客户端 %s 刷新工作流清单超时", clientIP)
	case <-ctx.Done():
		state.RemovePendingCommand(commandID)
		return ctx.Err()
	}
}

// RefreshClientInventories refreshes multiple client inventories. 批量刷新客户端工作流清单
func RefreshClientInventories(ctx context.Context, clientIPs []string, timeout time.Duration) error {
	uniqueClientIPs := normalizeClientIPs(clientIPs)
	if len(uniqueClientIPs) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(uniqueClientIPs))
	for _, clientIP := range uniqueClientIPs {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			errCh <- RefreshClientInventory(ctx, ip, timeout)
		}(clientIP)
	}

	wg.Wait()
	close(errCh)

	messages := make([]string, 0)
	for err := range errCh {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}
	if len(messages) > 0 {
		return gerror.New(strings.Join(messages, "；"))
	}
	return nil
}

// normalizeClientIPs deduplicates client ip list. 去重客户端 IP 列表
func normalizeClientIPs(clientIPs []string) []string {
	seen := make(map[string]struct{}, len(clientIPs))
	normalized := make([]string, 0, len(clientIPs))
	for _, clientIP := range clientIPs {
		clientIP = strings.TrimSpace(clientIP)
		if clientIP == "" {
			continue
		}
		if _, ok := seen[clientIP]; ok {
			continue
		}
		seen[clientIP] = struct{}{}
		normalized = append(normalized, clientIP)
	}
	return normalized
}
