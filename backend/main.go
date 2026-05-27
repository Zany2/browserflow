package main

import (
	"fmt"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"github.com/Zany2/browserflow/backend/internal/cmd"
	"github.com/Zany2/browserflow/backend/internal/configbootstrap"
)

func main() {
	if err := configbootstrap.Ensure(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "prepare config failed: %v\n", err)
		os.Exit(1)
	}

	cmd.Main.Run(gctx.GetInitCtx())
}
