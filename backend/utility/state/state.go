package state

import (
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
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
	PendingCommands = map[string]chan model.AgentCommandResult{}
	// AgentStatusListeners agent status subscribers 执行端状态订阅者
	AgentStatusListeners = map[chan []model.AgentStatus]struct{}{}
)
