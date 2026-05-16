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

> **A dual-mode browser automation platform.** BrowserFlow supports local Windows browser control and server-side task scheduling, integrating Automa workflows, browser/client agents, task dispatching, execution records, and LLM Skill invocation.

```bash
# Start backend
cd backend && go run .

# Start frontend
cd frontend && npm install && npm run dev

# Open console
http://localhost:5173
```

## What It Is

BrowserFlow turns Automa workflows from browser-local manual automations into managed, syncable, schedulable, and LLM-callable automation capabilities.

It has two runtime modes:

- **Windows mode**: the backend launches a controlled browser on the local machine, opens Browser Agent, and reads or runs Automa workflows from that browser.
- **Server mode**: the backend runs on a server as a scheduling center. Windows machines on the LAN or internet open the Client Agent page to connect, and the server manages clients, workflows, tasks, and execution records.

## Core Features

- **Dual runtime modes**: one frontend/backend codebase supports local `windows` mode and remote `server` scheduling mode.
- **Automa Bridge**: communicates with the Automa extension through page events to read, import, open, and run workflows.
- **Browser Agent**: binds to a backend-launched browser in Windows mode and receives workflow commands over WebSocket.
- **Client Agent**: lets any Windows client connect to a server-mode deployment through a web page, receive tasks, and execute local Automa workflows.
- **Workflow management**: supports file import, client sync, sync candidate comparison, syncability control, and backend workflow records.
- **Task scheduling**: combines workflow, client, parameters, and optional Cron expression into executable tasks in server mode.
- **Execution records**: stores dispatch status, client result, success/failure state, parameters, and returned data.
- **LLM Skills**: packages workflows or tasks as LLM-callable Skills with preflight checks for backend, agent, Automa plugin, and parameters.
- **Mode-aware routing**: the frontend disables pages that do not apply to the current backend runtime mode.

## Runtime Modes

| Capability | Windows Mode | Server Mode |
| --- | --- | --- |
| Runtime location | Local Windows machine | Backend on server, execution on Windows clients |
| Agent page | `/browser-agent` | `/client-agent` |
| Workflow source | Automa in the controlled browser | Synced from clients or imported into the server |
| Execution path | Backend sends commands to a Browser Agent | Server dispatches tasks to a Client Agent |
| Best for | Single-machine automation, local debugging, local LLM tools | Multi-client management, remote scheduling, centralized records |

### Windows Mode

Windows mode is designed for controlling a local browser on one machine.

Available pages:

- Browser
- Workflows
- LLM
- Chat
- Browser Agent

Typical flow:

1. The backend launches a local browser instance.
2. The browser opens `/browser-agent` automatically.
3. Browser Agent receives or resolves a `Browser ID` and connects to the backend over WebSocket.
4. The backend reads Automa workflows from the current browser through Browser Agent.
5. Users can open, run, or export workflow Skills.

### Server Mode

Server mode is designed for deploying the backend on a server and connecting multiple Windows clients as executors.

Available pages:

- Workflow Management
- Tasks
- Task Records
- Clients
- Client Agent

Typical flow:

1. Start the BrowserFlow backend and frontend on the server.
2. A Windows client opens the site and visits `/client-agent`.
3. Client Agent generates a stable `client_*` identifier and registers over WebSocket.
4. The server records client IP, online status, browser metadata, Automa plugin status, and Automa version.
5. The management UI syncs Automa workflows from clients into the server, or imports workflow files manually.
6. A task is created by binding a workflow, execution client, parameters, and optional Cron expression.
7. The server dispatches the task to an online client, and the client invokes local Automa to run it.
8. The server updates execution records for the management UI or LLM callers.

## Skill Guidelines

BrowserFlow Skills should make execution preconditions explicit so an LLM does not blindly send requests.

Before running a workflow or task, a Skill should check:

1. **Backend reachability**: call `/api/v1/app/runtime` and verify the service is reachable and in the expected runtime mode.
2. **Agent availability**: in Windows mode, verify the target Browser Agent; in Server mode, verify the target Client Agent or client IP is online.
3. **Automa plugin availability**: check `automa_installed`, plugin status, or client-reported plugin metadata.
4. **Required parameters**: if the workflow or task expects variables, the Skill must require the LLM caller to provide them.
5. **Execution request**: Windows mode can call the workflow run API; Server mode is better modeled as task execution through the scheduling API.

Current task results represent command dispatch and client-side acknowledgement. To wait until an Automa workflow fully completes, fails, or returns exported data, BrowserFlow needs deeper integration with Automa execution-state events or extension callbacks.

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

database:
  default:
    link: "pgsql:USER:PASSWORD@tcp(HOST:5432)/browserflow"

redis:
  default:
    address: HOST:6379
    db: 0
```

| Option | Description |
| --- | --- |
| `server.address` | Backend HTTP listen address |
| `app.mode` | Runtime mode, supports `windows` and `server` |
| `localStorage.path` | Local bbolt database file path |
| `frontend.url` | Frontend URL opened by backend-launched controlled browsers |
| `database.default.link` | SQL storage for server-mode tasks, clients, and related records |
| `redis.default` | Server-mode cache for online clients and workflow inventories |

## Requirements

- Go 1.25+
- Node.js and npm
- Chrome or Chromium
- Automa extension installed and enabled in the target browser for Automa-related features
- PostgreSQL and Redis when using server-mode client management, task scheduling, or workflow cache features

## Key Pages

| Page | Mode | Description |
| --- | --- | --- |
| `/browser` | Windows | Manage local controlled browser instances |
| `/workflows` | Windows | View Automa workflows from the current Browser Agent and export Skills |
| `/llm` | Windows | Configure LLM providers, models, API keys, and Base URLs |
| `/chat` | Windows | Chat with enabled local model configurations |
| `/browser-agent` | Windows | Browser executor page, usually opened automatically by the backend |
| `/automa` | Server | Manage server workflow records, imports, and client sync |
| `/tasks` | Server | Create and maintain task definitions |
| `/task-records` | Server | View task execution records and results |
| `/clients` | Server | View client online state, plugin status, browser metadata, and ban state |
| `/client-agent` | Server | Windows client executor page |

## Automa Bridge Events

| Event | Direction | Description |
| --- | --- | --- |
| `__automa-ext__` | Frontend to extension | Unified bridge request entry |
| `__automa-ext__get-workflows` | Extension to frontend | Returns local Automa workflows |
| `__automa-ext__add-workflow` | Extension to frontend | Returns workflow import result |
| `automa:execute-workflow` | Frontend to extension | Triggers Automa workflow execution |

If Automa is installed after a page is already open, its content script usually will not be injected into that existing page automatically. Agent pages handle this by refreshing or probing again.

## Project Structure

```text
browserflow/
├─ backend/                  GoFrame backend service
├─ frontend/                 Vue 3 frontend app
├─ docs/images/              README and documentation image assets
├─ third_party/automa/        Local Automa source snapshot, source only
├─ third_party/browserwing/  Local BrowserWing source snapshot, not committed
├─ go.work                   Go workspace
├─ README.md
└─ README_EN.md
```

## Development Commands

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

## Security Notes

When deploying server mode on a LAN or the internet, make sure the database, Redis, backend APIs, and frontend entry are protected by a trusted network or reverse proxy. Skill calls can trigger browser automation directly, so exposing them publicly without authentication or access control is not recommended.

## Git Notes

The following paths are local runtime data, external source snapshots, or personal configuration and should not be committed:

- `backend/data/`
- `backend/user_data/`
- `backend/logs/`
- `AGENTS.md`
- `CLAUDE.md`
- `third_party/automa/`
- `third_party/browserwing/`

If these files are already tracked by Git, adding them to `.gitignore` will not untrack them automatically. Use `git rm --cached` before committing.

`backend/manifest/config/config.yaml` may contain database, Redis, or local service addresses. Check it before committing to avoid exposing sensitive information.
