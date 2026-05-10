# <img src="docs/images/layout-logo.png" width="32" height="32" alt="BrowserFlow logo"> BrowserFlow

BrowserFlow 是一个围绕浏览器自动化、Automa 工作流、任务调度和本地大模型对话构建的管理工具。项目分为 Go 后端和 Vue 前端，支持两种运行模式：

- `windows`：本地桌面模式，侧重本机浏览器管理、Automa 工作流读取、大模型配置和对话。
- `server`：服务端调度模式，侧重客户端管理、工作流同步、任务下发和任务记录。

## 目录结构

```text
browserflow/
├─ backend/              GoFrame 后端服务
├─ frontend/             Vue 3 + Element Plus 前端
├─ automa/               Automa 本地源码快照，默认不纳入 Git
├─ browserwing/          BrowserWing 本地源码快照，默认不纳入 Git
├─ AUTOMA.md             Automa 本地改动说明
└─ README.md
```

## 技术栈

- 后端：Go、GoFrame、bbolt、Rod、Gorilla WebSocket、PostgreSQL、Redis
- 前端：Vue 3、Vite、Vue Router、Element Plus、Sass
- 浏览器自动化：Rod + Automa 扩展桥接

## 运行模式

运行模式由 `backend/manifest/config/config.yaml` 中的 `app.mode` 控制。

```yaml
app:
  mode: "windows"
```

### windows 模式

可用页面主要包括：

- 浏览器：管理本地/远程浏览器实例，并观察 Browser Agent 状态。
- 工作流：通过当前浏览器的 Automa bridge 读取本地 Automa 工作流。
- 大模型：管理本地大模型配置。
- 对话：基于已启用的大模型配置创建和管理对话会话。

禁用页面包括：工作流管理、任务、任务记录、客户端、客户端 Agent。

### server 模式

可用页面主要包括：

- 工作流管理：管理后端保存的 Automa 工作流。
- 任务：创建和下发任务。
- 任务记录：查看任务执行结果。
- 客户端：管理在线客户端。
- Client Agent：客户端执行端页面。

禁用页面包括：浏览器、大模型、对话、Browser Agent。

## 本地启动

### 1. 启动后端

```bash
cd backend
go run .
```

默认后端地址：

```text
http://localhost:8001
```

Swagger 地址：

```text
http://localhost:8001/swagger
```

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

默认前端地址：

```text
http://localhost:5173
```

前端开发服务会把 `/api` 代理到 `http://localhost:8001`。

## 常用配置

配置文件位置：

```text
backend/manifest/config/config.yaml
```

关键配置：

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

- `server.address`：后端监听地址。
- `app.mode`：运行模式，支持 `windows` 和 `server`。
- `localStorage.path`：本地 bbolt 数据文件路径。
- `frontend.url`：后端启动受控浏览器时打开的前端地址。

## Automa 桥接说明

项目通过 Automa content script 注入的页面 bridge 与扩展通信。前端向页面派发 `__automa-ext__` 事件，Automa 扩展回传工作流数据或执行结果。

常用事件：

- `__automa-ext__`：前端发送给 Automa bridge。
- `__automa-ext__get-workflows`：Automa 返回本地工作流列表。
- `__automa-ext__add-workflow`：Automa 返回工作流导入结果。
- `automa:execute-workflow`：前端触发 Automa 执行工作流。

注意：如果在浏览器页面已经打开后才安装 Automa 扩展，content script 通常不会自动注入到已打开页面。Browser Agent 会在连续检测不到 bridge 时刷新页面，以便新安装的扩展重新注入。

## 工作流相关页面

### `/workflows`

Windows 模式下的工作流列表，数据来自当前打开管理页面的浏览器 Automa bridge。它读取的是当前浏览器 profile 中的 Automa 本地工作流。

### `/automa`

Server 模式下的工作流管理页，数据来自后端数据库或同步缓存。它管理的是后端保存的工作流记录，不等同于当前浏览器实时 Automa 列表。

### `/browser-agent`

Browser Agent 页面由后端启动的受控浏览器打开，用于：

- 建立 WebSocket 长连接。
- 上报 Automa 安装状态和版本。
- 接收后端下发的工作流列表、打开、执行等命令。

## 大模型和对话

Windows 模式下可以在“大模型”页面配置模型提供商、模型名称、API Key 和 Base URL。

“对话”页面只使用已启用的大模型配置创建会话，不在对话页内新增模型配置。

## Git 忽略说明

以下内容默认作为本地文件或外部源码快照处理：

- `AGENTS.md`
- `CLAUDE.md`
- `automa/`
- `browserwing/`

Automa 本地源码改动记录在 `AUTOMA.md` 中。更新或覆盖 `automa/` 后，需要对照该文件重新检查本地改动。
