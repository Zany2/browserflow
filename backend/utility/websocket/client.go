package websockets

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Client websocket connection client WebSocket 连接客户端
type Client struct {
	// Ctx client context 连接上下文
	Ctx context.Context
	// clientIP client identity 客户端唯一标识
	clientIP string
	// clientID business client id 业务客户端标识
	clientID string
	// connectedAt connected time 建连时间
	connectedAt time.Time
	// identityMu protect business identity 保护业务标识
	identityMu sync.RWMutex

	// lastActiveTime last active time 最近活跃时间
	lastActiveTime int64
	// lastHeartbeatTime last heartbeat time 最近心跳时间
	lastHeartbeatTime int64
	// clientHeartbeatTime client reported heartbeat 客户端上报心跳时间
	clientHeartbeatTime int64

	// conn websocket connection 原始 WebSocket 连接
	conn *websocket.Conn
	// manager owner manager 所属管理器
	manager *WebSocketManager
	// message write queue 写队列
	message chan *messageData
	// closeOnce close queue once 只关闭一次消息通道
	closeOnce sync.Once
	// disconnectOnce cleanup once 只执行一次断连清理
	disconnectOnce sync.Once
}

// messageData outbound websocket message 待发送 WebSocket 消息
type messageData struct {
	// dataType websocket frame type 消息帧类型
	dataType int
	// data message payload 消息内容
	data []byte
}

// ClientIP get client ip 获取客户端标识
func (c *Client) ClientIP() string {
	return c.clientIP
}

// ClientID get business client id 获取业务客户端标识
func (c *Client) ClientID() string {
	c.identityMu.RLock()
	defer c.identityMu.RUnlock()
	return c.clientID
}

// BindClientID bind business client id 绑定业务客户端标识
func (c *Client) BindClientID(clientID string) {
	c.identityMu.Lock()
	defer c.identityMu.Unlock()
	c.clientID = clientID
}

// ConnectedAt get connected time 获取连接建立时间
func (c *Client) ConnectedAt() time.Time {
	return c.connectedAt
}

// LastActiveTime get last active time 获取最近活跃时间
func (c *Client) LastActiveTime() int64 {
	return atomic.LoadInt64(&c.lastActiveTime)
}

// LastHeartbeatTime get last heartbeat time 获取最近心跳时间
func (c *Client) LastHeartbeatTime() int64 {
	return atomic.LoadInt64(&c.lastHeartbeatTime)
}

// ClientHeartbeatTime get client heartbeat time 获取客户端心跳时间
func (c *Client) ClientHeartbeatTime() int64 {
	return atomic.LoadInt64(&c.clientHeartbeatTime)
}

// markActive update active timestamp 更新活跃时间
func (c *Client) markActive(now time.Time) {
	atomic.StoreInt64(&c.lastActiveTime, now.UnixMilli())
}

// markHeartbeat update heartbeat timestamps 更新心跳时间
func (c *Client) markHeartbeat(clientHeartbeatTime int64, now time.Time) {
	serverTime := now.UnixMilli()
	atomic.StoreInt64(&c.lastActiveTime, serverTime)
	atomic.StoreInt64(&c.lastHeartbeatTime, serverTime)
	if clientHeartbeatTime > 0 {
		atomic.StoreInt64(&c.clientHeartbeatTime, clientHeartbeatTime)
	}
}

// refreshReadDeadline refresh read deadline 刷新读超时
func (c *Client) refreshReadDeadline() error {
	return c.conn.SetReadDeadline(time.Now().Add(c.manager.heartbeatTimeout))
}

// Read read websocket messages 读取 WebSocket 消息
func (c *Client) Read() {
	defer c.manager.disConnect(c)

	if err := c.refreshReadDeadline(); err != nil {
		return
	}

	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		c.markActive(time.Now())
		if err = c.refreshReadDeadline(); err != nil {
			return
		}

		switch msgType {
		case websocket.TextMessage, websocket.BinaryMessage:
			c.manager.onMessage(c, msgType, msg)
		case websocket.CloseMessage:
			return
		}
	}
}

// Write write websocket messages 写入 WebSocket 消息
func (c *Client) Write() {
	defer c.manager.disConnect(c)

	for msg := range c.message {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.manager.writeWait)); err != nil {
			return
		}

		if err := c.conn.WriteMessage(msg.dataType, msg.data); err != nil {
			return
		}
	}
}

// closeMessageChannel close message channel 关闭消息通道
func (c *Client) closeMessageChannel() {
	c.closeOnce.Do(func() {
		close(c.message)
	})
}
