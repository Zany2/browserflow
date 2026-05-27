package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorCookies manages browser cookies. 管理浏览器 Cookie。
func (c *ControllerV1) BrowserExecutorCookies(ctx context.Context, req *v1.BrowserExecutorCookiesReq) (res *v1.BrowserExecutorCookiesRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		if failBrowserExecutor(ctx, nil, err) {
			return nil, nil
		}
		return nil, err
	}
	result, opErr := executor.Cookies(ctx, model.BrowserExecutorCookieOptions{
		Action:   req.Action,
		Name:     req.Name,
		Value:    req.Value,
		URL:      req.URL,
		Domain:   req.Domain,
		Path:     req.Path,
		Secure:   req.Secure,
		HTTPOnly: req.HTTPOnly,
		SameSite: req.SameSite,
		Expires:  req.Expires,
	})
	if opErr != nil {
		if failBrowserExecutor(ctx, result, opErr) {
			return nil, nil
		}
		return nil, opErr
	}
	return &v1.BrowserExecutorCookiesRes{Result: result}, nil
}
