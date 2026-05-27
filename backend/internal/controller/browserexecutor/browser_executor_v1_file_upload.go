package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorFileUpload uploads files to a file input. 上传文件到文件输入框。
func (c *ControllerV1) BrowserExecutorFileUpload(ctx context.Context, req *v1.BrowserExecutorFileUploadReq) (res *v1.BrowserExecutorFileUploadRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.FileUpload(ctx, req.Identifier, req.FilePaths)
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorFileUploadRes{Result: result}, nil
}
