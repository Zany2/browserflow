package middleware

import (
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gvalid"
)

// Cors cors middleware 跨域中间件
func Cors() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		r.Response.CORSDefault()
		r.Middleware.Next()
	}
}

// HandlerResponseMiddleware unified response middleware 统一响应中间件
func HandlerResponseMiddleware() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		r.Middleware.Next()

		if r.Response.BufferLength() > 0 {
			return
		}

		var (
			err = r.GetError()
			res = r.GetHandlerResponse()
		)

		if err != nil {
			// HandlerResponseMiddleware validate error 参数校验错误
			if _, ok := err.(gvalid.Error); ok {
				rr.FailedJsonWithCodeAndMessageExitAll(r, consts.CodeParamError, err.Error())
				return
			}

			// HandlerResponseMiddleware unexpected error 非预期错误
			g.Log().Line().Errorf(r.Context(), "%+v", err)
			rr.FailedJsonWithCodeAndMessageExitAll(r, consts.CodeServerError, consts.CodeMessageMap[consts.CodeServerError])
			return
		}

		// EmptyResponseFallback empty response fallback 空响应回退为空数组
		if g.IsNil(res) {
			res = []interface{}{}
		}

		rr.SuccessJsonWithDataExitAll(r, res)
	}
}
