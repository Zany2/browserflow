package browserexecutor

import (
	"archive/zip"
	"bytes"
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserExecutorExportSkill exports a browser executor skill zip bundle. 导出浏览器控制 Skill 压缩包。
func (c *ControllerV1) BrowserExecutorExportSkill(ctx context.Context, req *v1.BrowserExecutorExportSkillReq) (res *v1.BrowserExecutorExportSkillRes, err error) {
	request := g.RequestFromCtx(ctx)
	// Skill export is static and should not include current browser runtime state. Skill 是静态说明，不写入当前浏览器运行状态。
	bundle := browserexecutor.GenerateSkillBundle(browserexecutor.RequestBaseURL(request))

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, file := range bundle {
		entry, createErr := writer.Create(file.Path)
		if createErr != nil {
			_ = writer.Close()
			return nil, createErr
		}
		if _, writeErr := entry.Write([]byte(file.Content)); writeErr != nil {
			_ = writer.Close()
			return nil, writeErr
		}
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	request.Response.Header().Set("Content-Type", "application/zip")
	request.Response.Header().Set("Content-Disposition", "attachment; filename=browserflow-browser-executor.zip")
	request.Response.Write(buffer.Bytes())
	request.ExitAll()
	return nil, nil
}
