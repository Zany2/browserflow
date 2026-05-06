// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package chat

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/chat/v1"
)

type IChatV1 interface {
	ChatSessionList(ctx context.Context, req *v1.ChatSessionListReq) (res *v1.ChatSessionListRes, err error)
	ChatSessionCreate(ctx context.Context, req *v1.ChatSessionCreateReq) (res *v1.ChatSessionCreateRes, err error)
	ChatSessionDeleteMany(ctx context.Context, req *v1.ChatSessionDeleteManyReq) (res *v1.ChatSessionDeleteManyRes, err error)
	ChatSessionDetail(ctx context.Context, req *v1.ChatSessionDetailReq) (res *v1.ChatSessionDetailRes, err error)
	ChatSessionDelete(ctx context.Context, req *v1.ChatSessionDeleteReq) (res *v1.ChatSessionDeleteRes, err error)
	ChatMessageSend(ctx context.Context, req *v1.ChatMessageSendReq) (res *v1.ChatMessageSendRes, err error)
}
