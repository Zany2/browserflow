package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserExecutorExportSkill exports a SKILL.md file. BrowserExecutorExportSkill 鐎电厧鍤?SKILL.md 閺傚洣娆㈤妴?
func (c *ControllerV1) BrowserExecutorExportSkill(ctx context.Context, req *v1.BrowserExecutorExportSkillReq) (res *v1.BrowserExecutorExportSkillRes, err error) {
	request := g.RequestFromCtx(ctx)
	executor, execErr := browserexecutor.Current(ctx)
	if execErr != nil {
		rr.FailedJsonWithMessageExitAll(request, execErr.Error())
		return nil, nil
	}
	skill := browserexecutor.GenerateSkill(browserexecutor.RequestBaseURL(request), executor.Status(ctx))
	request.Response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	request.Response.Header().Set("Content-Disposition", "attachment; filename=SKILL_BROWSER_EXECUTOR.md")
	request.Response.Write(skill)
	request.ExitAll()
	return nil, nil
}
