package model

import (
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gorilla/websocket"
)

// AgentRegisterPayload websocket register payload WebSocket 注册负载
type AgentRegisterPayload struct {
	Role            string `json:"role"`
	BrowserID       string `json:"browser_id"`
	Token           string `json:"token"`
	AutomaInstalled bool   `json:"automa_installed"`
	AutomaVersion   string `json:"automa_version"`
}

// AgentStatusUpdatePayload websocket status update payload WebSocket 状态更新负载
type AgentStatusUpdatePayload struct {
	AutomaInstalled bool   `json:"automa_installed"`
	AutomaVersion   string `json:"automa_version"`
}

// AgentSession online agent session 执行端在线会话
type AgentSession struct {
	ClientID        string
	Role            string
	Token           string
	RemoteIP        string
	UserAgent       string
	AutomaInstalled bool
	AutomaVersion   string
	ConnectedAt     time.Time
	LastSeenAt      time.Time

	Conn      *websocket.Conn
	WriteMu   sync.Mutex
	PendingMu sync.Mutex
	Pending   map[string]chan g.Map
}

// AgentSubscriber status subscriber 状态订阅连接
type AgentSubscriber struct {
	Conn    *websocket.Conn
	WriteMu sync.Mutex
}

// AgentHub runtime hub state 执行端运行时状态
type AgentHub struct {
	Mu                 sync.RWMutex
	Sessions           map[string]*AgentSession
	AgentSubscribers   map[*websocket.Conn]*AgentSubscriber
	BrowserSubscribers map[*websocket.Conn]*AgentSubscriber
	Sequence           uint64
}
