package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/Zany2/browserflow/windows-worker/internal/app"
	"github.com/Zany2/browserflow/windows-worker/internal/config"
)

func main() {
	defer func() {
		if recovered := recover(); recovered != nil {
			writeCrashLog(fmt.Sprintf("panic: %v\n%s", recovered, debug.Stack()))
			os.Exit(1)
		}
	}()
	if err := app.Run(); err != nil {
		log.Println(err)
		writeCrashLog(err.Error())
		os.Exit(1)
	}
}

func writeCrashLog(message string) {
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		return
	}
	logPath := filepath.Join(filepath.Dir(configPath), "worker.log")
	if err = os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(logPath, []byte(message+"\n"), 0o644)
}
