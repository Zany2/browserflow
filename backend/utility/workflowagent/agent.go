package workflowagent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/util/guid"
)

// FormatWorkflowListError converts internal errors to user messages. 转换工作流列表错误为用户提示
func FormatWorkflowListError(err error) string {
	message := strings.TrimSpace(err.Error())
	switch message {
	case "no current browser running":
		return "当前没有运行中的浏览器，请先启动浏览器"
	case "browser agent is offline":
		return "浏览器执行端未在线，请确认 browser-agent 页面已打开"
	case "browser agent automa unavailable":
		return "未检测到 Automa，请确认扩展已安装并启用"
	case "browser agent workflow list failed":
		return "读取工作流列表失败"
	default:
		if message != "" {
			return message
		}
		return "读取工作流列表失败"
	}
}

// RequestWorkflowList sends list command to browser agent. 向浏览器执行端请求工作流列表
func RequestWorkflowList(ctx context.Context, browserID string) (string, []map[string]any, error) {
	agent, resolvedBrowserID, err := resolveBrowserAgent(browserID)
	if err != nil {
		return "", nil, err
	}
	if !agent.AutomaInstalled {
		return "", nil, errors.New("browser agent automa unavailable")
	}

	commandID := "cmd_" + guid.S()
	resultCh := make(chan model.AgentCommandResult, 1)
	state.AgentMu.Lock()
	state.PendingCommands[commandID] = resultCh
	agent.LastSeenAt = time.Now()
	state.AgentMu.Unlock()

	if sent := websockets.SendConnectionMessage(agent.ConnectionID, &model.WSResponse{
		Type:      "agent_command",
		BrowserID: resolvedBrowserID,
		CommandID: commandID,
		Command:   "automa.workflow.list",
		Payload:   map[string]any{},
	}); sent <= 0 {
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		return "", nil, errors.New("browser agent is offline")
	}

	select {
	case result := <-resultCh:
		if !result.Success {
			if strings.TrimSpace(result.Error) != "" {
				return "", nil, errors.New(result.Error)
			}
			return "", nil, errors.New("browser agent workflow list failed")
		}
		workflows, parseErr := parseWorkflowList(result.Data)
		if parseErr != nil {
			return "", nil, parseErr
		}
		return resolvedBrowserID, workflows, nil
	case <-ctx.Done():
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		return "", nil, ctx.Err()
	}
}

// ResolveRunAgent resolves target browser agent. 解析目标浏览器执行端
func ResolveRunAgent(browserID string) (*state.AgentConnection, string, error) {
	state.AgentMu.Lock()
	defer state.AgentMu.Unlock()

	resolvedBrowserID := browserID
	var agent *state.AgentConnection
	if resolvedBrowserID != "" {
		agent = state.AgentConnections[resolvedBrowserID]
		if agent == nil {
			return nil, "", errors.New("browser agent is offline")
		}
		return agent, resolvedBrowserID, nil
	}

	for id, item := range state.AgentConnections {
		agent = item
		resolvedBrowserID = id
		break
	}
	if agent == nil {
		return nil, "", errors.New("no browser agent online")
	}
	return agent, resolvedBrowserID, nil
}

// NormalizeRunTimeout normalizes wait timeout. 规范化等待超时时间
func NormalizeRunTimeout(timeout int) int {
	if timeout <= 0 {
		return 300
	}
	if timeout < 5 {
		return 5
	}
	return timeout
}

// resolveBrowserAgent chooses current or requested browser agent. 选择当前或指定的浏览器执行端
func resolveBrowserAgent(browserID string) (*state.AgentConnection, string, error) {
	resolvedBrowserID := strings.TrimSpace(browserID)
	if resolvedBrowserID == "" {
		state.BrowserMu.Lock()
		resolvedBrowserID = strings.TrimSpace(state.BrowserCurrentInstanceID)
		state.BrowserMu.Unlock()
	}
	if resolvedBrowserID == "" {
		return nil, "", errors.New("no current browser running")
	}

	state.AgentMu.Lock()
	defer state.AgentMu.Unlock()
	agent := state.AgentConnections[resolvedBrowserID]
	if agent == nil {
		return nil, "", errors.New("browser agent is offline")
	}
	return agent, resolvedBrowserID, nil
}

// parseWorkflowList unwraps agent command result. 解包执行端返回的工作流列表
func parseWorkflowList(data json.RawMessage) ([]map[string]any, error) {
	if len(data) == 0 {
		return []map[string]any{}, nil
	}

	var wrapped struct {
		Workflows json.RawMessage `json:"workflows"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && len(wrapped.Workflows) > 0 {
		data = wrapped.Workflows
	}

	var workflows []map[string]any
	if err := json.Unmarshal(data, &workflows); err != nil {
		return nil, err
	}
	if workflows == nil {
		return []map[string]any{}, nil
	}
	return workflows, nil
}
