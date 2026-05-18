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
  简体中文 · <a href="./README_EN.md">English</a>
</p>

> **双模式浏览器自动化平台。** BrowserFlow 支持 Windows 本地浏览器控制和 Server 服务端任务调度，集成 Automa 工作流、浏览器/客户端执行端、任务下发、执行记录和大模型 Skill 调用能力。

```bash
# 启动后端
cd backend && go run .

# 启动前端
cd frontend && npm install && npm run dev

# 打开管理台
http://localhost:5173
```

## 项目定位

BrowserFlow 把 Automa 工作流从“只能在单个浏览器里手动点”扩展成一个可管理、可同步、可调度、可被大模型调用的自动化平台。

它有两种运行方式：

- **Windows 模式**：后端在本机启动受控浏览器，打开 Browser Agent，读取并执行当前浏览器中的 Automa 工作流。
- **Server 模式**：后端部署在服务器上作为调度中心，局域网或互联网内的 Windows 电脑打开 Client Agent 页面后接入，服务端统一管理客户端、工作流、任务和执行记录。

## 核心能力

- **双运行模式**：`windows` 本地桌面模式和 `server` 服务端调度模式共用一套前后端。
- **Automa Bridge**：通过页面事件和浏览器扩展通信，读取、导入、打开并触发 Automa 工作流。
- **Browser Agent**：Windows 模式下绑定后端启动的浏览器实例，通过 WebSocket 接收工作流命令。
- **Client Agent**：Server 模式下由任意 Windows 客户端打开网页接入，接收服务端任务并调用本机 Automa 执行。
- **工作流管理**：支持文件导入、客户端同步、同步候选对比、可同步开关和服务端工作流记录管理。
- **任务调度**：Server 模式下可把工作流、客户端、参数和 Cron 表达式组合成任务。
- **执行记录**：记录任务下发、客户端回执、成功/失败状态、参数和结果数据。
- **大模型 Skill**：可将工作流或任务封装为大模型可调用的 Skill，执行前校验后端、执行端、Automa 插件和参数。
- **模式感知路由**：前端根据后端运行模式自动隐藏或禁用不适用页面。

## 运行模式

| 能力 | Windows 模式 | Server 模式 |
| --- | --- | --- |
| 运行位置 | 本机 Windows | 后端在服务器，执行端在 Windows 客户端 |
| 执行端页面 | `/browser-agent` | `/client-agent` |
| 工作流来源 | 当前受控浏览器中的 Automa | 客户端同步到服务端，或服务端导入 |
| 执行方式 | 后端向指定 Browser Agent 下发命令 | 服务端向指定 Client Agent 下发任务 |
| 适合场景 | 单机自动化、本地调试、大模型本机工具 | 多客户端管理、远程任务调度、集中执行记录 |

### Windows 模式

Windows 模式适合在一台电脑上控制本地浏览器。

可用页面：

- 浏览器
- 工作流
- 大模型
- 对话
- Browser Agent

典型流程：

1. 后端启动本地浏览器实例。
2. 浏览器自动打开 `/browser-agent`。
3. Browser Agent 生成或接收 `Browser ID`，并与后端建立 WebSocket 连接。
4. 后端通过 Browser Agent 读取当前浏览器 Automa 工作流列表。
5. 用户可以打开、执行或导出工作流 Skill。

### Server 模式

Server 模式适合把后端部署到服务器，由多台 Windows 客户端接入执行任务。

可用页面：

- 工作流管理
- 任务配置
- 执行记录
- 客户端
- Client Agent

典型流程：

1. 服务器启动 BrowserFlow 后端和前端。
2. Windows 客户端在浏览器中打开管理站点的 `/client-agent` 页面。
3. Client Agent 生成稳定的 `client_*` 标识，注册到服务端 WebSocket。
4. 服务端记录客户端 IP、在线状态、浏览器信息、Automa 插件状态和版本。
5. 管理端从客户端同步 Automa 工作流到服务端，或手动导入工作流文件。
6. 管理端创建任务，绑定工作流、执行客户端、参数和可选 Cron 表达式。
7. 服务端下发任务到在线客户端，客户端调用本机 Automa 执行并回传结果。
8. 服务端更新执行记录，供管理端或大模型查询。

## Skill 使用约定

BrowserFlow 的 Skill 应该把“能不能执行”检查写清楚，避免大模型直接盲发请求。

执行工作流或任务前建议按顺序检查：

1. **后端是否启动**：请求 `/api/v1/app/runtime`，确认服务可访问且运行模式符合 Skill 目标。
2. **执行端是否在线**：Windows 模式检查目标 Browser Agent；Server 模式检查目标 Client Agent 或客户端 IP 是否在线。
3. **Automa 插件是否可用**：检查 `automa_installed`、插件状态或客户端上报的插件信息。
4. **参数是否完整**：如果工作流/任务需要变量，Skill 必须要求大模型提供对应参数。
5. **再发起执行请求**：Windows 模式可调用工作流运行接口；Server 模式更适合调用任务执行接口，由服务端调度到指定客户端。

当前任务执行回执表示命令已成功下发并由客户端回传处理结果。若要持续等待 Automa 工作流真正完成、失败或返回导出数据，需要进一步接入 Automa 执行状态事件或扩展回调。

## 快速开始

### 1. 配置运行模式

编辑 `backend/manifest/config/config.yaml`：

```yaml
app:
  mode: "windows"
```

可选值：

- `windows`：本地桌面模式
- `server`：服务端调度模式

### 2. 启动后端

```bash
cd backend
go run .
```

默认后端地址：

```text
http://localhost:8001
```

Swagger：

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

开发环境中，Vite 会把 `/api` 转发到 `http://localhost:8001`。

## 配置

主配置文件：

```text
backend/manifest/config/config.yaml
```

常用配置：

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

| 配置项 | 说明 |
| --- | --- |
| `server.address` | 后端 HTTP 服务监听地址 |
| `app.mode` | 运行模式，支持 `windows` 和 `server` |
| `localStorage.path` | 本地 bbolt 数据文件路径 |
| `frontend.url` | 后端启动受控浏览器时打开的前端地址 |
| `database.default.link` | Server 模式任务、客户端等 SQL 数据存储配置 |
| `redis.default` | Server 模式客户端在线状态和工作流清单缓存配置 |

## 环境要求

- Go 1.25+
- Node.js 和 npm
- Chrome 或 Chromium
- 使用 Automa 功能时，目标浏览器需要安装并启用 Automa 扩展
- Server 模式使用客户端管理、任务调度或工作流缓存时，需要可用的 PostgreSQL 和 Redis

## 关键页面

| 页面 | 模式 | 说明 |
| --- | --- | --- |
| `/browser` | Windows | 管理本地受控浏览器实例 |
| `/workflows` | Windows | 查看当前 Browser Agent 中的 Automa 工作流，并支持导出 Skill |
| `/llm` | Windows | 配置大模型提供商、模型、API Key 和 Base URL |
| `/chat` | Windows | 使用已启用模型进行本地对话 |
| `/browser-agent` | Windows | 浏览器执行端页面，通常由后端自动打开 |
| `/automa` | Server | 管理服务端工作流记录，支持导入和客户端同步 |
| `/tasks` | Server | 创建和维护任务配置 |
| `/task-records` | Server | 查看任务执行记录和结果 |
| `/clients` | Server | 查看客户端在线、插件、浏览器和拉黑状态 |
| `/client-agent` | Server | Windows 客户端执行端页面 |

## Automa Bridge 事件

| 事件 | 方向 | 说明 |
| --- | --- | --- |
| `__automa-ext__` | 前端到扩展 | 统一 bridge 请求入口 |
| `__automa-ext__get-workflows` | 扩展到前端 | 返回本地 Automa 工作流列表 |
| `__automa-ext__add-workflow` | 扩展到前端 | 返回工作流导入结果 |
| `automa:execute-workflow` | 前端到扩展 | 触发 Automa 执行工作流 |

如果页面已经打开后才安装 Automa 扩展，content script 通常不会自动注入到已打开页面。Agent 页面会通过刷新或重新检测处理这个场景。

## 项目结构

```text
browserflow/
├─ backend/                  GoFrame 后端服务
├─ frontend/                 Vue 3 前端应用
├─ docs/images/              README 和文档图片资源
├─ third_party/automa/        Automa 本地源码快照，保留源码，不含依赖
├─ third_party/browserwing/  BrowserWing 本地源码快照，不进仓库
├─ go.work                   Go workspace
├─ README.md
└─ README_EN.md
```

## 开发命令

前端：

```bash
cd frontend
npm run dev
npm run build
npm run lint
npm run format
```

后端：

```bash
cd backend
go run .
```

## 安全说明

如果把 Server 模式部署到局域网或互联网，请确认数据库、Redis、后端接口和前端访问入口都处在可信网络或反向代理保护下。Skill 调用会直接触发浏览器自动化任务，不建议在没有认证和访问控制的情况下暴露到公网。

## 开源协议

Apache-2.0
