package websockets

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gorilla/websocket"
)

const (
	// defaultMaxConn default max connection count 默认最大连接数
	defaultMaxConn = 1000000
)

var (
	// once singleton initializer 全局只初始化一次
	once sync.Once
	// WsManage global manager 全局连接管理器
	WsManage *WebSocketManager
	// WsHandler global default handler 全局默认处理器
	WsHandler *WsHandlerFunc
)

// WsHandlerFunc default websocket handler 默认 WebSocket 处理器
type WsHandlerFunc struct {
	// mu protect context maps 保护上下文索引
	mu sync.RWMutex
	// ClientCtxMap client contexts 客户端上下文
	ClientCtxMap map[string]context.Context
	// ClientCtxCancel client cancel funcs 客户端取消函数
	ClientCtxCancel map[string]context.CancelFunc
}

// Init init websocket manager 初始化 WebSocket 管理器
func Init(ctx context.Context) {
	once.Do(func() {
		WsHandler = &WsHandlerFunc{
			ClientCtxMap:    make(map[string]context.Context),
			ClientCtxCancel: make(map[string]context.CancelFunc),
		}
		WsManage = InitWebSocketManagerWithConfig(WsHandler, loadConfig(ctx))
	})
}

// loadConfig load websocket config 加载 WebSocket 配置
func loadConfig(ctx context.Context) Config {
	return Config{
		HeartbeatInterval: time.Duration(g.Cfg().MustGet(ctx, "websocket.heartbeatInterval", 15).Int()) * time.Second,
		HeartbeatTimeout:  time.Duration(g.Cfg().MustGet(ctx, "websocket.heartbeatTimeout", 45).Int()) * time.Second,
		WriteWait:         time.Duration(g.Cfg().MustGet(ctx, "websocket.writeWait", 10).Int()) * time.Second,
		MessageBuffer:     g.Cfg().MustGet(ctx, "websocket.messageBuffer", 64).Int(),
		BucketCount:       g.Cfg().MustGet(ctx, "websocket.bucketCount", defaultBucketCount).Int(),
	}
}

// MaxConn get max websocket count 获取最大连接数
func MaxConn(ctx context.Context) int64 {
	return g.Cfg().MustGet(ctx, "websocket.maxConn", defaultMaxConn).Int64()
}

// BuildClientIdentity build client identity 构建客户端标识
func BuildClientIdentity(clientIP string) ClientIdentity {
	return ClientIdentity{
		ClientIP: clientIP,
	}
}

// SendClientMessage send structured websocket message 发送结构化消息
func SendClientMessage(clientIP string, in *model.WSResponse) int {
	if in == nil {
		return 0
	}
	if in.ServerTime <= 0 {
		in.ServerTime = time.Now().UnixMilli()
	}

	body, err := json.Marshal(in)
	if err != nil {
		return 0
	}
	if WsManage == nil {
		Init(context.Background())
	}
	return WsManage.SendMessageToClient(clientIP, body)
}

// SendRawClientMessage send raw websocket message 发送原始消息
func SendRawClientMessage(clientIP string, messageType int, payload []byte) int {
	if WsManage == nil {
		Init(context.Background())
	}

	switch messageType {
	case websocket.BinaryMessage:
		return WsManage.SendBinaryToClient(clientIP, payload)
	default:
		return WsManage.SendMessageToClient(clientIP, payload)
	}
}

// SendRawClientMessages send raw websocket messages in batch 批量发送原始消息
func SendRawClientMessages(clientIPs []string, messageType int, payload []byte) int {
	if WsManage == nil {
		Init(context.Background())
	}

	switch messageType {
	case websocket.BinaryMessage:
		return WsManage.SendBinaryToClients(clientIPs, payload)
	default:
		return WsManage.SendMessageToClients(clientIPs, payload)
	}
}

// OnMessage handle websocket message 处理 WebSocket 消息
func (ws *WsHandlerFunc) OnMessage(client *Client, messageType int, message []byte) {
	if client == nil {
		return
	}
	if messageType != websocket.TextMessage {
		g.Log().Line().Infof(client.Ctx, "收到非文本 WebSocket 消息：client_ip=%s message_type=%d", client.ClientIP(), messageType)
		return
	}

	var in model.WSRequest
	if err := json.Unmarshal(message, &in); err != nil {
		g.Log().Line().Infof(client.Ctx, "收到无法解析的 WebSocket 原始消息：client_ip=%s message=%s", client.ClientIP(), string(message))
		return
	}

	switch strings.ToLower(in.Type) {
	case model.WSMessageTypeHeartbeat:
		ws.handleHeartbeat(client, &in)
	case model.WSMessageTypeAgentRegister:
		ws.handleAgentRegister(client, &in)
	case model.WSMessageTypeAgentStatusUpdate:
		ws.handleAgentStatusUpdate(client, &in)
	case model.WSMessageTypeAgentResult:
		ws.handleAgentResult(client, &in)
	case model.WSMessageTypeWorkflowInventory:
		ws.handleWorkflowInventory(client, &in)
	case model.WSMessageTypePing:
		_ = SendClientMessage(client.ClientIP(), &model.WSResponse{Type: model.WSMessageTypePong})
	default:
		g.Log().Line().Infof(client.Ctx, "收到未处理的 WebSocket 消息：client_ip=%s type=%s", client.ClientIP(), in.Type)
	}
}

// handleHeartbeat handle heartbeat message 处理心跳消息
func (ws *WsHandlerFunc) handleHeartbeat(client *Client, in *model.WSRequest) {
	now := time.Now()
	client.markHeartbeat(in.ClientTime, now)
	workflowcache.TouchClient(client.Ctx, client.ClientIP())

	// Refresh client online state 刷新客户端在线状态
	if err := updateClientLastSeen(client, in); err != nil {
		g.Log().Line().Errorf(client.Ctx, "处理心跳时更新客户端状态失败：client_ip=%s err=%+v", client.ClientIP(), err)
	}

	_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
		Type:       model.WSMessageTypeHeartbeatAck,
		BrowserID:  in.BrowserID,
		ClientID:   resolveClientID(client, in),
		ClientIP:   client.ClientIP(),
		ClientTime: in.ClientTime,
		ServerTime: now.UnixMilli(),
		Data: g.Map{
			"last_heartbeat_time": client.LastHeartbeatTime(),
			"heartbeat_interval":  client.manager.heartbeatInterval.Milliseconds(),
			"heartbeat_timeout":   client.manager.heartbeatTimeout.Milliseconds(),
		},
	})
}

// handleAgentRegister handle agent register message 处理执行端注册消息
func (ws *WsHandlerFunc) handleAgentRegister(client *Client, in *model.WSRequest) {
	now := time.Now()
	client.markHeartbeat(now.UnixMilli(), now)

	// Resolve client id before persistence 保存前解析客户端标识
	clientID := resolveClientID(client, in)
	client.BindClientID(clientID)

	// Persist client information 保存客户端信息
	if err := saveClientRegister(client, in, clientID); err != nil {
		g.Log().Line().Errorf(client.Ctx, "处理客户端注册时保存客户端信息失败：client_ip=%s err=%+v", client.ClientIP(), err)
		_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
			Type:    model.WSMessageTypeError,
			Error:   "客户端信息保存失败",
			Message: err.Error(),
		})
		return
	}

	_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
		Type:            model.WSMessageTypeAgentRegistered,
		BrowserID:       in.BrowserID,
		ClientID:        clientID,
		ClientIP:        client.ClientIP(),
		Role:            in.Role,
		AutomaInstalled: in.AutomaInstalled,
		AutomaVersion:   in.AutomaVersion,
	})
}

// handleAgentStatusUpdate handle agent status update 处理执行端状态更新
func (ws *WsHandlerFunc) handleAgentStatusUpdate(client *Client, in *model.WSRequest) {
	now := time.Now()
	client.markHeartbeat(now.UnixMilli(), now)

	// Refresh client status in database 刷新数据库中的客户端状态
	if err := updateClientLastSeen(client, in); err != nil {
		g.Log().Line().Errorf(client.Ctx, "处理客户端状态上报时更新客户端状态失败：client_ip=%s err=%+v", client.ClientIP(), err)
	}

	_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
		Type:      model.WSMessageTypeAgentStatusUpdateAck,
		BrowserID: in.BrowserID,
		ClientID:  resolveClientID(client, in),
		ClientIP:  client.ClientIP(),
	})
}

// handleAgentResult handle agent result message 处理执行端命令结果
func (ws *WsHandlerFunc) handleAgentResult(client *Client, in *model.WSRequest) {
	now := time.Now()
	client.markHeartbeat(now.UnixMilli(), now)

	// Refresh last seen time after command result 命令结果上报后刷新最近活跃时间
	if err := updateClientLastSeen(client, in); err != nil {
		g.Log().Line().Errorf(client.Ctx, "处理客户端命令结果时更新客户端状态失败：client_ip=%s err=%+v", client.ClientIP(), err)
	}

	// Update task record when command id belongs to task execution 更新任务执行记录
	if strings.HasPrefix(in.CommandID, "task-record-") {
		recordID := gconv.Int64(strings.TrimPrefix(in.CommandID, "task-record-"))
		if recordID > 0 {
			resultJSON := "{}"
			if len(in.Data) > 0 {
				resultJSON = string(in.Data)
			}
			status := "success"
			if !in.Success {
				status = "failed"
			}
			_, err := dao.TaskRecords.Ctx(client.Ctx).
				WherePri(recordID).
				Data(g.Map{
					dao.TaskRecords.Columns().Status:       status,
					dao.TaskRecords.Columns().ResultJson:   resultJSON,
					dao.TaskRecords.Columns().ErrorMessage: strings.TrimSpace(in.Error),
					dao.TaskRecords.Columns().FinishedAt:   gtime.Now(),
					dao.TaskRecords.Columns().UpdatedAt:    gtime.Now(),
				}).
				Update()
			if err != nil {
				g.Log().Line().Errorf(client.Ctx, "更新任务执行记录失败：record_id=%d err=%+v", recordID, err)
			}
		}
	}

	_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
		Type:      model.WSMessageTypeAgentResultAck,
		BrowserID: in.BrowserID,
		ClientID:  resolveClientID(client, in),
		ClientIP:  client.ClientIP(),
		CommandID: in.CommandID,
	})
}

// handleWorkflowInventory caches client workflow inventory 缓存客户端工作流清单
func (ws *WsHandlerFunc) handleWorkflowInventory(client *Client, in *model.WSRequest) {
	now := time.Now()
	client.markHeartbeat(now.UnixMilli(), now)
	workflowcache.TouchClient(client.Ctx, client.ClientIP())

	// Convert workflow list to GoFrame maps 转换工作流列表
	workflows := make([]g.Map, 0, len(in.Workflows))
	for _, workflow := range in.Workflows {
		if workflow != nil {
			workflows = append(workflows, g.Map(workflow))
		}
	}

	if err := workflowcache.SaveInventory(client.Ctx, client.ClientIP(), workflows); err != nil {
		g.Log().Line().Errorf(client.Ctx, "处理工作流清单上报时保存缓存失败：client_ip=%s err=%+v", client.ClientIP(), err)
		_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
			Type:     model.WSMessageTypeError,
			ClientIP: client.ClientIP(),
			Error:    "workflow inventory save failed",
			Message:  err.Error(),
		})
		return
	}

	_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
		Type:     model.WSMessageTypeWorkflowInventoryAck,
		ClientIP: client.ClientIP(),
		ClientID: resolveClientID(client, in),
		Data: g.Map{
			"workflow_count": len(workflows),
		},
	})
}

// resolveClientID resolve business client id 解析业务客户端标识
func resolveClientID(client *Client, in *model.WSRequest) string {
	// Prefer explicit client id 优先使用前端明确上报的客户端标识
	if in != nil {
		if clientID := strings.TrimSpace(in.ClientID); clientID != "" {
			return clientID
		}
		if browserID := strings.TrimSpace(in.BrowserID); browserID != "" {
			return browserID
		}
	}

	// Reuse bound client id 复用已绑定的客户端标识
	if client != nil {
		if clientID := strings.TrimSpace(client.ClientID()); clientID != "" {
			return clientID
		}
		return strings.TrimSpace(client.ClientIP())
	}

	return ""
}

// saveClientRegister save client register info 保存客户端注册信息
func saveClientRegister(client *Client, in *model.WSRequest, clientID string) error {
	if client == nil || in == nil || strings.TrimSpace(client.ClientIP()) == "" {
		return nil
	}

	// Prepare columns and timestamps 准备字段和时间
	columns := dao.Clients.Columns()
	now := gtime.Now()
	clientName := strings.TrimSpace(in.ClientName)
	if clientName == "" {
		clientName = clientID
	}

	// Build shared save data 构建新增和更新共用数据
	saveData := g.Map{
		columns.ClientId:       clientID,
		columns.ClientName:     clientName,
		columns.ClientIp:       client.ClientIP(),
		columns.UserAgent:      strings.TrimSpace(in.UserAgent),
		columns.Status:         "online",
		columns.PluginStatus:   resolvePluginStatus(in.AutomaInstalled),
		columns.AutomaVersion:  strings.TrimSpace(in.AutomaVersion),
		columns.BrowserName:    strings.TrimSpace(in.BrowserName),
		columns.BrowserVersion: strings.TrimSpace(in.BrowserVersion),
		columns.OsName:         strings.TrimSpace(in.OsName),
		columns.OsVersion:      strings.TrimSpace(in.OsVersion),
		columns.Hostname:       strings.TrimSpace(in.Hostname),
		columns.LastSeenAt:     now,
		columns.ConnectedAt:    now,
		columns.UpdatedAt:      now,
	}

	// Query existing client by ip 按客户端 IP 查询已有记录
	record, err := dao.Clients.Ctx(client.Ctx).
		Where(columns.ClientIp, client.ClientIP()).
		One()
	if err != nil {
		return err
	}

	// Insert new client when missing 不存在时新增客户端
	if record.IsEmpty() {
		saveData[columns.FirstSeenAt] = now
		saveData[columns.CreatedAt] = now
		_, err = dao.Clients.Ctx(client.Ctx).Data(saveData).Insert()
		return err
	}

	// Keep custom name unless empty 保留自定义名称，仅空名称时由同步补全
	if strings.TrimSpace(gconv.String(record[columns.ClientName])) != "" {
		delete(saveData, columns.ClientName)
	}
	_, err = dao.Clients.Ctx(client.Ctx).
		Where(columns.ClientIp, client.ClientIP()).
		Data(saveData).
		Update()
	return err
}

// updateClientLastSeen update client active status 更新客户端活跃状态
func updateClientLastSeen(client *Client, in *model.WSRequest) error {
	if client == nil {
		return nil
	}

	// Resolve client ip for update 解析要更新的客户端 IP
	clientIP := strings.TrimSpace(client.ClientIP())
	if clientIP == "" {
		return nil
	}
	clientID := resolveClientID(client, in)
	client.BindClientID(clientID)

	// Build update data 构建更新数据
	columns := dao.Clients.Columns()
	now := gtime.Now()
	updateData := g.Map{
		columns.ClientIp:   client.ClientIP(),
		columns.Status:     "online",
		columns.LastSeenAt: now,
		columns.UpdatedAt:  now,
	}
	if in != nil && (in.Type == model.WSMessageTypeAgentRegister || in.Type == model.WSMessageTypeAgentStatusUpdate) {
		updateData[columns.PluginStatus] = resolvePluginStatus(in.AutomaInstalled)
		if automaVersion := strings.TrimSpace(in.AutomaVersion); automaVersion != "" || in.Type == model.WSMessageTypeAgentRegister {
			updateData[columns.AutomaVersion] = automaVersion
		}
	}

	// Update matched client 按客户端 IP 更新记录
	_, err := dao.Clients.Ctx(client.Ctx).
		Where(columns.ClientIp, clientIP).
		Data(updateData).
		Update()
	return err
}

// markClientOffline mark client disconnected 标记客户端离线
func markClientOffline(client *Client) error {
	if client == nil {
		return nil
	}

	// Resolve client ip 解析客户端 IP
	clientIP := strings.TrimSpace(client.ClientIP())
	if clientIP == "" {
		return nil
	}

	// Update offline status 更新离线状态
	columns := dao.Clients.Columns()
	now := gtime.Now()
	_, err := dao.Clients.Ctx(client.Ctx).
		Where(columns.ClientIp, clientIP).
		Data(g.Map{
			columns.Status:         "offline",
			columns.DisconnectedAt: now,
			columns.UpdatedAt:      now,
		}).
		Update()
	return err
}

// resolvePluginStatus resolve Automa plugin status 解析 Automa 插件状态
func resolvePluginStatus(automaInstalled bool) string {
	if automaInstalled {
		return "installed"
	}
	return "not_installed"
}

// OnOpen handle websocket open 处理连接建立
func (ws *WsHandlerFunc) OnOpen(client *Client) {
	if client == nil || client.Ctx == nil {
		return
	}

	ctx, cancel := context.WithCancel(client.Ctx)

	ws.mu.Lock()
	ws.ClientCtxMap[client.ClientIP()] = ctx
	ws.ClientCtxCancel[client.ClientIP()] = cancel
	ws.mu.Unlock()

	g.Log().Line().Infof(client.Ctx, "WebSocket 连接已建立：client_ip=%s", client.ClientIP())
}

// OnClose handle websocket close 处理连接关闭
func (ws *WsHandlerFunc) OnClose(client *Client) {
	if client == nil {
		return
	}

	// Mark database client offline 标记数据库客户端离线
	if err := markClientOffline(client); err != nil {
		g.Log().Line().Errorf(client.Ctx, "WebSocket 断开时标记客户端离线失败：client_ip=%s err=%+v", client.ClientIP(), err)
	}

	ws.mu.Lock()
	cancel, ok := ws.ClientCtxCancel[client.ClientIP()]
	delete(ws.ClientCtxMap, client.ClientIP())
	delete(ws.ClientCtxCancel, client.ClientIP())
	ws.mu.Unlock()

	if ok {
		cancel()
	}

	g.Log().Line().Infof(client.Ctx, "WebSocket 连接已断开：client_ip=%s", client.ClientIP())
}
