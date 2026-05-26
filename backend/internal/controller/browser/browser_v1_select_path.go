package browser

import (
	"context"
	"runtime"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserSelectBinPath opens a native file picker for browser executable. 打开原生文件选择器选择浏览器程序。
func (c *ControllerV1) BrowserSelectBinPath(ctx context.Context, req *v1.BrowserSelectBinPathReq) (res *v1.BrowserPathSelectRes, err error) {
	if runtime.GOOS != "windows" {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "当前功能仅支持 Windows 模式")
		return nil, nil
	}

	path, ok := selectBrowserBinPath()
	if !ok {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "已取消选择浏览器路径")
		return nil, nil
	}
	return &v1.BrowserPathSelectRes{Path: path}, nil
}

// BrowserSelectUserDataDir opens a native folder picker for browser data. 打开原生目录选择器选择用户数据目录。
func (c *ControllerV1) BrowserSelectUserDataDir(ctx context.Context, req *v1.BrowserSelectUserDataDirReq) (res *v1.BrowserPathSelectRes, err error) {
	if runtime.GOOS != "windows" {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "当前功能仅支持 Windows 模式")
		return nil, nil
	}

	path, ok := selectUserDataDir()
	if !ok {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "已取消选择用户目录")
		return nil, nil
	}
	return &v1.BrowserPathSelectRes{Path: path}, nil
}
