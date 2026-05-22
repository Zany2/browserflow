package workflows

import (
	"context"

	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
)

// requireWindowsMode blocks local browser-agent APIs outside Windows mode. 阻止本地 browser-agent 接口在非 Windows 模式调用。
func requireWindowsMode(ctx context.Context) bool {
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeWindows {
		return true
	}

	rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "当前接口仅支持 Windows 模式")
	return false
}
