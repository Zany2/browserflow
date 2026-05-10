<p align="center">
  <img width="96" height="96" alt="BrowserFlow" src="docs/images/layout-logo.png">
</p>

<h1 align="center">BrowserFlow</h1>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white" />
  <img alt="Vue" src="https://img.shields.io/badge/Vue-3-42B883?logo=vuedotjs&logoColor=white" />
  <img alt="Vite" src="https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white" />
  <img alt="Element Plus" src="https://img.shields.io/badge/Element%20Plus-2-409EFF" />
  <img alt="GoFrame" src="https://img.shields.io/badge/GoFrame-2.10-00ADD8" />
  <img alt="Automa" src="https://img.shields.io/badge/Automa-Bridge-6C5CE7" />
</p>

<p align="center">
  <a href="./README.md">简体中文</a> · English
</p>

> **Manage browser instances, Automa workflows, task dispatching, and local LLM chat from one console.** BrowserFlow supports both a local Windows desktop mode and a server scheduling mode, with Automa bridge integration for reading, syncing, and running browser workflows.

```bash
# Start backend
cd backend && go run .

# Start frontend
cd frontend && npm install && npm run dev

# Open
http://localhost:5173
```

## Highlights

**A lightweight management platform for Automa and browser agents**

- **Dual runtime modes**: one codebase supports local `windows` mode and remote `server` scheduling mode.
- **Browser instance management**: create, start, stop, and switch local or remote browser instances.
- **Browser Agent**: controlled browsers receive commands over WebSocket and report Automa status/version.
- **Automa Bridge**: communicate with the Automa extension through page events to read workflows and trigger runs.
- **Workflow management**: import files, sync from clients, compare sync candidates, and manage backend records.
- **Task dispatching**: send workflow tasks to online clients and persist execution records.
- **LLM configuration and chat**: manage model providers and create local chat sessions.
- **Mode-aware routing**: the frontend disables pages that do not apply to the current backend runtime mode.

## Requirements

- Go 1.25+
- Node.js and npm
- Chrome or Chromium
- Automa extension installed and enabled in the target browser for Automa-related features
- PostgreSQL and Redis configuration when using server-mode client cache and workflow syncing features

## Quick Start

### 1. Configure runtime mode

Edit `backend/manifest/config/config.yaml`:

```yaml
app:
  mode: "windows"
```

Available modes:

- `windows`: local desktop mode
- `server`: server scheduling mode

### 2. Start backend

```bash
cd backend
go run .
```

Default backend URL:

```text
http://localhost:8001
```

Swagger:

```text
http://localhost:8001/swagger
```

### 3. Start frontend

```bash
cd frontend
npm install
npm run dev
```

Default frontend URL:

```text
http://localhost:5173
```

In development, Vite proxies `/api` to `http://localhost:8001`.

## Runtime Modes

### Windows Mode

Designed for local usage, focusing on local browser control, current-browser Automa workflows, and local LLM chat.

Available pages:

- Browser
- Workflows
- LLM
- Chat

Disabled pages:

- Workflow Management
- Tasks
- Task Records
- Clients
- Client Agent

### Server Mode

Designed as a scheduling center for remote clients, backend workflow records, task dispatching, and execution results.

Available pages:

- Workflow Management
- Tasks
- Task Records
- Clients
- Client Agent

Disabled pages:

- Browser
- LLM
- Chat
- Browser Agent

## Usage Guide

### Browser Agent

`/browser-agent` is usually opened automatically by a controlled browser launched from the backend. It is responsible for:

1. Establishing a WebSocket connection with the backend.
2. Periodically checking whether the Automa bridge is available.
3. Reporting `automa_installed` and `automa_version`.
4. Receiving workflow list/open/run commands from the backend.
5. Reloading the page when the Automa bridge cannot be detected, so newly installed extensions can inject their content scripts.

### Workflows

BrowserFlow has two workflow entry points that serve different purposes:

- `/workflows`: Windows-mode page. Data comes from the Automa bridge in the browser currently opening the management UI.
- `/automa`: Server-mode page. Data comes from backend workflow records or sync cache.

### Automa Bridge Events

| Event | Direction | Description |
| --- | --- | --- |
| `__automa-ext__` | Frontend to extension | Unified bridge request entry |
| `__automa-ext__get-workflows` | Extension to frontend | Returns local Automa workflows |
| `__automa-ext__add-workflow` | Extension to frontend | Returns workflow import result |
| `automa:execute-workflow` | Frontend to extension | Triggers Automa workflow execution |

> If Automa is installed after a page is already open, the content script usually will not be injected into that existing page automatically. Browser Agent reloads itself after repeated bridge detection failures to handle this case.

### LLM Chat

In Windows mode:

1. Configure providers, models, API keys, and Base URLs on the LLM page.
2. Enable one or more model configurations.
3. Select an enabled model on the Chat page and create sessions.

The Chat page uses existing model configurations only. It does not create new LLM configurations.

## Configuration

Main configuration file:

```text
backend/manifest/config/config.yaml
```

Common options:

```yaml
server:
  address: ":8001"

app:
  mode: "windows"

localStorage:
  path: "data/browserflow.db"

frontend:
  url: "http://localhost:5173"
```

| Option | Description |
| --- | --- |
| `server.address` | Backend HTTP listen address |
| `app.mode` | Runtime mode, supports `windows` and `server` |
| `localStorage.path` | Local bbolt database file path |
| `frontend.url` | Frontend URL opened by backend-launched controlled browsers |

## Project Structure

```text
browserflow/
├─ backend/                  GoFrame backend service
├─ frontend/                 Vue 3 frontend app
├─ docs/images/              README and documentation image assets
├─ automa/                   Local Automa source snapshot, ignored by Git
├─ browserwing/              Local BrowserWing source snapshot, ignored by Git
├─ AUTOMA.md                 Local Automa patch notes
├─ go.work                   Go workspace
└─ README.md
```

## Development

Frontend:

```bash
cd frontend
npm run dev
npm run build
npm run lint
npm run format
```

Backend:

```bash
cd backend
go run .
```

## Git Notes

The following paths are local runtime data, external source snapshots, or personal configuration and should not be committed:

- `backend/data/`
- `backend/user_data/`
- `backend/logs/`
- `AGENTS.md`
- `CLAUDE.md`
- `automa/`
- `browserwing/`

If these files are already tracked by Git, adding them to `.gitignore` will not untrack them automatically. Use `git rm --cached` before committing.

`backend/manifest/config/config.yaml` may contain database, Redis, or local service addresses. Check it before committing to avoid exposing sensitive information.

## Automa Local Changes

`automa/` is treated as an external source snapshot and is not committed by default. Local modifications to Automa are documented in `AUTOMA.md`. After updating or replacing `automa/`, review that file and reapply required changes.
