package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserExecutorExportSkill exports a SKILL.md file. 导出浏览器控制 Skill 文件
func (c *ControllerV1) BrowserExecutorExportSkill(ctx context.Context, req *v1.BrowserExecutorExportSkillReq) (res *v1.BrowserExecutorExportSkillRes, err error) {
	request := g.RequestFromCtx(ctx)
	// Skill export is static and should not include current browser runtime state. Skill 是静态说明，不写入当前浏览器运行状态。
	skill := browserexecutor.GenerateSkill(browserexecutor.RequestBaseURL(request))
	request.Response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	request.Response.Header().Set("Content-Disposition", "attachment; filename=SKILL_BROWSER_EXECUTOR.md")
	request.Response.Write(skill)
	request.ExitAll()
	return nil, nil
}
