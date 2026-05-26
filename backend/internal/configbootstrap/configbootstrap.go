package configbootstrap

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

const (
	configFileName = "config.yaml"
)

var (
	executableDirFunc = executableDir
	runtimeGOOS       = runtime.GOOS
)

// Ensure prepares the runtime config before any business code reads g.Cfg(). 启动业务前准备运行配置。
func Ensure() error {
	if runtimeGOOS != "windows" {
		return nil
	}

	exeDir, err := executableDirFunc()
	if err != nil {
		return err
	}

	paths := []string{
		filepath.Join(".", configFileName),
		filepath.Join(exeDir, configFileName),
		filepath.Join("manifest", "config", configFileName),
		filepath.Join("backend", "manifest", "config", configFileName),
	}
	for _, path := range paths {
		if existsFile(path) {
			return useConfigFile(path)
		}
	}

	configPath := filepath.Join(exeDir, configFileName)
	if err = os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	if err = os.WriteFile(configPath, []byte(defaultWindowsConfig), 0644); err != nil {
		return err
	}
	return useConfigFile(configPath)
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func existsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func useConfigFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	adapter, err := gcfg.NewAdapterFile(absPath)
	if err != nil {
		return err
	}
	g.Cfg().SetAdapter(adapter)
	return nil
}

const defaultWindowsConfig = `# BrowserFlow backend configuration. BrowserFlow 后端配置。

# -----------------------------------------------------------------------------
# Common configuration for Windows mode. Windows 模式共用配置。
# -----------------------------------------------------------------------------

server:
  address: ":8001"             # HTTP listen address. HTTP 服务监听地址。
  name: backend                # Server name used by GoFrame. GoFrame 服务名称。
  clientMaxBodySize: 0         # Maximum request body size; 0 means no explicit limit. 请求体最大大小，0 表示不额外限制。
  # openapiPath: "/web.json"     # OpenAPI document path. OpenAPI 文档访问路径。
  # swaggerPath: "/swagger"      # Swagger UI path. Swagger UI 访问路径。

app:
  mode: "windows"              # Runtime mode: "windows" for local desktop edition. 运行模式："windows" 为本地桌面版。

# Logger configuration. 日志配置。
logger:
  path:                  "./logs/"             # Log file path; empty disables file logging and keeps console output only. 日志文件路径，为空表示关闭文件日志，仅输出到终端。
  file:                  "{Y-m-d}.log"         # Log file format. 日志文件格式。
  prefix:                ""                    # Log line prefix. 日志内容输出前缀。
  level:                 "ERROR"                 # Log output level. 日志输出级别。
  timeFormat:            "2006-01-02 15:04:05" # Log time format using the Go standard layout. 自定义日志时间格式，使用 Golang 标准时间格式。
  ctxKeys:               []                    # Context keys automatically printed into logs. 自定义 Context 变量名称，自动打印到日志中。
  header:                true                  # Whether to print log header information. 是否打印日志头信息。
  stdout:                true                  # Whether to also output logs to terminal. 日志是否同时输出到终端。
  rotateSize:            "20M"                 # Rotate logs by file size; 0 disables size rotation. 按日志文件大小滚动切分，0 表示关闭。
  rotateExpire:          "1d"                  # Rotate logs by time interval; 0 disables time rotation. 按日志文件时间间隔滚动切分，0 表示关闭。
  rotateBackupLimit:     7                     # Cleanup by rotated file count when rotation is enabled. 滚动切分开启时，按切分文件数量清理备份。
  rotateBackupExpire:    0                     # Cleanup by rotated file age when rotation is enabled. 滚动切分开启时，按切分文件有效期清理备份。
  rotateBackupCompress:  0                     # Compression level for rotated files, 0-9; 0 disables compression. 滚动切分文件压缩比 0-9，0 表示不压缩。
  rotateCheckInterval:   "1h"                  # Rotation check interval. 滚动切分检测间隔。
  stdoutColorDisabled:   false                 # Disable terminal color output. 关闭终端颜色输出。
  writerColorEnable:     false                 # Enable color output in log files. 日志文件是否带颜色。

frontend:
  url: "http://127.0.0.1:8001"  # Frontend base URL used by local browser launch URLs. 前端访问地址，用于本地浏览器启动地址。

websocket:
  heartbeatInterval: 15         # Heartbeat send interval in seconds. 心跳发送间隔，单位秒。
  heartbeatTimeout: 45          # Heartbeat timeout in seconds. 心跳超时时间，单位秒。
  writeWait: 10                 # WebSocket write timeout in seconds. WebSocket 写入超时时间，单位秒。
  messageBuffer: 64             # Outbound message buffer size per connection. 单连接发送消息缓冲区大小。
  bucketCount: 32               # Connection bucket count used by the WebSocket manager. WebSocket 管理器连接分桶数量。
  maxConn: 10000                # Maximum WebSocket connection count. WebSocket 最大连接数。

# -----------------------------------------------------------------------------
# Windows mode only. 仅 Windows 本地桌面模式使用。
# -----------------------------------------------------------------------------

localStorage:
  path: "data/browserflow.db"   # Local BoltDB file path for Windows mode business data. Windows 模式本地业务数据的 BoltDB 文件路径。
`
