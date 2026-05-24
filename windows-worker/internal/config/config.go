package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const appDirName = "BrowserFlowWorker"

const defaultServerURLPlaceholder = "http://192.168.1.100:8080"

type Config struct {
	ServerURL           string `json:"server_url"`
	ChromePath          string `json:"chrome_path"`
	NodeCount           int    `json:"node_count"`
	MachineID           string `json:"machine_id"`
	MachineName         string `json:"machine_name"`
	DataDir             string `json:"data_dir"`
	AutomaExtensionDir  string `json:"automa_extension_dir"`
	RequireAutomaFolder bool   `json:"require_automa_folder"`
}

func LoadOrCreate(path string) (Config, bool, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := defaultConfig()
		if err = Save(path, cfg); err != nil {
			return Config{}, false, err
		}
		return cfg, true, nil
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false, err
	}

	cfg := defaultConfig()
	if err = json.Unmarshal(body, &cfg); err != nil {
		return Config{}, false, err
	}
	cfg.normalize()
	return cfg, false, nil
}

func Save(path string, cfg Config) error {
	cfg.normalize()
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func DefaultConfigPath() (string, error) {
	dir := DefaultAppDir()
	if strings.TrimSpace(dir) == "" {
		return "", errors.New("无法获取用户数据目录")
	}
	return filepath.Join(dir, "config.json"), nil
}

func DefaultAppDir() string {
	dir, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(dir) == "" {
		dir = os.Getenv("LOCALAPPDATA")
	}
	if strings.TrimSpace(dir) == "" {
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			dir = home
		}
	}
	return filepath.Join(dir, appDirName)
}

func DefaultDataDir() string {
	return filepath.Join(DefaultAppDir(), "data")
}

func DefaultAutomaExtensionDir() string {
	return filepath.Join(DefaultAppDir(), "extensions", "automa")
}

func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return errors.New("请先填写服务端地址")
	}
	if c.NodeCount <= 0 {
		return errors.New("执行节点数必须大于 0")
	}
	if c.NodeCount > 8 {
		return errors.New("执行节点数不能超过 8")
	}
	if c.MachineID == "" {
		return errors.New("当前电脑标识不能为空，请重启软件后再试")
	}
	return nil
}

func (c *Config) normalize() {
	c.ServerURL = strings.TrimRight(strings.TrimSpace(c.ServerURL), "/")
	if c.ServerURL == strings.TrimRight(defaultServerURLPlaceholder, "/") {
		c.ServerURL = ""
	}
	c.ChromePath = strings.TrimSpace(c.ChromePath)
	c.MachineID = strings.TrimSpace(c.MachineID)
	c.MachineName = strings.TrimSpace(c.MachineName)
	c.DataDir = strings.TrimSpace(c.DataDir)
	c.AutomaExtensionDir = strings.TrimSpace(c.AutomaExtensionDir)
	if c.NodeCount <= 0 {
		c.NodeCount = 1
	}
	if c.MachineID == "" {
		c.MachineID = "bfw-" + randomHex(6)
	}
	if c.MachineName == "" {
		if hostname, err := os.Hostname(); err == nil {
			c.MachineName = hostname
		}
	}
}

func defaultConfig() Config {
	hostname, _ := os.Hostname()
	return Config{
		ServerURL:           "",
		ChromePath:          "",
		NodeCount:           1,
		MachineID:           "bfw-" + randomHex(6),
		MachineName:         hostname,
		DataDir:             DefaultDataDir(),
		AutomaExtensionDir:  DefaultAutomaExtensionDir(),
		RequireAutomaFolder: false,
	}
}

func randomHex(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "local"
	}
	return hex.EncodeToString(buffer)
}
