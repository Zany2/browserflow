package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
	"github.com/Zany2/browserflow/backend/utility/chatruntime"
)

// ChatSessionDetail returns chat session detail 获取对话会话详情
func (c *ControllerV1) ChatSessionDetail(ctx context.Context, req *v1.ChatSessionDetailReq) (res *v1.ChatSessionDetailRes, err error) {
	db, _, err := chatruntime.Ensure(ctx)
	if err != nil {
		return nil, err
	}

	sessionID := strings.TrimSpace(req.ID)
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	session, err := db.GetChatSession(sessionID)
	if err != nil {
		return nil, err
	}
	return &v1.ChatSessionDetailRes{Session: session}, nil
}
