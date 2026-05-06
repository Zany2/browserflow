package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
)

// ChatSessionDelete deletes chat session 删除对话会话
func (c *ControllerV1) ChatSessionDelete(ctx context.Context, req *v1.ChatSessionDeleteReq) (res *v1.ChatSessionDeleteRes, err error) {
	db, _, err := ensureRuntime(ctx)
	if err != nil {
		return nil, err
	}

	sessionID := strings.TrimSpace(req.ID)
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	if err = db.DeleteChatSession(sessionID); err != nil {
		return nil, err
	}
	return &v1.ChatSessionDeleteRes{Message: "删除成功"}, nil
}
