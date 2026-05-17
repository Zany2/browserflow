package workflows

import (
	"context"
	"errors"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
)

// WorkflowExecutionDetail gets workflow execution state 获取工作流执行状态
func (c *ControllerV1) WorkflowExecutionDetail(ctx context.Context, req *v1.WorkflowExecutionDetailReq) (res *v1.WorkflowExecutionDetailRes, err error) {
	execution, ok := workflowexecution.Get(req.ExecutionID)
	if !ok {
		return nil, errors.New("workflow execution not found")
	}

	return &v1.WorkflowExecutionDetailRes{Execution: execution}, nil
}
