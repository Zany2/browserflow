package browser

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
)

// BrowserStop stops current browser instance 停止当前浏览器实例
func (c *ControllerV1) BrowserStop(ctx context.Context, req *v1.BrowserStopReq) (res *v1.BrowserStopRes, err error) {
	state.BrowserMu.Lock()
	instanceID := state.BrowserCurrentInstanceID
	runtime, ok := state.BrowserInstances[instanceID]
	if !ok {
		state.BrowserMu.Unlock()
		return nil, errors.New("浏览器实例未运行")
	}
	delete(state.BrowserInstances, instanceID)
	state.BrowserCurrentInstanceID = ""
	for id := range state.BrowserInstances {
		state.BrowserCurrentInstanceID = id
		break
	}
	nextRuntime := state.BrowserInstances[state.BrowserCurrentInstanceID]
	status := model.BrowserStatus{Running: false}
	if nextRuntime != nil {
		startTime := nextRuntime.StartTime
		instanceCopy := *nextRuntime.Instance
		instanceCopy.IsActive = true
		status = model.BrowserStatus{Running: true, CurrentInstanceID: state.BrowserCurrentInstanceID, Instance: &instanceCopy, StartTime: &startTime, UptimeSeconds: int64(time.Since(startTime).Seconds()), ControlURL: nextRuntime.ControlURL, AgentURL: nextRuntime.AgentURL}
	}
	for listener := range state.BrowserStatusListeners {
		select {
		case listener <- status:
		default:
		}
	}
	state.BrowserMu.Unlock()

	if runtime.Browser != nil {
		_ = runtime.Browser.Close()
	}
	if runtime.Launcher != nil {
		if runtime.Instance != nil && strings.TrimSpace(runtime.Instance.UserDataDir) != "" {
			runtime.Launcher.Kill()
		} else {
			runtime.Launcher.Cleanup()
		}
	}
	return &v1.BrowserStopRes{Status: &status}, nil
}
