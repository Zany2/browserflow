package workflows

import (
	"context"
	"errors"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// WorkflowOpen opens workflow through browser agent 打开工作流
func (c *ControllerV1) WorkflowOpen(ctx context.Context, req *v1.WorkflowOpenReq) (res *v1.WorkflowOpenRes, err error) {
	if !requireWindowsMode(ctx) {
		return nil, nil
	}

	state.AgentMu.Lock()
	var agent *state.AgentConnection
	if req.BrowserID != "" {
		agent = state.AgentConnections[req.BrowserID]
		if agent == nil {
			state.AgentMu.Unlock()
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), workflowagent.FormatWorkflowListError(errors.New("browser agent is offline")))
			return nil, nil
		}
	} else {
		for _, item := range state.AgentConnections {
			agent = item
			break
		}
		if agent == nil {
			state.AgentMu.Unlock()
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "没有在线的浏览器执行端")
			return nil, nil
		}
	}
	commandID := "cmd_" + guid.S()
	resultCh := make(chan model.AgentCommandResult, 1)
	state.SetPendingCommand(commandID, resultCh)
	agent.LastSeenAt = time.Now()
	state.AgentMu.Unlock()

	if sent := websockets.SendConnectionMessage(agent.ConnectionID, &model.WSResponse{Type: "agent_command", BrowserID: agent.BrowserID, CommandID: commandID, Command: "automa.workflow.open", Payload: map[string]any{"id": req.ID}}); sent <= 0 {
		state.AgentMu.Lock()
		state.RemovePendingCommand(commandID)
		state.AgentMu.Unlock()
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), workflowagent.FormatWorkflowListError(errors.New("browser agent is offline")))
		return nil, nil
	}
	select {
	case result := <-resultCh:
		return &v1.WorkflowOpenRes{Result: &result}, nil
	case <-ctx.Done():
		state.AgentMu.Lock()
		state.RemovePendingCommand(commandID)
		state.AgentMu.Unlock()
		return nil, ctx.Err()
	}
}
