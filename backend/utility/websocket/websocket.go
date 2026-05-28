package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/tasklock"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gorilla/websocket"
)

const (
	// defaultMaxConn default max connection count
	defaultMaxConn = 1000000
)

var (
	// once singleton initializer
	once sync.Once
	// WsManage global manager
	WsManage *WebSocketManager
	// WsHandler global default handler
	WsHandler *WsHandlerFunc
)

// WsHandlerFunc default websocket handler.
type WsHandlerFunc struct {
	// mode runtime mode
	mode string
	// mu protect context maps
	mu sync.RWMutex
	// ClientCtxMap client contexts
	ClientCtxMap map[string]context.Context
	// ClientCtxCancel client cancel funcs
	ClientCtxCancel map[string]context.CancelFunc
}

// Init init websocket manager.
func Init(ctx context.Context) {
	once.Do(func() {
		WsHandler = &WsHandlerFunc{
			mode:            consts.ResolveRuntimeMode(ctx),
			ClientCtxMap:    make(map[string]context.Context),
			ClientCtxCancel: make(map[string]context.CancelFunc),
		}
		WsManage = InitWebSocketManagerWithConfig(WsHandler, loadConfig(ctx))
	})
}

// loadConfig load websocket config.
func loadConfig(ctx context.Context) Config {
	return Config{
		HeartbeatInterval: time.Duration(g.Cfg().MustGet(ctx, "websocket.heartbeatInterval", 15).Int()) * time.Second,
		HeartbeatTimeout:  time.Duration(g.Cfg().MustGet(ctx, "websocket.heartbeatTimeout", 45).Int()) * time.Second,
		WriteWait:         time.Duration(g.Cfg().MustGet(ctx, "websocket.writeWait", 10).Int()) * time.Second,
		MessageBuffer:     g.Cfg().MustGet(ctx, "websocket.messageBuffer", 64).Int(),
		BucketCount:       g.Cfg().MustGet(ctx, "websocket.bucketCount", defaultBucketCount).Int(),
	}
}

// MaxConn get max websocket count.
func MaxConn(ctx context.Context) int64 {
	return g.Cfg().MustGet(ctx, "websocket.maxConn", defaultMaxConn).Int64()
}

// BuildClientIdentity build client identity.
func BuildClientIdentity(clientIP string) ClientIdentity {
	return ClientIdentity{
		ConnectionID:     clientIP,
		ClientIP:         clientIP,
		RequireHeartbeat: true,
	}
}

// BuildNodeIdentity builds a server execution-node identity.
func BuildNodeIdentity(clientIP string, nodeID string) ClientIdentity {
	connectionID := NodeConnectionID(clientIP, nodeID)
	if connectionID == "" {
		connectionID = strings.TrimSpace(clientIP)
	}
	return ClientIdentity{
		ConnectionID:     connectionID,
		ClientIP:         strings.TrimSpace(clientIP),
		NodeID:           NormalizeNodeID(clientIP, nodeID),
		RequireHeartbeat: true,
	}
}

// BuildConnectionIdentity build arbitrary connection identity.
func BuildConnectionIdentity(connectionID, clientIP string, requireHeartbeat bool) ClientIdentity {
	return ClientIdentity{
		ConnectionID:     connectionID,
		ClientIP:         clientIP,
		RequireHeartbeat: requireHeartbeat,
	}
}

// SendClientMessage send structured websocket message.
func SendClientMessage(clientIP string, in *model.WSResponse) int {
	return sendStructuredMessage(clientIP, in)
}

// SendNodeMessage sends a message to an execution node, falling back to legacy ip.
func SendNodeMessage(nodeID string, clientIP string, in *model.WSResponse) int {
	if sent := sendStructuredMessage(NodeConnectionID(clientIP, nodeID), in); sent > 0 {
		return sent
	}
	return sendStructuredMessage(strings.TrimSpace(clientIP), in)
}

// NodeConnectionID builds the runtime identity for one IP-scoped node 浣跨敤 IP + node 鏍囪瘑鎵ц鑺傜偣
func NodeConnectionID(clientIP string, nodeID string) string {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = NormalizeNodeID(clientIP, nodeID)
	if clientIP == "" {
		return nodeID
	}
	if nodeID == "" || nodeID == clientIP {
		return clientIP
	}
	return clientIP + "|" + nodeID
}

// NormalizeNodeID strips duplicated IP prefixes from node id. 鍘绘帀 node_id 涓噸澶嶆嫾鎺ョ殑 IP 鍓嶇紑
func NormalizeNodeID(clientIP string, nodeID string) string {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" || nodeID == "" {
		return nodeID
	}

	prefix := clientIP + "|"
	for strings.HasPrefix(nodeID, prefix) {
		nodeID = strings.TrimSpace(strings.TrimPrefix(nodeID, prefix))
	}
	return nodeID
}

// SendConnectionMessage send structured message by connection id.
func SendConnectionMessage(connectionID string, in *model.WSResponse) int {
	return sendStructuredMessage(connectionID, in)
}

// sendStructuredMessage sends one JSON websocket message.
func sendStructuredMessage(connectionID string, in *model.WSResponse) int {
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
	return WsManage.SendMessageToClient(connectionID, body)
}

// SendRawClientMessage send raw websocket message.
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

// SendRawClientMessages send raw websocket messages in batch.
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

// OnMessage handle websocket message.
func (ws *WsHandlerFunc) OnMessage(client *Client, messageType int, message []byte) {
	if client == nil {
		return
	}
	if messageType != websocket.TextMessage {
		g.Log().Line().Infof(client.Ctx, "ignore non-text websocket message: client_ip=%s message_type=%d", client.ClientIP(), messageType)
		return
	}

	var in model.WSRequest
	if err := json.Unmarshal(message, &in); err != nil {
		g.Log().Line().Infof(client.Ctx, "parse websocket message failed: client_ip=%s message=%s", client.ClientIP(), string(message))
		return
	}

	if ws.mode != consts.RuntimeModeServer {
		ws.handleDesktopMessage(client, &in)
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
		_ = SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypePong})
	default:
		g.Log().Line().Infof(client.Ctx, "unknown websocket message type: client_ip=%s type=%s", client.ClientIP(), in.Type)
	}
}

// handleHeartbeat handle heartbeat message.
func (ws *WsHandlerFunc) handleHeartbeat(client *Client, in *model.WSRequest) {
	now := time.Now()
	bindNodeIdentity(client, in)
	client.markHeartbeat(in.ClientTime, now)
	workflowcache.TouchNode(client.Ctx, executionNodeID(client, in), resolveNodeName(in), "", client.ClientIP())
	if commandID := resolveExecutionCommandID(in); commandID != "" {
		if _, err := tasklock.RenewNode(client.Ctx, executionNodeID(client, in), client.ClientIP(), commandID); err != nil {
			g.Log().Line().Warningf(client.Ctx, "renew client task lock failed: client_ip=%s command_id=%s err=%+v", client.ClientIP(), commandID, err)
		}
	}

	// Refresh client online state.
	if err := updateClientLastSeen(client, in); err != nil {
		g.Log().Line().Errorf(client.Ctx, "update client last seen failed: client_ip=%s err=%+v", client.ClientIP(), err)
	}

	_ = SendNodeMessage(executionNodeID(client, in), client.ClientIP(), &model.WSResponse{
		Type:       model.WSMessageTypeHeartbeatAck,
		BrowserID:  in.BrowserID,
		ClientID:   resolveClientID(client, in),
		ClientIP:   client.ClientIP(),
		NodeID:     executionNodeID(client, in),
		NodeName:   resolveNodeName(in),
		ClientTime: in.ClientTime,
		ServerTime: now.UnixMilli(),
		Data: map[string]any{
			"last_heartbeat_time": client.LastHeartbeatTime(),
			"heartbeat_interval":  client.manager.heartbeatInterval.Milliseconds(),
			"heartbeat_timeout":   client.manager.heartbeatTimeout.Milliseconds(),
		},
	})
}

// RequestTaskRecovery asks online clients to report an existing locked task.
func RequestTaskRecovery(ctx context.Context) {
	Init(ctx)
	locks, err := tasklock.List(ctx)
	if err != nil {
		g.Log().Line().Warningf(ctx, "scan client task locks for recovery failed: %+v", err)
		return
	}
	for _, lockInfo := range locks {
		requestClientTaskRecovery(ctx, lockInfo)
	}
}

func requestClientTaskRecovery(ctx context.Context, lockInfo tasklock.LockInfo) {
	clientIP := strings.TrimSpace(lockInfo.ClientIP)
	nodeID := strings.TrimSpace(lockInfo.NodeID)
	commandID := strings.TrimSpace(lockInfo.CommandID)
	if clientIP == "" && nodeID == "" || !strings.HasPrefix(commandID, "task-record-") {
		return
	}
	connectionID := NodeConnectionID(clientIP, nodeID)
	if WsManage == nil || !WsManage.HasClient(connectionID) {
		return
	}

	recordID := tasklock.RecordIDFromCommand(commandID)
	if recordID <= 0 {
		return
	}

	persistence := persistenceHandler()
	if persistence == nil {
		return
	}
	canRecover, err := persistence.CanRecoverTask(ctx, recordID)
	if err != nil {
		g.Log().Line().Warningf(ctx, "query task record before recovery failed: record_id=%d err=%+v", recordID, err)
		return
	}
	if !canRecover {
		_ = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID)
		return
	}

	sent := SendNodeMessage(nodeID, clientIP, &model.WSResponse{
		Type:      model.WSMessageTypeAgentCommand,
		ClientIP:  clientIP,
		NodeID:    nodeID,
		CommandID: commandID,
		Command:   "task.status.query",
		Payload: map[string]any{
			"task_record_id": recordID,
			"task_id":        lockInfo.TaskID,
			"workflow_id":    lockInfo.WorkflowID,
			"command_id":     commandID,
			"execution_id":   commandID,
			"recover":        true,
		},
	})
	if sent <= 0 {
		g.Log().Line().Warningf(ctx, "send task recovery query failed: client_ip=%s command_id=%s", clientIP, commandID)
	}
}

// handleAgentRegister handle agent register message.
func (ws *WsHandlerFunc) handleAgentRegister(client *Client, in *model.WSRequest) {
	now := time.Now()
	bindNodeIdentity(client, in)
	ws.rebindRegisteredNode(client, executionNodeID(client, in))
	client.markHeartbeat(now.UnixMilli(), now)

	// Resolve client id before persistence.
	clientID := resolveClientID(client, in)
	client.BindClientID(clientID)

	// Persist client information.
	if err := saveClientRegister(client, in, clientID); err != nil {
		g.Log().Line().Errorf(client.Ctx, "save client register failed: client_ip=%s err=%+v", client.ClientIP(), err)
		_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
			Type:    model.WSMessageTypeError,
			Error:   "client information save failed",
			Message: err.Error(),
		})
		return
	}

	_ = SendNodeMessage(executionNodeID(client, in), client.ClientIP(), &model.WSResponse{
		Type:            model.WSMessageTypeAgentRegistered,
		BrowserID:       in.BrowserID,
		ClientID:        clientID,
		ClientIP:        client.ClientIP(),
		NodeID:          executionNodeID(client, in),
		NodeName:        resolveNodeName(in),
		Role:            in.Role,
		AutomaInstalled: in.AutomaInstalled,
		AutomaVersion:   in.AutomaVersion,
	})

	lockInfo, hasLock, err := tasklock.GetNode(client.Ctx, executionNodeID(client, in), client.ClientIP())
	if err != nil {
		g.Log().Line().Warningf(client.Ctx, "read client task lock after register failed: client_ip=%s err=%+v", client.ClientIP(), err)
		return
	}
	if hasLock {
		requestClientTaskRecovery(client.Ctx, lockInfo)
	}
}

func (ws *WsHandlerFunc) rebindRegisteredNode(client *Client, nodeID string) {
	if client == nil || WsManage == nil {
		return
	}
	nodeID = strings.TrimSpace(nodeID)
	connectionID := NodeConnectionID(client.ClientIP(), nodeID)
	if connectionID == "" || connectionID == strings.TrimSpace(client.ClientIP()) {
		return
	}

	previous, oldConnectionID, rebound := WsManage.RebindClientConnection(client, connectionID)
	if !rebound {
		return
	}

	ws.mu.Lock()
	if ctx, ok := ws.ClientCtxMap[oldConnectionID]; ok {
		ws.ClientCtxMap[connectionID] = ctx
		delete(ws.ClientCtxMap, oldConnectionID)
	}
	if cancel, ok := ws.ClientCtxCancel[oldConnectionID]; ok {
		ws.ClientCtxCancel[connectionID] = cancel
		delete(ws.ClientCtxCancel, oldConnectionID)
	}
	ws.mu.Unlock()

	if previous != nil && previous != client {
		previous.MarkSuperseded()
		WsManage.disConnect(previous)
	}
	g.Log().Line().Infof(client.Ctx, "WebSocket rebound to node: old_connection_id=%s connection_id=%s node_id=%s client_ip=%s", oldConnectionID, connectionID, nodeID, client.ClientIP())
}

// handleAgentStatusUpdate handle agent status update.
func (ws *WsHandlerFunc) handleAgentStatusUpdate(client *Client, in *model.WSRequest) {
	now := time.Now()
	bindNodeIdentity(client, in)
	client.markHeartbeat(now.UnixMilli(), now)

	// Refresh client status in database.
	if err := updateClientLastSeen(client, in); err != nil {
		g.Log().Line().Errorf(client.Ctx, "update client status failed: client_ip=%s err=%+v", client.ClientIP(), err)
	}

	_ = SendNodeMessage(executionNodeID(client, in), client.ClientIP(), &model.WSResponse{
		Type:      model.WSMessageTypeAgentStatusUpdateAck,
		BrowserID: in.BrowserID,
		ClientID:  resolveClientID(client, in),
		ClientIP:  client.ClientIP(),
		NodeID:    executionNodeID(client, in),
	})
}

// handleAgentResult handle agent result message.
func (ws *WsHandlerFunc) handleAgentResult(client *Client, in *model.WSRequest) {
	now := time.Now()
	bindNodeIdentity(client, in)
	client.markHeartbeat(now.UnixMilli(), now)
	browserID := in.BrowserID
	if browserID == "" {
		browserID = resolveClientID(client, in)
	}
	var resultData []byte
	if len(in.Data) > 0 {
		resultBytes, marshalErr := json.Marshal(in.Data)
		if marshalErr != nil {
			resultData = []byte(fmt.Sprintf(`{"error":%q}`, marshalErr.Error()))
		} else {
			resultData = resultBytes
		}
	}
	result := model.AgentCommandResult{BrowserID: browserID, CommandID: in.CommandID, Success: in.Success, Data: resultData, Error: in.Error}

	// Refresh last seen time after command result.
	if err := updateClientLastSeen(client, in); err != nil {
		g.Log().Line().Errorf(client.Ctx, "update client last seen after command result failed: client_ip=%s err=%+v", client.ClientIP(), err)
	}

	// Update task record when command id belongs to task execution.
	// Notify command waiter and update execution state.
	state.AgentMu.Lock()
	resultCh := state.PopPendingCommand(in.CommandID)
	if agent := state.AgentConnections[browserID]; agent != nil {
		agent.LastSeenAt = now
	}
	state.AgentMu.Unlock()
	workflowexecution.CompleteByCommand(in.CommandID, &result)

	if strings.HasPrefix(in.CommandID, "task-record-") {
		recordID := gconv.Int64(strings.TrimPrefix(in.CommandID, "task-record-"))
		if recordID > 0 {
			if err := updateTaskRecordFromAgentResult(client.Ctx, TaskRecordResultData{
				RecordID:  recordID,
				NodeID:    executionNodeID(client, in),
				ClientIP:  client.ClientIP(),
				CommandID: in.CommandID,
				Success:   in.Success,
				Result:    resultData,
				ErrorText: in.Error,
			}); err != nil {
				g.Log().Line().Errorf(client.Ctx, "update task record failed: record_id=%d err=%+v", recordID, err)
			}
		}
	}

	if resultCh != nil {
		resultCh <- result
	}

	_ = SendNodeMessage(executionNodeID(client, in), client.ClientIP(), &model.WSResponse{
		Type:      model.WSMessageTypeAgentResultAck,
		BrowserID: in.BrowserID,
		ClientID:  resolveClientID(client, in),
		ClientIP:  client.ClientIP(),
		NodeID:    executionNodeID(client, in),
		CommandID: in.CommandID,
	})
}

func updateTaskRecordFromAgentResult(ctx context.Context, data TaskRecordResultData) error {
	persistence := persistenceHandler()
	if persistence == nil {
		return nil
	}
	if err := persistence.UpdateTaskRecordResult(ctx, data); err != nil {
		return err
	}
	status := ResolveTaskRecordResultStatus(data.Success, data.Result)
	if status == "running" {
		if _, renewErr := tasklock.RenewNode(ctx, data.NodeID, data.ClientIP, data.CommandID); renewErr != nil {
			g.Log().Line().Warningf(ctx, "renew client task lock failed: client_ip=%s command_id=%s err=%+v", data.ClientIP, data.CommandID, renewErr)
		}
		return nil
	}
	if status == "success" || status == "failed" {
		if err := tasklock.ReleaseNode(ctx, data.NodeID, data.ClientIP, data.CommandID); err != nil {
			g.Log().Line().Warningf(ctx, "release client task lock failed: client_ip=%s command_id=%s err=%+v", data.ClientIP, data.CommandID, err)
		}
		if err := persistence.MarkNodeIdle(ctx, data.ClientIP, data.NodeID, data.CommandID, data.RecordID); err != nil {
			g.Log().Line().Warningf(ctx, "mark client node idle failed: client_ip=%s command_id=%s err=%+v", data.ClientIP, data.CommandID, err)
		}
	}
	return nil
}

func resolveExecutionCommandID(in *model.WSRequest) string {
	if in == nil {
		return ""
	}
	for _, value := range []string{in.ExecutionID, in.CommandID} {
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, "task-record-") {
			return value
		}
	}
	if len(in.Data) > 0 {
		for _, key := range []string{"execution_id", "executionId", "command_id", "commandId"} {
			value := strings.TrimSpace(gconv.String(in.Data[key]))
			if strings.HasPrefix(value, "task-record-") {
				return value
			}
		}
	}
	return ""
}

// resolveTaskRecordResultStatus maps agent result to task status.
func ResolveTaskRecordResultStatus(success bool, resultData []byte) string {
	if !success {
		return "failed"
	}

	var data map[string]any
	if len(resultData) > 0 && json.Unmarshal(resultData, &data) == nil {
		status := strings.ToLower(strings.TrimSpace(gconv.String(data["status"])))
		switch status {
		case "error", "failed", "fail", "timeout", "stopped", "cancelled", "canceled", "unknown", "not_found", "missing", "lost":
			return "failed"
		case "queued", "submitted", "pending", "running":
			return "running"
		case "success", "finished", "done", "completed":
			return "success"
		}
		if okValue, ok := data["ok"].(bool); ok && !okValue {
			return "failed"
		}
	}

	return "success"
}

func ResolveAutomaExecutionID(resultData []byte) string {
	var data map[string]any
	if len(resultData) == 0 || json.Unmarshal(resultData, &data) != nil {
		return ""
	}
	for _, key := range []string{"automa_execution_id", "automaExecutionId", "automa_run_id", "automaRunId", "history_id", "historyId"} {
		if value := strings.TrimSpace(gconv.String(data[key])); value != "" {
			return value
		}
	}
	return ""
}

// isLostTaskResultStatus reports recovery states where the client cannot find the execution.
func IsLostTaskResultStatus(resultData []byte) bool {
	var data map[string]any
	if len(resultData) == 0 || json.Unmarshal(resultData, &data) != nil {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(gconv.String(data["status"]))) {
	case "unknown", "not_found", "missing", "lost":
		return true
	default:
		return false
	}
}

// resolveTaskRecordResultMessage extracts readable failure message.
func ResolveTaskRecordResultMessage(resultData []byte) string {
	var data map[string]any
	if len(resultData) == 0 || json.Unmarshal(resultData, &data) != nil {
		return ""
	}
	for _, key := range []string{"message", "error", "error_message", "errorMessage"} {
		if value := strings.TrimSpace(gconv.String(data[key])); value != "" {
			return value
		}
	}
	return ""
}

// handleWorkflowInventory caches client workflow inventory.
func (ws *WsHandlerFunc) handleWorkflowInventory(client *Client, in *model.WSRequest) {
	now := time.Now()
	bindNodeIdentity(client, in)
	client.markHeartbeat(now.UnixMilli(), now)
	workflowcache.TouchNode(client.Ctx, executionNodeID(client, in), resolveNodeName(in), "", client.ClientIP())

	// Convert workflow list to GoFrame maps.
	workflows := make([]g.Map, 0, len(in.Workflows))
	for _, workflow := range in.Workflows {
		if workflow != nil {
			workflows = append(workflows, g.Map(workflow))
		}
	}

	if err := workflowcache.SaveNodeInventory(client.Ctx, executionNodeID(client, in), resolveNodeName(in), "", client.ClientIP(), workflows); err != nil {
		g.Log().Line().Errorf(client.Ctx, "save workflow inventory failed: client_ip=%s err=%+v", client.ClientIP(), err)
		_ = SendClientMessage(client.ClientIP(), &model.WSResponse{
			Type:     model.WSMessageTypeError,
			ClientIP: client.ClientIP(),
			Error:    "workflow inventory save failed",
			Message:  err.Error(),
		})
		return
	}

	_ = SendNodeMessage(executionNodeID(client, in), client.ClientIP(), &model.WSResponse{
		Type:     model.WSMessageTypeWorkflowInventoryAck,
		ClientIP: client.ClientIP(),
		NodeID:   executionNodeID(client, in),
		NodeName: resolveNodeName(in),
		ClientID: resolveClientID(client, in),
		Data: map[string]any{
			"workflow_count": len(workflows),
		},
	})
}

// resolveClientID resolve business client id.
func resolveClientID(client *Client, in *model.WSRequest) string {
	// Prefer explicit client id.
	if in != nil {
		if clientID := strings.TrimSpace(in.ClientID); clientID != "" {
			return clientID
		}
		if browserID := strings.TrimSpace(in.BrowserID); browserID != "" {
			return browserID
		}
	}

	// Reuse bound client id.
	if client != nil {
		if clientID := strings.TrimSpace(client.ClientID()); clientID != "" {
			return clientID
		}
		return strings.TrimSpace(client.ClientIP())
	}

	return ""
}

func bindNodeIdentity(client *Client, in *model.WSRequest) {
	if client == nil {
		return
	}
	nodeID := executionNodeID(client, in)
	client.BindNodeIdentity(nodeID)
}

func executionNodeID(client *Client, in *model.WSRequest) string {
	if in != nil {
		if nodeID := strings.TrimSpace(in.NodeID); nodeID != "" {
			return nodeID
		}
	}
	if client != nil {
		if nodeID := strings.TrimSpace(client.NodeID()); nodeID != "" {
			return nodeID
		}
		return strings.TrimSpace(client.ClientIP())
	}
	return ""
}

func resolveNodeName(in *model.WSRequest) string {
	if in == nil {
		return ""
	}
	return strings.TrimSpace(in.NodeName)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

// saveClientRegister save client register info.
func saveClientRegister(client *Client, in *model.WSRequest, clientID string) error {
	if client == nil || in == nil || strings.TrimSpace(client.ClientIP()) == "" {
		return nil
	}
	persistence := persistenceHandler()
	if persistence == nil {
		return nil
	}
	return persistence.SaveClientRegister(client.Ctx, NodeRegisterData{
		ClientIP:       client.ClientIP(),
		NodeID:         executionNodeID(client, in),
		NodeIndex:      in.NodeIndex,
		WorkerVersion:  in.WorkerVersion,
		ProfileDir:     in.ProfileDir,
		ExtensionDir:   in.ExtensionDir,
		UserAgent:      in.UserAgent,
		PluginStatus:   resolvePluginStatus(in.AutomaInstalled),
		AutomaVersion:  in.AutomaVersion,
		BrowserName:    in.BrowserName,
		BrowserVersion: in.BrowserVersion,
		OsName:         in.OsName,
		OsVersion:      in.OsVersion,
		Hostname:       in.Hostname,
		Capabilities:   in.Capabilities,
	})
}

// updateClientLastSeen update client active status.
func updateClientLastSeen(client *Client, in *model.WSRequest) error {
	if client == nil {
		return nil
	}

	// Resolve client ip for update.
	clientIP := strings.TrimSpace(client.ClientIP())
	if clientIP == "" {
		return nil
	}
	clientID := resolveClientID(client, in)
	client.BindClientID(clientID)
	bindNodeIdentity(client, in)
	nodeID := executionNodeID(client, in)

	persistence := persistenceHandler()
	if persistence == nil {
		return nil
	}
	data := NodeHeartbeatData{
		ClientIP: client.ClientIP(),
		NodeID:   nodeID,
	}
	if commandID := resolveExecutionCommandID(in); commandID != "" {
		data.CommandID = commandID
		data.RecordID = tasklock.RecordIDFromCommand(commandID)
	}
	if in != nil && (in.Type == model.WSMessageTypeAgentRegister || in.Type == model.WSMessageTypeAgentStatusUpdate) {
		data.PluginStatus = resolvePluginStatus(in.AutomaInstalled)
		data.AutomaVersion = in.AutomaVersion
		data.UpdatePluginInfo = true
	}
	return persistence.UpdateClientLastSeen(client.Ctx, data)
}

// markClientOffline mark client disconnected.
func markClientOffline(client *Client) error {
	if client == nil {
		return nil
	}

	// Resolve client ip.
	clientIP := strings.TrimSpace(client.ClientIP())
	nodeID := strings.TrimSpace(client.ExecutionNodeID())
	if clientIP == "" && nodeID == "" {
		return nil
	}

	persistence := persistenceHandler()
	if persistence == nil {
		return nil
	}
	return persistence.MarkClientOffline(client.Ctx, clientIP, nodeID)
}

// resolvePluginStatus resolve Automa plugin status.
func resolvePluginStatus(automaInstalled bool) string {
	if automaInstalled {
		return "installed"
	}
	return "not_installed"
}

// failClientRunningTask marks the disconnected client active task failed.
func failClientRunningTask(client *Client) {
	if client == nil {
		return
	}

	clientIP := strings.TrimSpace(client.ClientIP())
	nodeID := strings.TrimSpace(client.ExecutionNodeID())
	if clientIP == "" && nodeID == "" {
		return
	}

	lockInfo, hasLock, err := tasklock.GetNode(client.Ctx, nodeID, clientIP)
	if err != nil {
		g.Log().Line().Warningf(client.Ctx, "read client task lock on close failed: client_ip=%s err=%+v", clientIP, err)
		return
	}
	if !hasLock || !strings.HasPrefix(lockInfo.CommandID, "task-record-") {
		return
	}

	recordID := tasklock.RecordIDFromCommand(lockInfo.CommandID)
	if recordID > 0 {
		persistence := persistenceHandler()
		if persistence != nil {
			err = persistence.FailClientRunningTask(client.Ctx, recordID, "client disconnected, task execution was automatically ended")
		}
		if err != nil {
			g.Log().Line().Warningf(client.Ctx, "fail client running task on close failed: record_id=%d client_ip=%s err=%+v", recordID, clientIP, err)
		}
	}

	state.RemovePendingCommand(lockInfo.CommandID)
	if err = tasklock.ReleaseNode(client.Ctx, nodeID, clientIP, lockInfo.CommandID); err != nil {
		g.Log().Line().Warningf(client.Ctx, "release client task lock on close failed: client_ip=%s command_id=%s err=%+v", clientIP, lockInfo.CommandID, err)
	}
	if persistence := persistenceHandler(); persistence != nil {
		if err = persistence.MarkNodeIdle(client.Ctx, clientIP, nodeID, lockInfo.CommandID, recordID); err != nil {
			g.Log().Line().Warningf(client.Ctx, "mark client node idle on close failed: client_ip=%s command_id=%s err=%+v", clientIP, lockInfo.CommandID, err)
		}
	}
}

// OnOpen handle websocket open.
func (ws *WsHandlerFunc) OnOpen(client *Client) {
	if client == nil || client.Ctx == nil {
		return
	}

	ctx, cancel := context.WithCancel(client.Ctx)

	ws.mu.Lock()
	ws.ClientCtxMap[client.ConnectionID()] = ctx
	ws.ClientCtxCancel[client.ConnectionID()] = cancel
	ws.mu.Unlock()

	//g.Log().Line().Warningf(client.Ctx, "WebSocket connected: connection_id=%s client_ip=%s", client.ConnectionID(), client.ClientIP())
}

// OnClose handle websocket close.
func (ws *WsHandlerFunc) OnClose(client *Client) {
	if client == nil {
		return
	}

	if ws.mode == consts.RuntimeModeServer {
		if client.IsSuperseded() {
			g.Log().Line().Infof(client.Ctx, "WebSocket close skipped for superseded client: connection_id=%s client_ip=%s", client.ConnectionID(), client.ClientIP())
		} else {
			// Mark database client offline.
			if err := markClientOffline(client); err != nil {
				g.Log().Line().Errorf(client.Ctx, "WebSocket close mark client offline failed: client_ip=%s err=%+v", client.ClientIP(), err)
			}
			if err := workflowcache.ClearNode(client.Ctx, client.ExecutionIdentity()); err != nil {
				g.Log().Line().Errorf(client.Ctx, "WebSocket close clear workflow cache failed: client_ip=%s err=%+v", client.ClientIP(), err)
			}
			failClientRunningTask(client)
		}
	} else {
		ws.handleDesktopClose(client)
	}

	ws.mu.Lock()
	cancel, ok := ws.ClientCtxCancel[client.ConnectionID()]
	delete(ws.ClientCtxMap, client.ConnectionID())
	delete(ws.ClientCtxCancel, client.ConnectionID())
	ws.mu.Unlock()

	if ok {
		cancel()
	}

	g.Log().Line().Infof(client.Ctx, "WebSocket disconnected: connection_id=%s client_ip=%s", client.ConnectionID(), client.ClientIP())
}
