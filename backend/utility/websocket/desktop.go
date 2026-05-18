package websockets

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/Zany2/browserflow/backend/utility/workflowexecution"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// handleDesktopMessage routes Windows-mode websocket messages 处理 Windows 模式 WebSocket 消息
func (ws *WsHandlerFunc) handleDesktopMessage(client *Client, in *model.WSRequest) {
	switch strings.ToLower(in.Type) {
	case model.WSMessageTypePing:
		_ = SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypePong})
	case model.WSMessageTypeChatSend:
		ws.handleDesktopChatSend(client, in)
	case model.WSMessageTypeBrowserSubscribe:
		ws.handleBrowserSubscribe(client)
	case model.WSMessageTypeAgentStatusSubscribe:
		ws.handleAgentStatusSubscribe(client)
	case model.WSMessageTypeAgentRegister:
		ws.handleDesktopAgentRegister(client, in)
	case model.WSMessageTypeAgentStatusUpdate:
		ws.handleDesktopAgentStatusUpdate(client, in)
	case model.WSMessageTypeAgentResult:
		ws.handleDesktopAgentResult(client, in)
	default:
		_ = SendConnectionMessage(client.ConnectionID(), &model.WSResponse{
			Type:  model.WSMessageTypeError,
			Error: "不支持的 WebSocket 消息类型",
		})
	}
}

// connectionContext returns lifecycle context for one connection 获取单连接生命周期上下文
func (ws *WsHandlerFunc) connectionContext(client *Client) context.Context {
	if client == nil {
		return context.Background()
	}

	ws.mu.RLock()
	ctx := ws.ClientCtxMap[client.ConnectionID()]
	ws.mu.RUnlock()
	if ctx != nil {
		return ctx
	}
	return client.Ctx
}

// handleDesktopChatSend keeps the historical Windows WS chat path 处理 Windows 模式 WS 对话
func (ws *WsHandlerFunc) handleDesktopChatSend(client *Client, in *model.WSRequest) {
	ctx := ws.connectionContext(client)
	db, llmClient, err := ensureDesktopRuntime(ctx)
	if err == nil {
		session, sessionErr := db.GetChatSession(in.SessionID)
		if sessionErr == nil {
			config, configErr := db.GetLLMConfig(session.LLMConfigID)
			if configErr != nil {
				sessionErr = configErr
			} else {
				userMessage := model.ChatMessage{
					ID:        "msg_" + guid.S(),
					SessionID: in.SessionID,
					Role:      "user",
					Content:   strings.TrimSpace(in.Message),
					Timestamp: time.Now(),
				}
				session.Messages = append(session.Messages, userMessage)
				sessionErr = db.SaveChatSession(session)
				if sessionErr == nil {
					assistantMessage := model.ChatMessage{
						ID:        "msg_" + guid.S(),
						SessionID: in.SessionID,
						Role:      "assistant",
						Timestamp: time.Now(),
					}
					sessionErr = llmClient.StreamChat(ctx, config, session.Messages, func(chunk string) error {
						assistantMessage.Content += chunk
						if sent := SendConnectionMessage(client.ConnectionID(), &model.WSResponse{
							Type:      "chat_message",
							SessionID: in.SessionID,
							Content:   chunk,
							MessageID: assistantMessage.ID,
						}); sent <= 0 {
							return errors.New("websocket connection closed")
						}
						return nil
					})
					if sessionErr == nil {
						session.Messages = append(session.Messages, assistantMessage)
						sessionErr = db.SaveChatSession(session)
					}
					if sessionErr == nil {
						SendConnectionMessage(client.ConnectionID(), &model.WSResponse{
							Type:      model.WSMessageTypeChatDone,
							SessionID: in.SessionID,
							MessageID: assistantMessage.ID,
						})
					}
				}
			}
		}
		err = sessionErr
	}
	if err != nil {
		SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeError, Error: err.Error()})
	}
}

// handleBrowserSubscribe streams browser runtime status 订阅浏览器运行状态
func (ws *WsHandlerFunc) handleBrowserSubscribe(client *Client) {
	ch := make(chan model.BrowserStatus, 8)
	state.BrowserMu.Lock()
	state.BrowserStatusListeners[ch] = struct{}{}
	current := state.CurrentBrowserStatusLocked()
	state.BrowserMu.Unlock()
	SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeBrowserStatus, Browser: current})

	go func() {
		ctx := ws.connectionContext(client)
		defer func() {
			state.BrowserMu.Lock()
			delete(state.BrowserStatusListeners, ch)
			state.BrowserMu.Unlock()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case status := <-ch:
				if sent := SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeBrowserStatus, Browser: status}); sent <= 0 {
					return
				}
			}
		}
	}()
}

// handleAgentStatusSubscribe streams browser agent status 订阅执行端状态
func (ws *WsHandlerFunc) handleAgentStatusSubscribe(client *Client) {
	ch := make(chan []model.AgentStatus, 8)
	state.AgentMu.Lock()
	state.AgentStatusListeners[ch] = struct{}{}
	current := desktopAgentStatusesLocked()
	state.AgentMu.Unlock()
	SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeAgentStatus, Agents: current})

	go func() {
		ctx := ws.connectionContext(client)
		defer func() {
			state.AgentMu.Lock()
			delete(state.AgentStatusListeners, ch)
			state.AgentMu.Unlock()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case statuses := <-ch:
				if sent := SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeAgentStatus, Agents: statuses}); sent <= 0 {
					return
				}
			}
		}
	}()
}

// handleDesktopAgentRegister registers one Windows browser agent 注册 Windows 浏览器执行端
func (ws *WsHandlerFunc) handleDesktopAgentRegister(client *Client, in *model.WSRequest) {
	browserID := strings.TrimSpace(in.BrowserID)
	if browserID == "" {
		state.BrowserMu.Lock()
		browserID = strings.TrimSpace(state.BrowserCurrentInstanceID)
		state.BrowserMu.Unlock()
	}
	if browserID == "" {
		SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeError, Error: "browser agent missing browser_id"})
		return
	}

	now := time.Now()
	role := strings.ToLower(strings.TrimSpace(in.Role))
	if role == "" {
		role = "browser_agent"
	}
	state.AgentMu.Lock()
	state.AgentConnections[browserID] = &state.AgentConnection{
		BrowserID:       browserID,
		Role:            role,
		Token:           in.Token,
		ConnectionID:    client.ConnectionID(),
		AutomaInstalled: in.AutomaInstalled,
		AutomaVersion:   strings.TrimSpace(in.AutomaVersion),
		ConnectedAt:     now,
		LastSeenAt:      now,
	}
	status := model.AgentStatus{
		BrowserID:       browserID,
		Online:          true,
		AutomaInstalled: in.AutomaInstalled,
		AutomaVersion:   strings.TrimSpace(in.AutomaVersion),
		ConnectedAt:     now,
		LastSeenAt:      now,
	}
	statuses := desktopAgentStatusesLocked()
	broadcastDesktopAgentStatusesLocked(statuses)
	state.AgentMu.Unlock()

	SendConnectionMessage(client.ConnectionID(), &model.WSResponse{
		Type:      model.WSMessageTypeAgentRegistered,
		BrowserID: browserID,
		Agents:    []model.AgentStatus{status},
	})
}

// handleDesktopAgentStatusUpdate refreshes browser agent metadata 更新浏览器执行端状态
func (ws *WsHandlerFunc) handleDesktopAgentStatusUpdate(client *Client, in *model.WSRequest) {
	browserID := strings.TrimSpace(in.BrowserID)
	state.AgentMu.Lock()
	if browserID == "" {
		browserID = findDesktopAgentByConnectionLocked(client.ConnectionID())
	}
	if agent := state.AgentConnections[browserID]; agent != nil {
		agent.AutomaInstalled = in.AutomaInstalled
		agent.AutomaVersion = strings.TrimSpace(in.AutomaVersion)
		agent.LastSeenAt = time.Now()
	}
	statuses := desktopAgentStatusesLocked()
	broadcastDesktopAgentStatusesLocked(statuses)
	state.AgentMu.Unlock()

	SendConnectionMessage(client.ConnectionID(), &model.WSResponse{Type: model.WSMessageTypeAgentStatusUpdateAck})
}

// handleDesktopAgentResult resolves browser-agent command waiters 处理浏览器执行端命令结果
func (ws *WsHandlerFunc) handleDesktopAgentResult(client *Client, in *model.WSRequest) {
	browserID := strings.TrimSpace(in.BrowserID)
	state.AgentMu.Lock()
	if browserID == "" {
		browserID = findDesktopAgentByConnectionLocked(client.ConnectionID())
	}
	result := model.AgentCommandResult{
		BrowserID: browserID,
		CommandID: in.CommandID,
		Success:   in.Success,
		Data:      in.Data,
		Error:     in.Error,
	}
	resultCh := state.PendingCommands[in.CommandID]
	delete(state.PendingCommands, in.CommandID)
	if agent := state.AgentConnections[browserID]; agent != nil {
		agent.LastSeenAt = time.Now()
	}
	state.AgentMu.Unlock()

	workflowexecution.CompleteByCommand(in.CommandID, &result)
	if resultCh != nil {
		resultCh <- result
	}
	SendConnectionMessage(client.ConnectionID(), &model.WSResponse{
		Type:      model.WSMessageTypeAgentResultAck,
		CommandID: in.CommandID,
	})
}

// handleDesktopClose clears Windows browser agent state 处理 Windows 连接断开
func (ws *WsHandlerFunc) handleDesktopClose(client *Client) {
	state.AgentMu.Lock()
	browserID := findDesktopAgentByConnectionLocked(client.ConnectionID())
	if browserID == "" {
		state.AgentMu.Unlock()
		return
	}

	agent := state.AgentConnections[browserID]
	delete(state.AgentConnections, browserID)
	statuses := desktopAgentStatusesLocked()
	broadcastDesktopAgentStatusesLocked(statuses)
	state.AgentMu.Unlock()

	if strings.ToLower(strings.TrimSpace(agent.Role)) == "browser_agent" {
		scheduleBrowserRuntimeProbe(browserID)
	}
}

// findDesktopAgentByConnectionLocked resolves agent id by connection 调用方需持有 AgentMu
func findDesktopAgentByConnectionLocked(connectionID string) string {
	for browserID, agent := range state.AgentConnections {
		if agent != nil && agent.ConnectionID == connectionID {
			return browserID
		}
	}
	return ""
}

// desktopAgentStatusesLocked builds Windows browser-agent statuses 调用方需持有 AgentMu
func desktopAgentStatusesLocked() []model.AgentStatus {
	statuses := make([]model.AgentStatus, 0, len(state.AgentConnections))
	for browserID, agent := range state.AgentConnections {
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

// broadcastDesktopAgentStatusesLocked notifies agent subscribers 调用方需持有 AgentMu
func broadcastDesktopAgentStatusesLocked(statuses []model.AgentStatus) {
	for listener := range state.AgentStatusListeners {
		select {
		case listener <- statuses:
		default:
		}
	}
}

// ensureDesktopRuntime prepares local websocket chat dependencies 初始化本地 WS 对话依赖
func ensureDesktopRuntime(ctx context.Context) (*storage.BoltDB, *llm.Client, error) {
	state.DBMu.Lock()
	defer state.DBMu.Unlock()

	if state.DB == nil {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
		}
		db, err := storage.NewBoltDB(dbPath)
		if err != nil {
			return nil, nil, err
		}
		state.DB = db
	}
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}
	return state.DB, state.LLMClient, nil
}

// scheduleBrowserRuntimeProbe checks browser runtime after agent disconnect Agent 断开后延迟探测浏览器
func scheduleBrowserRuntimeProbe(browserID string) {
	browserID = strings.TrimSpace(browserID)
	if browserID == "" {
		return
	}

	go func() {
		time.Sleep(6 * time.Second)

		state.AgentMu.Lock()
		_, online := state.AgentConnections[browserID]
		state.AgentMu.Unlock()
		if online {
			return
		}

		state.BrowserMu.Lock()
		runtime := state.BrowserInstances[browserID]
		state.BrowserMu.Unlock()
		if runtime == nil || runtime.Browser == nil {
			return
		}
		if browserRuntimeHasAgentPage(runtime) {
			return
		}

		removedRuntime, _, removed := state.RemoveBrowserRuntime(browserID, runtime)
		if removed {
			state.CleanupBrowserRuntime(removedRuntime)
		}
	}()
}

// browserRuntimeHasAgentPage checks whether the managed agent tab still exists 检查执行端页面是否仍存在
func browserRuntimeHasAgentPage(runtime *state.BrowserRuntime) bool {
	if runtime == nil || runtime.Browser == nil {
		return false
	}

	browser := runtime.Browser.Timeout(1 * time.Second)
	defer browser.CancelTimeout()
	pages, err := browser.Pages()
	if err != nil {
		return false
	}

	for _, page := range pages {
		info, infoErr := page.Info()
		if infoErr != nil {
			continue
		}
		pageURL := strings.TrimSpace(info.URL)
		if pageURL != "" && strings.Contains(pageURL, "#/browser-agent") {
			return true
		}
	}
	return false
}
