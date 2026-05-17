package browserexecutor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
)

var (
	executorMu sync.Mutex
	executors  = map[string]*Executor{}
)

// Current returns the executor for the current browser instance. Current returns executor for current browser instance.
func Current(ctx context.Context) (*Executor, error) {
	state.BrowserMu.Lock()
	browserID := strings.TrimSpace(state.BrowserCurrentInstanceID)
	runtime := state.BrowserInstances[browserID]
	state.BrowserMu.Unlock()

	if browserID == "" || runtime == nil || runtime.Browser == nil {
		return nil, fmt.Errorf("当前没有运行中的浏览器实例")
	}

	executorMu.Lock()
	defer executorMu.Unlock()

	executor := executors[browserID]
	if executor == nil || executor.runtime != runtime {
		executor = NewExecutor(browserID, runtime)
		executors[browserID] = executor
	}
	return executor.WithContext(ctx), nil
}

// CurrentStatus returns executor availability without creating page actions. CurrentStatus returns executor availability.
func CurrentStatus(ctx context.Context) model.BrowserExecutorStatus {
	executor, err := Current(ctx)
	if err != nil {
		return model.BrowserExecutorStatus{Running: false, Message: err.Error()}
	}
	return executor.Status(ctx)
}

// Cleanup removes cached executor for a browser instance. Cleanup removes cached executor.
func Cleanup(browserID string) {
	executorMu.Lock()
	defer executorMu.Unlock()
	delete(executors, strings.TrimSpace(browserID))
}
