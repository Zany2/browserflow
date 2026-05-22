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

BrowserFlow is an automation platform built around Automa browser workflows. It turns workflows that normally live inside one browser into managed, remotely dispatched, scheduled, audited, result-returning, and LLM-callable automation capabilities.

The project supports two runtime modes:

- **Windows mode**: for local personal automation. The frontend, backend, and controlled browser run on the same Windows machine. Business data is stored in a local BoltDB file.
- **Server mode**: for intranet servers and multiple Windows execution clients. PostgreSQL stores business data, while Redis keeps client online state, workflow inventory cache, and per-client execution locks.

It is useful for systems that have no stable API and must be operated through a browser, especially internal enterprise systems, government intranet systems, private network platforms, data lookup workflows, and report export workflows.

## Contents

- [Core Value](#core-value)
- [Runtime Modes](#runtime-modes)
- [Quick Start](#quick-start)
- [Server Mode Flow](#server-mode-flow)
- [Task Scheduling](#task-scheduling)
- [Skill Execution](#skill-execution)
- [Result Return](#result-return)
- [Page Map](#page-map)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Development Commands](#development-commands)
- [Security And Deployment Notes](#security-and-deployment-notes)

## Core Value

BrowserFlow does not replace Automa. It turns Automa into a manageable, dispatchable, auditable execution capability.

- **Automate systems without APIs**: many intranet, legacy, and private network systems expose only browser pages. BrowserFlow can automate lookup, form filling, export, inspection, and reporting workflows.
- **Centralize multiple Windows clients**: some systems can only be accessed from specific machines, networks, certificates, or browser environments. Server mode lets those machines act as execution nodes.
- **Keep execution evidence**: every run can record trigger type, client, parameters, status, failure reason, result JSON, and result files.
- **Avoid browser concurrency corruption**: one client runs only one Automa workflow at a time, avoiding tab switching, variable conflicts, and context pollution.
- **Schedule tasks with Cron**: tasks can use Cron expressions and are registered into GoFrame gcron.
- **Expose workflows to LLMs**: both Windows and Server modes can export Automa Skills so an LLM can call workflows or task APIs.
- **Friendly to intranet deployment**: Server mode can run inside an internal network. Clients only need to keep `/client-agent` open.

## Runtime Modes

| Area | Windows Mode | Server Mode |
| --- | --- | --- |
| Purpose | Local personal automation | Intranet task scheduling center |
| Storage | BoltDB | PostgreSQL + Redis |
| Backend location | Local Windows machine | Server |
| Executor | Local controlled browser launched by backend | Multiple Windows clients |
| Agent page | `/browser-agent` | `/client-agent` |
| Workflow source | Automa workflows in the controlled browser | Synced from clients or imported into server |
| Task dispatch | Local workflow execution | Task definitions, Cron scheduling, client dispatch |
| Execution records | Local-oriented | Centralized records and queries |
| Best for | Local tools, personal automation, debugging, local LLM calls | Enterprise, government, public security, and other intranet deployments |

### Windows Mode

Windows mode is designed for packaging BrowserFlow as a local executable. The frontend can be bundled into the executable and served by the local backend.

Typical flow:

1. The backend starts a local controlled browser.
2. The browser opens `/browser-agent` automatically.
3. Browser Agent connects to the backend through WebSocket.
4. The backend reads Automa workflows from the current browser through Browser Agent.
5. The user can open, run, and export workflow Skills, or configure an LLM and chat locally.

Boundaries:

- Windows mode should not depend on PostgreSQL or Redis.
- Windows mode business data belongs only in the BoltDB file configured by `localStorage.path`.
- Windows mode is best for personal use, local execution, and workflow debugging.

### Server Mode

Server mode is designed for deploying BrowserFlow on an intranet server as a task scheduling center for multiple Windows clients.

Typical flow:

1. The server starts the BrowserFlow backend and frontend.
2. A Windows client opens `/client-agent`.
3. Client Agent generates a stable client identity and registers through WebSocket.
4. The server records client IP, online state, browser metadata, Automa extension state, and workflow inventory.
5. Administrators sync workflows from clients into the server or import workflow files manually.
6. Administrators create tasks with workflow, parameters, optional client, and optional Cron expression.
7. Manual triggers, Cron triggers, or Skill triggers select a client and send an execution command.
8. The client runs local Automa and returns state and result through WebSocket.
9. The server stores execution records and result files for UI queries or later LLM access.

Boundaries:

- PostgreSQL is the source of truth for Server mode business data.
- Redis stores online clients, workflow inventory cache, and client execution locks.
- Server mode should not use BoltDB as the data source for tasks, workflows, client sync, or execution records.

## Quick Start

### 1. Configure Runtime Mode

Edit:

```text
backend/manifest/config/config.yaml
```

Windows mode:

```yaml
app:
  mode: "windows"
```

Server mode:

```yaml
app:
  mode: "server"
```

### 2. Start Backend

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

### 3. Start Frontend

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

## Server Mode Flow

### 1. Client Access

A Windows client opens:

```text
http://<server-host>/client-agent
```

Client Agent will:

- Establish a WebSocket connection.
- Report client identity, IP, browser metadata, and Automa extension status.
- Send heartbeat messages.
- Report local Automa workflow inventory.
- Receive `task.execute` commands and invoke local Automa.

### 2. Workflow Management

Server mode stores workflows in PostgreSQL. Workflows can come from:

- Syncing Automa workflows from online clients.
- Manually importing Automa workflow files.
- Comparing client workflows with server records to detect synced or updated items.
- Marking workflows as protected to avoid unwanted overwrites.
- Exporting a Server-mode Skill for LLM-driven task API calls.

### 3. Task Configuration

A task definition includes:

- Task name and description.
- Automa workflow ID.
- Optional client IP or client ID.
- Execution parameters `params`.
- Cron expression.
- Enabled state.
- Whether to run once immediately after creation.

If no client IP is configured, Server mode finds an online client that owns the target workflow.

### 4. Execution Records

Every run writes to `task_records`:

- `trigger_type`: `manual`, `cron`, `task_create`, `skill`, `system`.
- `status`: `pending`, `queued`, `running`, `success`, `failed`.
- `client_ip`: actual execution client.
- `params_json`: parameters for this run.
- `result_json`: execution result.
- `error_message`: failure reason.
- `started_at`, `finished_at`: execution timestamps.

Large table results are stored in `task_record_files` and can be viewed or downloaded from the record detail page.

## Task Scheduling

### Manual Execution

The UI or API calls:

```text
POST /api/v1/tasks/{id}/execute
```

The backend reads the task, resolves parameters, chooses a client, acquires a lock, sends a WebSocket command, and creates an execution record.

### Cron Execution

When a task is enabled and has a Cron expression, the background scheduler registers it into GoFrame `gcron`.

Synchronization behavior:

- Sync once when the program starts.
- Sync database task configuration every 30 seconds.
- Read tasks with id-cursor pagination, 500 rows per batch.
- Select only `id` and `cron_expression`, avoiding full task record loading.
- Compare the desired database state with current gcron jobs, then add, remove, or update jobs.

Execution behavior:

- Tasks are registered with `gcron.AddSingleton`.
- If the previous run of the same cron job is still running, the next scheduled hit is skipped. It is not queued and not run concurrently.
- A Cron hit only starts dispatch. Actual execution still checks client online state, workflow ownership, and Redis locks.

### Client Selection

If a task specifies a client:

- Dispatch only to that client.
- If the client is offline, does not own the workflow, or is busy, a failed execution record is created with a readable reason.

If a task does not specify a client:

- The server queries Redis for online clients that own the workflow.
- It walks through candidates and tries to acquire a client lock.
- The first unlocked client receives the command.
- If all candidates are busy, a failed record is created with a reason like:

```text
All online clients that own the workflow have been checked, but all are busy; task execution failed.
```

### Client Lock

Server mode uses Redis locks per client:

- Lock granularity is the client.
- One client runs only one Automa workflow at a time.
- The lock stores task ID, record ID, workflow ID, and command ID.
- Client heartbeat renews the lock.
- Success or failure releases the lock.
- A background fallback sweep handles stale records and prevents permanent deadlocks.

## Skill Execution

BrowserFlow can export two types of Skills.

### Windows Workflow Skill

In Windows mode, the Skill targets workflows currently available in Browser Agent.

Endpoint:

```text
POST /api/v1/workflows/{workflow_id}/run
```

It supports:

- `wait_result`
- `return_data`
- local browser workflow execution

This is suitable for local LLM calls into local browser workflows.

### Server Workflow Skill

In Server mode, the Skill targets workflows stored in the database. It does not directly operate a browser. Instead, it calls task APIs.

Recommended pattern:

1. Create reusable tasks for frequently used workflows.
2. Let the Skill call the task execution API.
3. Use `trigger_type: "skill"` so records are easy to filter.

Execute an existing task:

```bash
curl -X POST 'http://localhost:8001/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{"trigger_type":"skill","client_ip":"","params":{}}'
```

Execute and wait for result:

```bash
curl -X POST 'http://localhost:8001/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{
    "trigger_type": "skill",
    "client_ip": "",
    "params": {},
    "wait_result": true,
    "timeout": 300,
    "return_data": {
      "variables": ["browserflow_output"],
      "include_table": true,
      "table_limit": 100,
      "include_history": false
    }
  }'
```

Notes:

- `wait_result=false`: the API returns once the task is dispatched.
- `wait_result=true`: the API waits for the client final `agent_result` and returns both record and result.
- HTTP wait timeout means the HTTP request stops waiting. It does not interrupt the real client execution.
- The final client result still writes back through WebSocket.

## Result Return

Automa workflow results are commonly returned in two forms.

### Variable Result

The recommended workflow output variable is:

```text
browserflow_output
```

Skill calls should request:

```json
{
  "return_data": {
    "variables": ["browserflow_output"]
  }
}
```

This lets an LLM read `browserflow_output` first and avoids returning unrelated variables.

### Table Result

For table data:

```json
{
  "return_data": {
    "include_table": true,
    "table_limit": 100
  }
}
```

Large table results are saved as files and recorded in `task_record_files`. The execution record detail page can display and download related files.

## Page Map

| Page | Mode | Description |
| --- | --- | --- |
| `/` | Shared | Home page and runtime-mode entry |
| `/browser` | Windows | Manage local controlled browser instances |
| `/workflows` | Windows | View Browser Agent Automa workflows, open or run them, and export Skills |
| `/llm` | Windows | Configure LLM providers, models, API keys, and Base URLs |
| `/chat` | Windows | Chat with enabled local model configurations |
| `/browser-agent` | Windows | Local browser executor page, usually opened automatically by the backend |
| `/automa` | Server | Manage server workflow records, imports, sync, and Skill export |
| `/tasks` | Server | Create and maintain task definitions |
| `/task-records` | Server | View execution records, results, result files, and failure reasons |
| `/clients` | Server | View client online state, extension state, browser metadata, and ban state |
| `/client-agent` | Server | Windows client executor page |

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
| `app.mode` | Runtime mode: `windows` or `server` |
| `localStorage.path` | Local BoltDB file path for Windows mode |
| `frontend.url` | Frontend URL opened by backend-launched controlled browsers |
| `database.default.link` | PostgreSQL connection for Server mode |
| `redis.default` | Redis connection for Server mode |

## Requirements

- Go 1.25+
- Node.js and npm
- Chrome or Chromium
- Automa browser extension
- PostgreSQL and Redis for Server mode

## Project Structure

```text
browserflow/
|-- backend/                 GoFrame backend service
|-- frontend/                Vue 3 frontend app
|-- docs/                    Documentation and image assets
|-- third_party/automa/      Local Automa source snapshot and BrowserFlow changes
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

## Security And Deployment Notes

BrowserFlow can trigger real browser automation. When deploying Server mode in enterprise, government, public security, or other intranet environments, pay attention to:

- Protect backend and frontend entry points with Nginx or a gateway.
- Do not expose `/client-agent` to untrusted networks.
- Allow PostgreSQL and Redis access only from trusted servers.
- Add authentication, authorization, IP allowlists, or reverse-proxy access controls for WebSocket and management APIs.
- Protect sensitive task parameters, execution results, and result files.
- Keep audit logs for task creation, updates, execution, and deletion.
- Add alerts for important failures, timeouts, and client offline events.
- Isolate permissions by department, business system, and execution client where needed.

## License

Apache-2.0
