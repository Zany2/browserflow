package browser

import (
	"time"

	"github.com/Zany2/browserflow/backend/utility/state"
)

const browserRuntimeProbeInterval = 2 * time.Second
const browserRuntimeProbeTimeout = 1 * time.Second
const browserRuntimeProbeFailureLimit = 2

// watchBrowserRuntime probes the browser CDP connection and clears stale runtime state. 监听浏览器连接，关闭后清理运行态。
func watchBrowserRuntime(instanceID string, runtime *state.BrowserRuntime) {
	if runtime == nil || runtime.Browser == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(browserRuntimeProbeInterval)
		defer ticker.Stop()

		failureCount := 0
		for range ticker.C {
			state.BrowserMu.Lock()
			currentRuntime, running := state.BrowserInstances[instanceID]
			state.BrowserMu.Unlock()
			if !running || currentRuntime != runtime {
				return
			}

			if browserRuntimeAlive(runtime) {
				failureCount = 0
				continue
			}

			failureCount += 1
			if failureCount >= browserRuntimeProbeFailureLimit {
				state.RemoveBrowserRuntime(instanceID, runtime)
				state.CleanupBrowserRuntime(runtime)
				return
			}
		}
	}()
}

func browserRuntimeAlive(runtime *state.BrowserRuntime) bool {
	if runtime == nil || runtime.Browser == nil {
		return false
	}

	// CDP probe 使用超时探测，避免浏览器关闭后请求长时间挂起
	browser := runtime.Browser.Timeout(browserRuntimeProbeTimeout)
	defer browser.CancelTimeout()
	_, err := browser.Version()
	return err == nil
}
