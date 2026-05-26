<template>
  <main class="agent-page">
    <section class="agent-console">
      <header class="agent-hero">
        <div>
          <p class="eyebrow">BrowserFlow Agent</p>
          <h1>本地浏览器执行端</h1>
          <p>保持此页面打开，BrowserFlow 会通过它读取和执行当前浏览器中的 Automa 工作流。</p>
        </div>
        <el-tag size="large" :type="connectionTagType">{{ connectionStatusText }}</el-tag>
      </header>

      <section class="status-grid" aria-label="执行端状态">
        <article class="status-card" :class="connectionStateClass">
          <span>后端连接</span>
          <strong>{{ connectionStatusText }}</strong>
          <small>{{ connectionNote }}</small>
        </article>
        <article class="status-card" :class="automaStateClass">
          <span>Automa</span>
          <strong>{{ automaStatusText }}</strong>
          <small>{{ automaNote }}</small>
        </article>
      </section>

      <section class="detail-panel">
        <div class="detail-row">
          <span>Browser ID</span>
          <strong>{{ browserId || '等待后端分配' }}</strong>
        </div>
        <div class="detail-row">
          <span>最后命令</span>
          <strong>{{ lastCommand || '暂无' }}</strong>
        </div>
        <div class="detail-row">
          <span>执行标识</span>
          <strong>{{ lastExecutionId || lastCommandId || '暂无' }}</strong>
        </div>
        <div class="detail-row">
          <span>执行结果</span>
          <strong>{{ lastResult || '等待命令' }}</strong>
        </div>
      </section>
    </section>
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import {
  getAutomaInfo,
  getAutomaWorkflows,
  openAutomaWorkflow,
  runAutomaWorkflow,
} from '@/services/automaBridge'
import { createAgentSocket } from '@/services/agentWs'

const route = useRoute()
const browserId = ref(resolveBrowserId())
const status = ref('connecting')
const automaChecked = ref(false)
const automaInstalled = ref(false)
const automaVersion = ref('')
const lastCommand = ref('')
const lastCommandId = ref('')
const lastExecutionId = ref('')
const lastResult = ref('')
const closeSocket = ref(null)

onMounted(() => {
  closeSocket.value = createAgentSocket({
    browserId: browserId.value,
    getAutomaInfo,
    getWorkflows: getAutomaWorkflows,
    onCommand: handleCommand,
    onStatus: (nextStatus) => {
      status.value = nextStatus
    },
    onRegistered: (payload) => {
      browserId.value = payload?.browser_id || ''
      updateAutomaInfo(payload)
    },
    onAutomaStatus: updateAutomaInfo,
    onMessage: handleSocketMessage,
    onError: (error) => {
      lastResult.value = error.message
      appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
    },
  })
})

onBeforeUnmount(() => {
  closeSocket.value?.()
})

function handleSocketMessage(payload) {
  if (payload?.type === 'agent_registered' || payload?.type === 'agent_status') {
    updateAutomaInfo(payload)
  }
  if (payload?.type === 'agent_command') {
    lastCommandId.value = payload.command_id || ''
    lastExecutionId.value = payload.payload?.execution_id || payload.payload?.executionId || ''
  }
}

async function handleCommand(command, payload, rawMessage) {
  lastCommand.value = command
  lastCommandId.value = rawMessage?.command_id || lastCommandId.value
  lastExecutionId.value = payload.execution_id || payload.executionId || lastExecutionId.value

  if (command === 'automa.workflow.list') {
    const workflows = await getAutomaWorkflows()
    lastResult.value = `读取到 ${workflows.length} 个工作流`
    return {
      ok: true,
      total: workflows.length,
      workflows,
    }
  }

  if (command === 'automa.workflow.run') {
    const waitResult = Boolean(payload.wait_result ?? payload.waitResult ?? false)
    const result = await runAutomaWorkflow({
      id: payload.id || payload.workflow_id || payload.workflowId,
      publicId: payload.publicId || payload.public_id,
      variables: payload.variables || payload.params || {},
      checkParams: payload.check_params ?? payload.checkParams ?? false,
      executionId: payload.execution_id || payload.executionId || '',
      waitResult,
      timeout: payload.timeout || 300,
      returnData: payload.return_data || payload.returnData || null,
    })
    lastResult.value = waitResult ? `执行完成：${result.status || ''}` : '已提交执行'
    return result
  }


  if (command === 'automa.workflow.open') {
    openAutomaWorkflow(payload.id || payload.workflow_id || payload.workflowId)
    lastResult.value = '已打开工作流'
    return { ok: true }
  }

  throw new Error(`不支持的 Agent 命令: ${command}`)
}

function resolveBrowserId() {
  // Browser id is bound when backend opens this agent page 浏览器 ID 由后端打开执行端页面时写入
  return String(route.query.browser_id || route.query.browserId || '').trim()
}

function updateAutomaInfo(payload) {
  const agent = Array.isArray(payload?.agents) ? payload.agents[0] : payload
  const installedValue = agent?.automa_installed ?? agent?.automaInstalled ?? agent?.installed
  if (installedValue !== undefined) {
    automaChecked.value = true
    automaInstalled.value = Boolean(installedValue)
  }
  automaVersion.value = String(agent?.automa_version || agent?.automaVersion || agent?.version || '').trim()
}

const automaStatusText = computed(() => {
  if (!automaChecked.value) return '检测中'
  if (!automaInstalled.value) return '未检测到'
  return `可用${automaVersion.value ? ` · ${automaVersion.value}` : ''}`
})

const connectionStatusText = computed(() => {
  if (status.value === 'online') return '已连接'
  if (status.value === 'reconnecting') return '重连中'
  if (status.value === 'offline') return '离线'
  return '连接中'
})

const connectionTagType = computed(() => {
  if (status.value === 'online') return 'success'
  if (status.value === 'offline') return 'danger'
  return 'warning'
})

const connectionStateClass = computed(() => {
  if (status.value === 'online') return 'is-success'
  if (status.value === 'offline') return 'is-danger'
  return 'is-warning'
})

const connectionNote = computed(() => {
  if (status.value === 'online') return '可以接收后端命令'
  if (status.value === 'reconnecting') return '正在恢复 WebSocket'
  if (status.value === 'offline') return '请检查后端服务'
  return '正在建立 WebSocket'
})

const automaStateClass = computed(() => {
  if (!automaChecked.value) return 'is-muted'
  return automaInstalled.value ? 'is-success' : 'is-warning'
})

const automaNote = computed(() => {
  if (!automaChecked.value) return '正在探测插件桥接'
  if (automaInstalled.value) return '插件桥接响应正常'
  return '未收到插件桥接响应'
})
</script>

<style scoped>
:global(html),
:global(body),
:global(#app) {
  width: 100%;
  height: 100%;
  min-height: 100%;
}

.agent-page {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  width: 100dvw;
  height: 100dvh;
  overflow: hidden;
  background: #f5f7fa;
}

.agent-console {
  position: fixed;
  top: 50vh;
  left: 50vw;
  top: 50dvh;
  left: 50dvw;
  display: grid;
  gap: 16px;
  width: min(760px, calc(100vw - 64px));
  width: min(760px, calc(100dvw - 64px));
  max-width: 760px;
  padding: 24px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  box-shadow: 0 20px 54px rgba(31, 45, 61, 0.08);
  transform: translate3d(-50%, -50%, 0);
}

.agent-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  gap: 16px;
}

.eyebrow {
  margin: 0 0 8px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  text-transform: uppercase;
}

.agent-hero h1 {
  margin: 0;
  color: #303133;
  font-size: 24px;
}

.agent-hero p {
  max-width: 560px;
  margin: 10px 0 0;
  color: #606266;
  line-height: 1.65;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.status-card {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #eef2f7;
  border-left-width: 4px;
  border-radius: 8px;
}

.status-card span {
  color: #909399;
  font-size: 13px;
  font-weight: 700;
}

.status-card strong {
  overflow: hidden;
  color: #111827;
  font-size: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-card small {
  overflow: hidden;
  color: #606266;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-card.is-success {
  border-left-color: #16a34a;
}

.status-card.is-warning {
  border-left-color: #f59e0b;
}

.status-card.is-danger {
  border-left-color: #dc2626;
}

.status-card.is-muted {
  border-left-color: #94a3b8;
}

.detail-panel {
  display: grid;
  gap: 10px;
}

.detail-row {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  min-width: 0;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid #eef2f7;
  border-radius: 6px;
}

.detail-row span {
  color: #909399;
  font-size: 13px;
}

.detail-row strong {
  overflow: hidden;
  color: #303133;
  font-size: 14px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .agent-console {
    width: calc(100dvw - 32px);
    padding: 18px;
  }

  .agent-hero,
  .status-grid {
    grid-template-columns: 1fr;
  }

  .detail-row {
    grid-template-columns: 1fr;
    gap: 4px;
  }
}
</style>
