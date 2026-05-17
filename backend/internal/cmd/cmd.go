package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gctx"
	"os"
	"time"

	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/controller/agents"
	"github.com/Zany2/browserflow/backend/internal/controller/app"
	"github.com/Zany2/browserflow/backend/internal/controller/automa"
	"github.com/Zany2/browserflow/backend/internal/controller/browser"
	"github.com/Zany2/browserflow/backend/internal/controller/browserexecutor"
	"github.com/Zany2/browserflow/backend/internal/controller/chat"
	"github.com/Zany2/browserflow/backend/internal/controller/clients"
	"github.com/Zany2/browserflow/backend/internal/controller/llm"
	"github.com/Zany2/browserflow/backend/internal/controller/taskrecords"
	"github.com/Zany2/browserflow/backend/internal/controller/tasks"
	"github.com/Zany2/browserflow/backend/internal/controller/workflows"
	"github.com/Zany2/browserflow/backend/internal/controller/ws"
	"github.com/Zany2/browserflow/backend/middleware"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gproc"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			runtimeMode := consts.ResolveRuntimeMode(ctx)

			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(
					middleware.Cors(),                      // Cors cross origin 跨域处理
					middleware.HandlerResponseMiddleware(), // Response middleware 统一返回
				)

				// Common routes shared by all runtime modes 所有运行模式共用路由
				group.Group("/app", func(group *ghttp.RouterGroup) {
					group.Bind(app.NewV1())
				})
				group.Group("/ws", func(group *ghttp.RouterGroup) {
					group.Bind(ws.NewV1())
				})
				group.Group("/agents", func(group *ghttp.RouterGroup) {
					group.Bind(agents.NewV1())
				})
				group.Group("/workflows", func(group *ghttp.RouterGroup) {
					group.Bind(workflows.NewV1())
				})

				if runtimeMode == consts.RuntimeModeServer {
					// Server routes for remote task and client management 服务端任务与客户端管理路由
					group.Group("/tasks", func(group *ghttp.RouterGroup) {
						group.Bind(tasks.NewV1())
					})
					group.Group("/task-records", func(group *ghttp.RouterGroup) {
						group.Bind(taskrecords.NewV1())
					})
					group.Group("/automa", func(group *ghttp.RouterGroup) {
						group.Bind(automa.NewV1())
					})
					group.Group("/clients", func(group *ghttp.RouterGroup) {
						group.Bind(clients.NewV1())
					})
				} else {
					// Desktop routes for local browser, model, and chat operation 本地浏览器、大模型和对话路由
					group.Group("/browser", func(group *ghttp.RouterGroup) {
						group.Bind(browser.NewV1())
					})
					group.Group("/browser-executor", func(group *ghttp.RouterGroup) {
						group.Bind(browserexecutor.NewV1())
					})
					group.Group("/llm", func(group *ghttp.RouterGroup) {
						group.Bind(llm.NewV1())
					})
					group.Group("/chat", func(group *ghttp.RouterGroup) {
						group.Bind(chat.NewV1())
					})
				}
			})

			if runtimeMode == consts.RuntimeModeServer {
				// Redis cache cleanup clears transient client inventory on server shutdown 服务端退出时清理临时客户端清单缓存
				gproc.AddSigHandlerShutdown(func(sig os.Signal) {
					cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if cleanupErr := workflowcache.ClearBrowserflowKeys(cleanupCtx); cleanupErr != nil {
						g.Log().Line().Error(gctx.New(), "清理 Redis 客户端缓存失败 ", cleanupErr.Error())
					}
				})
			}

			s.Run()
			return nil
		},
	}
)
