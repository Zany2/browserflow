package workflows

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowExecutionDetail gets workflow execution state 获取工作流执行状态
func (c *ControllerV1) WorkflowExecutionDetail(ctx context.Context, req *v1.WorkflowExecutionDetailReq) (res *v1.WorkflowExecutionDetailRes, err error) {
	if !requireWindowsMode(ctx) {
		return nil, nil
	}

	execution, ok := workflowexecution.Get(req.ExecutionID)
	if !ok {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流执行记录不存在")
		return nil, nil
	}

	return &v1.WorkflowExecutionDetailRes{Execution: execution}, nil
}
