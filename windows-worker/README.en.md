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
- Supports loading a local Automa extension directory.
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
5. Save the configuration.
6. On first use, click the install/update Automa button and select the Automa zip package.
7. Start the execution nodes.
8. Click the close-all-nodes button when the started Chrome nodes should be stopped.
9. If execution nodes were started and the window `X` is clicked, the app asks whether to close all nodes as well.

Each node opens a Chrome window with an isolated profile and navigates to the client entry page.

The current URL format is:

```text
{server_url}/?machine_id=...&machine_name=...&node_id=...#/client-agent
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
  worker.log
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
- `automa_extension_dir`: Automa extension directory.
- `require_automa_folder`: Whether the Automa extension directory must exist.

## Automa Extension

To load Automa automatically, click the install/update Automa button in the window and select an Automa zip package, for example:

```text
automa-chrome-v1.30.00.zip
```

The app extracts it to:

```text
%LOCALAPPDATA%\BrowserFlowWorker\extensions\automa
```

Chrome is launched with:

```text
--load-extension=extensions/automa
--disable-extensions-except=extensions/automa
```

If `extensions/automa` does not exist, startup is not blocked by default. To require the extension directory, set:

```json
"require_automa_folder": true
```

To upgrade Automa later, click install/update Automa again and select the newer zip package.

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
