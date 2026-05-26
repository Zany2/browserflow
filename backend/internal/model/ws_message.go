package model

const (
	// WSMessageTypePing ping request ping 请求
	WSMessageTypePing = "ping"
	// WSMessageTypePong ping response ping 响应
	WSMessageTypePong = "pong"
	// WSMessageTypeHeartbeat heartbeat request 心跳请求
	WSMessageTypeHeartbeat = "heartbeat"
	// WSMessageTypeHeartbeatAck heartbeat response 心跳响应
	WSMessageTypeHeartbeatAck = "heartbeat_ack"
	// WSMessageTypeChatSend chat stream request 聊天流请求
	WSMessageTypeChatSend = "chat_send"
	// WSMessageTypeChatDone chat stream done 聊天流完成
	WSMessageTypeChatDone = "chat_done"
	// WSMessageTypeBrowserSubscribe browser status subscription 浏览器状态订阅
	WSMessageTypeBrowserSubscribe = "browser_subscribe"
	// WSMessageTypeBrowserStatus browser status response 浏览器状态响应
	WSMessageTypeBrowserStatus = "browser_status"
	// WSMessageTypeAgentStatusSubscribe agent status subscription 执行端状态订阅
	WSMessageTypeAgentStatusSubscribe = "agent_status_subscribe"
	// WSMessageTypeAgentStatus agent status response 执行端状态响应
	WSMessageTypeAgentStatus = "agent_status"
	// WSMessageTypeAgentRegister agent register request 执行端注册请求
	WSMessageTypeAgentRegister = "agent_register"
	// WSMessageTypeAgentRegistered agent register response 执行端注册响应
	WSMessageTypeAgentRegistered = "agent_registered"
	// WSMessageTypeAgentStatusUpdate agent status update 执行端状态更新
	WSMessageTypeAgentStatusUpdate = "agent_status_update"
	// WSMessageTypeAgentStatusUpdateAck agent status update response 执行端状态更新响应
	WSMessageTypeAgentStatusUpdateAck = "agent_status_update_ack"
	// WSMessageTypeAgentCommand agent command request 执行端命令请求
	WSMessageTypeAgentCommand = "agent_command"
	// WSMessageTypeAgentResult agent command result 执行端命令结果
	WSMessageTypeAgentResult = "agent_result"
	// WSMessageTypeAgentResultAck agent command result response 执行端命令结果响应
	WSMessageTypeAgentResultAck = "agent_result_ack"
	// WSMessageTypeWorkflowInventory workflow inventory report 工作流清单上报
	WSMessageTypeWorkflowInventory = "workflow_inventory"
	// WSMessageTypeWorkflowInventoryAck workflow inventory response 工作流清单响应
	WSMessageTypeWorkflowInventoryAck = "workflow_inventory_ack"
	// WSMessageTypeClientBanned client ban notification 客户端拉黑通知
	WSMessageTypeClientBanned = "client_banned"
	// WSMessageTypeError error response 错误响应
	WSMessageTypeError = "error"
)

// WSRequest frontend websocket request 前端 WebSocket 请求
type WSRequest struct {
	// Type message type 消息类型
	Type string `json:"type"`
	// SessionID chat session id 聊天会话标识
	SessionID string `json:"session_id,omitempty"`
	// Message chat message 聊天消息
	Message string `json:"message,omitempty"`
	// Role client role 客户端角色
	Role string `json:"role,omitempty"`
	// BrowserID browser agent id 浏览器执行端标识
	BrowserID string `json:"browser_id,omitempty"`
	// ClientID client id 客户端标识
	ClientID string `json:"client_id,omitempty"`
	// ClientName client display name 客户端显示名称
	ClientName string `json:"client_name,omitempty"`
	// MachineID physical worker machine id 物理机器标识
	MachineID string `json:"machine_id,omitempty"`
	// MachineName physical worker machine name 物理机器名称
	MachineName string `json:"machine_name,omitempty"`
	// NodeID execution node id 执行节点标识
	NodeID string `json:"node_id,omitempty"`
	// NodeName execution node name 执行节点名称
	NodeName string `json:"node_name,omitempty"`
	// NodeIndex execution node index 执行节点序号
	NodeIndex int `json:"node_index,omitempty"`
	// WorkerVersion BrowserFlow Worker version Worker 版本
	WorkerVersion string `json:"worker_version,omitempty"`
	// ProfileDir Chrome profile directory Chrome 用户数据目录
	ProfileDir string `json:"profile_dir,omitempty"`
	// ExtensionDir Automa extension directory Automa 插件目录
	ExtensionDir string `json:"extension_dir,omitempty"`
	// Capabilities node capabilities 节点能力
	Capabilities map[string]any `json:"capabilities,omitempty"`
	// Token client token 客户端令牌
	Token string `json:"token,omitempty"`
	// UserAgent browser user agent 浏览器 User-Agent
	UserAgent string `json:"user_agent,omitempty"`
	// AutomaInstalled Automa installed flag Automa 插件安装状态
	AutomaInstalled bool `json:"automa_installed,omitempty"`
	// AutomaVersion Automa extension version Automa 插件版本
	AutomaVersion string `json:"automa_version,omitempty"`
	// BrowserName browser name 浏览器名称
	BrowserName string `json:"browser_name,omitempty"`
	// BrowserVersion browser version 浏览器版本
	BrowserVersion string `json:"browser_version,omitempty"`
	// OsName operating system name 操作系统名称
	OsName string `json:"os_name,omitempty"`
	// OsVersion operating system version 操作系统版本
	OsVersion string `json:"os_version,omitempty"`
	// Hostname device hostname 设备主机名
	Hostname string `json:"hostname,omitempty"`
	// CommandID command id 命令标识
	CommandID string `json:"command_id,omitempty"`
	// ExecutionID workflow execution id 工作流执行标识
	ExecutionID string `json:"execution_id,omitempty"`
	// Data result data 结果数据
	Data map[string]any `json:"data,omitempty"`
	// Workflows client workflow list 客户端工作流列表
	Workflows []map[string]any `json:"workflows,omitempty"`
	// Success command success flag 命令成功状态
	Success bool `json:"success,omitempty"`
	// Error error message 错误消息
	Error string `json:"error,omitempty"`
	// ClientTime client timestamp 客户端时间戳
	ClientTime int64 `json:"client_time,omitempty"`
}

// WSResponse backend websocket response 后端 WebSocket 响应
type WSResponse struct {
	// Type message type 消息类型
	Type string `json:"type"`
	// SessionID chat session id 聊天会话标识
	SessionID string `json:"session_id,omitempty"`
	// Content stream content 流式内容
	Content string `json:"content,omitempty"`
	// Error error message 错误消息
	Error string `json:"error,omitempty"`
	// Message readable message 可读消息
	Message string `json:"message,omitempty"`
	// MessageID chat message id 聊天消息标识
	MessageID string `json:"message_id,omitempty"`
	// Browser browser status 浏览器状态
	Browser any `json:"browser,omitempty"`
	// Agents agent status list 执行端状态列表
	Agents any `json:"agents,omitempty"`
	// BrowserID browser agent id 浏览器执行端标识
	BrowserID string `json:"browser_id,omitempty"`
	// ClientID client id 客户端标识
	ClientID string `json:"client_id,omitempty"`
	// ClientIP client ip 客户端 IP
	ClientIP string `json:"client_ip,omitempty"`
	// MachineID physical worker machine id 物理机器标识
	MachineID string `json:"machine_id,omitempty"`
	// MachineName physical worker machine name 物理机器名称
	MachineName string `json:"machine_name,omitempty"`
	// NodeID execution node id 执行节点标识
	NodeID string `json:"node_id,omitempty"`
	// NodeName execution node name 执行节点名称
	NodeName string `json:"node_name,omitempty"`
	// Role client role 客户端角色
	Role string `json:"role,omitempty"`
	// AutomaInstalled Automa installed flag Automa 插件安装状态
	AutomaInstalled bool `json:"automa_installed,omitempty"`
	// AutomaVersion Automa extension version Automa 插件版本
	AutomaVersion string `json:"automa_version,omitempty"`
	// CommandID command id 命令标识
	CommandID string `json:"command_id,omitempty"`
	// Command command name 命令名称
	Command string `json:"command,omitempty"`
	// Payload command payload 命令负载
	Payload map[string]any `json:"payload,omitempty"`
	// Data response data 响应数据
	Data map[string]any `json:"data,omitempty"`
	// ClientTime client timestamp 客户端时间戳
	ClientTime int64 `json:"client_time,omitempty"`
	// ServerTime server timestamp 服务端时间戳
	ServerTime int64 `json:"server_time,omitempty"`
}
