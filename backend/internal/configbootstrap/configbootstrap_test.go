package configbootstrap

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

func TestEnsureUsesWorkingConfigFirst(t *testing.T) {
	withTempWorkspace(t, func(root, exeDir string) {
		runtimeGOOS = "windows"
		writeFile(t, filepath.Join(root, configFileName), `app:
  mode: "server"
`)

		if err := Ensure(); err != nil {
			t.Fatalf("Ensure() error = %v", err)
		}
		if mode := g.Cfg().MustGet(context.Background(), "app.mode", "").String(); mode != "server" {
			t.Fatalf("app.mode = %q, want server", mode)
		}
		if existsFile(filepath.Join(exeDir, configFileName)) {
			t.Fatalf("unexpected generated config at exe dir")
		}
	})
}

func TestEnsureUsesManifestConfigBeforeGenerating(t *testing.T) {
	withTempWorkspace(t, func(root, exeDir string) {
		runtimeGOOS = "windows"
		writeFile(t, filepath.Join(root, "manifest", "config", configFileName), `app:
  mode: "server"
`)

		if err := Ensure(); err != nil {
			t.Fatalf("Ensure() error = %v", err)
		}
		if mode := g.Cfg().MustGet(context.Background(), "app.mode", "").String(); mode != "server" {
			t.Fatalf("app.mode = %q, want server", mode)
		}
		if existsFile(filepath.Join(exeDir, configFileName)) {
			t.Fatalf("unexpected generated config at exe dir")
		}
	})
}

func TestEnsureUsesBackendManifestConfigBeforeGenerating(t *testing.T) {
	withTempWorkspace(t, func(root, exeDir string) {
		runtimeGOOS = "windows"
		writeFile(t, filepath.Join(root, "backend", "manifest", "config", configFileName), `app:
  mode: "server"
`)

		if err := Ensure(); err != nil {
			t.Fatalf("Ensure() error = %v", err)
		}
		if mode := g.Cfg().MustGet(context.Background(), "app.mode", "").String(); mode != "server" {
			t.Fatalf("app.mode = %q, want server", mode)
		}
		if existsFile(filepath.Join(exeDir, configFileName)) {
			t.Fatalf("unexpected generated config at exe dir")
		}
	})
}

func TestEnsureGeneratesWindowsConfigInExeDir(t *testing.T) {
	withTempWorkspace(t, func(root, exeDir string) {
		runtimeGOOS = "windows"
		if err := Ensure(); err != nil {
			t.Fatalf("Ensure() error = %v", err)
		}

		configPath := filepath.Join(exeDir, configFileName)
		if !existsFile(configPath) {
			t.Fatalf("expected generated config at %s", configPath)
		}
		if mode := g.Cfg().MustGet(context.Background(), "app.mode", "").String(); mode != "windows" {
			t.Fatalf("app.mode = %q, want windows", mode)
		}
		if database := g.Cfg().MustGet(context.Background(), "database.default.link", "").String(); database != "" {
			t.Fatalf("database.default.link = %q, want empty", database)
		}
		if redis := g.Cfg().MustGet(context.Background(), "redis.default.address", "").String(); redis != "" {
			t.Fatalf("redis.default.address = %q, want empty", redis)
		}
		_ = root
	})
}

func TestEnsureSkipsNonWindows(t *testing.T) {
	withTempWorkspace(t, func(root, exeDir string) {
		runtimeGOOS = "linux"
		if err := Ensure(); err != nil {
			t.Fatalf("Ensure() error = %v", err)
		}
		if existsFile(filepath.Join(root, configFileName)) {
			t.Fatalf("unexpected generated config at working dir")
		}
		if existsFile(filepath.Join(exeDir, configFileName)) {
			t.Fatalf("unexpected generated config at exe dir")
		}
	})
}

func withTempWorkspace(t *testing.T, fn func(root, exeDir string)) {
	t.Helper()

	root := t.TempDir()
	exeDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(exeDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	oldExecutableDirFunc := executableDirFunc
	oldRuntimeGOOS := runtimeGOOS
	resetConfig(t)

	if err = os.Chdir(root); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	executableDirFunc = func() (string, error) {
		return exeDir, nil
	}

	t.Cleanup(func() {
		_ = os.Chdir(oldDir)
		executableDirFunc = oldExecutableDirFunc
		runtimeGOOS = oldRuntimeGOOS
		resetConfig(t)
	})

	fn(root, exeDir)
}

func resetConfig(t *testing.T) {
	t.Helper()

	adapter, err := gcfg.NewAdapterFile("config.yaml")
	if err != nil {
		t.Fatalf("NewAdapterFile() error = %v", err)
	}
	g.Cfg().SetAdapter(adapter)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
