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
  简体中文 · <a href="./README_EN.md">English</a>
</p>

BrowserFlow 是一个围绕 Automa 浏览器工作流构建的自动化平台。它把原本只能在单个浏览器里手动管理和运行的工作流，扩展成可以集中管理、远程调度、定时执行、记录审计、结果回传，并且可以被大模型 Skill 调用的能力。

项目支持两种运行模式：

- **Windows 模式**：面向个人、本机、单机自动化。前端、后端、受控浏览器运行在同一台 Windows 电脑上，业务数据使用本地 BoltDB。
- **Server 模式**：面向内网服务器和多 Windows 客户端调度。服务端使用 PostgreSQL 保存业务数据，使用 Redis 维护客户端在线状态、工作流清单缓存和客户端执行锁。

它适合处理那些“系统没有接口、只能通过浏览器操作、流程重复、需要集中记录”的场景，例如公司内部系统、政务内网系统、专网业务系统、数据查询和报表导出流程等。

## 目录

- [核心价值](#核心价值)
- [运行模式](#运行模式)
- [快速开始](#快速开始)
- [Server 模式完整流程](#server-模式完整流程)
- [任务调度逻辑](#任务调度逻辑)
- [Skill 调用逻辑](#skill-调用逻辑)
- [结果回传](#结果回传)
- [页面地图](#页面地图)
- [配置说明](#配置说明)
- [项目结构](#项目结构)
- [开发命令](#开发命令)
- [安全和部署建议](#安全和部署建议)

## 核心价值

BrowserFlow 的核心不是替代 Automa，而是把 Automa 变成一个可管理、可调度、可审计的自动化执行能力。

- **解决无 API 系统自动化**：很多内网系统、老系统和专网平台没有开放接口，BrowserFlow 可以通过浏览器工作流完成查询、填报、导出、巡检等操作。
- **集中调度多台 Windows 客户端**：某些系统只能在特定电脑、特定网络、特定证书或特定浏览器环境访问，Server 模式可以把这些电脑接入为执行节点。
- **保留执行痕迹**：任务执行会形成记录，包含触发方式、执行客户端、参数、状态、失败原因、结果 JSON 和结果文件。
- **避免浏览器并发污染**：同一个客户端一次只执行一个 Automa 工作流，避免多个工作流同时切换标签页、改变量、抢上下文。
- **支持定时任务**：任务可以配置 Cron 表达式，服务端把数据库任务配置同步到 GoFrame gcron。
- **支持大模型调用**：Windows 和 Server 模式都可以导出 Automa Skill，让大模型按说明调用工作流或任务 API。
- **内网部署友好**：Server 模式可以部署在内网，客户端只需要打开 `/client-agent` 页面并保持连接。

## 运行模式

| 项目 | Windows 模式 | Server 模式 |
| --- | --- | --- |
| 主要定位 | 本机个人自动化 | 内网任务调度中心 |
| 数据存储 | BoltDB | PostgreSQL + Redis |
| 后端位置 | Windows 本机 | 服务器 |
| 执行端 | 后端启动的本地受控浏览器 | 多台 Windows 客户端 |
| 执行端页面 | `/browser-agent` | `/client-agent` |
| 工作流来源 | 当前受控浏览器里的 Automa 工作流 | 客户端同步或服务端导入 |
| 任务调度 | 本地运行工作流 | 任务配置、Cron 调度、客户端分发 |
| 执行记录 | 偏本地使用 | 集中记录和查询 |
| 适合场景 | 本机工具、个人自动化、调试、LLM 本地调用 | 公司、政务、公安等内网多客户端调度 |

### Windows 模式

Windows 模式适合把 BrowserFlow 打包成一个本地可执行程序使用。前端可以被打包进可执行文件，由后端在本机提供页面服务。

典型流程：

1. 后端启动本地受控浏览器。
2. 浏览器自动打开 `/browser-agent`。
3. Browser Agent 通过 WebSocket 连接后端。
4. 后端通过 Browser Agent 读取当前浏览器里的 Automa 工作流。
5. 用户可以打开、执行、导出工作流 Skill，也可以配置本地大模型并通过聊天调用能力。

边界要求：

- Windows 模式不应依赖 PostgreSQL 或 Redis。
- Windows 模式本地业务数据只使用 `localStorage.path` 指向的 BoltDB 文件。
- Windows 模式更适合个人使用、单机运行和工作流调试。

### Server 模式

Server 模式适合部署在内网服务器上，作为多个 Windows 客户端的任务调度中心。

典型流程：

1. 服务器启动 BrowserFlow 后端和前端。
2. Windows 客户端访问站点并打开 `/client-agent`。
3. Client Agent 生成稳定的客户端标识，通过 WebSocket 注册到服务端。
4. 服务端记录客户端 IP、在线状态、浏览器信息、Automa 插件状态和工作流清单。
5. 管理员从客户端同步工作流到服务端，或在服务端导入工作流文件。
6. 管理员创建任务，绑定工作流、参数、可选客户端和可选 Cron 表达式。
7. 任务手动触发、Cron 触发或 Skill 触发后，服务端选择客户端并下发执行命令。
8. 客户端调用本机 Automa 执行工作流，并通过 WebSocket 回传状态和结果。
9. 服务端写入执行记录和结果文件，供页面查询或大模型后续读取。

边界要求：

- Server 模式业务数据以 PostgreSQL 为准。
- Redis 用于客户端在线状态、客户端工作流清单缓存、客户端执行锁。
- Server 模式不应使用 BoltDB 作为任务、工作流、客户端同步和执行记录的数据源。

## 快速开始

### 1. 配置运行模式

编辑：

```text
backend/manifest/config/config.yaml
```

Windows 模式：

```yaml
app:
  mode: "windows"
```

Server 模式：

```yaml
app:
  mode: "server"
```

### 2. 启动后端

```bash
cd backend
go run .
```

默认后端地址：

```text
http://localhost:8001
```

接口文档：

```text
http://localhost:8001/swagger
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

默认前端地址：

```text
http://localhost:5173
```

开发环境中，Vite 会把 `/api` 请求代理到 `http://localhost:8001`。

## Server 模式完整流程

### 1. 客户端接入

Windows 客户端打开：

```text
http://<server-host>/client-agent
```

Client Agent 会：

- 建立 WebSocket 连接。
- 上报客户端标识、IP、浏览器信息和 Automa 插件状态。
- 定期发送心跳。
- 上报本机 Automa 工作流清单。
- 接收 `task.execute` 命令并调用本机 Automa 执行。

### 2. 工作流管理

Server 模式下工作流存储在 PostgreSQL 中。来源包括：

- 从在线客户端同步 Automa 工作流。
- 手动导入 Automa 工作流文件。
- 对比客户端工作流和服务端工作流，判断是否已同步、是否有更新。
- 可设置保护状态，避免服务端工作流被客户端覆盖。
- 可导出 Server 模式 Skill，供大模型调用任务 API。

### 3. 任务配置

任务配置包含：

- 任务名称和说明。
- Automa 工作流 ID。
- 可选客户端 IP 或客户端 ID。
- 执行参数 `params`。
- Cron 表达式。
- 启用状态。
- 创建后是否立即执行一次。

如果任务没有配置客户端 IP，Server 模式执行时会自动查找在线且拥有该工作流的客户端。

### 4. 执行记录

每次执行都会写入 `task_records`：

- `trigger_type`：`manual`、`cron`、`task_create`、`skill`、`system`。
- `status`：`pending`、`queued`、`running`、`success`、`failed`。
- `client_ip`：实际执行客户端。
- `params_json`：本次执行参数。
- `result_json`：执行结果。
- `error_message`：失败原因。
- `started_at`、`finished_at`：执行时间。

表格类大结果会保存到 `task_record_files`，页面详情中可以查看和下载。

## 任务调度逻辑

### 手动执行

页面或 API 调用：

```text
POST /api/v1/tasks/{id}/execute
```

后端会读取任务配置，解析参数，选择客户端，加锁，下发 WebSocket 命令，并创建执行记录。

### Cron 执行

任务配置了 Cron 表达式并启用后，后台调度器会把任务注册到 GoFrame `gcron`。

调度同步逻辑：

- 程序启动时同步一次数据库任务配置到 gcron。
- 后台每 30 秒同步一次任务配置。
- 同步时按 `id` 游标分页读取，每批 500 条。
- 只读取 `id` 和 `cron_expression`，避免一次性加载完整任务记录。
- 同步结果会和当前 gcron 任务做对比，新增、删除或更新对应 job。

执行策略：

- 项目使用 `gcron.AddSingleton` 注册任务。
- 同一个 cron job 上一次还没结束时，下一次命中会被跳过，不排队、不并发。
- Cron 到点只代表触发执行，真正是否能执行还要经过客户端在线、工作流拥有关系和 Redis 锁检查。

### 客户端选择

如果任务指定了客户端：

- 只调度这个客户端。
- 客户端离线、没有该工作流或繁忙，会创建失败记录并写入原因。

如果任务没有指定客户端：

- 服务端查询 Redis 中拥有目标工作流的在线客户端。
- 遍历候选客户端，尝试获取客户端锁。
- 找到第一个未繁忙客户端后下发执行。
- 如果所有候选客户端都繁忙，会创建失败记录，原因类似：

```text
已遍历调度所有在线且拥有工作流的客户端，均处于繁忙状态，任务执行失败。
```

### 客户端锁

Server 模式使用 Redis 对每个客户端加执行锁：

- 锁粒度是客户端。
- 同一客户端同一时间只执行一个 Automa 工作流。
- 锁包含任务 ID、执行记录 ID、工作流 ID、命令 ID。
- 客户端心跳会续期锁。
- 任务成功或失败后释放锁。
- 后台兜底清理会处理长时间未完成的执行记录，避免永久锁死。

## Skill 调用逻辑

BrowserFlow 支持导出两类 Skill。

### Windows 模式工作流 Skill

Windows 模式下，Skill 面向当前 Browser Agent 中可用的 Automa 工作流。

调用方式：

- 运行工作流：

```text
POST /api/v1/workflows/{workflow_id}/run
```

- 支持 `wait_result`。
- 支持 `return_data`。
- 适合本机大模型直接调用本机浏览器工作流。

### Server 模式工作流 Skill

Server 模式下，Skill 面向数据库中已保存的 Automa 工作流。Skill 不直接操作浏览器，而是调用任务 API。

推荐方式：

1. 为常用工作流创建可复用任务。
2. Skill 调用任务执行接口。
3. 使用 `trigger_type: "skill"` 方便执行记录筛选。

执行已有任务：

```bash
curl -X POST 'http://localhost:8001/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{"trigger_type":"skill","client_ip":"","params":{}}'
```

执行并等待结果：

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

说明：

- `wait_result=false`：API 在任务下发成功后返回执行记录。
- `wait_result=true`：API 等待客户端最终回传结果，返回执行记录和 `result`。
- HTTP 等待超时只代表当前请求不再等待，不会中断客户端真实执行。
- 客户端最终结果仍会通过 WebSocket 写回执行记录。

## 结果回传

Automa 工作流结果分为两类。

### 变量结果

推荐工作流把业务结果写入变量：

```text
browserflow_output
```

Skill 调用时建议请求：

```json
{
  "return_data": {
    "variables": ["browserflow_output"]
  }
}
```

这样大模型可以优先读取 `browserflow_output`，避免返回无关变量。

### 表格结果

如果需要返回表格数据：

```json
{
  "return_data": {
    "include_table": true,
    "table_limit": 100
  }
}
```

大表格结果会被保存为文件，并记录到 `task_record_files`。执行记录详情页面可以查看关联文件并下载。

## 页面地图

| 页面 | 模式 | 说明 |
| --- | --- | --- |
| `/` | 通用 | 首页和运行模式入口 |
| `/browser` | Windows | 管理本地受控浏览器实例 |
| `/workflows` | Windows | 查看 Browser Agent 中的 Automa 工作流，支持打开、运行和导出 Skill |
| `/llm` | Windows | 配置大模型提供商、模型、API Key 和 Base URL |
| `/chat` | Windows | 使用已启用模型进行本地对话 |
| `/browser-agent` | Windows | 本地浏览器执行端页面，通常由后端自动打开 |
| `/automa` | Server | 管理服务端工作流记录，支持导入、同步、导出 Skill |
| `/tasks` | Server | 创建和维护任务配置 |
| `/task-records` | Server | 查看任务执行记录、结果、结果文件和失败原因 |
| `/clients` | Server | 查看客户端在线状态、插件状态、浏览器信息和拉黑状态 |
| `/client-agent` | Server | Windows 客户端执行端页面 |

## 配置说明

主配置文件：

```text
backend/manifest/config/config.yaml
```

常用配置：

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

| 配置项 | 说明 |
| --- | --- |
| `server.address` | 后端 HTTP 服务监听地址 |
| `app.mode` | 运行模式，支持 `windows` 和 `server` |
| `localStorage.path` | Windows 模式本地 BoltDB 文件路径 |
| `frontend.url` | 后端启动受控浏览器时打开的前端地址 |
| `database.default.link` | Server 模式 PostgreSQL 连接配置 |
| `redis.default` | Server 模式 Redis 连接配置 |

## 环境要求

- Go 1.25+
- Node.js 和 npm
- Chrome 或 Chromium
- Automa 浏览器扩展
- Server 模式需要 PostgreSQL 和 Redis

## 项目结构

```text
browserflow/
|-- backend/                 GoFrame 后端服务
|-- frontend/                Vue 3 前端应用
|-- docs/                    项目文档和图片资源
|-- third_party/automa/      Automa 本地源码快照和 BrowserFlow 本地改造
|-- go.work                  Go workspace
|-- README.md                中文说明
`-- README_EN.md             English README
```

## 开发命令

后端：

```bash
cd backend
go run .
```

前端：

```bash
cd frontend
npm run dev
npm run build
npm run lint
npm run format
```

数据库初始化或更新：

```text
backend/sql/browserflow.sql
```

## 安全和部署建议

BrowserFlow 可以触发真实浏览器自动化操作。Server 模式部署到公司、政务、公安等内网环境时，建议重点关注：

- 使用 Nginx 或网关保护后端和前端入口。
- 不要把 `/client-agent` 暴露到不可信网络。
- PostgreSQL 和 Redis 只允许可信服务器访问。
- 为 WebSocket 和管理 API 增加认证、授权、IP 白名单或反向代理访问控制。
- 对任务参数、执行结果、结果文件做权限控制和敏感信息保护。
- 保留任务创建、修改、执行、删除等审计日志。
- 对重要任务配置失败告警、超时告警和客户端离线告警。
- 对不同部门、不同业务系统、不同客户端执行节点做权限隔离。

## License

Apache-2.0
