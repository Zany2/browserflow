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
            {{ automaInstalled ? '可以接收工作流和任务指令' : '请先安装并启用 Automa 插件' }}
          </small>
        </article>

        <article class="overview-card">
          <span class="overview-label">当前 IP</span>
          <strong>{{ currentIp || '-' }}</strong>
          <small>优先显示后端注册回传的客户端 IP</small>
        </article>

        <article class="overview-card">
          <span class="overview-label">客户端标识</span>
          <strong>{{ browserId || '-' }}</strong>
          <small>{{ roleLabel }}</small>
        </article>
      </section>

      <section class="detail-grid">
        <article class="detail-card">
          <div class="detail-card__header">
            <h2>最近命令</h2>
          </div>

          <el-descriptions border :column="1">
            <el-descriptions-item label="命令">{{ lastCommand || '-' }}</el-descriptions-item>
            <el-descriptions-item label="结果">{{ lastResult || '-' }}</el-descriptions-item>
            <el-descriptions-item label="最后更新">{{ lastCommandTime || '-' }}</el-descriptions-item>
          </el-descriptions>
        </article>

        <article class="detail-card detail-card--messages">
          <div class="detail-card__header">
            <h2>最近消息</h2>
            <span>最多保留 20 条</span>
          </div>

          <div v-if="messageLogs.length === 0" class="message-empty">等待后端消息...</div>
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

    <AppDialog
      v-model="syncDialogVisible"
      title="同步后端工作流到本地 Automa"
      width="980px"
    >
      <div class="sync-dialog">
        <div class="sync-toolbar">
          <el-input
            v-model="syncKeyword"
            clearable
            placeholder="搜索后端工作流名称、ID、来源 IP"
            @clear="handleSyncSearch"
            @keyup.enter="handleSyncSearch"
          />
          <el-button :loading="syncListLoading" @click="handleSyncSearch">刷新列表</el-button>
          <el-button :loading="localWorkflowLoading" @click="loadLocalWorkflows">刷新本地</el-button>
        </div>

        <el-table
          v-loading="syncListLoading"
          class="adaptive-table"
          :data="syncableWorkflows"
          border
          height="420"
          row-key="id"
          @selection-change="handleSyncSelectionChange"
        >
          <el-table-column type="selection" width="40" reserve-selection />
          <el-table-column label="工作流名称" min-width="180">
            <template #default="{ row }">
              {{ row.name || '' }}
            </template>
          </el-table-column>
          <el-table-column label="工作流描述" min-width="180">
            <template #default="{ row }">
              {{ row.description || '' }}
            </template>
          </el-table-column>
          <el-table-column label="本地状态" width="96" align="center">
            <template #default="{ row }">
              <el-tag :type="getLocalWorkflowTagType(row)" effect="plain">
                {{ getLocalWorkflowText(row) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="更新时间" width="160" class-name="nowrap-column">
            <template #default="{ row }">
              {{ formatDate(getServerWorkflowUpdatedAt(row)) }}
            </template>
          </el-table-column>
        </el-table>

        <div class="sync-footer">
          <span class="sync-summary">已选择 {{ selectedSyncIds.length }} 个工作流</span>
          <AppPagination
            v-model:current-page="syncPageNum"
            v-model:page-size="syncPageSize"
            :total="syncTotal"
            layout="total, sizes, prev, pager, next"
          />
        </div>
      </div>

      <template #footer>
        <el-button @click="syncDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="syncing"
          :disabled="selectedSyncIds.length === 0"
          @click="handleSyncSelectedWorkflows"
        >
          同步到本地 Automa
        </el-button>
      </template>
    </AppDialog>
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import { getAutomaWorkflowDetail, listAutomaWorkflows } from '@/services/automa'
import { createAgentSocket } from '@/services/agentWs'
import { checkClient } from '@/services/client'
import { localCache } from '@/utils/storage'
import {
  getAutomaWorkflows,
  getAutomaInfo,
  importAutomaWorkflow,
  openAutomaWorkflow,
  runAutomaWorkflow,
} from '@/services/automaBridge'

const MAX_LOG_COUNT = 20
const CLIENT_AGENT_ID_STORAGE_KEY = 'browserflow_client_agent_id'
const pluginDownloadUrl = import.meta.env.VITE_AUTOMA_PLUGIN_DOWNLOAD_URL || ''

const route = useRoute()
const browserId = resolveClientAgentId()
const token = String(route.query.token || '')
const role = String(route.query.role || 'client_agent')

const status = ref('connecting')
const automaInstalled = ref(false)
const currentIp = ref(String(route.query.ip || ''))
const lastCommand = ref('')
const lastResult = ref('')
const lastCommandTime = ref('')
const blockedReason = ref('')
const messageLogs = ref([])
const syncDialogVisible = ref(false)
const syncKeyword = ref('')
const syncableWorkflows = ref([])
const selectedSyncRows = ref([])
const syncTotal = ref(0)
const syncPageNum = ref(1)
const syncPageSize = ref(10)
const localWorkflowLoading = ref(false)
const syncListLoading = ref(false)
const syncing = ref(false)
const localWorkflowMap = ref(new Map())
const closeSocketHandler = ref(null)
const pluginStateTimer = ref(0)

const connectionHint = computed(() => {
  if (blockedReason.value) return blockedReason.value
  if (status.value === 'online') return '已与后端保持连接'
  if (status.value === 'connecting') return '正在建立连接'
  if (status.value === 'reconnecting') return '连接已断开，正在自动重连'
  return '连接已断开，可手动恢复'
})

const roleLabel = computed(() => {
  return role === 'browser_agent' ? 'Browser Agent 兼容模式' : 'Client Agent'
})

const selectedSyncIds = computed(() => {
  return selectedSyncRows.value.map((item) => getServerWorkflowId(item)).filter(Boolean)
})

onMounted(() => {
  refreshAutomaState()
  connectSocket()
  pluginStateTimer.value = window.setInterval(refreshAutomaState, 2000)
})

onBeforeUnmount(() => {
  if (pluginStateTimer.value) {
    window.clearInterval(pluginStateTimer.value)
  }
  closeSocketHandler.value?.()
})

watch(syncPageNum, () => {
  if (!syncDialogVisible.value) return
  loadSyncableWorkflows()
})

watch(syncPageSize, () => {
  if (!syncDialogVisible.value) return
  loadFirstSyncPage()
})

async function refreshAutomaState() {
  const automaInfo = await getAutomaInfo()
  automaInstalled.value = automaInfo.installed
}

function connectSocket() {
  closeSocketHandler.value?.()
  closeSocketHandler.value = createAgentSocket({
    browserId,
    token,
    role,
    beforeConnect: ensureClientAllowedBeforeConnect,
    getAutomaInstalled: () => automaInstalled.value,
    getAutomaInfo,
    getWorkflows: getAutomaWorkflows,
    onCommand: handleCommand,
    onStatus: (nextStatus) => {
      status.value = nextStatus
      appendLog({
        type: 'system',
        title: '连接状态',
        message: getConnectionText(nextStatus),
      })
    },
    onRegistered: (payload) => {
      const ip = payload?.ip || payload?.client?.ip || payload?.client_ip || payload?.remote_ip || ''
      if (ip) currentIp.value = ip
      appendLog({
        type: 'system',
        title: '注册成功',
        message: JSON.stringify(payload, null, 2),
      })
    },
    onMessage: (payload) => {
      if (payload?.type === 'agent_registered' || payload?.type === 'heartbeat_ack') return
      appendLog({
        type: 'socket',
        title: payload?.type || 'socket.message',
        message: JSON.stringify(payload || {}, null, 2),
      })
    },
    onNoReconnect: handleNoReconnect,
    onError: (error) => {
      lastResult.value = error.message
      appendLog({
        type: 'error',
        title: '连接异常',
        message: error.message,
      })
      appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
    },
  })
}

function disconnectSocket() {
  closeSocketHandler.value?.()
  closeSocketHandler.value = null
  status.value = 'offline'
  blockedReason.value = ''
  appendLog({
    type: 'system',
    title: '连接操作',
    message: '已手动断开连接',
  })
}

function reconnectSocket() {
  blockedReason.value = ''
  appendLog({
    type: 'system',
    title: '连接操作',
    message: '正在重新连接后端',
  })
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
    lastResult.value = reason
    if (shouldNotify) {
      appendLog({
        type: 'error',
        title: '连接被拒绝',
        message: reason,
      })
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
  lastResult.value = reason
  if (shouldNotify) {
    appendLog({
      type: 'error',
      title: '拉黑暂停连接',
      message: reason,
    })
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: reason })
  }
}

async function handleCommand(command, payload) {
  lastCommand.value = command
  lastCommandTime.value = new Date().toLocaleString()
  appendLog({
    type: 'command',
    title: command,
    message: JSON.stringify(payload || {}, null, 2),
  })

  if (command === 'automa.workflow.list') {
    const workflows = await getAutomaWorkflows()
    lastResult.value = `已读取 ${workflows.length} 个工作流`
    return {
      ok: true,
      total: workflows.length,
      workflows,
    }
  }

  if (command === 'automa.workflow.open') {
    const workflowId = payload.id || payload.workflowId || payload.workflow_id || ''
    openAutomaWorkflow(workflowId)
    lastResult.value = `已打开工作流：${workflowId || '-'}`
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

    lastResult.value = waitResult
      ? `????????${result.status || '-'}`
      : taskId
        ? `????????${taskName || taskId}`
        : '??????????'

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

async function openSyncDialog() {
  if (!automaInstalled.value) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请先安装并启用 Automa 插件' })
    return
  }

  syncDialogVisible.value = true
  await Promise.all([loadSyncableWorkflows(), loadLocalWorkflows()])
}

async function loadSyncableWorkflows() {
  syncListLoading.value = true

  try {
    const data = await listAutomaWorkflows({
      keyword: syncKeyword.value.trim(),
      page_num: syncPageNum.value,
      page_size: syncPageSize.value,
      syncable: 1,
    })
    const workflowList = normalizeList(data, 'workflows').filter((item) => !item.is_deleted)
    syncableWorkflows.value = await enrichSyncableWorkflowHashes(workflowList)
    syncTotal.value = Number(data?.total || 0)
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '读取后端工作流失败' })
  } finally {
    syncListLoading.value = false
  }
}

async function loadLocalWorkflows() {
  localWorkflowLoading.value = true

  try {
    const workflows = await getAutomaWorkflows()
    localWorkflowMap.value = await buildLocalWorkflowMap(workflows)
  } catch (error) {
    appendLog({
      type: 'error',
      title: '读取本地工作流失败',
      message: error.message,
    })
  } finally {
    localWorkflowLoading.value = false
  }
}

function handleSyncSelectionChange(selection) {
  selectedSyncRows.value = selection
}

function handleSyncSearch() {
  loadFirstSyncPage()
}

function loadFirstSyncPage() {
  if (syncPageNum.value === 1) {
    loadSyncableWorkflows()
    return
  }

  syncPageNum.value = 1
}

async function handleSyncSelectedWorkflows() {
  syncing.value = true

  try {
    let syncedCount = 0
    for (const row of selectedSyncRows.value) {
      const workflowId = getServerWorkflowId(row)
      const data = await getAutomaWorkflowDetail(workflowId)
      const workflow = data?.workflow || data
      const result = await importAutomaWorkflow(extractWorkflowImportPayload(workflow))
      syncedCount += 1
      appendLog({
        type: 'system',
        title: '同步工作流',
        message: `已同步 ${result?.name || workflow?.name || workflowId}`,
      })
    }

    await loadLocalWorkflows()
    appMessage({ type: APP_MESSAGE_TYPE.success, message: `已同步 ${syncedCount} 个工作流` })
    syncDialogVisible.value = false
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '同步工作流失败' })
  } finally {
    syncing.value = false
  }
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

function normalizeList(data, fallbackKey) {
  const list = data?.list || data?.[fallbackKey] || []
  return Array.isArray(list) ? list : []
}

async function enrichSyncableWorkflowHashes(workflows) {
  return Promise.all(workflows.map(async (workflow) => {
    try {
      const detail = await getAutomaWorkflowDetail(getServerWorkflowId(workflow))
      const detailWorkflow = parseWorkflowPayload(detail?.raw_json || detail?.rawJson || detail?.normalized_json || detail?.normalizedJson)
      return {
        ...workflow,
        content_hash: workflow.content_hash || detail?.content_hash || detail?.contentHash || '',
        serverContentHash: detailWorkflow ? await createWorkflowContentHash(detailWorkflow) : '',
        updated_at_automa:
          workflow.updated_at_automa || detail?.updated_at_automa || detail?.updatedAtAutoma || 0,
      }
    } catch {
      return workflow
    }
  }))
}

function extractWorkflowImportPayload(workflow) {
  if (!workflow) {
    throw new Error('工作流详情为空，无法同步')
  }

  const rawWorkflow =
    workflow.raw_json ||
    workflow.rawJson ||
    workflow.normalized_json ||
    workflow.normalizedJson ||
    workflow.data

  const parsedWorkflow = parseWorkflowPayload(rawWorkflow)
  const payload = parsedWorkflow || {
    id: workflow.automa_id || workflow.workflow_id || workflow.id,
    name: workflow.name,
    description: workflow.description,
    drawflow: workflow.drawflow_json || workflow.drawflow || {},
    settings: workflow.settings_json || workflow.settings || {},
    table: workflow.table_json || workflow.table || [],
    dataColumns: workflow.data_columns_json || workflow.dataColumns || [],
    trigger: workflow.trigger_json || workflow.trigger || null,
    globalData: workflow.global_data || '',
  }

  const automaId = workflow.automa_id || workflow.automaId || workflow.workflow_id || payload.id
  if (automaId) {
    payload.id = automaId
  }

  const createdAt = resolveAutomaTimestamp(
    workflow.created_at_automa,
    workflow.createdAtAutoma,
    payload.createdAt,
  )
  const updatedAt = resolveAutomaTimestamp(
    workflow.updated_at_automa,
    workflow.updatedAtAutoma,
    payload.updatedAt,
  )
  if (createdAt) {
    payload.createdAt = createdAt
  }
  if (updatedAt) {
    payload.updatedAt = updatedAt
  }
  if (!payload.table && payload.dataColumns) {
    payload.table = payload.dataColumns
  }

  return payload
}

function parseWorkflowPayload(value) {
  if (!value) return null
  if (typeof value === 'object') return { ...value }
  if (typeof value !== 'string') return null

  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

function resolveAutomaTimestamp(...values) {
  for (const value of values) {
    if (!value) continue

    const numberValue = Number(value)
    if (Number.isFinite(numberValue) && numberValue > 0) return numberValue

    const dateValue = new Date(value).getTime()
    if (!Number.isNaN(dateValue)) return dateValue
  }

  return 0
}

function getServerWorkflowId(row) {
  return row?.id || row?.server_id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function getWorkflowDisplayId(row) {
  return row?.automa_id || row?.workflow_id || row?.workflowId || row?.id || ''
}

function getLocalWorkflowId(row) {
  return row?.id || row?.automaId || row?.automa_id || row?.workflowId || ''
}

function getLocalWorkflowText(row) {
  return getLocalWorkflowStatus(row).text
}

function getLocalWorkflowTagType(row) {
  return getLocalWorkflowStatus(row).type
}

function getLocalWorkflowStatus(row) {
  const localWorkflow = getMatchedLocalWorkflow(row)
  if (!localWorkflow) {
    return { text: '待同步', type: 'info' }
  }

  const serverHash = String(row?.serverContentHash || row?.content_hash || row?.contentHash || '').trim()
  const localHash = String(localWorkflow.contentHash || '').trim()
  if (serverHash && localHash && serverHash === localHash) {
    return { text: '已同步', type: 'success' }
  }
  if (serverHash && localHash) {
    return { text: '本地有差异', type: 'warning' }
  }

  return { text: '本地已存在', type: 'primary' }
}

function getMatchedLocalWorkflow(row) {
  return localWorkflowMap.value.get(getWorkflowDisplayId(row)) || null
}

async function buildLocalWorkflowMap(workflows) {
  const workflowMap = new Map()
  await Promise.all(workflows.map(async (workflow) => {
    const workflowId = getLocalWorkflowId(workflow)
    if (workflowId) {
      workflowMap.set(workflowId, {
        ...workflow,
        contentHash: await createWorkflowContentHash(workflow),
      })
    }
  }))

  return workflowMap
}

async function createWorkflowContentHash(workflow) {
  const coreWorkflowData = {
    id: normalizeHashText(workflow?.id),
    name: normalizeHashText(workflow?.name),
    icon: normalizeHashText(workflow?.icon),
    table: workflow?.table ?? workflow?.dataColumns ?? [],
    drawflow: normalizeHashDrawflow(parseHashJsonValue(workflow?.drawflow)),
    settings: workflow?.settings ?? {},
    globalData: workflow?.globalData ?? '',
    description: normalizeHashText(workflow?.description),
  }
  return sha256Hex(stableStringify(coreWorkflowData))
}

function parseHashJsonValue(value) {
  if (value === undefined) return null
  if (typeof value !== 'string') return value
  const text = value.trim()
  if (!text) return value

  try {
    return JSON.parse(text)
  } catch {
    return value
  }
}

function normalizeHashText(value) {
  return String(value || '').trim()
}

function normalizeHashDrawflow(value) {
  if (!value || typeof value !== 'object') return value

  const drawflow = cloneHashValue(value)
  if (Array.isArray(drawflow.edges)) {
    drawflow.edges = drawflow.edges.map(normalizeHashEdge)
  }

  return drawflow
}

function normalizeHashEdge(edge) {
  if (!edge || typeof edge !== 'object') return edge

  const nextEdge = { ...edge }
  delete nextEdge.sourceNode
  delete nextEdge.targetNode
  return nextEdge
}

function cloneHashValue(value) {
  if (Array.isArray(value)) return value.map(cloneHashValue)
  if (value && typeof value === 'object') {
    return Object.keys(value).reduce((nextValue, key) => {
      nextValue[key] = cloneHashValue(value[key])
      return nextValue
    }, {})
  }

  return value
}

function stableStringify(value) {
  if (value === undefined) return 'null'
  if (Array.isArray(value)) {
    return `[${value.map((item) => stableStringify(item)).join(',')}]`
  }
  if (value && typeof value === 'object') {
    return `{${Object.keys(value)
      .sort()
      .map((key) => `${goJsonStringify(key)}:${stableStringify(value[key])}`)
      .join(',')}}`
  }
  return typeof value === 'string' ? goJsonStringify(value) : JSON.stringify(value)
}

function goJsonStringify(value) {
  return JSON.stringify(value)
    .replace(/</g, '\\u003c')
    .replace(/>/g, '\\u003e')
    .replace(/&/g, '\\u0026')
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029')
}

async function sha256Hex(value) {
  if (window.crypto?.subtle) {
    try {
      const bytes = new TextEncoder().encode(value)
      const hashBuffer = await window.crypto.subtle.digest('SHA-256', bytes)
      return Array.from(new Uint8Array(hashBuffer))
        .map((item) => item.toString(16).padStart(2, '0'))
        .join('')
    } catch {
      return sha256HexFallback(value)
    }
  }

  return sha256HexFallback(value)
}

function sha256HexFallback(value) {
  const bytes = new TextEncoder().encode(value)
  const words = bytesToSha256Words(bytes)
  const bitLength = bytes.length * 8
  words[bitLength >> 5] |= 0x80 << (24 - (bitLength % 32))
  words[(((bitLength + 64) >> 9) << 4) + 15] = bitLength

  const hash = [
    0x6a09e667,
    0xbb67ae85,
    0x3c6ef372,
    0xa54ff53a,
    0x510e527f,
    0x9b05688c,
    0x1f83d9ab,
    0x5be0cd19,
  ]
  const constants = [
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
  ]

  for (let i = 0; i < words.length; i += 16) {
    const chunk = words.slice(i, i + 16)
    const state = hash.slice()
    for (let j = 0; j < 64; j += 1) {
      if (j >= 16) {
        const s0 = rotateRight(chunk[j - 15], 7) ^ rotateRight(chunk[j - 15], 18) ^ (chunk[j - 15] >>> 3)
        const s1 = rotateRight(chunk[j - 2], 17) ^ rotateRight(chunk[j - 2], 19) ^ (chunk[j - 2] >>> 10)
        chunk[j] = add32(chunk[j - 16], s0, chunk[j - 7], s1)
      }
      const s1 = rotateRight(state[4], 6) ^ rotateRight(state[4], 11) ^ rotateRight(state[4], 25)
      const ch = (state[4] & state[5]) ^ (~state[4] & state[6])
      const temp1 = add32(state[7], s1, ch, constants[j], chunk[j])
      const s0 = rotateRight(state[0], 2) ^ rotateRight(state[0], 13) ^ rotateRight(state[0], 22)
      const maj = (state[0] & state[1]) ^ (state[0] & state[2]) ^ (state[1] & state[2])
      const temp2 = add32(s0, maj)

      state[7] = state[6]
      state[6] = state[5]
      state[5] = state[4]
      state[4] = add32(state[3], temp1)
      state[3] = state[2]
      state[2] = state[1]
      state[1] = state[0]
      state[0] = add32(temp1, temp2)
    }

    for (let j = 0; j < 8; j += 1) {
      hash[j] = add32(hash[j], state[j])
    }
  }

  return hash.map((item) => (item >>> 0).toString(16).padStart(8, '0')).join('')
}

function bytesToSha256Words(bytes) {
  const words = []
  bytes.forEach((byte, index) => {
    words[index >> 2] = (words[index >> 2] || 0) | (byte << (24 - (index % 4) * 8))
  })
  return words
}

function rotateRight(value, shift) {
  return (value >>> shift) | (value << (32 - shift))
}

function add32(...values) {
  return values.reduce((sum, value) => (sum + (value | 0)) | 0, 0)
}

function getServerWorkflowUpdatedAt(row) {
  return row?.updated_at_automa || row?.updatedAtAutoma || row?.updated_at || row?.updatedAt
}

function getConnectionText(value) {
  if (value === 'online') return '已连接'
  if (value === 'connecting') return '连接中'
  if (value === 'reconnecting') return '自动重连中'
  return '已断开'
}

function getConnectionTagType(value) {
  if (value === 'online') return 'success'
  if (value === 'connecting' || value === 'reconnecting') return 'warning'
  return 'info'
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

function formatDate(value) {
  if (!value) return ''
  const timestamp = Number(value)
  const date = new Date(Number.isFinite(timestamp) ? normalizeTimestamp(timestamp) : value)
  if (Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = padDatePart(date.getMonth() + 1)
  const day = padDatePart(date.getDate())
  const hour = padDatePart(date.getHours())
  const minute = padDatePart(date.getMinutes())
  const second = padDatePart(date.getSeconds())
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

function normalizeTimestamp(value) {
  return value > 0 && value < 10000000000 ? value * 1000 : value
}

function padDatePart(value) {
  return String(value).padStart(2, '0')
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
  margin: 0;
  overflow: auto;
  color: #334155;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.sync-dialog {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sync-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: 12px;
}

.sync-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.sync-footer :deep(.app-pagination) {
  justify-content: flex-end;
  margin-top: 0;
}

.sync-summary {
  color: #64748b;
  white-space: nowrap;
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

  .overview-grid,
  .sync-toolbar,
  .sync-footer {
    grid-template-columns: 1fr;
  }

  .sync-footer {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
