package cmd

import (
	"context"
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
	"github.com/Zany2/browserflow/backend/utility/taskcron"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
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
					middleware.Cors(),                      // Cors handles cross-origin requests. 跨域处理中间件。
					middleware.HandlerResponseMiddleware(), // Response middleware wraps common responses. 统一响应处理中间件。
				)

				// Common routes are required by both Windows and Server runtimes. 公共路由同时服务 Windows 与 Server 模式。
				group.Group("/app", func(group *ghttp.RouterGroup) {
					// App exposes runtime metadata for frontend bootstrap and route gating. 应用接口提供前端启动与路由控制所需的运行信息。
					group.Bind(app.NewV1())
				})
				group.Group("/ws", func(group *ghttp.RouterGroup) {
					// WS is the shared transport entry for local agents, remote clients, and status subscriptions. WebSocket 是本地执行端、远程客户端与状态订阅共用的长连接入口。
					group.Bind(ws.NewV1())
				})
				group.Group("/agents", func(group *ghttp.RouterGroup) {
					// Agents reports online browser-agent state for pages that need live executors. 执行端接口返回在线 browser-agent 状态，供需要实时执行端的页面使用。
					group.Bind(agents.NewV1())
				})
				group.Group("/workflows", func(group *ghttp.RouterGroup) {
					// Workflows stays shared because both modes reuse workflow storage and agent dispatch APIs. 工作流接口保持公共，因为两种模式都会复用工作流数据与执行端调度能力。
					group.Bind(workflows.NewV1())
				})

				if runtimeMode == consts.RuntimeModeServer {
					// Server-only routes manage remote clients, stored workflows, and task execution. Server 专属路由负责远程客户端、服务端工作流资产与任务调度。
					group.Group("/tasks", func(group *ghttp.RouterGroup) {
						// Tasks manages scheduled/manual tasks and dispatches them to remote clients. 任务接口管理任务并调度到远程客户端执行。
						group.Bind(tasks.NewV1())
					})
					group.Group("/task-records", func(group *ghttp.RouterGroup) {
						// Task records exposes server-side execution history. 任务记录接口提供服务端执行历史查询。
						group.Bind(taskrecords.NewV1())
					})
					group.Group("/automa", func(group *ghttp.RouterGroup) {
						// Automa keeps the server workflow management surface. Automa 接口保留服务端工作流管理能力。
						group.Bind(automa.NewV1())
					})
					group.Group("/clients", func(group *ghttp.RouterGroup) {
						// Clients manages remote client inventory and administrative actions. 客户端接口负责远程客户端列表与管理操作。
						group.Bind(clients.NewV1())
					})
				} else {
					// Windows-only routes operate local desktop resources. Windows 专属路由只操作本机桌面资源。
					group.Group("/browser", func(group *ghttp.RouterGroup) {
						// Browser manages local browser instances started by the desktop runtime. 浏览器接口管理由桌面端启动的本地浏览器实例。
						group.Bind(browser.NewV1())
					})
					group.Group("/browser-executor", func(group *ghttp.RouterGroup) {
						// Browser executor controls the current local browser instance and is not a remote-dispatch proxy. 浏览器执行器只控制当前本地浏览器实例，不承担远程转发职责。
						group.Bind(browserexecutor.NewV1())
					})
					group.Group("/llm", func(group *ghttp.RouterGroup) {
						// LLM manages local model-provider configurations. 大模型接口管理本地模型提供商配置。
						group.Bind(llm.NewV1())
					})
					group.Group("/chat", func(group *ghttp.RouterGroup) {
						// Chat manages local model conversations and SSE streaming. 对话接口管理本地会话与 SSE 流式输出。
						group.Bind(chat.NewV1())
					})
				}
			})

			if runtimeMode == consts.RuntimeModeServer {
				// Task scheduler only belongs to Server mode because scheduled workflow dispatch is a server responsibility. 任务调度器仅在 Server 模式启动，因为定时工作流调度属于服务端职责。
				taskcron.StartCronScheduler(ctx)

				// Redis cleanup clears transient client inventory when the server exits. Server 退出时清理临时客户端清单缓存。
				gproc.AddSigHandlerShutdown(func(sig os.Signal) {
					cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					// Stop scheduler before process exit to avoid accepting new scheduled work during shutdown. 进程退出前停止调度器，避免关闭阶段继续接收新的定时任务。
					taskcron.StopCronScheduler()

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
