package chatruntime

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// Ensure prepares shared chat dependencies. ??????????
func Ensure(ctx context.Context) (*storage.BoltDB, *llm.Client, error) {
	state.DBMu.Lock()
	defer state.DBMu.Unlock()

	// Database singleton initializes once. ??????????????
	if state.DB == nil {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
		}
		db, err := storage.NewBoltDB(dbPath)
		if err != nil {
			return nil, nil, err
		}
		state.DB = db
	}

	// LLM client singleton initializes once. ???????????????
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}

	return state.DB, state.LLMClient, nil
}
