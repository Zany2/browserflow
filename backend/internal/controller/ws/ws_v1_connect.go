package ws

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/ws/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/gorilla/websocket"
)

var serverWSUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}}

// Connect upgrades HTTP request and handles websocket messages 建立 WebSocket 连接
func (c *ControllerV1) Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error) {
	request := g.RequestFromCtx(ctx)
	if consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer {
		websockets.Init(ctx)
		conn, err := serverWSUpgrader.Upgrade(request.Response.Writer, request.Request, nil)
		if err != nil {
			return nil, err
		}
		connCtx := context.WithoutCancel(ctx)
		websockets.WsManage.RegisterClient(connCtx, request.GetClientIp(), conn)
		request.ExitAll()
		return nil, nil
	}

	state.DBMu.Lock()
	if state.DB == nil {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
		}
		state.DB, err = storage.NewBoltDB(dbPath)
		if err != nil {
			state.DBMu.Unlock()
			return nil, err
		}
	}
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}
	db := state.DB
	llmClient := state.LLMClient
	state.DBMu.Unlock()

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(request.Response.Writer, request.Request, nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	defer request.ExitAll()

	client := &state.WSClient{Conn: conn}
	registeredAgentID := ""
	writeJSON := func(value any) error {
		client.WriteMu.Lock()
		defer client.WriteMu.Unlock()
		return client.Conn.WriteJSON(value)
	}
	browserStatus := func() model.BrowserStatus {
		runtime := state.BrowserInstances[state.BrowserCurrentInstanceID]
		if runtime == nil {
			return model.BrowserStatus{Running: false}
		}
		startTime := runtime.StartTime
		instanceCopy := *runtime.Instance
		instanceCopy.IsActive = true
		return model.BrowserStatus{
			Running:           true,
			CurrentInstanceID: state.BrowserCurrentInstanceID,
			Instance:          &instanceCopy,
			StartTime:         &startTime,
			UptimeSeconds:     int64(time.Since(startTime).Seconds()),
			ControlURL:        runtime.ControlURL,
			AgentURL:          runtime.AgentURL,
		}
	}
	agentStatuses := func() []model.AgentStatus {
		statuses := make([]model.AgentStatus, 0, len(state.AgentConnections))
		for browserID, agent := range state.AgentConnections {
			if agent == nil {
				statuses = append(statuses, model.AgentStatus{BrowserID: browserID, Online: false})
				continue
			}
			if strings.ToLower(strings.TrimSpace(agent.Role)) == "client_agent" {
				continue
			}
			statuses = append(statuses, model.AgentStatus{BrowserID: agent.BrowserID, Online: true, AutomaInstalled: agent.AutomaInstalled, AutomaVersion: agent.AutomaVersion, ConnectedAt: agent.ConnectedAt, LastSeenAt: agent.LastSeenAt})
		}
		return statuses
	}
	defer func() {
		if registeredAgentID == "" {
			return
		}
		state.AgentMu.Lock()
		currentAgent := state.AgentConnections[registeredAgentID]
		if currentAgent != nil && currentAgent.Client == client {
			currentAgentRole := strings.ToLower(strings.TrimSpace(currentAgent.Role))
			// Disconnect cleanup 只清理当前断开的连接，避免刷新重连时旧连接误删新连接
			delete(state.AgentConnections, registeredAgentID)
			statuses := agentStatuses()
			for listener := range state.AgentStatusListeners {
				select {
				case listener <- statuses:
				default:
				}
			}
			if currentAgentRole == "browser_agent" {
				scheduleBrowserRuntimeProbe(registeredAgentID)
			}
		}
		state.AgentMu.Unlock()
	}()

	for {
		var in model.WSRequest
		if err = conn.ReadJSON(&in); err != nil {
			return nil, err
		}
		switch in.Type {
		case "ping":
			err = writeJSON(model.WSResponse{Type: "pong"})
		case "chat_send":
			session, sessionErr := db.GetChatSession(in.SessionID)
			if sessionErr == nil {
				config, configErr := db.GetLLMConfig(session.LLMConfigID)
				if configErr != nil {
					sessionErr = configErr
				} else {
					userMessage := model.ChatMessage{ID: "msg_" + guid.S(), SessionID: in.SessionID, Role: "user", Content: strings.TrimSpace(in.Message), Timestamp: time.Now()}
					session.Messages = append(session.Messages, userMessage)
					sessionErr = db.SaveChatSession(session)
					if sessionErr == nil {
						assistantMessage := model.ChatMessage{ID: "msg_" + guid.S(), SessionID: in.SessionID, Role: "assistant", Timestamp: time.Now()}
						sessionErr = llmClient.StreamChat(ctx, config, session.Messages, func(chunk string) error {
							assistantMessage.Content += chunk
							return writeJSON(model.WSResponse{Type: "chat_message", SessionID: in.SessionID, Content: chunk, MessageID: assistantMessage.ID})
						})
						if sessionErr == nil {
							session.Messages = append(session.Messages, assistantMessage)
							sessionErr = db.SaveChatSession(session)
						}
						if sessionErr == nil {
							sessionErr = writeJSON(model.WSResponse{Type: "chat_done", SessionID: in.SessionID, MessageID: assistantMessage.ID})
						}
					}
				}
			}
			if sessionErr != nil {
				err = writeJSON(model.WSResponse{Type: "error", Error: sessionErr.Error()})
			}
		case "browser_subscribe":
			ch := make(chan model.BrowserStatus, 8)
			state.BrowserMu.Lock()
			state.BrowserStatusListeners[ch] = struct{}{}
			current := browserStatus()
			state.BrowserMu.Unlock()
			if err = writeJSON(model.WSResponse{Type: "browser_status", Browser: current}); err != nil {
				break
			}
			for {
				select {
				case <-ctx.Done():
					state.BrowserMu.Lock()
					delete(state.BrowserStatusListeners, ch)
					state.BrowserMu.Unlock()
					return nil, nil
				case status := <-ch:
					if err = writeJSON(model.WSResponse{Type: "browser_status", Browser: status}); err != nil {
						state.BrowserMu.Lock()
						delete(state.BrowserStatusListeners, ch)
						state.BrowserMu.Unlock()
						return nil, err
					}
				}
			}
		case "agent_status_subscribe":
			ch := make(chan []model.AgentStatus, 8)
			state.AgentMu.Lock()
			state.AgentStatusListeners[ch] = struct{}{}
			current := agentStatuses()
			state.AgentMu.Unlock()
			if err = writeJSON(model.WSResponse{Type: "agent_status", Agents: current}); err != nil {
				break
			}
			for {
				select {
				case <-ctx.Done():
					state.AgentMu.Lock()
					delete(state.AgentStatusListeners, ch)
					state.AgentMu.Unlock()
					return nil, nil
				case statuses := <-ch:
					if err = writeJSON(model.WSResponse{Type: "agent_status", Agents: statuses}); err != nil {
						state.AgentMu.Lock()
						delete(state.AgentStatusListeners, ch)
						state.AgentMu.Unlock()
						return nil, err
					}
				}
			}
		case "agent_register":
			if strings.TrimSpace(in.BrowserID) == "" {
				state.BrowserMu.Lock()
				in.BrowserID = strings.TrimSpace(state.BrowserCurrentInstanceID)
				state.BrowserMu.Unlock()
			}
			if strings.TrimSpace(in.BrowserID) == "" {
				err = writeJSON(model.WSResponse{Type: "error", Error: "browser agent missing browser_id"})
				break
			}
			now := time.Now()
			role := strings.ToLower(strings.TrimSpace(in.Role))
			if role == "" {
				role = "browser_agent"
			}
			state.AgentMu.Lock()
			state.AgentConnections[in.BrowserID] = &state.AgentConnection{BrowserID: in.BrowserID, Role: role, Token: in.Token, Client: client, AutomaInstalled: in.AutomaInstalled, AutomaVersion: strings.TrimSpace(in.AutomaVersion), ConnectedAt: now, LastSeenAt: now}
			registeredAgentID = in.BrowserID
			status := model.AgentStatus{BrowserID: in.BrowserID, Online: true, AutomaInstalled: in.AutomaInstalled, AutomaVersion: strings.TrimSpace(in.AutomaVersion), ConnectedAt: now, LastSeenAt: now}
			statuses := agentStatuses()
			for listener := range state.AgentStatusListeners {
				select {
				case listener <- statuses:
				default:
				}
			}
			state.AgentMu.Unlock()
			err = writeJSON(model.WSResponse{Type: "agent_registered", BrowserID: in.BrowserID, Agents: []model.AgentStatus{status}})
		case "agent_status_update":
			browserID := in.BrowserID
			if browserID == "" {
				browserID = registeredAgentID
			}
			state.AgentMu.Lock()
			if agent := state.AgentConnections[browserID]; agent != nil {
				agent.AutomaInstalled = in.AutomaInstalled
				agent.AutomaVersion = strings.TrimSpace(in.AutomaVersion)
				agent.LastSeenAt = time.Now()
			}
			statuses := agentStatuses()
			for listener := range state.AgentStatusListeners {
				select {
				case listener <- statuses:
				default:
				}
			}
			state.AgentMu.Unlock()
			err = writeJSON(model.WSResponse{Type: "agent_status_update_ack"})
		case "agent_result":
			browserID := in.BrowserID
			if browserID == "" {
				browserID = registeredAgentID
			}
			result := model.AgentCommandResult{BrowserID: browserID, CommandID: in.CommandID, Success: in.Success, Data: in.Data, Error: in.Error}
			state.AgentMu.Lock()
			resultCh := state.PendingCommands[in.CommandID]
			delete(state.PendingCommands, in.CommandID)
			if agent := state.AgentConnections[browserID]; agent != nil {
				agent.LastSeenAt = time.Now()
			}
			state.AgentMu.Unlock()
			if resultCh != nil {
				resultCh <- result
			}
			err = writeJSON(model.WSResponse{Type: "agent_result_ack", CommandID: in.CommandID})
		default:
			err = writeJSON(model.WSResponse{Type: "error", Error: "不支持的 WebSocket 消息类型"})
		}
		if err != nil {
			_ = writeJSON(model.WSResponse{Type: "error", Error: err.Error()})
		}
	}
	request.ExitAll()
	return nil, nil
}

// scheduleBrowserRuntimeProbe checks browser runtime after agent disconnect. Agent 断开后延迟探测浏览器是否仍存活。
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

		// Agent page probe 长时间没有执行端且页面不存在时，才认为被管理浏览器已不可用
		if browserRuntimeHasAgentPage(runtime) {
			return
		}

		removedRuntime, _, removed := state.RemoveBrowserRuntime(browserID, runtime)
		if removed {
			state.CleanupBrowserRuntime(removedRuntime)
		}
	}()
}

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
