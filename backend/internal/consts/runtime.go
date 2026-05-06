package consts

import (
	"context"
	"runtime"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	RuntimeModeServer  = "server"  // Server mode 服务端模式
	RuntimeModeWindows = "windows" // Windows mode 本地桌面模式
)

// ResolveRuntimeMode reads configured runtime mode with os fallback. 读取配置运行模式，配置为空时按系统兜底。
func ResolveRuntimeMode(ctx context.Context) string {
	mode := strings.TrimSpace(g.Cfg().MustGet(ctx, "app.mode", "").String())
	if mode == "" {
		mode = RuntimeModeServer
		if runtime.GOOS == RuntimeModeWindows {
			mode = RuntimeModeWindows
		}
	}

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case RuntimeModeServer:
		return RuntimeModeServer
	case RuntimeModeWindows, "window", "desktop", "local":
		return RuntimeModeWindows
	default:
		return RuntimeModeWindows
	}
}
