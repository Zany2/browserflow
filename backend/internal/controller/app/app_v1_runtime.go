package app

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/app/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/gogf/gf/v2/frame/g"
)

// Runtime returns runtime mode and disabled routes 返回运行模式和禁用路由
func (c *ControllerV1) Runtime(ctx context.Context, req *v1.RuntimeReq) (res *v1.RuntimeRes, err error) {
	mode := consts.ResolveRuntimeMode(ctx)
	frontendURL := g.Cfg().MustGet(ctx, "frontend.url", "").String()
	if frontendURL == "" {
		frontendURL = os.Getenv("FRONTEND_URL")
	}
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Choose disabled routes by runtime mode 按运行模式选择禁用路由
	disabledRoutes := []string{"/automa", "/tasks", "/task-records", "/clients"}
	if mode == consts.RuntimeModeServer {
		disabledRoutes = []string{"/browser", "/workflows", "/llm", "/chat", "/browser-agent"}
	} else {
		disabledRoutes = append(disabledRoutes, "/client-agent")
	}

	return &v1.RuntimeRes{
		Mode:           mode,
		DisabledRoutes: disabledRoutes,
		FrontendURL:    frontendURL,
	}, nil
}
