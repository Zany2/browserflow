package browser

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserInstanceDelete deletes browser instance 删除浏览器实例
func (c *ControllerV1) BrowserInstanceDelete(ctx context.Context, req *v1.BrowserInstanceDeleteReq) (res *v1.BrowserInstanceDeleteRes, err error) {
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

	instance, getErr := db.GetBrowserInstance(req.ID)
	if getErr == nil && instance.IsDefault {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "默认浏览器实例不能删除")
		return
	}
	state.BrowserMu.Lock()
	_, running := state.BrowserInstances[req.ID]
	state.BrowserMu.Unlock()
	if running {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "请先停止浏览器实例")
		return
	}
	if err = db.DeleteBrowserInstance(req.ID); err != nil {
		return nil, err
	}
	return &v1.BrowserInstanceDeleteRes{Message: "删除成功"}, nil
}
