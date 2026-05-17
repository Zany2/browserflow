package workflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/util/guid"
)

// WorkflowRun runs workflow through browser agent 运行工作流
func (c *ControllerV1) WorkflowRun(ctx context.Context, req *v1.WorkflowRunReq) (res *v1.WorkflowRunRes, err error) {
	agent, browserID, err := resolveWorkflowRunAgent(req.BrowserID)
	if err != nil {
		return nil, err
	}

	commandID := "cmd_" + guid.S()
	executionID := workflowexecution.NewID()
	resultCh := make(chan model.AgentCommandResult, 1)

	execution := workflowexecution.Create(model.WorkflowExecutionCreateInput{
		ExecutionID: executionID,
		WorkflowID:  req.ID,
		BrowserID:   browserID,
		CommandID:   commandID,
		WaitResult:  req.WaitResult,
		ReturnData:  req.ReturnData,
	})

	state.AgentMu.Lock()
	state.PendingCommands[commandID] = resultCh
	agent.LastSeenAt = time.Now()
	state.AgentMu.Unlock()

	timeout := normalizeWorkflowRunTimeout(req.Timeout)
	returnData := workflowexecution.NormalizeReturnData(req.ReturnData)
	workflowexecution.MarkRunning(executionID)

	agent.Client.WriteMu.Lock()
	err = agent.Client.Conn.WriteJSON(model.WSResponse{
		Type:      "agent_command",
		BrowserID: agent.BrowserID,
		CommandID: commandID,
		Command:   "automa.workflow.run",
		Payload: map[string]any{
			"id":           req.ID,
			"variables":    req.Variables,
			"check_params": false,
			"execution_id": executionID,
			"wait_result":  req.WaitResult,
			"timeout":      timeout,
			"return_data":  returnData,
		},
	})
	agent.Client.WriteMu.Unlock()
	if err != nil {
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		workflowexecution.MarkTimeout(executionID, err.Error())
		return nil, err
	}

	if !req.WaitResult {
		execution, _ = workflowexecution.Get(executionID)
		return &v1.WorkflowRunRes{Execution: execution}, nil
	}

	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()

	select {
	case result := <-resultCh:
		workflowexecution.Complete(executionID, &result)
		execution, _ = workflowexecution.Get(executionID)
		return &v1.WorkflowRunRes{Result: &result, Execution: execution}, nil
	case <-timer.C:
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		workflowexecution.MarkTimeout(executionID, fmt.Sprintf("workflow execution timeout after %d seconds", timeout))
		execution, _ = workflowexecution.Get(executionID)
		return &v1.WorkflowRunRes{Execution: execution}, nil
	case <-ctx.Done():
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		workflowexecution.MarkTimeout(executionID, ctx.Err().Error())
		return nil, ctx.Err()
	}
}

// resolveWorkflowRunAgent resolves target browser agent 解析目标浏览器执行端
func resolveWorkflowRunAgent(browserID string) (*state.AgentConnection, string, error) {
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

// normalizeWorkflowRunTimeout normalizes wait timeout 规范化等待超时
func normalizeWorkflowRunTimeout(timeout int) int {
	if timeout <= 0 {
		return 300
	}
	if timeout < 5 {
		return 5
	}
	return timeout
}
