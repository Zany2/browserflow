package state

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/gogf/gf/v2/os/gcache"
)

const (
	// pendingCommandCacheSize max pending command count 最大等待命令数量
	pendingCommandCacheSize = 500
	// pendingCommandTTL max pending command lifetime 等待命令最长保存时间
	pendingCommandTTL = 6 * time.Hour
)

// BrowserRuntime running browser holder 浏览器运行态对象
type BrowserRuntime struct {
	Instance   *model.BrowserInstance // Instance browser config 浏览器实例配置
	Browser    *rod.Browser           // Browser rod browser instance Rod 浏览器实例
	Launcher   *launcher.Launcher     // Launcher rod launcher Rod 启动器
	StartTime  time.Time              // StartTime browser start time 浏览器启动时间
	ControlURL string                 // ControlURL browser control url 浏览器控制地址
	AgentURL   string                 // AgentURL browser agent url 浏览器执行端地址
	AgentToken string                 // AgentToken browser agent token 浏览器执行端令牌
}

// AgentConnection online browser agent connection 在线浏览器执行端连接
type AgentConnection struct {
	BrowserID       string    // BrowserID browser instance id 浏览器实例ID
	Role            string    // Role agent role 执行端角色
	Token           string    // Token agent auth token 执行端认证令牌
	ConnectionID    string    // ConnectionID websocket connection id WebSocket 连接标识
	AutomaInstalled bool      // AutomaInstalled plugin installed status 插件安装状态
	AutomaVersion   string    // AutomaVersion plugin version 插件版本
	ConnectedAt     time.Time // ConnectedAt connection time 连接建立时间
	LastSeenAt      time.Time // LastSeenAt last active time 最近活跃时间
}

var (
	// DBMu protects local db singleton 保护本地数据库单例
	DBMu sync.Mutex
	// DB local file database 本地文件数据库
	DB *storage.BoltDB
	// LLMClient shared large model client 共享大模型客户端
	LLMClient *llm.Client

	// BrowserMu protects browser runtime state 保护浏览器运行状态
	BrowserMu sync.Mutex
	// BrowserCurrentInstanceID current browser instance id 当前浏览器实例ID
	BrowserCurrentInstanceID string
	// BrowserInstances active browser runtime map 活跃浏览器运行态映射
	BrowserInstances = map[string]*BrowserRuntime{}
	// BrowserStatusListeners browser status subscribers 浏览器状态订阅者
	BrowserStatusListeners = map[chan model.BrowserStatus]struct{}{}

	// AgentMu protects browser agent state 保护浏览器执行端状态
	AgentMu sync.Mutex
	// AgentConnections online agent connections 在线执行端连接
	AgentConnections = map[string]*AgentConnection{}
	// PendingCommands command result waiters 等待命令结果
	PendingCommands = gcache.New(pendingCommandCacheSize)
	// AgentStatusListeners agent status subscribers 执行端状态订阅者
	AgentStatusListeners = map[chan []model.AgentStatus]struct{}{}
)

// SetPendingCommand stores command waiter with ttl 保存带过期时间的命令等待通道
func SetPendingCommand(commandID string, resultCh chan model.AgentCommandResult) {
	if commandID == "" || resultCh == nil {
		return
	}
	_ = PendingCommands.Set(context.Background(), commandID, resultCh, pendingCommandTTL)
}

// PopPendingCommand removes and returns command waiter 取出并删除命令等待通道
func PopPendingCommand(commandID string) chan model.AgentCommandResult {
	if commandID == "" {
		return nil
	}
	value, _ := PendingCommands.Remove(context.Background(), commandID)
	if value == nil {
		return nil
	}
	resultCh, _ := value.Val().(chan model.AgentCommandResult)
	return resultCh
}

// RemovePendingCommand removes command waiter 删除命令等待通道
func RemovePendingCommand(commandID string) {
	if commandID == "" {
		return
	}
	_, _ = PendingCommands.Remove(context.Background(), commandID)
}

// RemoveAgentConnection removes one browser agent and broadcasts status 删除执行端连接并广播状态
func RemoveAgentConnection(browserID string) {
	browserID = strings.TrimSpace(browserID)
	if browserID == "" {
		return
	}

	AgentMu.Lock()
	delete(AgentConnections, browserID)
	statuses := agentStatusesLocked()
	broadcastAgentStatusesLocked(statuses)
	AgentMu.Unlock()
}

// agentStatusesLocked builds agent statuses 调用方需持有 AgentMu
func agentStatusesLocked() []model.AgentStatus {
	statuses := make([]model.AgentStatus, 0, len(AgentConnections))
	for browserID, agent := range AgentConnections {
		if agent == nil {
			statuses = append(statuses, model.AgentStatus{BrowserID: browserID, Online: false})
			continue
		}
		if strings.ToLower(strings.TrimSpace(agent.Role)) == "client_agent" {
			continue
		}
		statuses = append(statuses, model.AgentStatus{
			BrowserID:       agent.BrowserID,
			Online:          true,
			AutomaInstalled: agent.AutomaInstalled,
			AutomaVersion:   agent.AutomaVersion,
			ConnectedAt:     agent.ConnectedAt,
			LastSeenAt:      agent.LastSeenAt,
		})
	}
	return statuses
}

// broadcastAgentStatusesLocked notifies agent subscribers 调用方需持有 AgentMu
func broadcastAgentStatusesLocked(statuses []model.AgentStatus) {
	for listener := range AgentStatusListeners {
		select {
		case listener <- statuses:
		default:
		}
	}
}
