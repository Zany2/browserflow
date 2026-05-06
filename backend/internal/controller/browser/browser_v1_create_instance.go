package browser

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// BrowserInstanceCreate creates browser instance 创建浏览器实例
func (c *ControllerV1) BrowserInstanceCreate(ctx context.Context, req *v1.BrowserInstanceCreateReq) (res *v1.BrowserInstanceCreateRes, err error) {
	state.DBMu.Lock()
	if state.DB == nil {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
		}
		state.DB, err = storage.NewBoltDB(dbPath)
		if err != nil {
			state.DBMu.Unlock()
			return nil, err
		}
	}
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}
	db := state.DB
	state.DBMu.Unlock()

	instance := &model.BrowserInstance{
		ID:          "browser_" + guid.S(),
		Name:        req.BrowserInstancePayload.Name,
		Description: req.BrowserInstancePayload.Description,
		IsDefault:   req.BrowserInstancePayload.IsDefault,
		Type:        req.BrowserInstancePayload.Type,
		BinPath:     req.BrowserInstancePayload.BinPath,
		UserDataDir: req.BrowserInstancePayload.UserDataDir,
		ControlURL:  req.BrowserInstancePayload.ControlURL,
		UserAgent:   req.BrowserInstancePayload.UserAgent,
		Headless:    req.BrowserInstancePayload.Headless,
		NoSandbox:   req.BrowserInstancePayload.NoSandbox,
		LaunchArgs:  req.BrowserInstancePayload.LaunchArgs,
		Proxy:       req.BrowserInstancePayload.Proxy,
	}
	if strings.TrimSpace(instance.Name) == "" {
		instance.Name = "默认浏览器"
	}
	if instance.Type != "remote" {
		instance.Type = "local"
	}
	if instance.Type == "remote" && strings.TrimSpace(instance.ControlURL) == "" {
		return nil, errors.New("远程浏览器控制地址不能为空")
	}
	if err = db.SaveBrowserInstance(instance); err != nil {
		return nil, err
	}
	return &v1.BrowserInstanceCreateRes{Instance: instance}, nil
}
