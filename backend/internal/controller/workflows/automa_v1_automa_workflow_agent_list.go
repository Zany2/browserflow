package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// WorkflowAgentList reads workflows from the selected browser agent 从指定浏览器执行端读取工作流
func (c *ControllerV1) WorkflowAgentList(ctx context.Context, req *v1.WorkflowAgentListReq) (res *v1.WorkflowAgentListRes, err error) {
	browserID, workflows, err := requestAgentWorkflowList(ctx, req.BrowserID)
	if err != nil {
		// Business guard 业务前置条件不满足时直接返回可读提示，避免被中间件当作服务异常
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), formatAgentWorkflowListError(err))
		return nil, nil
	}

	return &v1.WorkflowAgentListRes{
		BrowserID: browserID,
		Workflows: workflows,
		Total:     len(workflows),
	}, nil
}

// formatAgentWorkflowListError converts internal errors to user messages 转换执行端列表错误提示
func formatAgentWorkflowListError(err error) string {
	message := strings.TrimSpace(err.Error())
	switch message {
	case "no current browser running":
		return "请先在浏览器页面启动一个浏览器实例"
	case "browser agent is offline":
		return "当前浏览器执行端未连接，请确认 browser-agent 页面已打开"
	case "browser agent workflow list failed":
		return "读取浏览器执行端工作流失败"
	default:
		if message != "" {
			return message
		}
		return "读取浏览器执行端工作流失败"
	}
}

// requestAgentWorkflowList sends list command to browser agent 向浏览器执行端发送工作流列表命令
func requestAgentWorkflowList(ctx context.Context, browserID string) (string, []map[string]any, error) {
	agent, resolvedBrowserID, err := resolveBrowserAgent(browserID)
	if err != nil {
		return "", nil, err
	}

	commandID := "cmd_" + guid.S()
	resultCh := make(chan model.AgentCommandResult, 1)
	state.AgentMu.Lock()
	state.PendingCommands[commandID] = resultCh
	agent.LastSeenAt = time.Now()
	state.AgentMu.Unlock()

	agent.Client.WriteMu.Lock()
	err = agent.Client.Conn.WriteJSON(model.WSResponse{
		Type:      "agent_command",
		BrowserID: resolvedBrowserID,
		CommandID: commandID,
		Command:   "automa.workflow.list",
		Payload:   map[string]any{},
	})
	agent.Client.WriteMu.Unlock()
	if err != nil {
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		return "", nil, err
	}

	select {
	case result := <-resultCh:
		if !result.Success {
			if strings.TrimSpace(result.Error) != "" {
				return "", nil, errors.New(result.Error)
			}
			return "", nil, errors.New("browser agent workflow list failed")
		}
		workflows, parseErr := parseAgentWorkflowList(result.Data)
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

// resolveBrowserAgent chooses current or requested browser agent 选择当前或指定浏览器执行端
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

// parseAgentWorkflowList unwraps agent command result 解析执行端返回的工作流列表
func parseAgentWorkflowList(data json.RawMessage) ([]map[string]any, error) {
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
