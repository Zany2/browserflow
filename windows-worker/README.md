# BrowserFlow Windows Worker

中文说明 | [English](README.en.md)

BrowserFlow Windows Worker 是一个 Windows 桌面执行器原型，用来让一台 Windows 电脑启动多个相互隔离的 Chrome 执行节点。

它面向不熟悉命令行的办公人员：用户双击 `BrowserFlowWorker.exe`，填写服务端地址，选择或自动检测 Chrome，然后点击启动即可。

## 功能

- 提供 Windows 原生 GUI 窗口，支持双击运行。
- 自动检测本机 Chrome，也支持手动选择 `chrome.exe`。
- 支持配置 BrowserFlow 服务端地址和执行节点数。
- 为每个执行节点创建独立 Chrome profile 目录。
- 每个节点启动一个独立 Chrome 窗口。
- 支持设置下载根目录，并为每个执行节点自动分配独立下载子目录。
- 支持从 zip 安装/更新 Automa，并在界面显示插件安装/更新时间。
- 兼容新版 Chrome 的 Automa 自动加载逻辑。
- 使用前端 `favicon.ico` 作为窗口图标和 exe 图标。

## 目录结构

```text
windows-worker/
  assets/
    favicon.ico
  cmd/browserflow-worker/
    main.go
  internal/
    app/
    chrome/
    config/
    worker/
  winres/
    icon-only.json
    winres.json
  go.mod
  README.md
  README.en.md
```

运行或编译后可能生成：

```text
windows-worker/
  BrowserFlowWorker.exe
```

`BrowserFlowWorker.exe` 可以直接放到桌面运行。配置、日志、运行数据和扩展目录不会保存到桌面或 exe 同目录，而是统一保存到当前 Windows 用户目录。

## 使用方式

1. 双击 `BrowserFlowWorker.exe`。
2. 填写 BrowserFlow 服务端地址，例如公司提供的管理端地址。
3. 点击“自动检测”检测 Chrome，或点击“浏览...”手动选择 `chrome.exe`。
4. 设置“执行节点数”。
5. 设置“下载目录”，可以手动填写，也可以点击旁边的“浏览...”选择目录。
6. 点击“保存配置”。
7. 首次使用时，点击“安装/更新 Automa”，选择 Automa zip 安装包。
8. 点击“启动执行节点”。
9. 需要停止已启动的 Chrome 节点时，点击“关闭所有执行节点”。
10. 如果启动过执行节点后点击窗口右上角 `X`，程序会询问是否同时关闭所有执行节点。

启动后，每个执行节点会使用独立 profile 打开一个 Chrome 窗口，并访问客户端入口。

当前生成的访问地址格式为：

```text
{server_url}/#/client-agent?node_id=node-1
```

## 配置文件

第一次运行 `BrowserFlowWorker.exe` 时，如果没有 `config.json`，程序会自动生成配置文件。

统一数据目录：

```text
%LOCALAPPDATA%\BrowserFlowWorker
```

通常对应：

```text
C:\Users\用户名\AppData\Local\BrowserFlowWorker
```

目录结构：

```text
%LOCALAPPDATA%\BrowserFlowWorker\
  config.json
  logs\
    worker.log
  downloads\
    node-1\
    node-2\
  data\
    profiles\
      node-1\
      node-2\
  extensions\
    automa\
```

示例：

```json
{
  "server_url": "",
  "chrome_path": "",
  "node_count": 1,
  "machine_id": "bfw-xxxxxx",
  "machine_name": "OFFICE-PC",
  "data_dir": "C:\\Users\\用户名\\AppData\\Local\\BrowserFlowWorker\\data",
  "download_dir": "C:\\Users\\用户名\\AppData\\Local\\BrowserFlowWorker\\downloads",
  "automa_extension_dir": "C:\\Users\\用户名\\AppData\\Local\\BrowserFlowWorker\\extensions\\automa",
  "require_automa_folder": false
}
```

字段说明：

- `server_url`: BrowserFlow 服务端地址。
- `chrome_path`: Chrome 完整路径；留空时程序会自动检测。
- `node_count`: 要启动的执行节点数量，当前限制为 1 到 8。
- `machine_id`: 当前 Windows 电脑的稳定 ID。
- `machine_name`: 当前 Windows 电脑显示名称。
- `data_dir`: profile 和本地数据目录。
- `download_dir`: 执行节点下载根目录，程序会按节点分为 `node-1`、`node-2` 等子目录。
- `automa_extension_dir`: Automa 扩展目录。
- `require_automa_folder`: 是否要求 Automa 扩展目录必须存在。

## 下载目录

界面中的“下载目录”是所有执行节点的下载根目录。启动节点时，程序会自动写入每个 Chrome profile 的下载设置：

```text
{download_dir}\node-1
{download_dir}\node-2
```

例如下载目录为：

```text
D:\BrowserFlowDownloads
```

则节点下载位置为：

```text
D:\BrowserFlowDownloads\node-1
D:\BrowserFlowDownloads\node-2
```

这样多个节点同时导出文件时不会混在同一个目录里。

如果不填写下载目录，默认使用：

```text
%LOCALAPPDATA%\BrowserFlowWorker\downloads
```

注意：下载设置是在启动执行节点前写入对应 Chrome profile 的。如果某个节点 Chrome 已经打开，需要关闭该节点后重新启动，新的下载目录才会生效。

## 启动性能

启动执行节点时会同时做几件事：创建或检查独立 profile、写入下载设置、加载 Automa 扩展、打开 Chrome 并进入客户端页面。因此第一次启动或一次启动多个节点时，Chrome 有几秒钟卡顿是正常现象。

建议：

- 普通办公电脑先从 1 到 2 个节点开始测试。
- 确认可用后再逐步增加到 3 到 4 个节点。
- 节点数越多，占用的内存、CPU 和磁盘 IO 越高。
- Chrome for Testing 第一次加载 Automa 扩展时通常会比后续启动更慢一点。

## Automa 扩展

如果需要自动加载 Automa 扩展，可以在界面中点击“安装/更新 Automa”，选择 Automa zip 安装包，例如：

```text
automa-chrome-v1.30.00.zip
```

程序会自动解压到：

```text
%LOCALAPPDATA%\BrowserFlowWorker\extensions\automa
```

界面底部会显示 Automa 插件状态：

```text
未安装（点击“安装/更新 Automa”选择 zip 安装包）
```

或：

```text
已安装，更新时间：2026-05-25 11:17:09
```

启动 Chrome 时，程序会优先使用 Chrome 启动参数加载扩展：

```text
--load-extension=extensions/automa
--disable-extensions-except=extensions/automa
```

新版 Chrome 对命令行加载扩展有额外限制，因此程序还会通过 Chrome DevTools Protocol 调用 `Extensions.loadUnpacked` 再加载一次解压后的 Automa 目录，然后跳转到客户端入口页面。

如果 `extensions/automa` 不存在，默认不会阻止启动。需要强制要求扩展存在时，把配置改为：

```json
"require_automa_folder": true
```

后续升级 Automa 时，再次点击“安装/更新 Automa”，选择新的 zip 安装包即可。

## Chrome 版本要求

Automa 当前安装包的 `manifest.json` 中声明了：

```json
"minimum_chrome_version": "116"
```

因此 BrowserFlowWorker 建议使用：

- 最低版本：Chrome 116 及以上。
- 首选推荐：Chrome for Testing 136.0.7103.94 Windows 64 位。
- 备用推荐：普通 Chrome 136.0.7103.114 及以下。
- Chrome 136 及以下：通常可以直接通过 `--load-extension` 加载 Automa。
- Chrome 137 及以上：普通 Google Chrome 对 `--load-extension` 做了限制，BrowserFlowWorker 会尝试使用 Chrome DevTools Protocol 的 `Extensions.loadUnpacked` 进行兼容加载，但普通正式版 Chrome 不保证可用。

推荐把 Chrome for Testing 作为 BrowserFlowWorker 的专用执行浏览器，不影响用户日常使用的普通 Chrome。

下载地址：

```text
https://storage.googleapis.com/chrome-for-testing-public/136.0.7103.94/win64/chrome-win64.zip
```

使用方式：

1. 下载 `chrome-win64.zip`。
2. 解压到固定目录，例如：

```text
C:\BrowserFlow\chrome-win64
```

3. 在 BrowserFlowWorker 中点击“浏览...”，选择：

```text
C:\BrowserFlow\chrome-win64\chrome.exe
```

4. 再点击“启动执行节点”。

如果用户电脑存在企业策略限制，例如禁用调试端口、禁止安装扩展、禁止加载 unpacked extension，自动加载可能会失败。这种场景建议使用 Chrome for Testing、Chromium，或由管理员手动允许/安装 Automa 扩展。

## 编译准备

安装 `go-winres`：

```bash
GOBIN=D:/Go/bin go install github.com/tc-hib/go-winres@latest
```

确认安装成功：

```bash
ls /d/Go/bin/go-winres.exe
```

仓库根目录存在 `go.work`，当前不包含 `windows-worker` 模块。编译这个独立模块时必须关闭 workspace 模式：

```bash
GOWORK=off
```

如果你的 `go env GOOS` 不是 `windows`，也必须显式指定：

```bash
GOOS=windows GOARCH=amd64
```

否则即使输出文件名叫 `.exe`，生成出来的也可能不是 Windows 可执行文件。

## 编译带图标 exe

在 Git Bash 中执行：

```bash
cd /g/code/browserflow/windows-worker
rm -f BrowserFlowWorker*.exe
rm -f cmd/browserflow-worker/*.syso
rm -f rsrc*.syso
rm -rf extracted*

/d/Go/bin/go-winres.exe simply \
  --arch amd64 \
  --icon assets/favicon.ico \
  --manifest gui \
  --file-description "BrowserFlow Worker" \
  --product-name "BrowserFlow Worker" \
  --original-filename "BrowserFlowWorker.exe" \
  --out cmd/browserflow-worker/rsrc

GOOS=windows GOARCH=amd64 GOWORK=off go build -a -ldflags="-H windowsgui" -o BrowserFlowWorker.exe ./cmd/browserflow-worker

rm -f cmd/browserflow-worker/*.syso
```

在 PowerShell 中执行：

```powershell
Set-Location G:\code\browserflow\windows-worker
Remove-Item -Force -ErrorAction SilentlyContinue BrowserFlowWorker*.exe, cmd\browserflow-worker\*.syso, rsrc*.syso
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue extracted*

& D:\Go\bin\go-winres.exe simply `
  --arch amd64 `
  --icon assets\favicon.ico `
  --manifest gui `
  --file-description "BrowserFlow Worker" `
  --product-name "BrowserFlow Worker" `
  --original-filename "BrowserFlowWorker.exe" `
  --out cmd\browserflow-worker\rsrc

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GOWORK = "off"
go build -a -ldflags="-H windowsgui" -o BrowserFlowWorker.exe ./cmd/browserflow-worker

Remove-Item -Force -ErrorAction SilentlyContinue cmd\browserflow-worker\*.syso
```

说明：

- `go-winres simply` 会生成编译期资源文件，用来写入 exe 图标。
- `-H windowsgui` 用来避免双击 exe 时弹出命令行窗口。
- `-a` 强制重新编译，避免缓存影响资源测试。
- 最后的 `Remove-Item` 或 `rm` 只清理临时 `.syso` 文件，不影响已经生成的 exe。

## 验证资源

如需确认 exe 内部是否写入图标资源，可以执行：

```bash
cd /g/code/browserflow/windows-worker
rm -rf extracted-final
/d/Go/bin/go-winres.exe extract --dir extracted-final BrowserFlowWorker.exe
find extracted-final -type f
```

如果能抽取出资源文件，说明资源已写入 exe。

Windows 资源管理器可能缓存旧图标。如果图标看起来没有变化，可以复制成新文件名后再查看：

```bash
cp BrowserFlowWorker.exe BrowserFlowWorker-check.exe
```

## 服务端配套要求

这个 Worker 只是让一台电脑可以启动多个独立 Chrome 节点。要让服务端真正支持同机多节点并发调度，服务端调度身份需要从单纯的 `client_ip` 升级为稳定的节点身份。

建议：

- `machine_id` 表示物理 Windows 电脑。
- `node_id` 表示某一个独立 Chrome 执行节点。
- `client_ip` 只用于展示和辅助定位。
- Redis 执行锁 key 使用 `node_id`。
- 客户端在线清单、工作流清单缓存、任务记录建议保存 `node_id`。

否则，同一台电脑上的多个节点仍然会因为 IP 相同而被服务端当作同一个客户端。

## 常见问题

### 编译时报 go.work 相关错误

错误类似：

```text
directory cmd\browserflow-worker is contained in a module that is not one of the workspace modules listed in go.work
```

解决：

```bash
GOOS=windows GOARCH=amd64 GOWORK=off go build -a -ldflags="-H windowsgui" -o BrowserFlowWorker.exe ./cmd/browserflow-worker
```

### 双击 exe 报错或打不开

先确认当前 Go 目标平台：

```bash
go env GOOS GOARCH
```

如果不是：

```text
windows
amd64
```

请使用完整命令重新编译：

```bash
GOOS=windows GOARCH=amd64 GOWORK=off go build -a -ldflags="-H windowsgui" -o BrowserFlowWorker.exe ./cmd/browserflow-worker
```

### Chrome 检测不到

在界面里点击“浏览...”手动选择：

```text
C:\Program Files\Google\Chrome\Application\chrome.exe
```

或：

```text
C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
```
