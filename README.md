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

> **把浏览器实例、Automa 工作流、任务调度和本地大模型对话放进同一个管理台。** BrowserFlow 支持本地 Windows 桌面模式和服务端调度模式，可通过 Automa bridge 读取、同步并执行浏览器工作流。

```bash
# 启动后端
cd backend && go run .

# 启动前端
cd frontend && npm install && npm run dev

# 打开
http://localhost:5173
```

## Highlights

**面向 Automa 和浏览器执行端的轻量管理平台**

- **双运行模式**：`windows` 本地桌面模式和 `server` 服务端调度模式共用一套前后端。
- **浏览器实例管理**：创建、启动、停止和切换本地或远程浏览器实例。
- **Browser Agent**：受控浏览器通过 WebSocket 接收命令，并上报 Automa 安装状态和版本。
- **Automa Bridge**：通过页面事件桥接当前浏览器中的 Automa 扩展，读取本地工作流并触发执行。
- **工作流管理**：支持文件导入、客户端同步、同步候选对比和后端记录管理。
- **任务调度**：服务端模式下可向在线客户端下发任务并记录执行结果。
- **大模型与对话**：本地模式下维护模型配置，并基于启用模型创建对话会话。
- **模式感知路由**：前端根据后端运行模式自动禁用不适用页面。

## Requirements

- Go 1.25+
- Node.js 和 npm
- Chrome 或 Chromium
- 使用 Automa 功能时，需要在目标浏览器中安装并启用 Automa 扩展
- Server 模式使用客户端缓存、工作流同步等能力时，需要可用的 PostgreSQL 和 Redis 配置

## Quick Start

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

默认服务地址：

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

## Runtime Modes

### Windows Mode

适合本机使用，关注本地浏览器、当前浏览器 Automa 工作流和本地大模型对话。

可用页面：

- 浏览器
- 工作流
- 大模型
- 对话

禁用页面：

- 工作流管理
- 任务
- 任务记录
- 客户端
- Client Agent

### Server Mode

适合作为调度中心，关注客户端、服务端工作流记录、任务下发和任务结果。

可用页面：

- 工作流管理
- 任务
- 任务记录
- 客户端
- Client Agent

禁用页面：

- 浏览器
- 大模型
- 对话
- Browser Agent

## Usage Guide

### Browser Agent

`/browser-agent` 通常由后端启动的受控浏览器自动打开。它负责：

1. 与后端建立 WebSocket 长连接。
2. 周期检测 Automa bridge 是否可用。
3. 上报 `automa_installed` 和 `automa_version`。
4. 接收后端下发的工作流列表、打开和执行命令。
5. 当连续检测不到 Automa bridge 时刷新页面，让新安装的扩展重新注入。

### Workflows

BrowserFlow 中有两个容易混淆的工作流入口：

- `/workflows`：Windows 模式页面，数据来自**当前打开管理页面的浏览器**中的 Automa bridge。
- `/automa`：Server 模式页面，数据来自后端保存的工作流记录或同步缓存。

### Automa Bridge Events

| 事件 | 方向 | 说明 |
| --- | --- | --- |
| `__automa-ext__` | 前端到扩展 | 统一 bridge 请求入口 |
| `__automa-ext__get-workflows` | 扩展到前端 | 返回本地 Automa 工作流列表 |
| `__automa-ext__add-workflow` | 扩展到前端 | 返回工作流导入结果 |
| `automa:execute-workflow` | 前端到扩展 | 触发 Automa 执行工作流 |

> 如果页面已经打开后才安装 Automa 扩展，content script 通常不会自动注入到已打开页面。Browser Agent 的自刷新逻辑就是为了解决这个场景。

### LLM Chat

Windows 模式下：

1. 在“大模型”页面配置提供商、模型、API Key 和 Base URL。
2. 启用一个或多个模型配置。
3. 在“对话”页面选择已启用模型并创建会话。

对话页只使用已有配置，不提供新增模型配置入口。

## Configuration

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
```

| 配置项 | 说明 |
| --- | --- |
| `server.address` | 后端 HTTP 服务监听地址 |
| `app.mode` | 运行模式，支持 `windows` 和 `server` |
| `localStorage.path` | 本地 bbolt 数据文件路径 |
| `frontend.url` | 后端启动受控浏览器时打开的前端地址 |

## Project Structure

```text
browserflow/
├─ backend/                  GoFrame 后端服务
├─ frontend/                 Vue 3 前端应用
├─ docs/images/              README 和文档图片资源
├─ automa/                   Automa 本地源码快照，默认不提交
├─ browserwing/              BrowserWing 本地源码快照，默认不提交
├─ AUTOMA.md                 Automa 本地源码改动记录
├─ go.work                   Go workspace
└─ README.md
```

## Development

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

## Git Notes

以下内容属于本地运行数据、外部源码快照或个人配置，不应提交：

- `backend/data/`
- `backend/user_data/`
- `backend/logs/`
- `AGENTS.md`
- `CLAUDE.md`
- `automa/`
- `browserwing/`

如果这些文件已经被 Git 跟踪，仅加入 `.gitignore` 不会自动取消跟踪，需要执行 `git rm --cached` 后再提交。

`backend/manifest/config/config.yaml` 可能包含数据库、Redis 或本地服务地址。提交前请确认其中没有不应公开的敏感信息。

## Automa Local Changes

`automa/` 是外部源码快照，默认不提交。本项目对 Automa 的本地改动记录在 `AUTOMA.md` 中。更新或覆盖 `automa/` 后，需要对照该文件检查并重新应用必要改动。
