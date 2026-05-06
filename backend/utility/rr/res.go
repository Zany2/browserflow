// Author daixk 2023-12-04 09:06:38
package rr

import (
	"context"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CommonRes common response 通用响应结构
type CommonRes struct {
	Code    int         `json:"code"`    // Code response code 响应码
	Message string      `json:"message"` // Message response message 响应消息
	Data    interface{} `json:"data"`    // Data response data 响应数据
}

// CommonResModel common list response 通用列表响应结构
type CommonResModel[T any] struct {
	Total int64 `json:"total"` // Total list total count 列表总数
	List  []T   `json:"list"`  // List list data 列表数据
}

// Json write json response 写入 JSON 响应
func Json(r *ghttp.Request, code int, message string, data interface{}) {
	r.Response.WriteJson(CommonRes{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// SuccessJsonExitAll write success response and exit 写入成功响应并退出
func SuccessJsonExitAll(r *ghttp.Request) {
	Json(r, consts.CodeOK, consts.CodeMessageMap[consts.CodeOK], g.Array{})
	r.ExitAll()
}

// SuccessJsonWithMessageExitAll write success message response and exit 写入带消息的成功响应并退出
func SuccessJsonWithMessageExitAll(r *ghttp.Request, message string) {
	Json(r, consts.CodeOK, message, g.Array{})
	r.ExitAll()
}

// SuccessJsonWithDataExitAll write success data response and exit 写入带数据的成功响应并退出
func SuccessJsonWithDataExitAll(r *ghttp.Request, data interface{}) {
	Json(r, consts.CodeOK, consts.CodeMessageMap[consts.CodeOK], data)
	r.ExitAll()
}

// SuccessJsonWithMessageAndData write success message and data response and exit 写入带消息和数据的成功响应并退出
func SuccessJsonWithMessageAndData(r *ghttp.Request, message string, data interface{}) {
	Json(r, consts.CodeOK, message, data)
	r.ExitAll()
}

// FailedJsonExitAll write failed response and exit 写入失败响应并退出
func FailedJsonExitAll(r *ghttp.Request) {
	Json(r, consts.CodeServerError, consts.CodeMessageMap[consts.CodeServerError], g.Array{})
	r.ExitAll()
}

// FailedJsonWithMessageExitAll write failed message response and exit 写入带消息的失败响应并退出
func FailedJsonWithMessageExitAll(r *ghttp.Request, message string) {
	Json(r, consts.CodeServerError, message, g.Array{})
	r.ExitAll()
}

// FailedJsonWithCodeAndMessageExitAll write failed code and message response and exit 写入带错误码和消息的失败响应并退出
func FailedJsonWithCodeAndMessageExitAll(r *ghttp.Request, code int, message string) {
	Json(r, code, message, g.Array{})
	r.ExitAll()
}

// Success build success response 构建成功响应
func Success() *CommonRes {
	return &CommonRes{
		Code:    consts.CodeOK,
		Message: consts.CodeMessageMap[consts.CodeOK],
		Data:    g.Array{},
	}
}

// SuccessWithMessage build success message response 构建带消息的成功响应
func SuccessWithMessage(message string) *CommonRes {
	return &CommonRes{
		Code:    consts.CodeOK,
		Message: message,
		Data:    g.Array{},
	}
}

// SuccessWithData build success data response 构建带数据的成功响应
func SuccessWithData(data interface{}) *CommonRes {
	return &CommonRes{
		Code:    consts.CodeOK,
		Message: consts.CodeMessageMap[consts.CodeOK],
		Data:    data,
	}
}

// SuccessWithMessageAndData build success message and data response 构建带消息和数据的成功响应
func SuccessWithMessageAndData(message string, data interface{}) *CommonRes {
	return &CommonRes{
		Code:    consts.CodeOK,
		Message: message,
		Data:    data,
	}
}

// Failed build failed response 构建失败响应
func Failed() *CommonRes {
	return &CommonRes{
		Code:    consts.CodeServerError,
		Message: consts.CodeMessageMap[consts.CodeServerError],
		Data:    g.Array{},
	}
}

// FailedWithMessage build failed message response 构建带消息的失败响应
func FailedWithMessage(message string) *CommonRes {
	return &CommonRes{
		Code:    consts.CodeServerError,
		Message: message,
		Data:    g.Array{},
	}
}

// FailedWithCodeAndMessage build failed code and message response 构建带错误码和消息的失败响应
func FailedWithCodeAndMessage(code int, message string) *CommonRes {
	return &CommonRes{
		Code:    code,
		Message: message,
		Data:    g.Array{},
	}
}

// FailedWithMessageAndData build failed message and data response 构建带消息和数据的失败响应
func FailedWithMessageAndData(message string, data interface{}) *CommonRes {
	return &CommonRes{
		Code:    consts.CodeServerError,
		Message: message,
		Data:    data,
	}
}

// CtxFailedJsonExitAll write failed response from context 写入基于上下文的失败响应
func CtxFailedJsonExitAll(ctx context.Context) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeServerError, consts.CodeMessageMap[consts.CodeServerError], g.Array{})
}

// CtxFailedJsonWithMessageExitAll write failed message response from context 写入基于上下文的带消息失败响应
func CtxFailedJsonWithMessageExitAll(ctx context.Context, message string) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeServerError, message, g.Array{})
}

// CtxFailedJsonWithCodeAndDataExitAll write failed code and message response from context 写入基于上下文的带错误码和消息失败响应
func CtxFailedJsonWithCodeAndDataExitAll(ctx context.Context, code int, message string) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeServerError, message, g.Array{})
}

// CtxSuccessJsonExitAll write success response from context 写入基于上下文的成功响应
func CtxSuccessJsonExitAll(ctx context.Context) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeOK, consts.CodeMessageMap[consts.CodeOK], g.Array{})
}

// CtxSuccessJsonWithMessageExitAll write success message response from context 写入基于上下文的带消息成功响应
func CtxSuccessJsonWithMessageExitAll(ctx context.Context, message string) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeOK, message, g.Array{})
}

// CtxSuccessJsonWithCodeAndDataExitAll write success code and message response from context 写入基于上下文的带错误码和消息成功响应
func CtxSuccessJsonWithCodeAndDataExitAll(ctx context.Context, code int, message string) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeOK, message, g.Array{})
}

// CtxSuccessJsonWithMessageAndDataExitAll write success message and data response from context 写入基于上下文的带消息和数据成功响应
func CtxSuccessJsonWithMessageAndDataExitAll(ctx context.Context, message string, data interface{}) {
	request := g.RequestFromCtx(ctx)
	Json(request, consts.CodeOK, message, data)
}
