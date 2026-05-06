package chat

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
)

// ChatSessionDeleteMany deletes chat sessions in batch 批量删除对话会话
func (c *ControllerV1) ChatSessionDeleteMany(ctx context.Context, req *v1.ChatSessionDeleteManyReq) (res *v1.ChatSessionDeleteManyRes, err error) {
	db, _, err := ensureRuntime(ctx)
	if err != nil {
		return nil, err
	}

	for _, sessionID := range req.IDs {
		sessionID = strings.TrimSpace(sessionID)
		if sessionID == "" {
			continue
		}
		if err = db.DeleteChatSession(sessionID); err != nil {
			return nil, err
		}
	}
	return &v1.ChatSessionDeleteManyRes{Message: "批量删除成功"}, nil
}
