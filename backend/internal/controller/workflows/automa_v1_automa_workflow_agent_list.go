package workflows

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowAgentList reads workflows from the selected browser agent 从指定浏览器执行端读取工作流
func (c *ControllerV1) WorkflowAgentList(ctx context.Context, req *v1.WorkflowAgentListReq) (res *v1.WorkflowAgentListRes, err error) {
	browserID, workflows, err := workflowagent.RequestWorkflowList(ctx, req.BrowserID)
	if err != nil {
		// Business guard 业务前置条件不满足时直接返回可读提示，避免被中间件当作服务异常
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), workflowagent.FormatWorkflowListError(err))
		return nil, nil
	}

	return &v1.WorkflowAgentListRes{
		BrowserID: browserID,
		Workflows: workflows,
		Total:     len(workflows),
	}, nil
}
