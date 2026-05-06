package model

import (
	"encoding/json"
	"time"
)

// LLMConfig large model config 大模型配置
type LLMConfig struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	APIKey    string    `json:"api_key,omitempty"`
	Model     string    `json:"model"`
	BaseURL   string    `json:"base_url"`
	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LLMProvider large model provider info 大模型厂商信息
type LLMProvider struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	BaseURL    string   `json:"base_url"`
	Models     []string `json:"models"`
	Compatible bool     `json:"compatible"`
}

// ChatMessage chat message 对话消息
type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatSession chat session 对话会话
type ChatSession struct {
	ID          string        `json:"id"`
	LLMConfigID string        `json:"llm_config_id"`
	Messages    []ChatMessage `json:"messages"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// StreamChunk stream response chunk 流式响应片段
type StreamChunk struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	Error     string `json:"error,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

// BrowserInstance browser instance config 浏览器实例配置
type BrowserInstance struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	IsActive    bool      `json:"is_active"`
	Type        string    `json:"type"`
	BinPath     string    `json:"bin_path"`
	UserDataDir string    `json:"user_data_dir"`
	ControlURL  string    `json:"control_url"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Headless    *bool     `json:"headless,omitempty"`
	NoSandbox   *bool     `json:"no_sandbox,omitempty"`
	LaunchArgs  []string  `json:"launch_args,omitempty"`
	Proxy       string    `json:"proxy,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BrowserStatus browser runtime status 浏览器运行状态
type BrowserStatus struct {
	Running           bool             `json:"running"`
	CurrentInstanceID string           `json:"current_instance_id"`
	Instance          *BrowserInstance `json:"instance,omitempty"`
	StartTime         *time.Time       `json:"start_time,omitempty"`
	UptimeSeconds     int64            `json:"uptime_seconds"`
	ControlURL        string           `json:"control_url,omitempty"`
	AgentURL          string           `json:"agent_url,omitempty"`
}

// AutomaWorkflowSnapshot latest workflow snapshot 最新工作流快照
type AutomaWorkflowSnapshot struct {
	ID        string          `json:"id"`
	Workflows json.RawMessage `json:"workflows"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// AutomaWorkflowRecord saved workflow record 本地工作流记录
type AutomaWorkflowRecord struct {
	ID              int64     `json:"id"`
	AutomaID        string    `json:"automa_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Source          int       `json:"source"`
	SourceIP        string    `json:"source_ip"`
	SourceUserAgent string    `json:"source_user_agent"`
	AutomaVersion   string    `json:"automa_version"`
	ExtVersion      string    `json:"ext_version"`
	CreatedAtAutoma int64     `json:"created_at_automa"`
	UpdatedAtAutoma int64     `json:"updated_at_automa"`
	IsDisabled      bool      `json:"is_disabled"`
	IsProtected     bool      `json:"is_protected"`
	NodeCount       int       `json:"node_count"`
	EdgeCount       int       `json:"edge_count"`
	RawJSON         string    `json:"raw_json"`
	NormalizedJSON  string    `json:"normalized_json"`
	ContentHash     string    `json:"content_hash"`
	Revision        int       `json:"revision"`
	FirstSyncedAt   time.Time `json:"first_synced_at"`
	LastSyncedAt    time.Time `json:"last_synced_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AgentStatus browser agent status 浏览器执行端状态
type AgentStatus struct {
	BrowserID       string    `json:"browser_id"`
	Online          bool      `json:"online"`
	AutomaInstalled bool      `json:"automa_installed"`
	ConnectedAt     time.Time `json:"connected_at"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

// AgentCommandResult browser agent command result 浏览器执行端命令结果
type AgentCommandResult struct {
	BrowserID string          `json:"browser_id"`
	CommandID string          `json:"command_id"`
	Success   bool            `json:"success"`
	Data      json.RawMessage `json:"data,omitempty"`
	Error     string          `json:"error,omitempty"`
}
