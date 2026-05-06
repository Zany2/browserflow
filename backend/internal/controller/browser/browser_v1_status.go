package browser

import (
	"context"
	"time"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
)

// BrowserStatus returns browser runtime status 获取浏览器运行状态
func (c *ControllerV1) BrowserStatus(ctx context.Context, req *v1.BrowserStatusReq) (res *v1.BrowserStatusRes, err error) {
	state.BrowserMu.Lock()
	runtime := state.BrowserInstances[state.BrowserCurrentInstanceID]
	status := model.BrowserStatus{Running: false}
	if runtime != nil {
		startTime := runtime.StartTime
		instanceCopy := *runtime.Instance
		instanceCopy.IsActive = true
		status = model.BrowserStatus{
			Running:           true,
			CurrentInstanceID: state.BrowserCurrentInstanceID,
			Instance:          &instanceCopy,
			StartTime:         &startTime,
			UptimeSeconds:     int64(time.Since(startTime).Seconds()),
			ControlURL:        runtime.ControlURL,
			AgentURL:          runtime.AgentURL,
		}
	}
	state.BrowserMu.Unlock()
	return &v1.BrowserStatusRes{Status: &status}, nil
}
