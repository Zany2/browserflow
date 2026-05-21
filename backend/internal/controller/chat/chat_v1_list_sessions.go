package chat

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
	"github.com/Zany2/browserflow/backend/utility/chatruntime"
)

// ChatSessionList returns chat sessions 获取对话会话列表
func (c *ControllerV1) ChatSessionList(ctx context.Context, req *v1.ChatSessionListReq) (res *v1.ChatSessionListRes, err error) {
	db, _, err := chatruntime.Ensure(ctx)
	if err != nil {
		return nil, err
	}

	sessions, err := db.ListChatSessions()
	if err != nil {
		return nil, err
	}
	return &v1.ChatSessionListRes{Sessions: sessions}, nil
}
