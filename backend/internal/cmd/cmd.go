package cmd

import (
	"context"
	"os"
	"time"

	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/controller/agents"
	"github.com/Zany2/browserflow/backend/internal/controller/app"
	"github.com/Zany2/browserflow/backend/internal/controller/browser"
	"github.com/Zany2/browserflow/backend/internal/controller/browserexecutor"
	"github.com/Zany2/browserflow/backend/internal/controller/chat"
	"github.com/Zany2/browserflow/backend/internal/controller/clients"
	"github.com/Zany2/browserflow/backend/internal/controller/llm"
	"github.com/Zany2/browserflow/backend/internal/controller/taskrecords"
	"github.com/Zany2/browserflow/backend/internal/controller/tasks"
	"github.com/Zany2/browserflow/backend/internal/controller/workflows"
	"github.com/Zany2/browserflow/backend/internal/controller/ws"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/middleware"
	"github.com/Zany2/browserflow/backend/utility/taskcron"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtime"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			runtimeMode := consts.ResolveRuntimeMode(ctx)

			if runtimeMode == consts.RuntimeModeServer {
				// Reset stale node state before accepting reconnects. 启动监听前重置遗留在线节点状态
				now := gtime.Now()
				clientColumns := dao.Clients.Columns()
				if _, err = dao.Clients.Ctx(ctx).
					Where(clientColumns.Status, "online").
					Data(do.Clients{
						Status:              "offline",
						BusyStatus:          "idle",
						CurrentExecutionId:  "",
						CurrentTaskRecordId: 0,
						DisconnectedAt:      now,
					}).
					Update(); err != nil {
					return err
				}

				if err = workflowcache.ClearBrowserflowKeys(ctx); err != nil {
					return err
				}
				tasks.RecoverActiveTaskState(ctx)
			}

			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(
					middleware.Cors(),
					middleware.HandlerResponseMiddleware(),
				)

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
					// Server routes manage remote nodes, tasks, records, and server-side workflow assets. Server 路由管理远程节点和任务调度
					group.Group("/tasks", func(group *ghttp.RouterGroup) {
						group.Bind(tasks.NewV1())
					})
					group.Group("/task-records", func(group *ghttp.RouterGroup) {
						group.Bind(taskrecords.NewV1())
					})
					group.Group("/automa", func(group *ghttp.RouterGroup) {
						group.Bind(workflows.NewV1())
						group.Bind(workflows.NewServerV1())
					})
					group.Group("/clients", func(group *ghttp.RouterGroup) {
						group.Bind(clients.NewV1())
					})
				} else {
					// Windows routes operate local desktop resources only. Windows 路由只操作本机资源
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
				// Start scheduler and recovery after stale online state has been cleared. 清理后启动调度与恢复
				taskcron.StartCronScheduler(ctx)
				websockets.RequestTaskRecovery(ctx)

				gproc.AddSigHandlerShutdown(func(sig os.Signal) {
					cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					taskcron.StopCronScheduler()

					now := gtime.Now()
					clientColumns := dao.Clients.Columns()
					if _, cleanupErr := dao.Clients.Ctx(cleanupCtx).
						Where(clientColumns.Status, "online").
						Data(do.Clients{
							Status:              "offline",
							BusyStatus:          "idle",
							CurrentExecutionId:  "",
							CurrentTaskRecordId: 0,
							DisconnectedAt:      now,
						}).
						Update(); cleanupErr != nil {
						g.Log().Line().Error(gctx.New(), "标记服务端客户端离线失败 ", cleanupErr.Error())
					}
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
