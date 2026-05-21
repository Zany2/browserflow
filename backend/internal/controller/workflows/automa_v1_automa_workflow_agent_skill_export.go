package workflows

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/Zany2/browserflow/backend/utility/workflowskill"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowAgentExportSkill exports agent workflows as SKILL.md 导出执行端工作流为 SKILL.md
func (c *ControllerV1) WorkflowAgentExportSkill(ctx context.Context, req *v1.WorkflowAgentExportSkillReq) (res *v1.WorkflowAgentExportSkillRes, err error) {
	browserID, workflows, err := workflowagent.RequestWorkflowList(ctx, req.BrowserID)
	if err != nil {
		// Business guard 业务前置条件不满足时返回可读提示
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), workflowagent.FormatWorkflowListError(err))
		return nil, nil
	}

	workflows = workflowskill.FilterWorkflows(workflows, req.Scope, req.WorkflowIDs)
	if len(workflows) == 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "没有可导出的工作流")
		return nil, nil
	}

	request := g.RequestFromCtx(ctx)

	// Raw markdown response 原始 Markdown 响应，绕过统一 JSON 包装
	request.Response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	request.Response.Header().Set("Content-Disposition", workflowskill.ContentDisposition(workflowskill.FileName))
	request.Response.Write(workflowskill.GenerateMarkdown(workflows, workflowskill.BaseURL(request.Host, request.TLS != nil), browserID))

	request.ExitAll()
	return nil, nil
}
