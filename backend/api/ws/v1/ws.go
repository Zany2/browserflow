package v1

import "github.com/gogf/gf/v2/frame/g"

// ConnectReq 建立长连接
type ConnectReq struct {
	g.Meta `path:"/" method:"get" tags:"长连接" summary:"建立长连接" description:"客户端通过该接口建立长连接，用于注册执行端并接收后端指令"`
}

// ConnectRes 长连接占位响应
type ConnectRes struct{}
