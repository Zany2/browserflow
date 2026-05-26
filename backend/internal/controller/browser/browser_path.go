package browser

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-rod/rod/lib/launcher"
)

func resolveLocalBrowserBinPath(configuredPath string) (string, string) {
	binPath := strings.Trim(strings.TrimSpace(configuredPath), `"`)
	if binPath != "" {
		if info, err := os.Stat(binPath); err == nil && !info.IsDir() {
			return binPath, ""
		}
		if _, err := os.Stat(binPath + ".lnk"); err == nil {
			return "", fmt.Sprintf("浏览器路径指向了快捷方式：%s.lnk。请填写真实浏览器程序路径，例如 C:\\Users\\daixk\\AppData\\Local\\Google\\Chrome\\Application\\chrome.exe。", binPath)
		}
		return "", fmt.Sprintf("浏览器路径不存在：%s。请在浏览器配置中重新选择 chrome.exe 或 msedge.exe，或清空浏览器路径后再启动。", binPath)
	}

	if found, ok := launcher.LookPath(); ok {
		return found, ""
	}
	return "", "未找到可用的 Chrome、Chromium 或 Edge，请先安装浏览器，或在浏览器配置中手动填写 chrome.exe / msedge.exe 路径。"
}
