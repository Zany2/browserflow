# BrowserFlow Windows Worker

[中文说明](README.md) | English

BrowserFlow Windows Worker is a Windows desktop runner prototype for starting multiple isolated Chrome execution nodes on one Windows machine.

It is designed for office users who are not comfortable with command lines. They can double-click `BrowserFlowWorker.exe`, fill in the server URL, select or auto-detect Chrome, and start execution nodes from a small native window.

## Features

- Native Windows GUI, suitable for double-click usage.
- Auto-detects local Chrome and also supports manually selecting `chrome.exe`.
- Configures BrowserFlow server URL and execution node count.
- Creates an isolated Chrome profile directory for each node.
- Starts one independent Chrome window for each node.
- Supports a configurable download root and creates a separate download folder for each node.
- Supports installing/updating Automa from a zip package and displays the plugin install/update time.
- Supports Automa auto-loading compatibility for newer Chrome versions.
- Uses the frontend `favicon.ico` as both the window icon and exe icon.

## Directory Layout

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

Runtime and build output may include:

```text
windows-worker/
  BrowserFlowWorker.exe
```

`BrowserFlowWorker.exe` can be placed on the desktop. Configuration, logs, runtime data, and extension files are not written next to the exe or onto the desktop. They are stored together under the current Windows user's profile directory.

## Usage

1. Double-click `BrowserFlowWorker.exe`.
2. Enter the BrowserFlow server URL provided by your administrator.
3. Click the auto-detect button for Chrome, or browse and select `chrome.exe` manually.
4. Set the execution node count.
5. Set the download directory. You can type it manually or use the browse button next to the field.
6. Save the configuration.
7. On first use, click the install/update Automa button and select the Automa zip package.
8. Start the execution nodes.
9. Click the close-all-nodes button when the started Chrome nodes should be stopped.
10. If execution nodes were started and the window `X` is clicked, the app asks whether to close all nodes as well.

Each node opens a Chrome window with an isolated profile and navigates to the client entry page.

The current URL format is:

```text
{server_url}/#/client-agent?node_id=node-1
```

## Configuration

On first launch, the app creates `config.json` if it does not exist.

Unified data directory:

```text
%LOCALAPPDATA%\BrowserFlowWorker
```

Usually:

```text
C:\Users\Username\AppData\Local\BrowserFlowWorker
```

Directory layout:

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

Example:

```json
{
  "server_url": "",
  "chrome_path": "",
  "node_count": 1,
  "machine_id": "bfw-xxxxxx",
  "machine_name": "OFFICE-PC",
  "data_dir": "C:\\Users\\Username\\AppData\\Local\\BrowserFlowWorker\\data",
  "download_dir": "C:\\Users\\Username\\AppData\\Local\\BrowserFlowWorker\\downloads",
  "automa_extension_dir": "C:\\Users\\Username\\AppData\\Local\\BrowserFlowWorker\\extensions\\automa",
  "require_automa_folder": false
}
```

Fields:

- `server_url`: BrowserFlow server URL.
- `chrome_path`: Full path to Chrome. Leave empty to auto-detect.
- `node_count`: Number of execution nodes to start. Currently limited to 1 through 8.
- `machine_id`: Stable ID for the current Windows machine.
- `machine_name`: Display name for the current Windows machine.
- `data_dir`: Local profile and data directory.
- `download_dir`: Download root directory for execution nodes. The app creates per-node folders such as `node-1` and `node-2`.
- `automa_extension_dir`: Automa extension directory.
- `require_automa_folder`: Whether the Automa extension directory must exist.

## Download Directory

The download directory field is the root download directory for all execution nodes. When nodes start, the app writes the download setting into each Chrome profile:

```text
{download_dir}\node-1
{download_dir}\node-2
```

For example, if the download directory is:

```text
D:\BrowserFlowDownloads
```

node downloads go to:

```text
D:\BrowserFlowDownloads\node-1
D:\BrowserFlowDownloads\node-2
```

This keeps exported files from different nodes separated.

If the field is left empty, the default is:

```text
%LOCALAPPDATA%\BrowserFlowWorker\downloads
```

The download setting is written into the Chrome profile before the node starts. If a node is already running, close and start it again for the new download directory to take effect.

## Startup Performance

When starting nodes, the app creates or checks isolated profiles, writes download settings, loads the Automa extension, starts Chrome, and navigates to the client page. A short delay or brief Chrome lag is normal, especially on first startup or when multiple nodes are started together.

Recommendations:

- Start with 1 or 2 nodes on a regular office PC.
- Increase to 3 or 4 nodes only after confirming the machine is stable.
- More nodes use more memory, CPU, and disk IO.
- Chrome for Testing may be slower the first time it loads the Automa extension.

## Automa Extension

To load Automa automatically, click the install/update Automa button in the window and select an Automa zip package, for example:

```text
automa-chrome-v1.30.00.zip
```

The app extracts it to:

```text
%LOCALAPPDATA%\BrowserFlowWorker\extensions\automa
```

The window displays the Automa plugin status:

```text
Not installed
```

or:

```text
Installed, updated at: 2026-05-25 11:17:09
```

Chrome is first launched with extension loading arguments:

```text
--load-extension=extensions/automa
--disable-extensions-except=extensions/automa
```

Newer Chrome versions add restrictions around command-line extension loading. BrowserFlowWorker therefore also uses Chrome DevTools Protocol `Extensions.loadUnpacked` to load the extracted Automa directory, then navigates the node to the client entry page.

If `extensions/automa` does not exist, startup is not blocked by default. To require the extension directory, set:

```json
"require_automa_folder": true
```

To upgrade Automa later, click install/update Automa again and select the newer zip package.

## Chrome Version Requirements

The current Automa package declares this in `manifest.json`:

```json
"minimum_chrome_version": "116"
```

Recommended BrowserFlowWorker Chrome policy:

- Minimum version: Chrome 116 or later.
- Preferred recommendation: Chrome for Testing 136.0.7103.94 for Windows 64-bit.
- Fallback recommendation: regular Chrome 136.0.7103.114 or earlier.
- Chrome 136 or earlier: Automa can usually be loaded through `--load-extension`.
- Chrome 137 or later: regular Google Chrome restricts `--load-extension`. BrowserFlowWorker attempts Chrome DevTools Protocol `Extensions.loadUnpacked` for compatibility, but regular stable Chrome does not guarantee that this method is available.

Use Chrome for Testing as the dedicated BrowserFlowWorker execution browser. This does not affect the user's daily regular Chrome.

Download URL:

```text
https://storage.googleapis.com/chrome-for-testing-public/136.0.7103.94/win64/chrome-win64.zip
```

Usage:

1. Download `chrome-win64.zip`.
2. Extract it to a stable directory, for example:

```text
C:\BrowserFlow\chrome-win64
```

3. In BrowserFlowWorker, click browse and select:

```text
C:\BrowserFlow\chrome-win64\chrome.exe
```

4. Click start execution nodes.

If the user's computer is controlled by enterprise policy, automatic loading may still fail when debugging ports, extension installation, or unpacked extensions are blocked. In that case, use Chrome for Testing, Chromium, or ask the administrator to allow/install the Automa extension.

## Build Prerequisites

Install `go-winres`:

```bash
GOBIN=D:/Go/bin go install github.com/tc-hib/go-winres@latest
```

Verify it:

```bash
ls /d/Go/bin/go-winres.exe
```

The repository root has a `go.work` file that does not include the `windows-worker` module. Disable workspace mode when building this module:

```bash
GOWORK=off
```

If your `go env GOOS` is not `windows`, explicitly set:

```bash
GOOS=windows GOARCH=amd64
```

Otherwise a file named `.exe` may still be built for the wrong operating system.

## Build Iconed Exe

Run in Git Bash:

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

Run in PowerShell:

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

Notes:

- `go-winres simply` creates a compile-time resource file for the exe icon.
- `-H windowsgui` prevents a console window from appearing on double-click.
- `-a` forces a rebuild and avoids stale resource cache issues.
- The final cleanup removes temporary `.syso` files only. It does not affect the generated exe.

## Verify Resources

To confirm resources were embedded:

```bash
cd /g/code/browserflow/windows-worker
rm -rf extracted-final
/d/Go/bin/go-winres.exe extract --dir extracted-final BrowserFlowWorker.exe
find extracted-final -type f
```

If resource files are extracted, the exe contains embedded resources.

Windows Explorer may cache old icons. If the icon does not appear to change, copy the exe to a new filename and check again:

```bash
cp BrowserFlowWorker.exe BrowserFlowWorker-check.exe
```

## Server Requirements

This Worker only starts multiple isolated Chrome nodes on one machine. For true same-machine multi-node scheduling, the server should identify dispatch targets by stable node identity instead of only `client_ip`.

Recommended model:

- `machine_id` identifies a physical Windows machine.
- `node_id` identifies one isolated Chrome execution node.
- `client_ip` is only display and diagnostic metadata.
- Redis execution lock keys should use `node_id`.
- Client inventory, workflow cache, and task records should store `node_id`.

Otherwise, multiple nodes on the same machine will still be treated as the same client because they share the same IP.

## Troubleshooting

### go.work Build Error

Error example:

```text
directory cmd\browserflow-worker is contained in a module that is not one of the workspace modules listed in go.work
```

Fix:

```bash
GOOS=windows GOARCH=amd64 GOWORK=off go build -a -ldflags="-H windowsgui" -o BrowserFlowWorker.exe ./cmd/browserflow-worker
```

### Exe Does Not Open

Check the Go target:

```bash
go env GOOS GOARCH
```

Expected:

```text
windows
amd64
```

If not, rebuild with:

```bash
GOOS=windows GOARCH=amd64 GOWORK=off go build -a -ldflags="-H windowsgui" -o BrowserFlowWorker.exe ./cmd/browserflow-worker
```

### Chrome Is Not Detected

Use the browse button and select:

```text
C:\Program Files\Google\Chrome\Application\chrome.exe
```

or:

```text
C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
```
