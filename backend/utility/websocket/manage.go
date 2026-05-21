package websockets

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// defaultHeartbeatInterval default heartbeat interval 默认心跳间隔
	defaultHeartbeatInterval = 15 * time.Second
	// defaultHeartbeatTimeout default heartbeat timeout 默认心跳超时
	defaultHeartbeatTimeout = 45 * time.Second
	// defaultWriteWait default write wait 默认写入超时
	defaultWriteWait = 10 * time.Second
	// defaultMessageBuffer default message buffer 默认消息缓冲区
	defaultMessageBuffer = 64
	// defaultBucketCount default bucket count 默认分桶数量
	defaultBucketCount = 12
)

// Config websocket manager config WebSocket 管理器配置
type Config struct {
	// HeartbeatInterval heartbeat interval 心跳间隔
	HeartbeatInterval time.Duration
	// HeartbeatTimeout heartbeat timeout 心跳超时
	HeartbeatTimeout time.Duration
	// WriteWait write timeout 写入超时
	WriteWait time.Duration
	// MessageBuffer message channel buffer 消息缓冲区大小
	MessageBuffer int
	// BucketCount client bucket count 分桶数量
	BucketCount int
}

// ClientIdentity websocket client identity WebSocket 客户端身份
type ClientIdentity struct {
	// ConnectionID internal connection identity 内部连接标识
	ConnectionID string
	// ClientIP unique client identity 客户端唯一标识
	ClientIP string
	// RequireHeartbeat read timeout guard 是否要求心跳保活
	RequireHeartbeat bool
}

// ClientSnapshot websocket client snapshot WebSocket 客户端快照
type ClientSnapshot struct {
	// ConnectionID internal connection identity 内部连接标识
	ConnectionID string `json:"connection_id"`
	// ClientIP unique client identity 客户端唯一标识
	ClientIP string `json:"client_ip"`
	// ConnectedAt connected time 建连时间
	ConnectedAt time.Time `json:"connected_at"`
	// LastActiveTime last active time 最近活跃时间
	LastActiveTime int64 `json:"last_active_time"`
	// LastHeartbeatTime last heartbeat time 最近心跳时间
	LastHeartbeatTime int64 `json:"last_heartbeat_time"`
	// ClientHeartbeatTime client reported heartbeat 客户端上报心跳时间
	ClientHeartbeatTime int64 `json:"client_heartbeat_time"`
}

// HandlerFunc websocket lifecycle handler WebSocket 生命周期处理器
type HandlerFunc interface {
	OnOpen(client *Client)
	OnMessage(client *Client, messageType int, message []byte)
	OnClose(client *Client)
}

// clientBucket sharded client bucket 分片客户端桶
type clientBucket struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// WebSocketManager websocket connection manager WebSocket 连接管理器
type WebSocketManager struct {
	// count active connection count 在线连接数
	count int64
	// clientBuckets sharded client storage 分桶连接存储
	clientBuckets []*clientBucket
	// Handler connection callback 连接回调处理器
	Handler HandlerFunc
	// heartbeatInterval configured heartbeat interval 心跳间隔
	heartbeatInterval time.Duration
	// heartbeatTimeout configured heartbeat timeout 心跳超时
	heartbeatTimeout time.Duration
	// writeWait configured write timeout 写入超时
	writeWait time.Duration
	// messageBuffer client write queue size 写队列长度
	messageBuffer int
	// bucketCount total bucket count 分桶数量
	bucketCount uint32
}

// InitWebSocketManager init manager 初始化管理器
func InitWebSocketManager(handler HandlerFunc) *WebSocketManager {
	return InitWebSocketManagerWithConfig(handler, Config{})
}

// InitWebSocketManagerWithConfig init manager with config 使用配置初始化管理器
func InitWebSocketManagerWithConfig(handler HandlerFunc, cfg Config) *WebSocketManager {
	cfg = normalizeConfig(cfg)
	return &WebSocketManager{
		clientBuckets:     newClientBuckets(cfg.BucketCount),
		Handler:           handler,
		heartbeatInterval: cfg.HeartbeatInterval,
		heartbeatTimeout:  cfg.HeartbeatTimeout,
		writeWait:         cfg.WriteWait,
		messageBuffer:     cfg.MessageBuffer,
		bucketCount:       uint32(cfg.BucketCount),
	}
}

// normalizeConfig normalize config 归一化配置
func normalizeConfig(cfg Config) Config {
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = defaultHeartbeatInterval
	}
	if cfg.HeartbeatTimeout <= 0 {
		cfg.HeartbeatTimeout = defaultHeartbeatTimeout
	}
	if cfg.WriteWait <= 0 {
		cfg.WriteWait = defaultWriteWait
	}
	if cfg.MessageBuffer <= 0 {
		cfg.MessageBuffer = defaultMessageBuffer
	}
	if cfg.BucketCount <= 0 {
		cfg.BucketCount = defaultBucketCount
	}
	return cfg
}

// newClientBuckets create client buckets 创建连接分桶
func newClientBuckets(bucketCount int) []*clientBucket {
	buckets := make([]*clientBucket, bucketCount)
	for i := range buckets {
		buckets[i] = &clientBucket{
			clients: make(map[string]*Client),
		}
	}
	return buckets
}

// bucketIndex hash connection id to bucket 按连接标识分桶
func (m *WebSocketManager) bucketIndex(key string) uint32 {
	if m.bucketCount == 0 {
		return 0
	}

	var hash uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}

	return hash % m.bucketCount
}

// clientBucket get bucket by connection id 获取连接分桶
func (m *WebSocketManager) clientBucket(connectionID string) *clientBucket {
	return m.clientBuckets[m.bucketIndex(connectionID)]
}

// loadClient load client by connection id 按连接标识加载连接
func (m *WebSocketManager) loadClient(connectionID string) (*Client, bool) {
	if connectionID == "" {
		return nil, false
	}

	bucket := m.clientBucket(connectionID)
	bucket.mu.RLock()
	client, ok := bucket.clients[connectionID]
	bucket.mu.RUnlock()
	return client, ok
}

// rangeClients iterate all clients 遍历全部连接
func (m *WebSocketManager) rangeClients(handler func(connectionID string, client *Client) bool) {
	for i := range m.clientBuckets {
		bucket := m.clientBuckets[i]

		bucket.mu.RLock()
		pairs := make([]struct {
			connectionID string
			client       *Client
		}, 0, len(bucket.clients))
		for connectionID, client := range bucket.clients {
			pairs = append(pairs, struct {
				connectionID string
				client       *Client
			}{
				connectionID: connectionID,
				client:       client,
			})
		}
		bucket.mu.RUnlock()

		for _, pair := range pairs {
			if !handler(pair.connectionID, pair.client) {
				return
			}
		}
	}
}

// onConnect handle new connection 处理新连接
func (m *WebSocketManager) onConnect(client *Client) bool {
	if client == nil || client.connectionID == "" {
		if client != nil && client.conn != nil {
			_ = client.conn.Close()
		}
		return false
	}

	// onConnect replace same connection id 同一连接标识的新连接顶掉旧连接
	if previous, replaced := m.addClient(client); replaced {
		m.disConnect(previous)
	}
	if m.Handler != nil {
		m.Handler.OnOpen(client)
	}
	return true
}

// onMessage handle inbound message 处理入站消息
func (m *WebSocketManager) onMessage(client *Client, messageType int, message []byte) {
	if m.Handler != nil {
		m.Handler.OnMessage(client, messageType, message)
	}
}

// disConnect cleanup client connection 回收连接资源
func (m *WebSocketManager) disConnect(client *Client) {
	if client == nil {
		return
	}

	client.disconnectOnce.Do(func() {
		if m.deleteClientIfMatch(client.connectionID, client) {
			atomic.AddInt64(&m.count, -1)
		}
		client.closeMessageChannel()
		_ = client.conn.Close()
		if m.Handler != nil && client.connectionID != "" {
			m.Handler.OnClose(client)
		}
	})
}

// deleteClientIfMatch delete exact client instance 仅删除当前连接实例
func (m *WebSocketManager) deleteClientIfMatch(connectionID string, client *Client) bool {
	if connectionID == "" {
		return false
	}

	bucket := m.clientBucket(connectionID)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	existing, ok := bucket.clients[connectionID]
	if !ok || existing != client {
		return false
	}
	delete(bucket.clients, connectionID)
	return true
}

// addClient store client and return previous one 写入连接并返回旧连接
func (m *WebSocketManager) addClient(client *Client) (*Client, bool) {
	bucket := m.clientBucket(client.connectionID)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if previous, ok := bucket.clients[client.connectionID]; ok {
		bucket.clients[client.connectionID] = client
		return previous, true
	}

	bucket.clients[client.connectionID] = client
	atomic.AddInt64(&m.count, 1)
	return nil, false
}

// enqueue push message into write queue 推送消息到写队列
func (m *WebSocketManager) enqueue(client *Client, msg *messageData) {
	select {
	case client.message <- msg:
	default:
		m.disConnect(client)
	}
}

// RegisterClient register websocket client 注册 WebSocket 客户端
func (m *WebSocketManager) RegisterClient(ctx context.Context, clientIP string, conn *websocket.Conn) {
	m.RegisterClientWithIdentity(ctx, ClientIdentity{
		ConnectionID:     clientIP,
		ClientIP:         clientIP,
		RequireHeartbeat: true,
	}, conn)
}

// RegisterClientWithIdentity register client with identity 按身份注册客户端
func (m *WebSocketManager) RegisterClientWithIdentity(ctx context.Context, identity ClientIdentity, conn *websocket.Conn) {
	connectionID := identity.ConnectionID
	if connectionID == "" {
		connectionID = identity.ClientIP
	}
	client := &Client{
		Ctx:              ctx,
		connectionID:     connectionID,
		clientIP:         identity.ClientIP,
		connectedAt:      time.Now(),
		conn:             conn,
		manager:          m,
		requireHeartbeat: identity.RequireHeartbeat,
		message:          make(chan *messageData, m.messageBuffer),
	}
	client.markActive(client.connectedAt)

	if !m.onConnect(client) {
		return
	}

	go client.Read()
	go client.Write()
}

// TouchClient refresh client active time 刷新客户端活跃时间
func (m *WebSocketManager) TouchClient(connectionID string) bool {
	client, ok := m.loadClient(connectionID)
	if !ok {
		return false
	}
	client.markActive(time.Now())
	return true
}

// GetClientCtx get client context 获取连接上下文
func (m *WebSocketManager) GetClientCtx(connectionID string) context.Context {
	if client, ok := m.loadClient(connectionID); ok {
		return client.Ctx
	}
	return nil
}

// GetClientSnapshot get client snapshot 获取连接快照
func (m *WebSocketManager) GetClientSnapshot(connectionID string) (*ClientSnapshot, bool) {
	if client, ok := m.loadClient(connectionID); ok {
		return buildSnapshot(client), true
	}
	return nil, false
}

// ListClientConnections list one client connection 列出指定连接
func (m *WebSocketManager) ListClientConnections(connectionID string) []ClientSnapshot {
	if client, ok := m.loadClient(connectionID); ok {
		return []ClientSnapshot{*buildSnapshot(client)}
	}
	return nil
}

// HasClient check client exists 检查连接是否存在
func (m *WebSocketManager) HasClient(connectionID string) bool {
	_, ok := m.loadClient(connectionID)
	return ok
}

// GetWSCount get websocket count 获取连接数
func (m *WebSocketManager) GetWSCount() int64 {
	return atomic.LoadInt64(&m.count)
}

// GetClientConnectionCount get client connection count 获取指定连接数
func (m *WebSocketManager) GetClientConnectionCount(connectionID string) int {
	if m.HasClient(connectionID) {
		return 1
	}
	return 0
}

// SendMessageToClient send text message 发送文本消息
func (m *WebSocketManager) SendMessageToClient(connectionID string, message []byte) int {
	return m.sendMessageToClient(connectionID, websocket.TextMessage, message)
}

// SendMessageToClients send text message to clients 批量发送文本消息
func (m *WebSocketManager) SendMessageToClients(clientIPs []string, message []byte) int {
	return m.sendMessageToClients(clientIPs, websocket.TextMessage, message)
}

// BroadcastMessage broadcast text message 广播文本消息
func (m *WebSocketManager) BroadcastMessage(message []byte, excluded ...string) {
	excludedMap := make(map[string]struct{}, len(excluded))
	for i := range excluded {
		excludedMap[excluded[i]] = struct{}{}
	}

	m.rangeClients(func(connectionID string, client *Client) bool {
		if _, ok := excludedMap[connectionID]; ok {
			return true
		}
		m.enqueue(client, &messageData{data: message, dataType: websocket.TextMessage})
		return true
	})
}

// CloseClientConnection close one client connection 关闭指定连接
func (m *WebSocketManager) CloseClientConnection(connectionID string) {
	if client, ok := m.loadClient(connectionID); ok {
		m.disConnect(client)
	}
}

// CloseClientConnections close client connections 关闭指定连接
func (m *WebSocketManager) CloseClientConnections(connectionID string) int {
	if client, ok := m.loadClient(connectionID); ok {
		m.disConnect(client)
		return 1
	}
	return 0
}

// CloseAll close all connections 关闭全部连接
func (m *WebSocketManager) CloseAll() int {
	count := 0
	m.rangeClients(func(_ string, client *Client) bool {
		m.disConnect(client)
		count++
		return true
	})
	return count
}

// SendBinaryToClient send binary message 发送二进制消息
func (m *WebSocketManager) SendBinaryToClient(clientIP string, binary []byte) int {
	return m.sendMessageToClient(clientIP, websocket.BinaryMessage, binary)
}

// SendBinaryToClients send binary message to clients 批量发送二进制消息
func (m *WebSocketManager) SendBinaryToClients(clientIPs []string, binary []byte) int {
	return m.sendMessageToClients(clientIPs, websocket.BinaryMessage, binary)
}

// BroadcastBinary broadcast binary message 广播二进制消息
func (m *WebSocketManager) BroadcastBinary(binary []byte, excluded ...string) {
	excludedMap := make(map[string]struct{}, len(excluded))
	for i := range excluded {
		excludedMap[excluded[i]] = struct{}{}
	}

	m.rangeClients(func(connectionID string, client *Client) bool {
		if _, ok := excludedMap[connectionID]; ok {
			return true
		}
		m.enqueue(client, &messageData{data: binary, dataType: websocket.BinaryMessage})
		return true
	})
}

// sendMessageToClients send message to unique clients 向去重后的客户端发送消息
func (m *WebSocketManager) sendMessageToClients(clientIPs []string, messageType int, payload []byte) int {
	if len(clientIPs) == 0 {
		return 0
	}

	sent := 0
	visited := make(map[string]struct{}, len(clientIPs))
	for _, clientIP := range clientIPs {
		if clientIP == "" {
			continue
		}
		if _, ok := visited[clientIP]; ok {
			continue
		}
		visited[clientIP] = struct{}{}
		sent += m.sendMessageToClient(clientIP, messageType, payload)
	}
	return sent
}

// sendMessageToClient send message to one local client 向单个本地连接发送消息
func (m *WebSocketManager) sendMessageToClient(connectionID string, messageType int, payload []byte) int {
	if connectionID == "" {
		return 0
	}
	if client, ok := m.loadClient(connectionID); ok {
		m.enqueue(client, &messageData{data: payload, dataType: messageType})
		return 1
	}
	return 0
}

// buildSnapshot build snapshot data 构建连接快照
func buildSnapshot(client *Client) *ClientSnapshot {
	return &ClientSnapshot{
		ConnectionID:        client.connectionID,
		ClientIP:            client.clientIP,
		ConnectedAt:         client.connectedAt,
		LastActiveTime:      client.LastActiveTime(),
		LastHeartbeatTime:   client.LastHeartbeatTime(),
		ClientHeartbeatTime: client.ClientHeartbeatTime(),
	}
}
