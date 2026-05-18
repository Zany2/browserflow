package workflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/util/guid"
)

// WorkflowRun runs workflow through browser agent 运行工作流
func (c *ControllerV1) WorkflowRun(ctx context.Context, req *v1.WorkflowRunReq) (res *v1.WorkflowRunRes, err error) {
	agent, browserID, err := workflowagent.ResolveRunAgent(req.BrowserID)
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

	timeout := workflowagent.NormalizeRunTimeout(req.Timeout)
	returnData := workflowexecution.NormalizeReturnData(req.ReturnData)
	workflowexecution.MarkRunning(executionID)

	if sent := websockets.SendConnectionMessage(agent.ConnectionID, &model.WSResponse{
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
	}); sent <= 0 {
		state.AgentMu.Lock()
		delete(state.PendingCommands, commandID)
		state.AgentMu.Unlock()
		workflowexecution.MarkTimeout(executionID, "browser agent is offline")
		return nil, errors.New("browser agent is offline")
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
