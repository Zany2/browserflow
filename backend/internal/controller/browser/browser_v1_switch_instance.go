package browser

import (
	"context"
	"time"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserInstanceSwitch switches current browser instance 切换当前浏览器实例
func (c *ControllerV1) BrowserInstanceSwitch(ctx context.Context, req *v1.BrowserInstanceSwitchReq) (res *v1.BrowserInstanceSwitchRes, err error) {
	state.BrowserMu.Lock()
	runtime, ok := state.BrowserInstances[req.ID]
	if !ok {
		state.BrowserMu.Unlock()
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "浏览器实例未运行")
		return nil, nil
	}
	state.BrowserCurrentInstanceID = req.ID
	startTime := runtime.StartTime
	instanceCopy := *runtime.Instance
	instanceCopy.IsActive = true
	status := model.BrowserStatus{Running: true, CurrentInstanceID: state.BrowserCurrentInstanceID, Instance: &instanceCopy, StartTime: &startTime, UptimeSeconds: int64(time.Since(startTime).Seconds()), ControlURL: runtime.ControlURL, AgentURL: runtime.AgentURL}
	for listener := range state.BrowserStatusListeners {
		select {
		case listener <- status:
		default:
		}
	}
	state.BrowserMu.Unlock()
	return &v1.BrowserInstanceSwitchRes{Status: &status}, nil
}
