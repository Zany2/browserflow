package chatruntime

import (
	"context"
	"os"

	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
)

// Ensure prepares shared chat dependencies. 初始化共享聊天依赖。
func Ensure(ctx context.Context) (*storage.BoltDB, *llm.Client, error) {
	state.DBMu.Lock()
	defer state.DBMu.Unlock()

	// Database singleton initializes once. 数据库单例只初始化一次。
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

	// LLM client singleton initializes once. LLM 客户端单例只初始化一次。
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}

	return state.DB, state.LLMClient, nil
}
