package websockets

import (
	"context"
	"sync"
)

// NodeRegisterData describes one client node registration event.
type NodeRegisterData struct {
	ClientIP       string
	NodeID         string
	NodeIndex      int
	WorkerVersion  string
	ProfileDir     string
	ExtensionDir   string
	UserAgent      string
	PluginStatus   string
	AutomaVersion  string
	BrowserName    string
	BrowserVersion string
	OsName         string
	OsVersion      string
	Hostname       string
	Capabilities   map[string]any
}

// NodeHeartbeatData describes one client node activity event.
type NodeHeartbeatData struct {
	ClientIP         string
	NodeID           string
	CommandID        string
	RecordID         int64
	PluginStatus     string
	AutomaVersion    string
	UpdatePluginInfo bool
}

// TaskRecordResultData describes a task execution result reported by a node.
type TaskRecordResultData struct {
	RecordID  int64
	NodeID    string
	ClientIP  string
	CommandID string
	Success   bool
	Result    []byte
	ErrorText string
}

// Persistence persists server-mode websocket events outside the utility package.
type Persistence interface {
	CanRecoverTask(ctx context.Context, recordID int64) (bool, error)
	SaveClientRegister(ctx context.Context, data NodeRegisterData) error
	UpdateClientLastSeen(ctx context.Context, data NodeHeartbeatData) error
	MarkClientOffline(ctx context.Context, clientIP string, nodeID string) error
	UpdateTaskRecordResult(ctx context.Context, data TaskRecordResultData) error
	FailClientRunningTask(ctx context.Context, recordID int64, message string) error
	MarkNodeIdle(ctx context.Context, clientIP string, nodeID string, commandID string, recordID int64) error
}

var persistenceState struct {
	mu      sync.RWMutex
	handler Persistence
}

// RegisterPersistence registers websocket event persistence. 注册 WebSocket 事件持久化实现
func RegisterPersistence(handler Persistence) {
	persistenceState.mu.Lock()
	defer persistenceState.mu.Unlock()
	persistenceState.handler = handler
}

func persistenceHandler() Persistence {
	persistenceState.mu.RLock()
	defer persistenceState.mu.RUnlock()
	return persistenceState.handler
}
