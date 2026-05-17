# Automa 本地源码改动记录

这些是对 `third_party/automa/` 上游源码的本地改动。更新、重新拉取或覆盖 Automa 源码后，需要检查并重新应用。
以后凡是修改 `third_party/automa/src`，都必须先用成对注释包起来，再把修改记录追加到本文件，方便后续和上游源码对照。

## 源码改动标记规范

所有对 `third_party/automa/src` 的本地源码改动，都需要用成对注释包起来，方便后续和上游源码合并、排查或重新应用：

```js
// BrowserFlow local change start: short English intent 简短中文意图
// changed code
// BrowserFlow local change end
```

要求：

- 注释必须包住实际改动代码，不能只写在附近。
- 注释使用 `// English 中文` 风格，保持和项目注释规范一致。
- 每一类源码改动都需要同步记录到本文档，至少包含文件路径、目的、行为变化和关键代码位置。

## 导出工作流 JSON 时保留原始标识和时间戳

- File: `third_party/automa/src/utils/workflowData.js`
- Purpose: 导出 `.automa.json` 时保留 Automa 本地工作流的 `id`、`createdAt`、`updatedAt`。
- Backend fit: 本项目后端导入、文件导入和客户端同步时，将 JSON 顶层 `id` 保存为数据库 `automa_id`，并将 `createdAt`、`updatedAt` 分别保存为 `created_at_automa`、`updated_at_automa`。
- Update rule: 后端更新已有工作流时会比较 JSON `updatedAt` 和数据库 `updated_at_automa`，避免旧导出文件覆盖数据库中更新的工作流。

### 字段说明

```json
{
  "id": "Automa 本地工作流 ID，对应后端 automa_id",
  "createdAt": 1710000000000,
  "updatedAt": 1710000100000
}
```

`createdAt` 和 `updatedAt` 是 Automa 原始时间戳，通常为毫秒时间戳。它们不是本项目服务端记录的 `created_at` / `updated_at`，后者表示服务端数据库记录的创建和更新时间。

### 完整修改代码

```js
import browser from 'webextension-polyfill';
import { useWorkflowStore } from '@/stores/workflow';
import { registerWorkflowTrigger } from './workflowTrigger';
import {
  parseJSON,
  fileSaver,
  openFilePicker,
  findTriggerBlock,
} from './helper';

const contextMenuPermission =
  BROWSER_TYPE === 'firefox' ? 'menus' : 'contextMenus';
const checkPermission = (permissions) =>
  browser.permissions.contains({ permissions });
const requiredPermissions = {
  trigger: {
    name: contextMenuPermission,
    hasPermission({ data }) {
      const permissions = [];

      if (data.triggers) {
        data.triggers.forEach((trigger) => {
          if (trigger.type !== 'context-menu') return;

          permissions.push(contextMenuPermission);
        });
      } else if (data.type === 'context-menu') {
        permissions.push(contextMenuPermission);
      }

      return checkPermission(permissions);
    },
  },
  clipboard: {
    name: 'clipboardRead',
    hasPermission() {
      const clipboardPermissions = ['clipboardRead'];
      if (BROWSER_TYPE === 'firefox')
        clipboardPermissions.push('clipboardWrite');

      return checkPermission(clipboardPermissions);
    },
  },
  notification: {
    name: 'notifications',
    hasPermission() {
      return checkPermission(['notifications']);
    },
  },
  'handle-download': {
    name: 'downloads',
    hasPermission() {
      return checkPermission(['downloads']);
    },
  },
  'save-assets': {
    name: 'downloads',
    hasPermission() {
      return checkPermission(['downloads']);
    },
  },
  cookie: {
    name: 'cookies',
    hasPermission() {
      return checkPermission(['cookies']);
    },
  },
};

export async function getWorkflowPermissions(drawflow) {
  let blocks = [];
  const permissions = [];
  const drawflowData =
    typeof drawflow === 'string' ? parseJSON(drawflow) : drawflow;

  if (drawflowData.nodes) {
    blocks = drawflowData.nodes;
  } else {
    blocks = Object.values(drawflowData.drawflow?.Home?.data || {});
  }

  for (const block of blocks) {
    const name = block.label || block.name;
    const permission = requiredPermissions[name];

    if (permission && !permissions.includes(permission.name)) {
      const hasPermission = await permission.hasPermission(block);
      if (!hasPermission) permissions.push(permission.name);
    }
  }

  return permissions;
}

export function importWorkflow(attrs = {}) {
  return new Promise((resolve, reject) => {
    openFilePicker(['application/json'], attrs)
      .then((files) => {
        const handleOnLoadReader = ({ target }) => {
          const workflow = JSON.parse(target.result);
          const workflowStore = useWorkflowStore();

          if (workflow.includedWorkflows) {
            Object.keys(workflow.includedWorkflows).forEach((workflowId) => {
              const isWorkflowExists = Boolean(
                workflowStore.workflows[workflowId]
              );

              if (isWorkflowExists) return;

              const currentWorkflow = workflow.includedWorkflows[workflowId];
              currentWorkflow.table =
                currentWorkflow.table || currentWorkflow.dataColumns;
              delete currentWorkflow.dataColumns;

              workflowStore.insert(
                {
                  ...currentWorkflow,
                  id: workflowId,
                  createdAt: Date.now(),
                },
                { duplicateId: true }
              );
            });

            delete workflow.includedWorkflows;
          }

          workflow.table = workflow.table || workflow.dataColumns;
          delete workflow.dataColumns;

          if (typeof workflow.drawflow === 'string') {
            workflow.drawflow = parseJSON(workflow.drawflow, {});
          }

          workflowStore
            .insert({
              ...workflow,
              createdAt: Date.now(),
            })
            .then((result) => {
              Object.values(result).forEach((item) => {
                const triggerBlock = findTriggerBlock(item.drawflow);
                registerWorkflowTrigger(item.id, triggerBlock);
              });

              resolve(result);
            });
        };

        files.forEach((file) => {
          const reader = new FileReader();

          reader.onload = handleOnLoadReader;
          reader.readAsText(file);
        });
      })
      .catch((error) => {
        console.error(error);
        reject(error);
      });
  });
}

const defaultValue = {
  name: '',
  icon: '',
  table: [],
  settings: {},
  globalData: '',
  dataColumns: [],
  description: '',
  drawflow: { nodes: [], edges: [] },
  version: browser.runtime.getManifest().version,
};

export function convertWorkflow(workflow, additionalKeys = []) {
  if (!workflow) return null;

  const keys = [
    'name',
    'icon',
    'table',
    'version',
    'drawflow',
    'settings',
    'globalData',
    'description',
    ...additionalKeys,
  ];
  const content = {
    extVersion: browser.runtime.getManifest().version,
  };

  keys.forEach((key) => {
    content[key] = workflow[key] ?? defaultValue[key];
  });

  return content;
}
function findIncludedWorkflows(
  { drawflow },
  store,
  maxDepth = 3,
  workflows = {}
) {
  if (maxDepth === 0) return workflows;

  const flow = parseJSON(drawflow, drawflow);
  const blocks = flow?.drawflow?.Home.data ?? flow.nodes ?? null;
  if (!blocks) return workflows;

  const checkWorkflow = (type, workflowId) => {
    if (type !== 'execute-workflow' || workflows[workflowId]) return;

    const workflow = store.getById(workflowId);
    if (workflow) {
      workflows[workflowId] = convertWorkflow(workflow, [
        'id',
        'createdAt',
        'updatedAt',
      ]);
      findIncludedWorkflows(workflow, store, maxDepth - 1, workflows);
    }
  };

  if (flow.nodes) {
    flow.nodes.forEach((node) => {
      checkWorkflow(node.label, node.data.workflowId);
    });
  } else {
    Object.values(blocks).forEach(({ data, name }) => {
      checkWorkflow(name, data.workflowId);
    });
  }

  return workflows;
}
export function exportWorkflow(workflow) {
  if (workflow.isProtected) return;

  const workflowStore = useWorkflowStore();
  const includedWorkflows = findIncludedWorkflows(workflow, workflowStore);
  const content = convertWorkflow(workflow, ['id', 'createdAt', 'updatedAt']);

  content.includedWorkflows = includedWorkflows;

  const blob = new Blob([JSON.stringify(content)], {
    type: 'application/json',
  });
  const url = URL.createObjectURL(blob);

  fileSaver(`${workflow.name}.automa.json`, url);
}

export default {
  export: exportWorkflow,
  import: importWorkflow,
};
```

### 关键改动点

主工作流导出：

```js
const content = convertWorkflow(workflow, ['id', 'createdAt', 'updatedAt']);
```

子工作流导出：

```js
workflows[workflowId] = convertWorkflow(workflow, [
  'id',
  'createdAt',
  'updatedAt',
]);
```

这样导出的主工作流和 `includedWorkflows` 中的子工作流都会保留原始 `id`、`createdAt`、`updatedAt`。

## 页面桥接导入工作流时保留原始标识和时间戳

- File: `third_party/automa/src/content/services/webService.js`
- Purpose: 前端客户端执行页通过 `__automa-ext__` 发送 `add-workflow` 时，导入到本地 Automa 要保留服务端工作流 payload 中的 `id`、`createdAt`、`updatedAt`。
- Problem: 上游默认逻辑会用 `nanoid()` 重新生成工作流 ID，并用 `Date.now()` 覆盖 `createdAt`，前端也拿不到导入是否成功的回执，所以“同步到本地”容易表现为没有反应。
- Behavior: 如果 payload 带 `id`，直接作为本地工作流 ID；如果本地已有同 ID 工作流，则覆盖更新；如果没有 `id`，才回退到 `nanoid()`。
- Ack: 导入成功或失败后，通过 `__automa-ext__add-workflow` 回传 `requestId`、`ok`、`workflow` 或 `error`，前端据此提示成功/失败。
- Init: 桥接初始化不能只依赖 `DOMContentLoaded`。扩展在页面已加载后重载或重新注入时，`DOMContentLoaded` 已经错过，会导致前端等待 `add-workflow` 回执超时。

### 字段说明

```json
{
  "requestId": "前端生成的桥接请求 ID，用于匹配导入回执",
  "workflow": {
    "id": "Automa 本地工作流 ID，对应后端 automa_id",
    "createdAt": 1710000000000,
    "updatedAt": 1710000100000
  }
}
```

`createdAt` 和 `updatedAt` 支持毫秒时间戳或可被 `Date` 解析的时间字符串。无法解析时才使用当前时间兜底。

### 完整修改代码

在 `sendMessageBack` 后增加时间戳归一化 helper：

```js
// normalizeWorkflowTimestamp keeps imported Automa timestamps 保留导入工作流的原始时间
function normalizeWorkflowTimestamp(value, fallbackValue = Date.now()) {
  const numberValue = Number(value);
  if (Number.isFinite(numberValue) && numberValue > 0) return numberValue;

  const dateValue = new Date(value).getTime();
  if (!Number.isNaN(dateValue)) return dateValue;

  return fallbackValue;
}
```

替换 `add-workflow` 监听逻辑：

```js
webListener.on('add-workflow', async ({ workflow, requestId }) => {
  try {
    const { workflows: storedWorkflows } = await browser.storage.local.get(
      'workflows'
    );
    const workflowsStorage = storedWorkflows || {};

    const workflowId = workflow.id || nanoid();
    const now = Date.now();
    const workflowData = {
      ...workflow,
      id: workflowId,
      dataColumns: workflow.dataColumns || [],
      createdAt: normalizeWorkflowTimestamp(workflow.createdAt, now),
      updatedAt: normalizeWorkflowTimestamp(workflow.updatedAt, now),
      table: workflow.table || workflow.dataColumns || [],
    };

    workflowData.drawflow =
      typeof workflowData.drawflow === 'string'
        ? parseJSON(workflowData.drawflow, workflowData.drawflow)
        : workflowData.drawflow;

    if (Array.isArray(workflowsStorage)) {
      const workflowIndex = workflowsStorage.findIndex(
        (item) => item.id === workflowId
      );

      if (workflowIndex === -1) {
        workflowsStorage.push(workflowData);
      } else {
        workflowsStorage[workflowIndex] = workflowData;
      }
    } else {
      workflowsStorage[workflowId] = workflowData;
    }

    await browser.storage.local.set({ workflows: workflowsStorage });
    sendMessage(
      'workflow:added',
      { workflowId, workflowData },
      'background'
    );
    sendMessageBack('add-workflow', {
      ok: true,
      requestId,
      workflow: workflowData,
      workflow_id: workflowId,
      name: workflowData.name,
    });
  } catch (error) {
    console.error(error);
    sendMessageBack('add-workflow', {
      ok: false,
      requestId,
      error: error.message,
    });
  }
});
```

将原本的 `window.addEventListener('DOMContentLoaded', async () => { ... })` 包装为可立即执行的初始化函数：

```js
async function initWebServiceBridge() {
  try {
    document.body.setAttribute(
      'data-atm-ext-installed',
      browser.runtime.getManifest().version
    );

    const { workflows } = await browser.storage.local.get('workflows');
    const db = await openDB('automa', 1, {
      upgrade(event) {
        event.createObjectStore('store');
      },
    });

    await db.put('store', workflows, 'workflows');

    const webListener = initWebListener();

    // Keep existing webListener.on(...) registrations here 保留原有监听注册
  } catch (error) {
    console.error(error);
  }
}

if (document.readyState === 'loading') {
  window.addEventListener('DOMContentLoaded', initWebServiceBridge, {
    once: true,
  });
} else {
  initWebServiceBridge();
}
```

### 关联前端调用

前端不是 Automa 源码，但和上面的回执配套：

- File: `frontend/src/constants/automa.js`
- Event: `importWorkflowResponse: '__automa-ext__add-workflow'`
- File: `frontend/src/services/automaBridge.js`
- Behavior: `importAutomaWorkflow(workflow)` 发送 `add-workflow` 时带 `requestId`，并等待 `__automa-ext__add-workflow` 回执；`ok === false` 或超时会抛出错误。

前端同步到本地前会将服务端详情 payload 处理为 Automa 可识别的字段：

```js
payload.id = workflow.automa_id || workflow.automaId || workflow.workflow_id || payload.id;
payload.createdAt = workflow.created_at_automa || workflow.createdAtAutoma || payload.createdAt;
payload.updatedAt = workflow.updated_at_automa || workflow.updatedAtAutoma || payload.updatedAt;
```

这样服务端工作流同步到本地 Automa 后，本地 ID 和两个 Automa 原始时间戳不会被桥接层改写。

## 外部执行工作流时跳过 Automa 参数页

- File: `third_party/automa/src/content/services/shortcutListener.js`
- File: `third_party/automa/src/workflowEngine/WorkflowEngine.js`
- Purpose: BrowserFlow 工作流页面已经弹出自己的参数输入框，并把参数通过 `data.variables` 传给 Automa。Automa 源码需要保留外部传入的 `options.checkParams = false`，避免再次弹出 Automa 自带的参数输入页。
- Problem: 上游 `automa:execute-workflow` 监听只会把 `detail.data` 写入 `workflow.options.data`。如果不透传 `detail.options`，`WorkflowEngine` 会使用默认 `checkParams = true`，有触发器参数时就会打开 Automa 参数页。
- Behavior: `shortcutListener` 将外部事件中的 `detail.options` 和 `detail.data` 一起写入 `workflow.options`；`WorkflowEngine` 使用 `this.options?.checkParams ?? true` 判断是否需要弹 Automa 参数页。

### 关键源码标记

`third_party/automa/src/content/services/shortcutListener.js`：

```js
// BrowserFlow local change start: forward external run options 透传外部执行参数
workflow.options = {
  // BrowserFlow options allow external callers to finish param handling.
  // BrowserFlow options 允许外部调用方完成参数处理后跳过 Automa 参数页。
  ...(detail.options || {}),
  data: detail.data || {},
};
// BrowserFlow local change end
```

`third_party/automa/src/workflowEngine/WorkflowEngine.js`：

```js
// BrowserFlow local change start: allow external callers to skip Automa params page 允许外部调用跳过 Automa 参数页
const checkParams = this.options?.checkParams ?? true;
const hasParams =
  checkParams && triggerBlock.data?.parameters?.length > 0;
// BrowserFlow local change end
```

### 关联 BrowserFlow 调用

前端不是 Automa 源码，但和上面的源码改动配套：

- File: `frontend/src/services/automaBridge.js`
- Behavior: `runAutomaWorkflow({ id, variables, checkParams = false })` 发送 `automa:execute-workflow` 事件时，把 `checkParams` 放入 `detail.options`，把参数放入 `detail.data.variables`。
- File: `frontend/src/views/browser-agent/BrowserAgentView.vue`
- Behavior: Browser Agent 收到后端 `automa.workflow.run` 时，会把 `payload.check_params` 转为 `checkParams`，默认值是 `false`。
- File: `backend/internal/controller/workflows/automa_v1_automa_workflow_run.go`
- Behavior: 后端执行工作流命令下发 `check_params: false`，表示参数已经由 BrowserFlow 侧处理完成。

## BrowserFlow execution completion and returned data
- File: `third_party/automa/src/background/index.js`
- File: `third_party/automa/src/workflowEngine/WorkflowManager.js`
- File: `third_party/automa/src/content/services/webService.js`
- Purpose: BrowserFlow exported Automa Skills support both async dispatch and sync wait. When sync wait is requested, Automa sends the terminal status and selected output data back to the browser-agent page, and browser-agent returns it to the backend through the existing WebSocket command result.
- Behavior: BrowserFlow sends `browserFlowRequestId`, `browserFlowWaitResult`, and `browserFlowReturnData` through workflow options. Automa records the source tab id as `browserFlowSourceTabId`, emits a result from `engine.on('destroyed')`, and forwards `browserflow:workflow-result` to the page event `__browserflow_automa_workflow_result__`.
- Data rule: Workflows should write business output to the `browserflow_output` variable. Sync Skill calls should request and read that variable first instead of returning all variables, table rows, or logs.
