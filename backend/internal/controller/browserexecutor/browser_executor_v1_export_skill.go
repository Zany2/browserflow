package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserExecutorExportSkill exports a SKILL.md file. BrowserExecutorExportSkill 导出 SKILL.md 文件
func (c *ControllerV1) BrowserExecutorExportSkill(ctx context.Context, req *v1.BrowserExecutorExportSkillReq) (res *v1.BrowserExecutorExportSkillRes, err error) {
	request := g.RequestFromCtx(ctx)
	// Skill export 导出仅生成控制说明，运行状态在 Skill 执行前再检查
	skill := browserexecutor.GenerateSkill(browserexecutor.RequestBaseURL(request), browserexecutor.CurrentStatus(ctx))
	request.Response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	request.Response.Header().Set("Content-Disposition", "attachment; filename=SKILL_BROWSER_EXECUTOR.md")
	request.Response.Write(skill)
	request.ExitAll()
	return nil, nil
}
