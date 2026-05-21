package ws

import (
	"context"
	"net/http"

	"github.com/Zany2/browserflow/backend/api/ws/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}}

// Connect upgrades HTTP request and registers websocket client 建立并注册 WebSocket 连接
func (c *ControllerV1) Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error) {
	request := g.RequestFromCtx(ctx)
	conn, err := wsUpgrader.Upgrade(request.Response.Writer, request.Request, nil)
	if err != nil {
		return nil, err
	}

	websockets.Init(ctx)
	clientIP := request.GetClientIp()
	identity := websockets.BuildClientIdentity(clientIP)
	if consts.ResolveRuntimeMode(ctx) != consts.RuntimeModeServer {
		// Desktop connections need unique ids because one machine opens multiple sockets. 桌面端同一机器会建立多条连接
		identity = websockets.BuildConnectionIdentity("conn_"+guid.S(), clientIP, false)
	}
	websockets.WsManage.RegisterClientWithIdentity(context.WithoutCancel(ctx), identity, conn)
	request.ExitAll()
	return nil, nil
}
