package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/browser/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// BrowserInstanceStart starts browser instance 启动浏览器实例
func (c *ControllerV1) BrowserInstanceStart(ctx context.Context, req *v1.BrowserInstanceStartReq) (res *v1.BrowserInstanceStartRes, err error) {
	state.DBMu.Lock()
	if state.DB == nil {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
		}
		state.DB, err = storage.NewBoltDB(dbPath)
		if err != nil {
			state.DBMu.Unlock()
			return nil, err
		}
	}
	if state.LLMClient == nil {
		state.LLMClient = llm.NewClient()
	}
	db := state.DB
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = g.Cfg().MustGet(ctx, "frontend.url", "http://localhost:5173").String()
	}
	state.DBMu.Unlock()

	instance, err := db.GetBrowserInstance(req.ID)
	if err != nil {
		return nil, err
	}
	state.BrowserMu.Lock()
	if _, ok := state.BrowserInstances[instance.ID]; ok {
		state.BrowserCurrentInstanceID = instance.ID
		runtime := state.BrowserInstances[state.BrowserCurrentInstanceID]
		startTime := runtime.StartTime
		instanceCopy := *runtime.Instance
		instanceCopy.IsActive = true
		status := model.BrowserStatus{Running: true, CurrentInstanceID: state.BrowserCurrentInstanceID, Instance: &instanceCopy, StartTime: &startTime, UptimeSeconds: int64(time.Since(startTime).Seconds()), ControlURL: runtime.ControlURL, AgentURL: runtime.AgentURL}
		for listener := range state.BrowserStatusListeners {
			select {
			case listener <- status:
			default:
			}
		}
		state.BrowserMu.Unlock()
		return &v1.BrowserInstanceStartRes{Status: &status}, nil
	}
	state.BrowserMu.Unlock()

	var runtime *state.BrowserRuntime
	agentToken := "agent_" + guid.S()
	agentURL := fmt.Sprintf("%s/#/browser-agent?browserId=%s&token=%s", strings.TrimRight(frontendURL, "/"), url.QueryEscape(instance.ID), url.QueryEscape(agentToken))
	if instance.Type == "remote" {
		controlURL := strings.TrimSpace(instance.ControlURL)
		if controlURL == "" {
			return nil, fmt.Errorf("远程浏览器控制地址不能为空")
		}
		if !strings.HasPrefix(controlURL, "ws://") && !strings.HasPrefix(controlURL, "wss://") {
			// Resolve HTTP debug endpoint to WebSocket debugger URL 解析 HTTP 调试端点为 WebSocket 调试地址
			versionURL := strings.TrimRight(controlURL, "/") + "/json/version"
			httpClient := &http.Client{Timeout: 5 * time.Second}
			resp, httpErr := httpClient.Get(versionURL)
			if httpErr != nil {
				return nil, fmt.Errorf("解析远程浏览器 WebSocket 地址失败: %w", httpErr)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("解析远程浏览器 WebSocket 地址失败: %s", strings.TrimSpace(string(body)))
			}
			var version struct {
				WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
			}
			if decodeErr := json.NewDecoder(resp.Body).Decode(&version); decodeErr != nil {
				return nil, fmt.Errorf("解析远程浏览器 WebSocket 地址失败: %w", decodeErr)
			}
			if strings.TrimSpace(version.WebSocketDebuggerURL) == "" {
				return nil, fmt.Errorf("远程浏览器未返回 webSocketDebuggerUrl")
			}
			controlURL = strings.TrimSpace(version.WebSocketDebuggerURL)
		}
		browser := rod.New().ControlURL(controlURL)
		if err = browser.Connect(); err != nil {
			return nil, err
		}
		_ = (&proto.BrowserGrantPermissions{
			Permissions: []proto.BrowserPermissionType{
				proto.BrowserPermissionTypeClipboardReadWrite,
				proto.BrowserPermissionTypeClipboardSanitizedWrite,
			},
		}).Call(browser)
		agentPage, pageErr := browser.Page(proto.TargetCreateTarget{URL: agentURL})
		if pageErr != nil {
			_ = browser.Close()
			return nil, pageErr
		}
		if userAgent := strings.TrimSpace(instance.UserAgent); userAgent != "" {
			agentPage = agentPage.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: userAgent})
		}
		runtime = &state.BrowserRuntime{Instance: instance, Browser: browser, StartTime: time.Now(), ControlURL: controlURL, AgentURL: agentURL, AgentToken: agentToken}
	} else {
		launcherInstance := launcher.New()
		binPath := strings.TrimSpace(instance.BinPath)
		if binPath == "" {
			for _, path := range []string{
				"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
				"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
			} {
				if _, statErr := os.Stat(path); statErr == nil {
					binPath = path
					break
				}
			}
		}
		if binPath != "" {
			launcherInstance.Bin(binPath)
		}
		if instance.UserDataDir != "" {
			if mkdirErr := os.MkdirAll(instance.UserDataDir, 0o755); mkdirErr != nil {
				return nil, mkdirErr
			}
			launcherInstance.UserDataDir(instance.UserDataDir)
		}
		headless := false
		if instance.Headless != nil {
			headless = *instance.Headless
		}
		launcherInstance.Headless(headless)
		var proxyUsername, proxyPassword string
		if instance.Proxy != "" {
			proxyValue := strings.TrimSpace(instance.Proxy)
			parsedProxy, parseErr := url.Parse(proxyValue)
			if parseErr != nil {
				return nil, parseErr
			}
			if parsedProxy.User != nil {
				proxyUsername = parsedProxy.User.Username()
				proxyPassword, _ = parsedProxy.User.Password()
				proxyValue = fmt.Sprintf("%s://%s", parsedProxy.Scheme, parsedProxy.Host)
			}
			launcherInstance.Proxy(proxyValue)
		}
		if instance.NoSandbox != nil && *instance.NoSandbox {
			launcherInstance.Set(flags.Flag("no-sandbox"))
		}
		launchArgs := instance.LaunchArgs
		if len(launchArgs) == 0 {
			launchArgs = []string{
				"--disable-blink-features=AutomationControlled",
				"--excludeSwitches=enable-automation",
				"--no-first-run",
				"--no-default-browser-check",
				"--disable-infobars",
			}
		}
		if headless {
			launchArgs = append(launchArgs,
				"--disable-background-timer-throttling",
				"--disable-backgrounding-occluded-windows",
				"--disable-renderer-backgrounding",
				"--disable-ipc-flooding-protection",
			)
		}
		hasStartMaximized, hasWindowPosition, hasWindowSize := false, false, false
		for _, arg := range launchArgs {
			key, _, _ := strings.Cut(strings.TrimSpace(strings.TrimPrefix(arg, "--")), "=")
			if key == "start-maximized" {
				hasStartMaximized = true
			}
			if key == "window-position" {
				hasWindowPosition = true
			}
			if key == "window-size" {
				hasWindowSize = true
			}
		}
		if !hasStartMaximized {
			launcherInstance.Set(flags.Flag("start-maximized"))
		}
		if !hasWindowPosition {
			launcherInstance.Set(flags.Flag("window-position"), "0,0")
		}
		if !hasWindowSize {
			launcherInstance.Set(flags.Flag("window-size"), "1920,1080")
		}
		for _, arg := range launchArgs {
			arg = strings.TrimSpace(strings.TrimPrefix(arg, "--"))
			if arg == "" {
				continue
			}
			key, value, hasValue := strings.Cut(arg, "=")
			if hasValue {
				launcherInstance.Set(flags.Flag(key), value)
			} else {
				launcherInstance.Set(flags.Flag(key))
			}
		}
		controlURL, launchErr := launcherInstance.Launch()
		if launchErr != nil {
			return nil, launchErr
		}
		browser := rod.New().ControlURL(controlURL)
		if err = browser.Connect(); err != nil {
			if strings.TrimSpace(instance.UserDataDir) != "" {
				launcherInstance.Kill()
			} else {
				launcherInstance.Cleanup()
			}
			return nil, err
		}
		if proxyUsername != "" && proxyPassword != "" {
			go browser.HandleAuth(proxyUsername, proxyPassword)()
		}
		_ = (&proto.BrowserGrantPermissions{
			Permissions: []proto.BrowserPermissionType{
				proto.BrowserPermissionTypeClipboardReadWrite,
				proto.BrowserPermissionTypeClipboardSanitizedWrite,
			},
		}).Call(browser)
		agentPage, pageErr := browser.Page(proto.TargetCreateTarget{URL: agentURL, NewWindow: true})
		if pageErr != nil {
			_ = browser.Close()
			if strings.TrimSpace(instance.UserDataDir) != "" {
				launcherInstance.Kill()
			} else {
				launcherInstance.Cleanup()
			}
			return nil, pageErr
		}
		if userAgent := strings.TrimSpace(instance.UserAgent); userAgent != "" {
			agentPage = agentPage.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: userAgent})
		}
		left, top, width, height := 0, 0, 1920, 1080
		if err = agentPage.SetWindow(&proto.BrowserBounds{Left: &left, Top: &top, Width: &width, Height: &height}); err != nil {
			_ = browser.Close()
			if strings.TrimSpace(instance.UserDataDir) != "" {
				launcherInstance.Kill()
			} else {
				launcherInstance.Cleanup()
			}
			return nil, err
		}
		if err = agentPage.SetWindow(&proto.BrowserBounds{WindowState: proto.BrowserWindowStateMaximized}); err != nil {
			_ = browser.Close()
			if strings.TrimSpace(instance.UserDataDir) != "" {
				launcherInstance.Kill()
			} else {
				launcherInstance.Cleanup()
			}
			return nil, err
		}
		runtime = &state.BrowserRuntime{Instance: instance, Browser: browser, Launcher: launcherInstance, StartTime: time.Now(), ControlURL: controlURL, AgentURL: agentURL, AgentToken: agentToken}
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	state.BrowserMu.Lock()
	state.BrowserInstances[instance.ID] = runtime
	state.BrowserCurrentInstanceID = instance.ID
	startTime := runtime.StartTime
	instanceCopy := *runtime.Instance
	instanceCopy.IsActive = true
	status := model.BrowserStatus{Running: true, CurrentInstanceID: state.BrowserCurrentInstanceID, Instance: &instanceCopy, StartTime: &startTime, UptimeSeconds: int64(time.Since(startTime).Seconds()), ControlURL: runtime.ControlURL, AgentURL: runtime.AgentURL}
	for listener := range state.BrowserStatusListeners {
		select {
		case listener <- status:
		default:
		}
	}
	state.BrowserMu.Unlock()
	return &v1.BrowserInstanceStartRes{Status: &status}, nil
}
