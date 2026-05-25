package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Launcher struct {
	ChromePath          string
	ServerURL           string
	MachineID           string
	MachineName         string
	NodeCount           int
	DataDir             string
	DownloadDir         string
	AutomaExtensionDir  string
	RequireAutomaFolder bool
}

type Node struct {
	NodeID     string
	ProfileDir string
	URL        string
	PID        int
}

func (l Launcher) Start() ([]Node, error) {
	if strings.TrimSpace(l.ChromePath) == "" {
		return nil, errors.New("ChromePath 不能为空")
	}
	if strings.TrimSpace(l.ServerURL) == "" {
		return nil, errors.New("ServerURL 不能为空")
	}
	if l.NodeCount <= 0 {
		return nil, errors.New("NodeCount 必须大于 0")
	}
	if l.RequireAutomaFolder {
		if info, err := os.Stat(l.AutomaExtensionDir); err != nil || !info.IsDir() {
			return nil, fmt.Errorf("Automa 扩展目录不可用：%s", l.AutomaExtensionDir)
		}
	}
	extensionDir, hasExtension, err := resolveAutomaExtensionDir(l.AutomaExtensionDir)
	if err != nil {
		l.writeLog("resolve Automa extension failed: %v", err)
		return nil, err
	}
	l.writeLog("start nodes chrome=%q server_url=%q data_dir=%q extension_dir=%q has_extension=%v node_count=%d", l.ChromePath, l.ServerURL, l.DataDir, extensionDir, hasExtension, l.NodeCount)

	nodes := make([]Node, 0, l.NodeCount)
	for index := 1; index <= l.NodeCount; index++ {
		nodeID := fmt.Sprintf("node-%d", index)
		profileDir := filepath.Join(l.DataDir, "profiles", fmt.Sprintf("node-%d", index))
		if err := os.MkdirAll(profileDir, 0o755); err != nil {
			l.writeLog("node=%s create profile failed profile=%q error=%v", nodeID, profileDir, err)
			return nil, err
		}
		if err := ensureProfileAvailableStrict(profileDir, nodeID); err != nil {
			l.writeLog("node=%s profile is locked profile=%q error=%v", nodeID, profileDir, err)
			return nil, err
		}
		downloadDir := l.nodeDownloadDir(nodeID)
		if err := configureDownloadDir(profileDir, downloadDir); err != nil {
			l.writeLog("node=%s configure download dir failed profile=%q download_dir=%q error=%v", nodeID, profileDir, downloadDir, err)
			return nil, err
		}

		agentURL, err := l.agentURL(nodeID)
		if err != nil {
			l.writeLog("node=%s build agent url failed error=%v", nodeID, err)
			return nil, err
		}

		args := []string{
			"--user-data-dir=" + profileDir,
			"--no-first-run",
			"--no-default-browser-check",
			"--new-window",
			agentURL,
		}
		if hasExtension {
			args = append([]string{
				"--remote-debugging-port=0",
				"--enable-unsafe-extension-debugging",
				"--disable-features=DisableLoadExtensionCommandLineSwitch",
				"--disable-extensions-except=" + extensionDir,
				"--load-extension=" + extensionDir,
			}, args...)
		}

		cmd := exec.Command(l.ChromePath, args...)
		if err = cmd.Start(); err != nil {
			l.writeLog("node=%s start chrome failed args=%s error=%v", nodeID, quoteArgs(args), err)
			return nil, err
		}
		l.writeLog("node=%s started pid=%d profile=%q url=%q args=%s", nodeID, cmd.Process.Pid, profileDir, agentURL, quoteArgs(args))
		if hasExtension {
			if err = loadExtensionAndNavigate(profileDir, extensionDir, agentURL, l.writeLog); err != nil {
				l.writeLog("node=%s CDP load Automa skipped error=%v", nodeID, err)
			}
		}
		nodes = append(nodes, Node{
			NodeID:     nodeID,
			ProfileDir: profileDir,
			URL:        agentURL,
			PID:        cmd.Process.Pid,
		})
	}
	return nodes, nil
}

func resolveAutomaExtensionDir(extensionDir string) (string, bool, error) {
	extensionDir = strings.TrimSpace(extensionDir)
	if extensionDir == "" {
		return "", false, nil
	}
	info, err := os.Stat(extensionDir)
	if err != nil {
		return "", false, nil
	}
	if !info.IsDir() {
		return "", false, fmt.Errorf("Automa 扩展路径不是目录：%s", extensionDir)
	}
	manifestPath := filepath.Join(extensionDir, "manifest.json")
	if _, err = os.Stat(manifestPath); err != nil {
		return "", false, fmt.Errorf("Automa 扩展目录缺少 manifest.json：%s", extensionDir)
	}
	return filepath.Clean(extensionDir), true, nil
}

func ensureProfileAvailable(profileDir string, nodeID string) error {
	lockPath := filepath.Join(profileDir, "lockfile")
	if _, err := os.Stat(lockPath); err != nil {
		return nil
	}
	return fmt.Errorf("%s 的 Chrome 配置目录正在使用中，请先关闭旧的执行节点后再启动", nodeID)
}

func Stop(nodes []Node) int {
	stopped := 0
	for _, node := range nodes {
		if node.PID <= 0 {
			continue
		}
		cmd := exec.Command("taskkill", "/PID", strconv.Itoa(node.PID), "/T", "/F")
		if err := cmd.Run(); err == nil {
			stopped++
			continue
		}
		if process, err := os.FindProcess(node.PID); err == nil {
			if err = process.Kill(); err == nil {
				stopped++
			}
		}
	}
	return stopped
}

func (l Launcher) agentURL(nodeID string) (string, error) {
	base, err := normalizeServerURL(l.ServerURL)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("node_id", nodeID)
	base.Fragment = "/client-agent?" + query.Encode()
	return base.String(), nil
}

func normalizeServerURL(serverURL string) (*url.URL, error) {
	base, err := url.Parse(strings.TrimSpace(serverURL))
	if err != nil {
		return nil, err
	}
	base.RawQuery = ""
	base.ForceQuery = false
	base.Fragment = ""
	base.Path = strings.TrimRight(base.Path, "/") + "/"
	return base, nil
}

func (l Launcher) nodeDownloadDir(nodeID string) string {
	baseDir := strings.TrimSpace(l.DownloadDir)
	if baseDir == "" {
		baseDir = filepath.Join(filepath.Dir(l.DataDir), "downloads")
	}
	return filepath.Join(baseDir, nodeID)
}

func configureDownloadDir(profileDir string, downloadDir string) error {
	if strings.TrimSpace(downloadDir) == "" {
		return nil
	}
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return err
	}
	defaultDir := filepath.Join(profileDir, "Default")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		return err
	}
	preferencesPath := filepath.Join(defaultDir, "Preferences")
	preferences := map[string]any{}
	if body, err := os.ReadFile(preferencesPath); err == nil && len(body) > 0 {
		if err = json.Unmarshal(body, &preferences); err != nil {
			return err
		}
	}
	download, _ := preferences["download"].(map[string]any)
	if download == nil {
		download = map[string]any{}
	}
	download["default_directory"] = downloadDir
	download["directory_upgrade"] = true
	download["prompt_for_download"] = false
	preferences["download"] = download

	body, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(preferencesPath, body, 0o644)
}

func ensureProfileAvailableStrict(profileDir string, nodeID string) error {
	lockNames := []string{"lockfile", "SingletonLock", "SingletonCookie", "SingletonSocket"}
	for _, name := range lockNames {
		lockPath := filepath.Join(profileDir, name)
		if _, err := os.Stat(lockPath); err == nil {
			return fmt.Errorf("%s 的 Chrome 配置目录正在使用中，请先关闭旧的执行节点后再启动", nodeID)
		}
	}
	return nil
}

func (l Launcher) writeLog(format string, args ...any) {
	logPath := filepath.Join(filepath.Dir(l.DataDir), "logs", "worker.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	line := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(file, "%s %s\n", time.Now().Format("2006-01-02 15:04:05"), line)
}

func quoteArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t") {
			quoted = append(quoted, strconv.Quote(arg))
			continue
		}
		quoted = append(quoted, arg)
	}
	return strings.Join(quoted, " ")
}
