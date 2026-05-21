<template>
  <main class="client-agent-page">
    <section class="client-agent-shell">
      <header class="shell-header">
        <div class="header-copy">
          <h1>客户端执行页</h1>
          <p>页面保持 WebSocket 长连接，接收后端指令并调度本地 Automa 工作流。</p>
        </div>

        <div class="header-actions">
          <el-button @click="refreshAutomaState">检查插件</el-button>
          <el-button type="primary" :disabled="!automaInstalled" @click="openSyncDialog">
            同步工作流
          </el-button>
          <el-button :disabled="status !== 'online'" @click="disconnectSocket">断开连接</el-button>
          <el-button type="success" :disabled="status === 'online'" @click="reconnectSocket">
            恢复连接
          </el-button>
        </div>
      </header>

      <section class="overview-grid">
        <article class="overview-card">
          <span class="overview-label">连接状态</span>
          <el-tag :type="getConnectionTagType(status)" effect="plain">
            {{ getConnectionText(status) }}
          </el-tag>
          <small>{{ connectionHint }}</small>
        </article>

        <article class="overview-card">
          <span class="overview-label">Automa 插件</span>
          <el-tag :type="automaInstalled ? 'success' : 'danger'" effect="plain">
            {{ automaInstalled ? '已安装' : '未安装' }}
          </el-tag>
          <a
            v-if="!automaInstalled && pluginDownloadUrl"
            class="plugin-link"
            :href="pluginDownloadUrl"
            target="_blank"
            rel="noreferrer"
          >
            下载插件
          </a>
          <small v-else>
            {{ automaStatusHint }}
          </small>
        </article>

        <article class="overview-card">
          <span class="overview-label">当前 IP</span>
          <strong>{{ currentIp || '' }}</strong>
          <small>优先显示后端注册回传的客户端 IP</small>
        </article>

        <article class="overview-card">
          <span class="overview-label">客户端标识</span>
          <strong>{{ browserId || '' }}</strong>
          <small>{{ roleLabel }}</small>
        </article>
      </section>

      <section class="detail-grid">
        <article class="detail-card">
          <div class="detail-card__header">
            <h2>执行概览</h2>
            <span>只汇总后端下发的任务</span>
          </div>

          <el-descriptions class="task-overview-descriptions" border :column="1" label-width="88px">
            <el-descriptions-item label="最后任务">
              <span class="single-line-text" :title="latestTaskName">{{ latestTaskName || '' }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="最后状态">
              <el-tag v-if="latestTaskStatus" :type="getTaskStatusTagType(latestTaskStatus)" effect="plain">
                {{ getTaskStatusText(latestTaskStatus) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最后更新">
              <span class="single-line-text" :title="latestTaskUpdatedAt">{{ latestTaskUpdatedAt || '' }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </article>

        <article class="detail-card detail-card--messages">
          <div class="detail-card__header">
            <h2>最新任务消息</h2>
            <span>仅保留最近 10 条任务下发和最终状态</span>
          </div>

          <div v-if="messageLogs.length === 0" class="message-empty">等待后端任务...</div>
          <div v-else class="message-list">
            <article v-for="item in messageLogs" :key="item.id" class="message-item">
              <div class="message-item__head">
                <strong>{{ item.title }}</strong>
                <span>{{ item.time }}</span>
              </div>
              <el-tag :type="getLogTagType(item.type)" size="small" effect="plain">
                {{ getLogTypeText(item.type) }}
              </el-tag>
              <pre>{{ item.message }}</pre>
            </article>
          </div>
        </article>
      </section>
    </section>

    <ClientAgentSyncDialog v-model="syncDialogVisible" />
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import { createAgentSocket } from '@/services/agentWs'
import { checkClient } from '@/services/client'
import { localCache } from '@/utils/storage'
import ClientAgentSyncDialog from './components/ClientAgentSyncDialog.vue'
import {
  getAutomaWorkflows,
  getAutomaInfo,
  openAutomaWorkflow,
  runAutomaWorkflow,
} from '@/services/automaBridge'

const MAX_LOG_COUNT = 10
const CLIENT_AGENT_ID_STORAGE_KEY = 'browserflow_client_agent_id'
const pluginDownloadUrl = import.meta.env.VITE_AUTOMA_PLUGIN_DOWNLOAD_URL || ''

const route = useRoute()
const browserId = resolveClientAgentId()
const token = String(route.query.token || '')
const role = String(route.query.role || 'client_agent')

const status = ref('connecting')
const automaInstalled = ref(false)
const automaCheckedAt = ref('')
const currentIp = ref(String(route.query.ip || ''))
const latestTaskName = ref('')
const latestTaskStatus = ref('')
const latestTaskUpdatedAt = ref('')
const blockedReason = ref('')
const messageLogs = ref([])
const syncDialogVisible = ref(false)
const closeSocketHandler = ref(null)

const connectionHint = computed(() => {
  if (blockedReason.value) return blockedReason.value
  if (status.value === 'online') return '已与后端保持连接'
  if (status.value === 'connecting') return '正在建立连接'
  if (status.value === 'reconnecting') return '连接已断开，正在自动重连'
  if (status.value === 'manual_offline') return '已手动断开连接，可点击恢复连接'
  return '连接已断开，可手动恢复'
})

const automaStatusHint = computed(() => {
  const baseText = automaInstalled.value ? '可以接收工作流和任务指令' : '请先安装并启用 Automa 插件'
  return automaCheckedAt.value ? `${baseText}，最后检测：${automaCheckedAt.value}` : baseText
})

const roleLabel = computed(() => {
  return role === 'browser_agent' ? 'Browser Agent 兼容模式' : '目前仅做展示使用'
})

onMounted(() => {
  refreshAutomaState()
  connectSocket()
})

onBeforeUnmount(() => {
  closeSocketHandler.value?.()
})

async function refreshAutomaState() {
  const automaInfo = await getAutomaInfo()
  automaInstalled.value = automaInfo.installed
  automaCheckedAt.value = new Date().toLocaleTimeString()
  return automaInfo
}

function connectSocket() {
  closeSocketHandler.value?.()
  closeSocketHandler.value = createAgentSocket({
    browserId,
    token,
    role,
    enableHeartbeat: true,
    enableWorkflowInventory: true,
    enableAutomaStatusPolling: false,
    reportAutomaStatusOnWindowLoad: true,
    beforeConnect: ensureClientAllowedBeforeConnect,
    getAutomaInstalled: () => automaInstalled.value,
    getAutomaInfo: refreshAutomaState,
    getWorkflows: getAutomaWorkflows,
    onCommand: handleCommand,
    onCommandResult: handleCommandResult,
    onStatus: (nextStatus) => {
      status.value = nextStatus
    },
    onRegistered: (payload) => {
      const ip = payload?.ip || payload?.client?.ip || payload?.client_ip || payload?.remote_ip || ''
      if (ip) currentIp.value = ip
    },
    onMessage: () => {},
    onNoReconnect: handleNoReconnect,
    onError: (error) => {
      appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
    },
  })
}

function disconnectSocket() {
  closeSocketHandler.value?.()
  closeSocketHandler.value = null
  status.value = 'manual_offline'
  blockedReason.value = ''
}

function reconnectSocket() {
  blockedReason.value = ''
  connectSocket()
}

async function ensureClientAllowedBeforeConnect() {
  const data = await checkClient({
    client_id: browserId,
    _t: Date.now(),
  })
  if (data?.allowed === false || data?.is_banned) {
    const reason = data?.reason || '当前客户端 IP 已被拉黑，将持续检测，解除拉黑后自动重连'
    const shouldNotify = blockedReason.value !== reason
    blockedReason.value = reason
    if (shouldNotify) {
      appMessage({ type: APP_MESSAGE_TYPE.warning, message: reason })
    }
    return false
  }

  blockedReason.value = ''
  return true
}

function handleNoReconnect(payload) {
  const reason =
    payload?.error ||
    payload?.data?.reason ||
    payload?.message ||
    '当前客户端已被拉黑，将持续检测，解除拉黑后自动重连'
  const shouldNotify = blockedReason.value !== reason
  blockedReason.value = reason
  if (shouldNotify) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: reason })
  }
}

async function handleCommand(command, payload) {
  if (isTaskCommand(command)) {
    updateTaskOverview(payload, '已下发')
    appendTaskLog({
      type: 'command',
      title: '收到后端任务',
      command,
      payload,
    })
  }

  if (command === 'automa.workflow.list') {
    const workflows = await getAutomaWorkflows()
    return {
      ok: true,
      total: workflows.length,
      workflows,
    }
  }

  if (command === 'automa.workflow.open') {
    const workflowId = payload.id || payload.workflowId || payload.workflow_id || ''
    openAutomaWorkflow(workflowId)
    return { ok: true, workflow_id: workflowId }
  }

  if (command === 'automa.workflow.run' || command === 'task.execute' || command === 'task.run') {
    const workflowId = payload.id || payload.workflowId || payload.workflow_id || ''
    const publicId = payload.publicId || payload.public_id || ''
    const variables = payload.variables || payload.params || {}
    const taskId = payload.task_id || ''
    const taskName = payload.task_name || ''
    const waitResult = Boolean(payload.wait_result ?? payload.waitResult ?? false)

    const result = await runAutomaWorkflow({
      id: workflowId,
      publicId,
      variables,
      checkParams: payload.check_params ?? payload.checkParams ?? false,
      executionId: payload.execution_id || payload.executionId || taskId || '',
      waitResult,
      timeout: payload.timeout || 300,
      returnData: payload.return_data || payload.returnData || null,
    })

    updateTaskOverview(payload, waitResult ? resolveTaskStatus(result, true) : '执行中')

    return {
      ...result,
      task_id: taskId,
      task_name: taskName,
      workflow_id: workflowId,
      public_id: publicId,
      variables,
    }
  }


  throw new Error(`不支持的客户端命令: ${command}`)
}

function handleCommandResult({ command, payload, result, success, error, async: asyncResult }) {
  if (!isTaskCommand(command)) return

  const status = resolveTaskStatus(result, success)
  const isIntermediateStatus = !asyncResult && success && isIntermediateTaskStatus(status)
  updateTaskOverview(payload, isIntermediateStatus ? '执行中' : status)
  if (isIntermediateStatus) return

  const title = success ? (asyncResult ? '任务最终结果' : '任务执行结果') : '任务执行失败'
  const message = success
    ? formatTaskResultMessage(command, payload, result)
    : `${formatTaskName(payload)}执行失败：${error || '未知错误'}`
  appendTaskLog({
    type: success ? 'command' : 'error',
    title,
    command,
    payload,
    result,
    message,
  })
}

async function openSyncDialog() {
  if (!automaInstalled.value) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请先安装并启用 Automa 插件' })
    return
  }

  syncDialogVisible.value = true
}

function appendTaskLog({ type, title, command, payload, result, message }) {
  appendLog({
    type,
    title,
    message: message || JSON.stringify({
      command,
      task: formatTaskName(payload),
      payload: payload || {},
      result: result || null,
    }, null, 2),
  })
}

function updateTaskOverview(payload = {}, statusText = '') {
  latestTaskName.value = formatTaskName(payload)
  latestTaskStatus.value = statusText
  latestTaskUpdatedAt.value = new Date().toLocaleString()
}

function isTaskCommand(command) {
  return command === 'automa.workflow.run' || command === 'task.execute' || command === 'task.run'
}

function resolveTaskStatus(result = {}, success = true) {
  if (!success) return 'failed'
  const status = String(result?.status || '').trim()
  if (status) return status
  return result?.ok === false ? 'failed' : 'success'
}

function isIntermediateTaskStatus(status = '') {
  return ['queued', 'submitted', 'pending', 'running'].includes(String(status).toLowerCase())
}

function formatTaskName(payload = {}) {
  return (
    payload.task_name ||
    payload.taskName ||
    payload.workflow_name ||
    payload.workflowName ||
    payload.workflow_id ||
    payload.workflowId ||
    payload.id ||
    payload.public_id ||
    payload.publicId ||
    '未命名任务'
  )
}

function formatTaskResultMessage(command, payload = {}, result = {}) {
  const status = result?.status || (result?.ok === false ? 'failed' : 'success')
  const taskName = formatTaskName(payload)
  return JSON.stringify({
    command,
    task: taskName,
    status,
    result,
  }, null, 2)
}

function appendLog({ type, title, message }) {
  messageLogs.value = [
    {
      id: `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
      type,
      title,
      message,
      time: new Date().toLocaleTimeString(),
    },
    ...messageLogs.value,
  ].slice(0, MAX_LOG_COUNT)
}

function getConnectionText(value) {
  if (value === 'online') return '已连接'
  if (value === 'connecting') return '连接中'
  if (value === 'reconnecting') return '自动重连中'
  if (value === 'manual_offline') return '手动断开'
  return '已断开'
}

function getConnectionTagType(value) {
  if (value === 'online') return 'success'
  if (value === 'connecting' || value === 'reconnecting') return 'warning'
  if (value === 'manual_offline') return 'info'
  return 'info'
}

function getTaskStatusTagType(value) {
  const statusText = String(value || '').toLowerCase()
  if (['success', 'completed', 'complete', 'done'].includes(statusText)) return 'success'
  if (['failed', 'failure', 'error', 'timeout'].includes(statusText)) return 'danger'
  if (['已下发', '执行中', 'queued', 'submitted', 'pending', 'running'].includes(statusText)) return 'warning'
  return 'info'
}

function getTaskStatusText(value) {
  const statusText = String(value || '')
  const lowerStatusText = statusText.toLowerCase()
  if (['success', 'completed', 'complete', 'done'].includes(lowerStatusText)) return '执行成功'
  if (['failed', 'failure', 'error'].includes(lowerStatusText)) return '执行失败'
  if (lowerStatusText === 'timeout') return '执行超时'
  if (['queued', 'submitted', 'pending', 'running'].includes(lowerStatusText)) return '执行中'
  return statusText
}

function getLogTagType(type) {
  if (type === 'error') return 'danger'
  if (type === 'command') return 'success'
  if (type === 'socket') return 'warning'
  return 'info'
}

function getLogTypeText(type) {
  if (type === 'error') return '异常'
  if (type === 'command') return '命令'
  if (type === 'socket') return '消息'
  return '系统'
}

function resolveClientAgentId() {
  const cachedValue = String(localCache.get(CLIENT_AGENT_ID_STORAGE_KEY, '') || '').trim()
  if (cachedValue) return cachedValue

  // Persist generated id so reconnects keep the same logical client.
  const generatedValue = `client_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
  localCache.set(CLIENT_AGENT_ID_STORAGE_KEY, generatedValue)
  return generatedValue
}

</script>

<style scoped lang="scss">
:global(html),
:global(body),
:global(#app) {
  width: 100%;
  height: 100%;
  min-height: 100%;
  overflow: hidden;
}

.client-agent-page {
  position: fixed;
  inset: 0;
  overflow: hidden;
  background:
    radial-gradient(circle at top right, rgba(14, 116, 144, 0.14), transparent 28%),
    linear-gradient(180deg, #f7fafc 0%, #edf3f8 100%);
}

.client-agent-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: min(1320px, calc(100vw - 32px));
  height: calc(100dvh - 32px);
  margin: 16px auto;
  padding: 20px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid #d9e2ec;
  border-radius: 20px;
  box-shadow: 0 20px 48px rgba(15, 23, 42, 0.08);
}

.shell-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.header-copy h1 {
  margin: 0;
  color: #0f172a;
  font-size: 28px;
}

.header-copy p {
  margin: 8px 0 0;
  color: #64748b;
}

.header-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  flex-shrink: 0;
}

.overview-card,
.detail-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  padding: 18px;
  background: #f8fbfd;
  border: 1px solid #dde7f0;
  border-radius: 16px;
}

.overview-label {
  color: #64748b;
  font-size: 13px;
}

.overview-card strong {
  color: #0f172a;
  font-size: 20px;
}

.overview-card small {
  color: #64748b;
}

.plugin-link {
  color: #0f766e;
  font-weight: 600;
}

.detail-grid {
  display: grid;
  grid-template-columns: 400px minmax(0, 1fr);
  gap: 16px;
  min-height: 0;
  flex: 1;
}

.detail-card {
  min-height: 0;
}

.task-overview-descriptions {
  :deep(.el-descriptions__label) {
    width: 88px;
    min-width: 88px;
    white-space: nowrap;
  }

  :deep(.el-descriptions__content) {
    min-width: 0;
    text-align: left;
  }
}

.single-line-text {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.detail-card__header h2 {
  margin: 0;
  color: #0f172a;
  font-size: 18px;
}

.detail-card__header span {
  color: #64748b;
  font-size: 13px;
}

.detail-card--messages {
  overflow: hidden;
}

.message-empty {
  display: grid;
  place-items: center;
  flex: 1;
  min-height: 160px;
  color: #64748b;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #ffffff;
}

.message-list {
  display: grid;
  gap: 12px;
  overflow: auto;
  padding-right: 4px;
}

.message-item {
  display: grid;
  gap: 10px;
  padding: 14px 16px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
}

.message-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.message-item__head strong {
  color: #0f172a;
}

.message-item__head span {
  color: #64748b;
  font-size: 12px;
}

.message-item pre {
  max-height: 180px;
  margin: 0;
  overflow: auto;
  color: #334155;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 1080px) {
  .overview-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .client-agent-shell {
    width: calc(100vw - 16px);
    height: calc(100dvh - 16px);
    margin: 8px auto;
    padding: 16px;
  }

  .shell-header,
  .header-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .overview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
