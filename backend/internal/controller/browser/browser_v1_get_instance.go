package browser

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserInstanceDetail returns browser instance detail 获取浏览器实例详情
func (c *ControllerV1) BrowserInstanceDetail(ctx context.Context, req *v1.BrowserInstanceDetailReq) (res *v1.BrowserInstanceDetailRes, err error) {
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

	instance, err := db.GetBrowserInstance(req.ID)
	if err != nil {
		return nil, err
	}
	state.BrowserMu.Lock()
	_, instance.IsActive = state.BrowserInstances[instance.ID]
	state.BrowserMu.Unlock()
	return &v1.BrowserInstanceDetailRes{Instance: instance}, nil
}
