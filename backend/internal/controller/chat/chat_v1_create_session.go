package chat

import (
	"context"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/chatruntime"
	"github.com/gogf/gf/v2/util/guid"
)

// ChatSessionCreate creates chat session 创建对话会话
func (c *ControllerV1) ChatSessionCreate(ctx context.Context, req *v1.ChatSessionCreateReq) (res *v1.ChatSessionCreateRes, err error) {
	db, _, err := chatruntime.Ensure(ctx)
	if err != nil {
		return nil, err
	}

	configID := strings.TrimSpace(req.LLMConfigID)
	if configID == "" {
		// Use default model config when request omits one 未指定配置时使用默认大模型配置
		config, err := db.GetDefaultLLMConfig()
		if err != nil {
			return nil, err
		}
		configID = config.ID
	} else {
		if _, err = db.GetLLMConfig(configID); err != nil {
			return nil, err
		}
	}
	session := &model.ChatSession{
		ID:          "chat_" + guid.S(),
		LLMConfigID: configID,
		Messages:    []model.ChatMessage{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err = db.SaveChatSession(session); err != nil {
		return nil, err
	}
	return &v1.ChatSessionCreateRes{Session: session}, nil
}
