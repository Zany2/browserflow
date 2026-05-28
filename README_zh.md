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
  <a href="./README.md">English</a> · 简体中文
</p>

BrowserFlow 是一个浏览器工作流自动化平台，可以把网页业务流程沉淀成可复用、可调度、可审计的自动化能力，并将原本只能在单个浏览器里手动管理和运行的工作流，扩展成集中管理、远程调度、定时执行、记录审计和结果回传的完整流程。BrowserFlow 支持集成 Automa 浏览器工作流，也支持导出大模型 Skill，让自动化流程可以被大模型按说明调用，适合把浏览器里的业务操作连接到更大的自动化系统中。

项目支持两种运行模式：

- **Windows 模式**：面向本机桌面自动化。后端启动并连接本地受控浏览器，前端提供浏览器管理、工作流列表、大模型配置和本地对话能力；Server 侧的任务、执行记录、客户端管理页面会被禁用。业务数据保存在 `localStorage.path` 指向的本地 BoltDB 文件中，不依赖 PostgreSQL 或 Redis。
- **Server 模式**：面向多客户端集中调度。服务端提供工作流管理、任务配置、执行记录和客户端管理能力；Windows 本地浏览器、大模型配置和本地对话页面会被禁用。业务数据以 PostgreSQL 为准，Redis 用于维护在线客户端节点、工作流清单缓存、任务调度状态和客户端节点执行锁；任务可按自动、IP、节点或分组策略下发到打开 `/client-agent` 的 Windows 客户端。

它适合承载需要通过浏览器完成的业务流程，例如跨系统操作、内部管理后台、政务或专网业务系统、数据查询、信息采集、表单处理和报表导出等场景；也适合需要把浏览器流程统一管理、分发到指定客户端执行、保留执行记录，或进一步接入大模型调用的自动化任务。

## 核心价值

BrowserFlow 的核心是把浏览器里的业务流程变成可以被管理、调度、执行、记录和集成的自动化能力。

- **沉淀浏览器业务流程**：把查询、填报、采集、导出、巡检等网页操作整理成可复用的工作流能力，而不是停留在临时脚本或人工步骤里。
- **统一管理工作流资产**：支持工作流导入、同步、对比、保护和版本信息维护，便于在团队或多客户端场景下保持流程一致。
- **集中调度执行客户端**：Server 模式可以把多台 Windows 客户端接入为执行节点，并按自动、IP、节点或分组策略分发任务。
- **保留执行痕迹**：任务执行会形成记录，包含触发方式、执行客户端、参数、状态、失败原因、结果 JSON 和结果文件。
- **避免浏览器并发污染**：同一个客户端节点同一时间只执行一个浏览器工作流，避免多个流程同时切换标签页、改变量或抢占上下文。
- **支持定时任务**：任务可以配置 Cron 表达式，服务端把数据库任务配置同步到 GoFrame gcron。
- **连接大模型和外部自动化**：工作流和任务可以导出为 Skill，让大模型或其他自动化系统按说明触发浏览器流程并读取结果。

## 运行模式

| 项目 | Windows 模式 | Server 模式 |
| --- | --- | --- |
| 主要定位 | 本机桌面自动化 | 多客户端集中调度 |
| 数据存储 | BoltDB | PostgreSQL + Redis |
| 后端位置 | Windows 本机 | 服务器 |
| 执行端 | 后端启动的本地受控浏览器 | 打开 `/client-agent` 的 Windows 客户端节点 |
| 执行端页面 | `/browser-agent` | `/client-agent` |
| 主要页面 | 浏览器、工作流、大模型、对话 | 工作流管理、任务配置、执行记录、客户端 |
| 工作流来源 | 当前受控浏览器里的工作流 | 客户端同步或服务端导入 |
| 调度方式 | 本机直接运行 | 自动、IP、节点、分组调度，支持 Cron |
| 执行记录 | 面向本地执行查看 | 集中记录、查询和结果文件管理 |
| 适合场景 | 本机工具、个人自动化、调试、本地大模型调用 | 团队协作、多客户端执行、集中调度、统一审计 |

## 快速开始

### Windows 模式

Windows 模式面向本机浏览器自动化，适合个人使用、工作流调试和本地大模型调用。当前先按源码编译方式使用：先编译前端 `dist`，再将 `dist` 作为静态资源嵌入后端 Windows 可执行文件，最后直接启动可执行文件访问本地控制台。

首次使用前需要准备：

- Windows 10/11。
- Go 1.25+。
- Node.js 和 npm。
- Chrome 或 Chromium。
- Automa 浏览器扩展。

#### 1. 编译前端

```bash
cd frontend
npm install
npm run build
```

执行完成后会生成 `frontend/dist`。

#### 2. 将前端 dist 编译进后端可执行文件

后端会从 `backend/internal/web/dist` 嵌入静态资源。前端编译完成后，先把生成的文件移动到该目录，再编译后端可执行文件：

```bash
cd ..
mkdir -p backend/internal/web/dist
mv frontend/dist/* backend/internal/web/dist/
cd backend
go build -o BrowserFlow.exe .
```

如果使用 Windows PowerShell，可使用：

```powershell
New-Item -ItemType Directory -Force backend\internal\web\dist | Out-Null
Move-Item frontend\dist\* backend\internal\web\dist\
Set-Location backend
go build -o BrowserFlow.exe .
```

如果在非 Windows 环境交叉编译 Windows 可执行文件，可使用：

```bash
GOOS=windows GOARCH=amd64 go build -o BrowserFlow.exe .
```

#### 3. 启动 Windows 模式

```bash
./BrowserFlow.exe --port 8001
```

也可以省略 `--port` 使用配置文件中的端口；如果首次启动时传入 `--port`，生成的 Windows 配置文件会写入该端口。启动后访问：

```text
http://127.0.0.1:8001
```

基本使用流程：

1. 启动 BrowserFlow Windows 模式。
2. 打开或等待自动打开本地控制台。
3. 在“浏览器”页面启动受控浏览器。
4. 在受控浏览器中准备需要执行的浏览器工作流。
5. 回到“工作流”页面查看、运行或导出 Skill。
6. 如需大模型调用，在“大模型”页面配置模型，然后在“对话”页面发起调用。

### Server 模式

Server 模式建议部署在服务器上，用于统一管理工作流、任务配置和执行记录，并调度多个 Windows 客户端作为浏览器自动化执行节点。

Server 模式依赖 PostgreSQL 和 Redis。Windows 客户端通过打开 `/client-agent` 接入服务端，成为可调度的执行节点。

启动后，基本使用流程是：接入 Windows 客户端、同步或导入工作流、创建任务、手动执行或配置 Cron 定时触发，并在执行记录中查看结果。

#### Docker Compose 部署

Server 模式建议部署在 Linux 服务器上，并使用 `docker compose` 运行 BrowserFlow、PostgreSQL、Redis 和 Nginx。当前部署方式是先在服务器上编译前端和后端，再由 `docker-compose.yaml` 挂载编译后的后端可执行文件和 Server 配置启动服务。

服务器需要准备：

- Linux 服务器。
- Git。
- Go 1.25+。
- Node.js 和 npm。
- Docker 和 Docker Compose。

拉取源码：

```bash
git clone https://github.com/Zany2/browserflow.git
cd browserflow
```

编译前端，并把前端产物移动到后端嵌入目录：

```bash
cd frontend
npm install
npm run build

cd ..
mkdir -p backend/internal/web/dist
mv frontend/dist/* backend/internal/web/dist/
```

编译 Linux 后端可执行文件：

```bash
cd backend
go build -o browserflow .
cd ..
```

启动 Server 模式：

```bash
docker compose up -d
```

默认访问地址：

```text
http://服务器IP:8001
```

当前 `docker-compose.yaml` 会启动：

- `postgres`：PostgreSQL，首次启动时会加载 `backend/sql/public.sql` 初始化表结构。
- `redis`：Redis，默认开启 AOF 持久化，并使用 `browserflow` 作为示例密码。
- `browserflow`：BrowserFlow Server，挂载 `backend/browserflow` 可执行文件和 `deploy/server/config.yaml`。
- `nginx`：统一对外暴露 HTTP 入口，代理前端页面、API 和 WebSocket。

部署相关文件：

- `docker-compose.yaml`：Server 模式容器编排。
- `deploy/server/config.yaml`：Server 模式后端配置，包含 PostgreSQL、Redis、日志和 WebSocket 配置。
- `deploy/nginx/browserflow.conf`：Nginx 反向代理配置。

如果修改了前端或后端代码，需要重新执行前端构建、移动 `dist`、重新编译 `backend/browserflow`，然后重启服务：

```bash
docker compose restart browserflow nginx
```

停止服务：

```bash
docker compose down
```

生产部署前，建议修改 `docker-compose.yaml` 和 `deploy/server/config.yaml` 中的 PostgreSQL 密码、Redis 密码、`frontend.url`、Nginx `server_name` 等配置。

#### 源码运行

源码运行方式更适合开发、测试和自定义部署。可以参考 `deploy/server/config.yaml` 配置 Server 模式、PostgreSQL 和 Redis：

```bash
git clone https://github.com/Zany2/browserflow.git
cd browserflow
```

配置 Server 模式、PostgreSQL 和 Redis 后，启动后端：

```bash
cd backend
go run .
```

启动前端：

```bash
cd frontend
npm install
npm run dev
```

## 项目结构

```text
browserflow/
|-- backend/                         GoFrame 后端服务
|   |-- api/                         GoFrame API 请求/响应定义
|   |-- internal/                    控制器、服务注册、数据模型和业务实现
|   |   `-- web/                     嵌入式前端静态资源入口
|   |-- manifest/config/             后端配置文件
|   |-- middleware/                  HTTP 中间件
|   |-- sql/                         数据库初始化和结构脚本
|   `-- utility/                     浏览器执行、工作流、任务调度、WebSocket 等工具模块
|-- frontend/                        Vue 3 + Vite 前端应用
|   |-- src/api/                     前端请求封装
|   |-- src/components/              通用组件
|   |-- src/composables/             组合式逻辑
|   |-- src/layouts/                 页面布局
|   |-- src/router/                  前端路由和模式路由守卫
|   |-- src/services/                业务 API、WebSocket 和前端服务
|   |-- src/styles/                  全局样式和共享页面类
|   |-- src/utils/                   通用工具函数
|   `-- src/views/                   页面视图
|-- windows-worker/                  Windows 客户端/Worker 可执行程序源码和资源
|   |-- cmd/                         Worker 入口
|   |-- internal/                    Worker 内部实现
|   |-- assets/                      图标和静态资源
|   |-- chrome-packages/             浏览器相关包资源
|   `-- winres/                      Windows 可执行文件资源配置
|-- docs/                            项目文档和图片资源
|-- docs/images/                     README 和文档图片
|-- deploy/                          Server 模式部署配置
|   |-- nginx/                       Nginx 反向代理配置
|   `-- server/                      Server 模式后端配置
|-- third_party/automa/              Automa 本地源码快照和 BrowserFlow 本地改造
|-- workflows/                       示例 Automa 工作流文件
|-- .agents/                         项目本地 coding-agent 技能和配置
|-- docker-compose.yaml              Server 模式 Docker Compose 部署文件
|-- go.work                          Go workspace
|-- LICENSE                          开源许可证
|-- README.md                        English README
`-- README_zh.md                     中文说明
```

## 安全和部署建议

BrowserFlow 可以触发真实浏览器自动化操作，也可能处理账号、业务参数、查询结果、导出文件等敏感信息。部署到生产环境前，建议至少关注以下事项：

- **访问入口保护**：使用 Nginx、网关或反向代理统一暴露服务入口，并为管理页面、API 和 WebSocket 增加认证、授权、IP 白名单或单点登录。
- **客户端入口控制**：不要把 `/client-agent` 暴露给不可信用户或网络。只有被授权的 Windows 客户端才应该接入为执行节点。
- **数据库和 Redis 隔离**：PostgreSQL 和 Redis 不应直接暴露到公网，建议只允许 BrowserFlow 服务端和可信运维网络访问。
- **任务权限管理**：不同部门、业务系统或执行节点建议做权限隔离，避免用户创建或执行超出授权范围的浏览器任务。
- **敏感数据保护**：任务参数、Cookie、Token、账号密码、执行结果和结果文件都可能包含敏感数据，应结合业务要求做加密、脱敏、访问控制和保留周期管理。
- **执行节点安全**：Server 模式下，Windows 客户端会真实执行浏览器操作，应使用受控账号、专用机器或隔离环境运行，避免和个人办公浏览器混用。
- **并发和锁保护**：同一客户端节点同一时间只应执行一个浏览器工作流，避免多个任务同时切换标签页、改变量或抢占浏览器上下文。
- **审计和追踪**：保留任务创建、修改、启停、执行、删除、失败原因、执行客户端和操作者信息，便于后续审计和问题追踪。
- **告警和恢复**：对关键任务失败、执行超时、客户端离线、Redis 锁异常、数据库连接异常等情况配置告警和恢复流程。
- **生产配置管理**：不要把真实数据库密码、Redis 密码、API Key 或业务账号提交到代码仓库，建议通过环境变量、密钥管理服务或部署平台配置注入。
- **网络和证书**：生产环境建议启用 HTTPS/WSS，并根据部署环境配置可信证书、代理超时和 WebSocket 长连接参数。
- **发布前验证**：上线前在隔离环境验证任务影响范围、失败回滚方式、数据导出位置、结果文件访问权限和客户端断线后的恢复行为。

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

## 许可证

BrowserFlow 使用 [Apache-2.0](LICENSE) 协议开源。
