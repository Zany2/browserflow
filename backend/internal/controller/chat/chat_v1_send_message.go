package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/chatruntime"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// ChatMessageSend sends chat message through stream 发送流式对话消息
func (c *ControllerV1) ChatMessageSend(ctx context.Context, req *v1.ChatMessageSendReq) (res *v1.ChatMessageSendRes, err error) {
	sessionID := strings.TrimSpace(req.ID)
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, fmt.Errorf("消息不能为空")
	}

	db, llmClient, err := chatruntime.Ensure(ctx)
	if err != nil {
		return nil, err
	}

	request := g.RequestFromCtx(ctx)
	// Prepare server-sent event response 准备 SSE 流式响应
	request.Response.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	request.Response.Header().Set("Cache-Control", "no-cache")
	request.Response.Header().Set("Connection", "keep-alive")
	request.Response.Header().Set("X-Accel-Buffering", "no")

	session, err := db.GetChatSession(sessionID)
	if err == nil {
		configID := strings.TrimSpace(req.LLMConfigID)
		if configID == "" {
			configID = session.LLMConfigID
		}
		config, configErr := db.GetLLMConfig(configID)
		if configErr != nil {
			err = configErr
		} else {
			userMessage := model.ChatMessage{
				ID:        "msg_" + guid.S(),
				SessionID: sessionID,
				Role:      "user",
				Content:   message,
				Timestamp: time.Now(),
			}
			session.Messages = append(session.Messages, userMessage)
			err = db.SaveChatSession(session)
			if err == nil {
				// Stream assistant chunks and persist final message 流式输出助手回复并保存最终消息
				assistantMessage := model.ChatMessage{
					ID:        "msg_" + guid.S(),
					SessionID: sessionID,
					Role:      "assistant",
					Timestamp: time.Now(),
				}
				err = llmClient.StreamChat(ctx, config, session.Messages, func(chunk string) error {
					assistantMessage.Content += chunk
					data, _ := json.Marshal(model.StreamChunk{Type: "message", Content: chunk, MessageID: assistantMessage.ID})
					request.Response.Writef("data: %s\n\n", data)
					request.Response.Flush()
					return nil
				})
				if err == nil {
					session.Messages = append(session.Messages, assistantMessage)
					err = db.SaveChatSession(session)
				}
				if err == nil {
					data, _ := json.Marshal(model.StreamChunk{Type: "done", MessageID: assistantMessage.ID})
					request.Response.Writef("data: %s\n\n", data)
					request.Response.Flush()
				}
			}
		}
	}
	if err != nil {
		// Emit stream error event 输出流式错误事件
		data, _ := json.Marshal(model.StreamChunk{Type: "error", Error: err.Error()})
		request.Response.Writef("data: %s\n\n", data)
		request.Response.Flush()
	}
	request.ExitAll()
	return nil, nil
}
