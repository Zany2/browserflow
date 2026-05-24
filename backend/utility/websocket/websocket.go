package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/tasklock"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
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
func BuildNodeIdentity(clientIP string, machineID string, nodeID string) ClientIdentity {
	connectionID := strings.TrimSpace(nodeID)
	if connectionID == "" {
		connectionID = strings.TrimSpace(clientIP)
	}
	return ClientIdentity{
		ConnectionID:     connectionID,
		ClientIP:         strings.TrimSpace(clientIP),
		MachineID:        strings.TrimSpace(machineID),
		NodeID:           strings.TrimSpace(nodeID),
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
	if sent := sendStructuredMessage(strings.TrimSpace(nodeID), in); sent > 0 {
		return sent
	}
	return sendStructuredMessage(strings.TrimSpace(clientIP), in)
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
	workflowcache.TouchNode(client.Ctx, executionNodeID(client, in), resolveNodeName(in), resolveMachineID(client, in), client.ClientIP())
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
		MachineID:  resolveMachineID(client, in),
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
	connectionID := firstNonEmpty(nodeID, clientIP)
	if WsManage == nil || !WsManage.HasClient(connectionID) {
		return
	}

	recordID := tasklock.RecordIDFromCommand(commandID)
	if recordID <= 0 {
		return
	}

	columns := dao.TaskRecords.Columns()
	record, err := dao.TaskRecords.Ctx(ctx).
		Fields(columns.Id, columns.Status).
		WherePri(recordID).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		One()
	if err != nil {
		g.Log().Line().Warningf(ctx, "query task record before recovery failed: record_id=%d err=%+v", recordID, err)
		return
	}
	if record.IsEmpty() {
		_ = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID)
		return
	}

	sent := SendNodeMessage(nodeID, clientIP, &model.WSResponse{
		Type:      model.WSMessageTypeAgentCommand,
		ClientIP:  clientIP,
		MachineID: lockInfo.MachineID,
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
		MachineID:       resolveMachineID(client, in),
		MachineName:     resolveMachineName(in),
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
		MachineID: resolveMachineID(client, in),
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
			if err := updateTaskRecordFromAgentResult(client.Ctx, recordID, executionNodeID(client, in), client.ClientIP(), in.CommandID, in.Success, resultData, in.Error); err != nil {
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
		MachineID: resolveMachineID(client, in),
		NodeID:    executionNodeID(client, in),
		CommandID: in.CommandID,
	})
}

// resolveExecutionCommandID returns current task command id from heartbeat payload.
func updateTaskRecordFromAgentResult(ctx context.Context, recordID int64, nodeID string, clientIP string, commandID string, success bool, resultData []byte, errorText string) error {
	resultJSON := "{}"
	if len(resultData) > 0 {
		resultJSON = string(resultData)
	}
	resultJSON = saveTaskRecordResultFiles(ctx, recordID, clientIP, resultJSON)
	status := resolveTaskRecordResultStatus(success, resultData)
	errorMessage := strings.TrimSpace(errorText)
	if errorMessage == "" && status == "failed" {
		errorMessage = resolveTaskRecordResultMessage(resultData)
		if errorMessage == "" && isLostTaskResultStatus(resultData) {
			errorMessage = "client has no local execution state for this task"
		}
	}
	updateData := do.TaskRecords{
		Status:       status,
		ResultJson:   resultJSON,
		ErrorMessage: errorMessage,
	}
	if automaExecutionID := resolveAutomaExecutionID(resultData); automaExecutionID != "" {
		updateData.AutomaExecutionId = automaExecutionID
	}
	if status == "running" {
		updateData.StartedAt = gtime.Now()
	}
	if status == "success" || status == "failed" {
		updateData.FinishedAt = gtime.Now()
	}

	columns := dao.TaskRecords.Columns()
	_, err := dao.TaskRecords.Ctx(ctx).
		WherePri(recordID).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		Data(updateData).
		Update()
	if err != nil {
		return err
	}
	if status == "running" {
		if _, renewErr := tasklock.RenewNode(ctx, nodeID, clientIP, commandID); renewErr != nil {
			g.Log().Line().Warningf(ctx, "renew client task lock failed: client_ip=%s command_id=%s err=%+v", clientIP, commandID, renewErr)
		}
		return nil
	}
	if status == "success" || status == "failed" {
		if err = tasklock.ReleaseNode(ctx, nodeID, clientIP, commandID); err != nil {
			g.Log().Line().Warningf(ctx, "release client task lock failed: client_ip=%s command_id=%s err=%+v", clientIP, commandID, err)
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
func resolveTaskRecordResultStatus(success bool, resultData []byte) string {
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

func resolveAutomaExecutionID(resultData []byte) string {
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
func isLostTaskResultStatus(resultData []byte) bool {
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
func resolveTaskRecordResultMessage(resultData []byte) string {
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
	workflowcache.TouchNode(client.Ctx, executionNodeID(client, in), resolveNodeName(in), resolveMachineID(client, in), client.ClientIP())

	// Convert workflow list to GoFrame maps.
	workflows := make([]g.Map, 0, len(in.Workflows))
	for _, workflow := range in.Workflows {
		if workflow != nil {
			workflows = append(workflows, g.Map(workflow))
		}
	}

	if err := workflowcache.SaveNodeInventory(client.Ctx, executionNodeID(client, in), resolveNodeName(in), resolveMachineID(client, in), client.ClientIP(), workflows); err != nil {
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
		Type:      model.WSMessageTypeWorkflowInventoryAck,
		ClientIP:  client.ClientIP(),
		MachineID: resolveMachineID(client, in),
		NodeID:    executionNodeID(client, in),
		NodeName:  resolveNodeName(in),
		ClientID:  resolveClientID(client, in),
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
	machineID := resolveMachineID(client, in)
	nodeID := executionNodeID(client, in)
	client.BindNodeIdentity(machineID, nodeID)
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

func resolveMachineID(client *Client, in *model.WSRequest) string {
	if in != nil {
		if machineID := strings.TrimSpace(in.MachineID); machineID != "" {
			return machineID
		}
	}
	if client != nil {
		return strings.TrimSpace(client.MachineID())
	}
	return ""
}

func resolveMachineName(in *model.WSRequest) string {
	if in == nil {
		return ""
	}
	return strings.TrimSpace(in.MachineName)
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

	// Prepare columns and timestamps.
	columns := dao.Clients.Columns()
	now := gtime.Now()
	clientName := strings.TrimSpace(in.ClientName)
	if clientName == "" {
		clientName = clientID
	}
	machineID := resolveMachineID(client, in)
	machineName := resolveMachineName(in)
	nodeID := executionNodeID(client, in)
	nodeName := resolveNodeName(in)
	if nodeName == "" {
		nodeName = clientName
	}
	if err := saveClientMachine(client, in, machineID, machineName, now); err != nil {
		return err
	}
	capabilitiesJSON := "{}"
	if len(in.Capabilities) > 0 {
		if body, err := json.Marshal(in.Capabilities); err == nil {
			capabilitiesJSON = string(body)
		}
	}

	// Build shared save data.
	saveData := do.Clients{
		ClientId:         clientID,
		ClientName:       clientName,
		ClientIp:         client.ClientIP(),
		MachineId:        machineID,
		MachineName:      machineName,
		NodeId:           nodeID,
		NodeName:         nodeName,
		NodeIndex:        in.NodeIndex,
		WorkerVersion:    strings.TrimSpace(in.WorkerVersion),
		ProfileDir:       strings.TrimSpace(in.ProfileDir),
		ExtensionDir:     strings.TrimSpace(in.ExtensionDir),
		UserAgent:        strings.TrimSpace(in.UserAgent),
		Status:           "online",
		BusyStatus:       "idle",
		PluginStatus:     resolvePluginStatus(in.AutomaInstalled),
		AutomaVersion:    strings.TrimSpace(in.AutomaVersion),
		BrowserName:      strings.TrimSpace(in.BrowserName),
		BrowserVersion:   strings.TrimSpace(in.BrowserVersion),
		OsName:           strings.TrimSpace(in.OsName),
		OsVersion:        strings.TrimSpace(in.OsVersion),
		Hostname:         strings.TrimSpace(in.Hostname),
		CapabilitiesJson: capabilitiesJSON,
		LastSeenAt:       now,
		ConnectedAt:      now,
	}

	queryColumn := columns.ClientIp
	queryValue := client.ClientIP()
	if nodeID != "" {
		queryColumn = columns.NodeId
		queryValue = nodeID
	}

	record, err := dao.Clients.Ctx(client.Ctx).
		Where(queryColumn, queryValue).
		One()
	if err != nil {
		return err
	}

	// Insert new client when missing.
	if record.IsEmpty() {
		saveData.FirstSeenAt = now
		_, err = dao.Clients.Ctx(client.Ctx).Data(saveData).Insert()
		return err
	}

	// Keep custom name unless empty.
	if strings.TrimSpace(gconv.String(record[columns.ClientName])) != "" {
		saveData.ClientName = nil
	}
	_, err = dao.Clients.Ctx(client.Ctx).
		Where(queryColumn, queryValue).
		Data(saveData).
		Update()
	return err
}

func saveClientMachine(client *Client, in *model.WSRequest, machineID string, machineName string, now *gtime.Time) error {
	if client == nil || in == nil || strings.TrimSpace(machineID) == "" {
		return nil
	}
	columns := dao.ClientMachines.Columns()
	saveData := do.ClientMachines{
		MachineId:     machineID,
		MachineName:   machineName,
		ClientIp:      client.ClientIP(),
		Hostname:      strings.TrimSpace(in.Hostname),
		OsName:        strings.TrimSpace(in.OsName),
		OsVersion:     strings.TrimSpace(in.OsVersion),
		WorkerVersion: strings.TrimSpace(in.WorkerVersion),
		Status:        "online",
		LastSeenAt:    now,
	}
	record, err := dao.ClientMachines.Ctx(client.Ctx).Where(columns.MachineId, machineID).One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		saveData.FirstSeenAt = now
		_, err = dao.ClientMachines.Ctx(client.Ctx).Data(saveData).Insert()
		return err
	}
	_, err = dao.ClientMachines.Ctx(client.Ctx).Where(columns.MachineId, machineID).Data(saveData).Update()
	return err
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
	machineID := resolveMachineID(client, in)
	nodeID := executionNodeID(client, in)

	// Build update data.
	columns := dao.Clients.Columns()
	now := gtime.Now()
	updateData := do.Clients{
		ClientIp:   client.ClientIP(),
		MachineId:  machineID,
		NodeId:     nodeID,
		Status:     "online",
		LastSeenAt: now,
	}
	if in != nil && (in.Type == model.WSMessageTypeAgentRegister || in.Type == model.WSMessageTypeAgentStatusUpdate) {
		updateData.PluginStatus = resolvePluginStatus(in.AutomaInstalled)
		if automaVersion := strings.TrimSpace(in.AutomaVersion); automaVersion != "" || in.Type == model.WSMessageTypeAgentRegister {
			updateData.AutomaVersion = automaVersion
		}
	}

	// Update matched client.
	queryColumn := columns.ClientIp
	queryValue := clientIP
	if nodeID != "" {
		queryColumn = columns.NodeId
		queryValue = nodeID
	}
	_, err := dao.Clients.Ctx(client.Ctx).
		Where(queryColumn, queryValue).
		Data(updateData).
		Update()
	if err == nil && machineID != "" {
		machineColumns := dao.ClientMachines.Columns()
		_, err = dao.ClientMachines.Ctx(client.Ctx).
			Where(machineColumns.MachineId, machineID).
			Data(do.ClientMachines{
				ClientIp:   client.ClientIP(),
				Status:     "online",
				LastSeenAt: now,
			}).
			Update()
	}
	return err
}

// markClientOffline mark client disconnected.
func markClientOffline(client *Client) error {
	if client == nil {
		return nil
	}

	// Resolve client ip.
	clientIP := strings.TrimSpace(client.ClientIP())
	nodeID := strings.TrimSpace(client.ExecutionIdentity())
	if clientIP == "" && nodeID == "" {
		return nil
	}

	// Update offline status.
	columns := dao.Clients.Columns()
	queryColumn := columns.ClientIp
	queryValue := clientIP
	if nodeID != "" {
		queryColumn = columns.NodeId
		queryValue = nodeID
	}
	_, err := dao.Clients.Ctx(client.Ctx).
		Where(queryColumn, queryValue).
		Data(do.Clients{
			Status:              "offline",
			BusyStatus:          "idle",
			CurrentExecutionId:  "",
			CurrentTaskRecordId: 0,
			DisconnectedAt:      gtime.Now(),
		}).
		Update()
	return err
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
	nodeID := strings.TrimSpace(client.ExecutionIdentity())
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
		columns := dao.TaskRecords.Columns()
		_, err = dao.TaskRecords.Ctx(client.Ctx).
			WherePri(recordID).
			WhereIn(columns.Status, []string{"pending", "queued", "running"}).
			Data(do.TaskRecords{
				Status:       "failed",
				ErrorMessage: "client disconnected, task execution was automatically ended",
				FinishedAt:   gtime.Now(),
			}).
			Update()
		if err != nil {
			g.Log().Line().Warningf(client.Ctx, "fail client running task on close failed: record_id=%d client_ip=%s err=%+v", recordID, clientIP, err)
		}
	}

	state.RemovePendingCommand(lockInfo.CommandID)
	if err = tasklock.ReleaseNode(client.Ctx, nodeID, clientIP, lockInfo.CommandID); err != nil {
		g.Log().Line().Warningf(client.Ctx, "release client task lock on close failed: client_ip=%s command_id=%s err=%+v", clientIP, lockInfo.CommandID, err)
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

	g.Log().Line().Infof(client.Ctx, "WebSocket connected: connection_id=%s client_ip=%s", client.ConnectionID(), client.ClientIP())
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
