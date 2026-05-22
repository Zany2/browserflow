<p align="center">
  <img src="docs/images/layout-logo.png" alt="BrowserFlow" width="100" height="100">
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

BrowserFlow is a dual-mode browser automation platform. It turns Automa workflows into managed, syncable, schedulable, traceable, and LLM-callable automation capabilities.

It supports two runtime modes:

- **Windows mode**: for local personal use. The backend, frontend, and controlled browser run on the same Windows machine, and business data is stored in a local BoltDB file.
- **Server mode**: for server deployment and multi-client scheduling. The backend runs as a scheduling center, business data uses PostgreSQL, and Redis stores online state, workflow inventories, and per-client task locks.

## Quick Start

### 1. Configure runtime mode

Edit `backend/manifest/config/config.yaml`:

```yaml
app:
  mode: "windows"
```

Available values:

- `windows`: local personal mode, backed by BoltDB.
- `server`: server scheduling mode, backed by PostgreSQL and Redis.

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

In development, Vite proxies `/api` requests to `http://localhost:8001`.

## Core Features

- **Dual runtime modes**: one frontend/backend codebase supports local Windows mode and server scheduling mode.
- **Automa Bridge**: communicates with the browser extension through page events to read, import, open, and run Automa workflows.
- **Browser Agent**: connects the backend-launched browser in Windows mode and receives local workflow commands over WebSocket.
- **Client Agent**: lets Windows clients join a server-mode scheduling center through a web page, receive tasks, and execute local Automa workflows.
- **Workflow management**: supports file import, client sync, sync candidate comparison, protected state, syncability state, and server-side workflow records.
- **Task configuration**: combines workflow, client, parameters, and Cron expression into executable tasks in server mode.
- **Task dispatching**: supports dispatching to a configured client, or scanning online clients that own the target workflow when no client is configured.
- **Execution records**: stores dispatching, running, success, failure, timeout, parameters, return data, and failure reasons.
- **Per-client task locks**: uses Redis in server mode to prevent one client from running multiple Automa workflows concurrently.
- **LLM features**: supports local LLM configuration, chat, and workflow Skill export in Windows mode.
- **Mode-aware routing**: the frontend hides or blocks pages that do not apply to the active backend runtime mode.

## Runtime Boundaries

| Area | Windows Mode | Server Mode |
| --- | --- | --- |
| Purpose | Single-machine personal use | Server scheduling center |
| Main data source | BoltDB file configured by `localStorage.path` | PostgreSQL and Redis |
| Backend location | Local Windows machine | Server |
| Executor | Local controlled browser launched by the backend | Windows clients visiting `/client-agent` |
| Agent page | `/browser-agent` | `/client-agent` |
| Workflow source | Automa workflows in the controlled browser | Synced from clients or imported into the server |
| Task dispatch | Local workflow execution | Server task configuration, Cron scheduling, and records |
| Best for | Local automation, debugging, personal LLM tools | Multi-client management, remote scheduling, centralized records |

### Windows Mode

Windows mode is designed for packaging BrowserFlow as a local executable. The frontend can be bundled with the executable and served on the same machine.

Typical flow:

1. The backend launches a local controlled browser instance.
2. The browser opens `/browser-agent` automatically.
3. Browser Agent resolves or generates a `Browser ID` and connects to the backend over WebSocket.
4. The backend reads Automa workflows from the current browser through Browser Agent.
5. Users can open, run, and export workflows, or configure an LLM and chat locally.

Windows mode should not depend on PostgreSQL or Redis. Local business data belongs in the BoltDB file configured by `localStorage.path`.

### Server Mode

Server mode is designed for deploying BrowserFlow on a server as a scheduling center for multiple Windows clients.

Typical flow:

1. Start the BrowserFlow backend and frontend on the server.
2. A Windows client visits the site and opens `/client-agent`.
3. Client Agent generates a stable `client_*` identifier and registers over WebSocket.
4. The server records client IP, online state, browser metadata, Automa plugin state, and workflow inventory.
5. The management UI syncs Automa workflows from clients into the server, or imports workflow files manually.
6. A task is created by binding a workflow, client, parameters, and optional Cron expression.
7. The server dispatches the task to an online client, and the client invokes local Automa to run it and report state and results.
8. The server writes execution records for UI queries or LLM-driven flows.

In server mode, PostgreSQL is the source of truth for business data. Redis is used for online clients, workflow inventory cache, and per-client task locks.

## Server Dispatch Rules

- If a task has a configured client IP, dispatch only to that client.
- If the target client is offline, does not own the workflow, or is busy, the run creates a failed execution record with a clear reason.
- If a task has no configured client IP, the server scans online clients that own the workflow and selects an unlocked client.
- If all candidate clients are busy, the run creates a failed record with a reason such as: `已遍历调度所有在线且拥有工作流的客户端，均处于繁忙状态，任务执行失败`.
- Cron expressions trigger tasks at their scheduled time points. Scheduling does not wait until the previous run finishes before counting the next interval; the per-client lock prevents concurrent execution on the same client.
- Client task locks have a TTL fallback so disconnects, server failures, or missing callbacks do not leave permanent locks.
- A background scanner marks stale running records as failed or timed out and releases the related client lock.

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
  mode: "server"

localStorage:
  path: "data/browserflow.db"

frontend:
  url: "http://localhost:5173"

database:
  default:
    link: "pgsql:USER:PASSWORD@tcp(HOST:5432)/browserflow"

redis:
  default:
    address: HOST:6379
    db: 0
    pass: ""
```

| Option | Description |
| --- | --- |
| `server.address` | Backend HTTP listen address |
| `app.mode` | Runtime mode, supports `windows` and `server` |
| `localStorage.path` | Local BoltDB file path for Windows mode |
| `frontend.url` | Frontend URL opened by backend-launched controlled browsers |
| `database.default.link` | PostgreSQL connection for server mode |
| `redis.default` | Redis connection for server mode |

## Requirements

- Go 1.25+
- Node.js and npm
- Chrome or Chromium
- Automa extension installed and enabled in the target browser for Automa-related features
- PostgreSQL and Redis for server mode

## Page Map

| Page | Mode | Description |
| --- | --- | --- |
| `/` | Shared | Home page and runtime-mode entry |
| `/browser` | Windows | Manage local controlled browser instances |
| `/workflows` | Windows | View workflows from Browser Agent, open or run them, and export Skills |
| `/llm` | Windows | Configure LLM providers, models, API keys, and Base URLs |
| `/chat` | Windows | Chat with enabled local model configurations |
| `/browser-agent` | Windows | Local browser executor page, usually opened automatically by the backend |
| `/automa` | Server | Manage server workflow records, imports, and client sync |
| `/tasks` | Server | Create and maintain task definitions |
| `/task-records` | Server | View task execution records, results, and failure reasons |
| `/clients` | Server | View client online state, plugin state, browser metadata, and ban state |
| `/client-agent` | Server | Windows client executor page |

## Automa Bridge Events

| Event | Direction | Description |
| --- | --- | --- |
| `__automa-ext__` | Frontend to extension | Unified bridge request entry |
| `__automa-ext__get-workflows` | Extension to frontend | Returns local Automa workflows |
| `__automa-ext__add-workflow` | Extension to frontend | Returns workflow import result |
| `automa:execute-workflow` | Frontend to extension | Triggers Automa workflow execution |

If Automa is installed after an agent page is already open, its content script usually will not be injected into the existing page automatically. Refresh the agent page or reopen the controlled browser in that case.

## Project Structure

```text
browserflow/
|-- backend/                 GoFrame backend service
|-- frontend/                Vue 3 frontend app
|-- docs/                    Documentation and README image assets
|-- third_party/automa/      Local Automa source snapshot
|-- go.work                  Go workspace
|-- README.md                Chinese README
`-- README_EN.md             English README
```

## Development Commands

Backend:

```bash
cd backend
go run .
```

Frontend:

```bash
cd frontend
npm run dev
npm run build
npm run lint
npm run format
```

Database initialization or update:

```text
backend/sql/browserflow.sql
```

## Security Notes

When deploying server mode on a LAN or the internet, make sure PostgreSQL, Redis, backend APIs, the frontend entry, and `/client-agent` access are protected by a trusted network or reverse proxy. BrowserFlow can trigger real browser automation tasks, so exposing it publicly without authentication, authorization, and network isolation is not recommended.

## License

Apache-2.0
