package state

import (
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
)

// CurrentBrowserStatusLocked builds browser status while BrowserMu is held. 构造当前浏览器运行状态，调用方需持有 BrowserMu。
func CurrentBrowserStatusLocked() model.BrowserStatus {
	runtime := BrowserInstances[BrowserCurrentInstanceID]
	status := model.BrowserStatus{Running: false}
	if runtime == nil {
		return status
	}

	startTime := runtime.StartTime
	instanceCopy := *runtime.Instance
	instanceCopy.IsActive = true
	return model.BrowserStatus{
		Running:           true,
		CurrentInstanceID: BrowserCurrentInstanceID,
		Instance:          &instanceCopy,
		StartTime:         &startTime,
		UptimeSeconds:     int64(time.Since(startTime).Seconds()),
		ControlURL:        runtime.ControlURL,
		AgentURL:          runtime.AgentURL,
	}
}

// BroadcastBrowserStatusLocked sends status to browser subscribers. 广播浏览器状态，调用方需持有 BrowserMu。
func BroadcastBrowserStatusLocked(status model.BrowserStatus) {
	for listener := range BrowserStatusListeners {
		select {
		case listener <- status:
		default:
		}
	}
}

// RemoveBrowserRuntime removes a browser runtime and broadcasts the next current status. 移除浏览器运行态并广播最新状态。
func RemoveBrowserRuntime(instanceID string, expected *BrowserRuntime) (*BrowserRuntime, model.BrowserStatus, bool) {
	BrowserMu.Lock()
	currentRuntime, running := BrowserInstances[instanceID]
	if !running || (expected != nil && currentRuntime != expected) {
		BrowserMu.Unlock()
		return nil, model.BrowserStatus{Running: false}, false
	}

	delete(BrowserInstances, instanceID)
	if BrowserCurrentInstanceID == instanceID {
		BrowserCurrentInstanceID = ""
		for id := range BrowserInstances {
			BrowserCurrentInstanceID = id
			break
		}
	}

	status := CurrentBrowserStatusLocked()
	BroadcastBrowserStatusLocked(status)
	BrowserMu.Unlock()
	return currentRuntime, status, true
}

// CleanupBrowserRuntime releases launcher resources for a closed browser. 释放已关闭浏览器的启动器资源。
func CleanupBrowserRuntime(runtime *BrowserRuntime) {
	if runtime == nil || runtime.Launcher == nil {
		return
	}

	if runtime.Instance != nil && strings.TrimSpace(runtime.Instance.UserDataDir) != "" {
		runtime.Launcher.Kill()
		return
	}
	runtime.Launcher.Cleanup()
}
