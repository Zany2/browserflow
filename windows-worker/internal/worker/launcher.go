package worker

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Launcher struct {
	ChromePath          string
	ServerURL           string
	MachineID           string
	MachineName         string
	NodeCount           int
	DataDir             string
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

	nodes := make([]Node, 0, l.NodeCount)
	for index := 1; index <= l.NodeCount; index++ {
		nodeID := fmt.Sprintf("%s-node-%d", l.MachineID, index)
		profileDir := filepath.Join(l.DataDir, "profiles", fmt.Sprintf("node-%d", index))
		if err := os.MkdirAll(profileDir, 0o755); err != nil {
			return nil, err
		}

		agentURL, err := l.agentURL(nodeID)
		if err != nil {
			return nil, err
		}

		args := []string{
			"--user-data-dir=" + profileDir,
			"--no-first-run",
			"--no-default-browser-check",
			"--new-window",
			agentURL,
		}
		if l.AutomaExtensionDir != "" {
			if info, err := os.Stat(l.AutomaExtensionDir); err == nil && info.IsDir() {
				args = append([]string{
					"--disable-extensions-except=" + l.AutomaExtensionDir,
					"--load-extension=" + l.AutomaExtensionDir,
				}, args...)
			}
		}

		cmd := exec.Command(l.ChromePath, args...)
		if err = cmd.Start(); err != nil {
			return nil, err
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
	base, err := url.Parse(strings.TrimRight(l.ServerURL, "/") + "/")
	if err != nil {
		return "", err
	}
	base.Fragment = "/client-agent"
	query := url.Values{}
	query.Set("machine_id", l.MachineID)
	query.Set("machine_name", l.MachineName)
	query.Set("node_id", nodeID)
	base.RawQuery = query.Encode()
	return base.String(), nil
}
