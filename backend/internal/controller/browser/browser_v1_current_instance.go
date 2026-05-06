package browser

import (
	"context"
	"time"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
)

// BrowserInstanceCurrent returns current browser instance 获取当前浏览器实例
func (c *ControllerV1) BrowserInstanceCurrent(ctx context.Context, req *v1.BrowserInstanceCurrentReq) (res *v1.BrowserInstanceCurrentRes, err error) {
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
	return &v1.BrowserInstanceCurrentRes{Status: &status, Instance: status.Instance}, nil
}
