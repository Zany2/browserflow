package automa

import (
	"context"
	"errors"
	"time"

	"github.com/Zany2/browserflow/backend/api/automa/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/gogf/gf/v2/util/guid"
)

// AutomaWorkflowOpen opens workflow through browser agent 打开工作流
func (c *ControllerV1) AutomaWorkflowOpen(ctx context.Context, req *v1.AutomaWorkflowOpenReq) (res *v1.AutomaWorkflowOpenRes, err error) {
	state.AgentMu.Lock()
	var agent *state.AgentConnection
	if req.BrowserID != "" {
		agent = state.AgentConnections[req.BrowserID]
		if agent == nil {
			state.AgentMu.Unlock()
			return nil, errors.New("browser agent is offline")
		}
	} else {
		for _, item := range state.AgentConnections {
			agent = item
			break
		}
		if agent == nil {
			state.AgentMu.Unlock()
			return nil, errors.New("no browser agent online")
		}
	}
	commandID := "cmd_" + guid.S()
	resultCh := make(chan model.AgentCommandResult, 1)
	state.PendingCommands[commandID] = resultCh
	agent.LastSeenAt = time.Now()
	state.AgentMu.Unlock()

	agent.Client.WriteMu.Lock()
	err = agent.Client.Conn.WriteJSON(model.WSResponse{Type: "agent_command", BrowserID: agent.BrowserID, CommandID: commandID, Command: "automa.workflow.open", Payload: map[string]any{"id": req.ID}})
	agent.Client.WriteMu.Unlock()
	if err != nil {
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		return nil, err
	}
	select {
	case result := <-resultCh:
		return &v1.AutomaWorkflowOpenRes{Result: &result}, nil
	case <-ctx.Done():
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		return nil, ctx.Err()
	}
}
