package clientops

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/frame/g"
)

// RequestIP reads real client ip from request. 读取请求真实 IP
func RequestIP(ctx context.Context) string {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return ""
	}
	return strings.TrimSpace(request.GetClientIp())
}

// TargetConnectionID returns websocket identity for one client node. 构建节点连接标识
func TargetConnectionID(clientIP string, nodeID string) string {
	return websockets.NodeConnectionID(clientIP, nodeID)
}

// NotifyBanned tells connected node to stop reconnecting. 通知节点已被拉黑
func NotifyBanned(ctx context.Context, clientIP string, nodeID string, clientID string, reason string) int {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" && nodeID == "" {
		return 0
	}

	websockets.Init(ctx)
	return websockets.SendNodeMessage(nodeID, clientIP, &model.WSResponse{
		Type:     model.WSMessageTypeClientBanned,
		ClientID: strings.TrimSpace(clientID),
		ClientIP: clientIP,
		NodeID:   nodeID,
		Message:  "客户端已被拉黑，将持续检测，解除拉黑后自动重连",
		Error:    strings.TrimSpace(reason),
		Data: map[string]any{
			"no_reconnect": true,
			"reason":       strings.TrimSpace(reason),
		},
	})
}

// CloseConnection closes current websocket connection. 关闭当前 WebSocket 连接
func CloseConnection(ctx context.Context, clientIP string, nodeID string) int {
	connectionID := TargetConnectionID(clientIP, nodeID)
	if connectionID == "" {
		return 0
	}

	websockets.Init(ctx)
	return websockets.WsManage.CloseClientConnections(connectionID)
}
