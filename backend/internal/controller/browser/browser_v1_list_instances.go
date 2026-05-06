package browser

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserInstanceList returns browser instances 获取浏览器实例列表
func (c *ControllerV1) BrowserInstanceList(ctx context.Context, req *v1.BrowserInstanceListReq) (res *v1.BrowserInstanceListRes, err error) {
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

	instances, err := db.ListBrowserInstances()
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		headless := false
		defaultInstance := &model.BrowserInstance{
			ID:          "default",
			Name:        "默认浏览器",
			Description: "默认本地 Chrome 浏览器",
			IsDefault:   true,
			Type:        "local",
			Headless:    &headless,
			LaunchArgs:  []string{"--disable-blink-features=AutomationControlled", "--disable-infobars"},
		}
		if err = db.SaveBrowserInstance(defaultInstance); err != nil {
			return nil, err
		}
		instances = []*model.BrowserInstance{defaultInstance}
	}
	state.BrowserMu.Lock()
	for _, instance := range instances {
		_, instance.IsActive = state.BrowserInstances[instance.ID]
	}
	state.BrowserMu.Unlock()
	return &v1.BrowserInstanceListRes{Instances: instances}, nil
}
