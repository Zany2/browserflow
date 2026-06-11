package browserruntime

import (
	"time"

	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
	"github.com/Zany2/browserflow/backend/utility/state"
)

const browserRuntimeProbeInterval = 2 * time.Second
const browserRuntimeProbeTimeout = 1 * time.Second
const browserRuntimeProbeFailureLimit = 2

// Watch probes the browser CDP connection and clears stale runtime state. 探测浏览器连接并清理失效运行态。
func Watch(instanceID string, runtime *state.BrowserRuntime) {
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

			if runtimeAlive(runtime) {
				failureCount = 0
				continue
			}

			failureCount += 1
			if failureCount >= browserRuntimeProbeFailureLimit {
				state.RemoveBrowserRuntime(instanceID, runtime)
				browserexecutor.Cleanup(instanceID)
				state.CleanupBrowserRuntime(runtime)
				return
			}
		}
	}()
}

// runtimeAlive checks whether CDP still responds. 检查 CDP 是否仍可响应。
func runtimeAlive(runtime *state.BrowserRuntime) bool {
	if runtime == nil || runtime.Browser == nil {
		return false
	}

	// CDP probe uses timeout to avoid hanging after browser shutdown. 使用超时避免浏览器关闭后探测阻塞。
	browser := runtime.Browser.Timeout(browserRuntimeProbeTimeout)
	defer browser.CancelTimeout()
	_, err := browser.Version()
	return err == nil
}
