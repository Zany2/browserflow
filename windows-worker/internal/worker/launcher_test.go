package worker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLauncherAgentURL(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		want      string
	}{
		{
			name:      "plain server",
			serverURL: "http://127.0.0.1:8080",
			want:      "http://127.0.0.1:8080/#/client-agent?node_id=node-1",
		},
		{
			name:      "trailing slash",
			serverURL: "http://127.0.0.1:8080/",
			want:      "http://127.0.0.1:8080/#/client-agent?node_id=node-1",
		},
		{
			name:      "existing hash route",
			serverURL: "http://127.0.0.1:8080/#/client-agent?node_id=node-9",
			want:      "http://127.0.0.1:8080/#/client-agent?node_id=node-1",
		},
		{
			name:      "existing query",
			serverURL: "http://127.0.0.1:8080/?foo=bar",
			want:      "http://127.0.0.1:8080/#/client-agent?node_id=node-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Launcher{ServerURL: tt.serverURL}.agentURL("node-1")
			if err != nil {
				t.Fatalf("agentURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("agentURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveAutomaExtensionDir(t *testing.T) {
	extensionDir := t.TempDir()
	if _, ok, err := resolveAutomaExtensionDir(extensionDir); err == nil || ok {
		t.Fatalf("resolveAutomaExtensionDir() without manifest ok=%v err=%v, want error", ok, err)
	}

	if err := os.WriteFile(filepath.Join(extensionDir, "manifest.json"), []byte(`{"manifest_version":3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok, err := resolveAutomaExtensionDir(extensionDir)
	if err != nil {
		t.Fatalf("resolveAutomaExtensionDir() error = %v", err)
	}
	if !ok || got == "" {
		t.Fatalf("resolveAutomaExtensionDir() = %q, %v, want directory", got, ok)
	}
}
