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

BrowserFlow 是一个双模式浏览器自动化平台，把 Automa 工作流扩展为可管理、可同步、可调度、可记录、可被大模型调用的自动化能力。

它支持两种运行模式：

- **Windows 模式**：面向个人本地使用。后端、前端和受控浏览器运行在同一台 Windows 电脑上，业务数据存储在本地 BoltDB 文件中。
- **Server 模式**：面向服务器部署和多客户端调度。后端作为任务调度中心运行在服务器上，业务数据使用 PostgreSQL，在线状态、工作流清单和客户端任务锁使用 Redis。

## 快速开始

### 1. 配置运行模式

编辑 `backend/manifest/config/config.yaml`：

```yaml
app:
  mode: "windows"
```

可选值：

- `windows`：本地个人模式，只依赖 BoltDB。
- `server`：服务端调度模式，依赖 PostgreSQL 和 Redis。

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

## 核心能力

- **双运行模式**：同一套前后端代码支持 Windows 本地模式和 Server 调度模式。
- **Automa Bridge**：通过页面事件和浏览器扩展通信，读取、导入、打开并触发 Automa 工作流。
- **Browser Agent**：Windows 模式下连接后端启动的受控浏览器，通过 WebSocket 接收本地工作流命令。
- **Client Agent**：Server 模式下由 Windows 客户端打开网页接入调度中心，接收任务并调用本机 Automa 执行。
- **工作流管理**：支持文件导入、客户端同步、同步候选对比、保护状态、可同步状态和服务端工作流记录管理。
- **任务配置**：Server 模式下把工作流、客户端、参数和 Cron 表达式组合成可手动或定时执行的任务。
- **任务调度**：Server 模式下支持指定客户端执行，也支持未指定客户端时遍历在线且拥有目标工作流的客户端。
- **执行记录**：记录任务下发、执行中、成功、失败、超时、参数、返回数据和失败原因。
- **客户端任务锁**：Server 模式下使用 Redis 对每个客户端加锁，避免同一个客户端同时执行多个 Automa 工作流。
- **大模型能力**：Windows 模式下支持本地大模型配置、对话和把工作流导出为可调用 Skill。
- **模式感知路由**：前端根据后端运行模式自动屏蔽不适用页面。

## 运行模式边界

| 项目 | Windows 模式 | Server 模式 |
| --- | --- | --- |
| 定位 | 单机个人使用 | 服务器任务调度中心 |
| 主要数据源 | BoltDB，本地文件由 `localStorage.path` 配置 | PostgreSQL 和 Redis |
| 后端部署位置 | Windows 本机 | 服务器 |
| 执行端 | 后端启动的本地受控浏览器 | 访问 `/client-agent` 的 Windows 客户端 |
| 执行端页面 | `/browser-agent` | `/client-agent` |
| 工作流来源 | 当前受控浏览器中的 Automa 工作流 | 客户端同步到服务端，或服务端导入 |
| 任务调度 | 本地工作流运行 | Server 任务配置、Cron 调度和执行记录 |
| 适合场景 | 本地自动化、本地调试、个人大模型工具 | 多客户端管理、远程调度、集中记录 |

### Windows 模式

Windows 模式适合把 BrowserFlow 打包成一个本地可执行程序，前端资源可以随可执行文件一起运行在本机。

典型流程：

1. 后端启动本地受控浏览器实例。
2. 浏览器自动打开 `/browser-agent`。
3. Browser Agent 获取或生成 `Browser ID`，并与后端建立 WebSocket 连接。
4. 后端通过 Browser Agent 读取当前浏览器中的 Automa 工作流。
5. 用户可以打开、执行、导出工作流，也可以配置大模型并进行本地对话。

Windows 模式下不应依赖 PostgreSQL 和 Redis；本地业务数据使用 `localStorage.path` 指向的 BoltDB 文件。

### Server 模式

Server 模式适合把 BrowserFlow 部署在服务器上，作为多个 Windows 客户端的任务调度中心。

典型流程：

1. 服务器启动 BrowserFlow 后端和前端。
2. Windows 客户端访问站点并打开 `/client-agent`。
3. Client Agent 生成稳定的 `client_*` 标识，并通过 WebSocket 注册到服务端。
4. 服务端记录客户端 IP、在线状态、浏览器信息、Automa 插件状态和工作流清单。
5. 管理端从客户端同步 Automa 工作流，或手动导入工作流文件。
6. 管理端创建任务，绑定工作流、客户端、参数和可选 Cron 表达式。
7. 服务端把任务下发给在线客户端，客户端调用本机 Automa 执行并回传状态和结果。
8. 服务端写入执行记录，供页面查询或大模型调用链路使用。

Server 模式下业务数据以 PostgreSQL 为准；Redis 用于在线客户端、工作流清单缓存和客户端任务锁。

## Server 调度规则

- 如果任务配置了客户端 IP，只调度到该客户端。
- 如果目标客户端离线、没有目标工作流或正处于繁忙锁定状态，本次执行会产生失败记录，并写入明确失败原因。
- 如果任务没有配置客户端 IP，服务端会遍历在线且拥有该工作流的客户端，选择未加锁的客户端执行。
- 如果所有候选客户端都繁忙，本次执行会产生失败记录，失败原因类似：`已遍历调度所有在线且拥有工作流的客户端，均处于繁忙状态，任务执行失败`。
- Cron 表达式用于定时触发任务。调度触发时不会等待上一轮自然结束后再重新计时，而是按 Cron 时间点触发；客户端锁负责拦截同一客户端上的并发执行。
- 客户端任务锁有 TTL 兜底，避免客户端断线、服务端异常或回调丢失导致永久锁死。
- 后台会扫描超时的执行记录，将长时间没有完成回调的执行标记为失败或超时，并释放对应客户端锁。

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
- 使用 Automa 相关能力时，目标浏览器需要安装并启用 Automa 扩展
- Server 模式需要可用的 PostgreSQL 和 Redis

## 页面地图

| 页面 | 模式 | 说明 |
| --- | --- | --- |
| `/` | 通用 | 首页和运行模式入口 |
| `/browser` | Windows | 管理本地受控浏览器实例 |
| `/workflows` | Windows | 查看 Browser Agent 中的 Automa 工作流，支持打开、执行和导出 Skill |
| `/llm` | Windows | 配置大模型提供商、模型、API Key 和 Base URL |
| `/chat` | Windows | 使用已启用模型进行本地对话 |
| `/browser-agent` | Windows | 本地浏览器执行端页面，通常由后端自动打开 |
| `/automa` | Server | 管理服务端工作流记录，支持导入和客户端同步 |
| `/tasks` | Server | 创建和维护任务配置 |
| `/task-records` | Server | 查看任务执行记录、结果和失败原因 |
| `/clients` | Server | 查看客户端在线、插件、浏览器和拉黑状态 |
| `/client-agent` | Server | Windows 客户端执行端页面 |

## Automa Bridge 事件

| 事件 | 方向 | 说明 |
| --- | --- | --- |
| `__automa-ext__` | 前端到扩展 | 统一 bridge 请求入口 |
| `__automa-ext__get-workflows` | 扩展到前端 | 返回本地 Automa 工作流列表 |
| `__automa-ext__add-workflow` | 扩展到前端 | 返回工作流导入结果 |
| `automa:execute-workflow` | 前端到扩展 | 触发 Automa 执行工作流 |

如果页面已经打开后才安装 Automa 扩展，content script 通常不会自动注入到已打开页面。遇到这种情况，请刷新 Agent 页面或重新打开受控浏览器。

## 项目结构

```text
browserflow/
|-- backend/                 GoFrame 后端服务
|-- frontend/                Vue 3 前端应用
|-- docs/                    项目文档和 README 图片资源
|-- third_party/automa/      Automa 本地源码快照
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

## 安全提醒

Server 模式如果部署到局域网或公网，请确保 PostgreSQL、Redis、后端接口、前端入口和 `/client-agent` 访问入口处在可信网络或反向代理保护下。BrowserFlow 可以触发真实浏览器自动化任务，不建议在没有认证、授权和网络隔离的情况下直接暴露到公网。

## License

Apache-2.0
