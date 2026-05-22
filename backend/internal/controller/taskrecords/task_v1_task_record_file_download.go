package taskrecords

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const taskRecordDownloadBaseDir = "data/task-record-files"

// TaskRecordFileDownload downloads one task record result file.
func (c *ControllerV1) TaskRecordFileDownload(ctx context.Context, req *v1.TaskRecordFileDownloadReq) (res *v1.TaskRecordFileDownloadRes, err error) {
	fileID := gconv.Int64(req.ID)
	if fileID <= 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "结果文件不存在")
		return nil, nil
	}

	columns := dao.TaskRecordFiles.Columns()
	record, err := dao.TaskRecordFiles.Ctx(ctx).
		WherePri(fileID).
		Where(columns.DeletedAt + " IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "结果文件不存在")
		return nil, nil
	}

	filePath := strings.TrimSpace(gconv.String(record[columns.FilePath]))
	safePath, err := resolveTaskRecordDownloadPath(filePath)
	if err != nil {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "结果文件路径无效")
		return nil, nil
	}
	info, err := os.Stat(safePath)
	if err != nil || info.IsDir() {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "结果文件不存在或已被清理")
		return nil, nil
	}

	fileName := strings.TrimSpace(gconv.String(record[columns.FileName]))
	if fileName == "" {
		fileName = filepath.Base(safePath)
	}
	mimeType := strings.TrimSpace(gconv.String(record[columns.MimeType]))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	request := g.RequestFromCtx(ctx)
	request.Response.Header().Set("Content-Type", mimeType)
	request.Response.Header().Set("Content-Disposition", taskRecordDownloadDisposition(fileName))
	request.Response.ServeFile(safePath)
	return nil, nil
}

func resolveTaskRecordDownloadPath(filePath string) (string, error) {
	filePath = filepath.Clean(filepath.FromSlash(strings.TrimSpace(filePath)))
	if filePath == "." || filepath.IsAbs(filePath) {
		return "", fmt.Errorf("invalid file path")
	}

	workspace, err := os.Getwd()
	if err != nil {
		return "", err
	}
	baseDir, err := filepath.Abs(filepath.Join(workspace, taskRecordDownloadBaseDir))
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(workspace, filePath))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(baseDir, target)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("file path outside task record result directory")
	}
	return target, nil
}

func taskRecordDownloadDisposition(fileName string) string {
	escapedName := strings.ReplaceAll(fileName, `"`, "")
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, escapedName, url.PathEscape(fileName))
}
