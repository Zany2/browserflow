package browserexecutor

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
)

// failBrowserExecutor writes readable business errors for LLM callers.
func failBrowserExecutor(ctx context.Context, result interface{}, err error) bool {
	if err == nil {
		return false
	}

	message := strings.TrimSpace(err.Error())
	switch typed := result.(type) {
	case *model.BrowserExecutorOperationResult:
		if typed != nil && strings.TrimSpace(typed.Error) != "" {
			message = strings.TrimSpace(typed.Error)
		}
	case *model.BrowserExecutorSnapshotResult:
		if typed != nil && strings.TrimSpace(typed.Error) != "" {
			message = strings.TrimSpace(typed.Error)
		}
	case *model.BrowserExecutorObserveResult:
		if typed != nil && strings.TrimSpace(typed.Error) != "" {
			message = strings.TrimSpace(typed.Error)
		}
	}
	if message == "" {
		message = "浏览器操作失败"
	}
	if !isBrowserExecutorReadableError(message) {
		return false
	}
	rr.FailedJsonWithMessageAndDataExitAll(g.RequestFromCtx(ctx), message, g.Map{
		"result": result,
	})
	return true
}

// isBrowserExecutorReadableError keeps expected browser-operation errors visible.
func isBrowserExecutorReadableError(message string) bool {
	readablePatterns := []string{
		"当前没有运行中的浏览器实例",
		"browser runtime is unavailable",
		"no controllable business page",
		"no BrowserFlow browser-agent page",
		"identifier",
		"element not found",
		"RefID",
		"BackendNodeID",
		"form field",
		"form fields",
		"refusing to",
		"file paths cannot be empty",
		"cookie name cannot be empty",
		"unknown cookies action",
		"storage key cannot be empty",
		"tab index out of range",
		"unknown tabs action",
		"unknown mouse action",
		"unknown window action",
		"unknown act intent",
		"unknown batch operation",
		"element has no visible box",
		"Accessibility Tree is empty",
	}
	for _, pattern := range readablePatterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}
