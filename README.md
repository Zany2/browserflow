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
  English · <a href="./README_zh.md">简体中文</a>
</p>

BrowserFlow is a browser workflow automation platform that turns web-based business processes into reusable, schedulable, and auditable automation capabilities. It extends workflows that previously had to be managed and run manually inside a single browser into a complete flow with centralized management, remote dispatch, scheduled execution, execution records, and result return. BrowserFlow supports integration with Automa browser workflows and can export LLM Skills, so automation flows can be invoked by large language models and connected to broader automation systems.

The project supports two runtime modes:

- **Windows mode**: for local desktop automation. The backend launches and connects to a local controlled browser, while the frontend provides browser management, workflow lists, LLM configuration, and local chat. Server-side task, execution record, and client management pages are disabled. Business data is stored in the local BoltDB file configured by `localStorage.path`, without PostgreSQL or Redis.
- **Server mode**: for centralized multi-client dispatch. The server provides workflow management, task configuration, execution records, and client management. Local Windows browser, LLM configuration, and local chat pages are disabled. PostgreSQL is the source of truth for business data, while Redis keeps online client nodes, workflow inventory cache, task dispatch state, and per-client-node execution locks. Tasks can be dispatched to Windows clients that keep `/client-agent` open by automatic, IP, node, or group strategies.

It is suitable for browser-based business processes such as cross-system operations, internal admin systems, government or private-network business systems, data lookup, information collection, form processing, and report export. It is also useful when browser flows need to be managed centrally, dispatched to specific clients, recorded for audit, or connected to LLM-driven automation.

## Core Value

BrowserFlow turns browser-based business processes into automation capabilities that can be managed, dispatched, executed, recorded, and integrated.

- **Capture browser business flows**: turn web operations such as lookup, form filling, collection, export, and inspection into reusable workflow capabilities instead of temporary scripts or manual steps.
- **Manage workflow assets centrally**: support workflow import, sync, comparison, protection, and version metadata, making it easier to keep flows consistent across teams or multiple clients.
- **Dispatch execution clients centrally**: Server mode can connect multiple Windows clients as execution nodes and dispatch tasks by automatic, IP, node, or group strategies.
- **Keep execution traces**: task runs create records with trigger type, execution client, parameters, status, failure reason, result JSON, and result files.
- **Avoid browser concurrency conflicts**: one client node runs only one browser workflow at a time, avoiding tab switching, variable changes, or context contention across multiple flows.
- **Support scheduled tasks**: tasks can use Cron expressions, and the server synchronizes database task configuration into GoFrame gcron.
- **Connect LLMs and external automation**: workflows and tasks can be exported as Skills, allowing LLMs or other automation systems to trigger browser flows and read results.

## Runtime Modes

| Area | Windows Mode | Server Mode |
| --- | --- | --- |
| Purpose | Local desktop automation | Centralized multi-client dispatch |
| Storage | BoltDB | PostgreSQL + Redis |
| Backend location | Local Windows machine | Server |
| Executor | Local controlled browser launched by backend | Windows client nodes opened on `/client-agent` |
| Executor page | `/browser-agent` | `/client-agent` |
| Main pages | Browser, Workflows, LLM, Chat | Workflow Management, Task Configuration, Execution Records, Clients |
| Workflow source | Workflows in the current controlled browser | Synced from clients or imported into the server |
| Dispatch method | Run locally | Automatic, IP, node, or group dispatch, with Cron support |
| Execution records | Local execution viewing | Centralized records, queries, and result file management |
| Best for | Local tools, personal automation, debugging, local LLM calls | Team collaboration, multi-client execution, centralized dispatch, unified audit |

## Quick Start

### Windows Mode

Windows mode targets local browser automation and is suitable for personal use, workflow debugging, and local LLM calls. For now, use a source build flow: build the frontend `dist`, embed `dist` as static assets into the backend Windows executable, and then run the executable to open the local console.

Before first use, prepare:

- Windows 10/11.
- Go 1.25+.
- Node.js and npm.
- Chrome or Chromium.
- Automa browser extension.

#### 1. Build The Frontend

```bash
cd frontend
npm install
npm run build
```

This generates `frontend/dist`.

#### 2. Embed dist Into The Backend Executable

The backend embeds static assets from `backend/internal/web/dist`. After building the frontend, move the generated files into that directory, then build the backend executable:

```bash
cd ..
mkdir -p backend/internal/web/dist
mv frontend/dist/* backend/internal/web/dist/
cd backend
go build -o BrowserFlow.exe .
```

On Windows PowerShell, use:

```powershell
New-Item -ItemType Directory -Force backend\internal\web\dist | Out-Null
Move-Item frontend\dist\* backend\internal\web\dist\
Set-Location backend
go build -o BrowserFlow.exe .
```

For cross-compiling a Windows executable from a non-Windows environment:

```bash
GOOS=windows GOARCH=amd64 go build -o BrowserFlow.exe .
```

#### 3. Start Windows Mode

```bash
./BrowserFlow.exe --port 8001
```

You can omit `--port` to use the configured port. If `--port` is provided on first startup, the generated Windows config file will use that port. Then open:

```text
http://127.0.0.1:8001
```

Basic usage flow:

1. Start BrowserFlow in Windows mode.
2. Open, or wait for it to open, the local console.
3. Start a controlled browser from the Browser page.
4. Prepare the browser workflow to execute inside the controlled browser.
5. Return to the Workflows page to view, run, or export a Skill.
6. For LLM calls, configure a model on the LLM page, then invoke it from the Chat page.

### Server Mode

Server mode is recommended to be deployed on a server. It manages workflows, task definitions, and execution records in one place, and dispatches browser automation tasks to multiple Windows clients as execution nodes.

Server mode depends on PostgreSQL and Redis. Windows clients connect to the server by opening `/client-agent`, then become dispatchable execution nodes.

After startup, the basic flow is: connect Windows clients, sync or import workflows, create tasks, run them manually or with Cron scheduling, and view results in execution records.

#### Docker Compose Deployment

Server mode is recommended to run on a Linux server with `docker compose` starting BrowserFlow, PostgreSQL, Redis, and Nginx together. The current deployment flow builds the frontend and backend on the server first, then `docker-compose.yaml` mounts the compiled backend binary and Server configuration.

Prepare the server with:

- Linux server.
- Git.
- Go 1.25+.
- Node.js and npm.
- Docker and Docker Compose.

Clone the source:

```bash
git clone https://github.com/Zany2/browserflow.git
cd browserflow
```

Build the frontend and move the output into the backend embed directory:

```bash
cd frontend
npm install
npm run build

cd ..
mkdir -p backend/internal/web/dist
mv frontend/dist/* backend/internal/web/dist/
```

Build the Linux backend executable:

```bash
cd backend
go build -o browserflow .
cd ..
```

Start Server mode:

```bash
docker compose up -d
```

Default access URL:

```text
http://SERVER_IP:8001
```

The current `docker-compose.yaml` starts:

- `postgres`: PostgreSQL, loading `backend/sql/public.sql` on first startup to initialize tables.
- `redis`: Redis with AOF persistence enabled and `browserflow` as the example password.
- `browserflow`: BrowserFlow Server, mounting the `backend/browserflow` executable and `deploy/server/config.yaml`.
- `nginx`: the public HTTP entry point, proxying frontend pages, APIs, and WebSocket traffic.

Deployment files:

- `docker-compose.yaml`: Server mode container orchestration.
- `deploy/server/config.yaml`: Server backend configuration for PostgreSQL, Redis, logs, and WebSocket.
- `deploy/nginx/browserflow.conf`: Nginx reverse proxy configuration.

After changing frontend or backend code, rebuild the frontend, move `dist`, rebuild `backend/browserflow`, then restart services:

```bash
docker compose restart browserflow nginx
```

Stop services:

```bash
docker compose down
```

Before production deployment, update the PostgreSQL password, Redis password, `frontend.url`, Nginx `server_name`, and related settings in `docker-compose.yaml` and `deploy/server/config.yaml`.

#### Source Run

Running from source is better suited for development, testing, and custom deployments. Use `deploy/server/config.yaml` as a reference for Server mode, PostgreSQL, and Redis configuration:

```bash
git clone https://github.com/Zany2/browserflow.git
cd browserflow
```

Configure Server mode, PostgreSQL, and Redis, then start the backend:

```bash
cd backend
go run .
```

Start the frontend:

```bash
cd frontend
npm install
npm run dev
```

## Project Structure

```text
browserflow/
|-- backend/                         GoFrame backend service
|   |-- api/                         GoFrame API request/response definitions
|   |-- internal/                    Controllers, service registration, models, and business logic
|   |   `-- web/                     Embedded frontend static asset entry
|   |-- manifest/config/             Backend configuration files
|   |-- middleware/                  HTTP middleware
|   |-- sql/                         Database initialization and schema scripts
|   `-- utility/                     Browser execution, workflows, task dispatch, WebSocket utilities
|-- frontend/                        Vue 3 + Vite frontend app
|   |-- src/api/                     Frontend request wrappers
|   |-- src/components/              Shared components
|   |-- src/composables/             Composable logic
|   |-- src/layouts/                 Page layouts
|   |-- src/router/                  Frontend routes and runtime-mode guards
|   |-- src/services/                Business APIs, WebSocket, and frontend services
|   |-- src/styles/                  Global styles and shared page classes
|   |-- src/utils/                   Shared utility functions
|   `-- src/views/                   Page views
|-- windows-worker/                  Windows client/worker executable source and resources
|   |-- cmd/                         Worker entrypoint
|   |-- internal/                    Worker internals
|   |-- assets/                      Icons and static assets
|   |-- chrome-packages/             Browser-related package resources
|   `-- winres/                      Windows executable resource configuration
|-- docs/                            Project documentation and image assets
|-- docs/images/                     README and documentation images
|-- deploy/                          Server mode deployment configuration
|   |-- nginx/                       Nginx reverse proxy configuration
|   `-- server/                      Server backend configuration
|-- third_party/automa/              Local Automa source snapshot and BrowserFlow changes
|-- workflows/                       Example Automa workflow files
|-- .agents/                         Project-local coding-agent skills and configuration
|-- docker-compose.yaml              Server mode Docker Compose deployment file
|-- go.work                          Go workspace
|-- LICENSE                          Open-source license
|-- README.md                        English README
`-- README_zh.md                     Chinese README
```

## Security And Deployment Notes

BrowserFlow can trigger real browser automation and may process sensitive information such as accounts, business parameters, query results, and exported files. Before deploying to production, consider at least the following:

- **Protect service entry points**: expose services through Nginx, a gateway, or a reverse proxy, and add authentication, authorization, IP allowlists, or SSO for management pages, APIs, and WebSocket connections.
- **Control client entry**: do not expose `/client-agent` to untrusted users or networks. Only authorized Windows clients should be allowed to join as execution nodes.
- **Isolate database and Redis**: PostgreSQL and Redis should not be exposed directly to the public internet. Allow access only from the BrowserFlow server and trusted operations networks.
- **Manage task permissions**: isolate permissions by department, business system, or execution node to prevent users from creating or running browser tasks beyond their authorization.
- **Protect sensitive data**: task parameters, Cookies, Tokens, account passwords, execution results, and result files may contain sensitive data. Apply encryption, masking, access control, and retention policies according to business requirements.
- **Secure execution nodes**: in Server mode, Windows clients perform real browser actions. Use controlled accounts, dedicated machines, or isolated environments, and avoid mixing them with personal office browsers.
- **Protect concurrency and locks**: one client node should run only one browser workflow at a time to avoid tasks switching tabs, modifying variables, or competing for browser context.
- **Audit and trace**: keep records for task creation, updates, enable/disable changes, execution, deletion, failure reasons, execution clients, and operators for later audit and troubleshooting.
- **Alert and recover**: configure alerts and recovery procedures for critical task failures, execution timeouts, client offline events, Redis lock anomalies, and database connection failures.
- **Manage production configuration**: do not commit real database passwords, Redis passwords, API keys, or business accounts to the repository. Prefer environment variables, secret management services, or deployment-platform configuration injection.
- **Network and certificates**: enable HTTPS/WSS in production and configure trusted certificates, proxy timeouts, and WebSocket long-connection parameters for the deployment environment.
- **Verify before release**: validate task impact scope, rollback behavior, data export locations, result file permissions, and recovery after client disconnection in an isolated environment before going live.

## Star History

<picture>
  <source
    media="(prefers-color-scheme: dark)"
    srcset="https://api.star-history.com/svg?repos=Zany2/browserflow&type=Date&theme=dark"
  />
  <source
    media="(prefers-color-scheme: light)"
    srcset="https://api.star-history.com/svg?repos=Zany2/browserflow&type=Date"
  />
  <img
    alt="Star History Chart"
    src="https://api.star-history.com/svg?repos=Zany2/browserflow&type=Date"
  />
</picture>

## License

BrowserFlow is open source under the [Apache-2.0](LICENSE) license.
