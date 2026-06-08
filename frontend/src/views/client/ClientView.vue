<template>
  <section class="client-page server-list-page">
    <header class="page-actions server-list-actions">
      <el-button type="primary" :icon="RefreshRight" @click="loadClients">刷新</el-button>
    </header>

    <section class="client-panel server-list-panel">
      <div class="client-filters server-list-filters">
        <div class="filter-item filter-item--keyword">
          <span class="filter-label">关键字</span>
          <el-input v-model="keywordFilter" clearable placeholder="客户端 IP、执行节点、客户端名称" />
        </div>

        <div class="filter-item filter-item--status">
          <span class="filter-label">状态</span>
          <el-select v-model="statusFilter" clearable placeholder="全部">
            <el-option label="全部" value="" />
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="已拉黑" value="banned" />
          </el-select>
        </div>

        <el-button @click="resetFilters">重置</el-button>
        <el-button type="warning" :disabled="selectedClientIds.length === 0 || batchOfflineLoading"
          :loading="batchOfflineLoading" @click="handleBatchOffline">
          下线重连
        </el-button>
        <el-button type="danger" :disabled="selectedClientIds.length === 0 || batchBanLoading"
          :loading="batchBanLoading" @click="handleBatchBan">
          拉黑
        </el-button>
        <AppSelectionSummary :count="selectedClientIds.length" unit="客户端" />
      </div>

      <el-table ref="clientTableRef" v-loading="loading" class="client-table server-list-table"
        :data="pagedClients" border height="100%" :row-key="getClientId" empty-text="暂无客户端"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column label="客户端 IP" width="124" show-overflow-tooltip>
          <template #default="{ row }">{{ getClientIp(row) || '' }}</template>
        </el-table-column>
        <el-table-column label="执行节点 ID" width="108" show-overflow-tooltip>
          <template #default="{ row }">{{ getNodeId(row) }}</template>
        </el-table-column>
        <el-table-column label="客户端名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ getClientName(row) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="74" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row)" effect="plain">{{ getStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="忙闲状态" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="getBusyStatusTagType(row)" effect="plain">{{ getBusyStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 状态" width="104" align="center">
          <template #default="{ row }">
            <el-tag :type="getAutomaTagType(row)" effect="plain">{{ getAutomaStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 版本" width="96" show-overflow-tooltip>
          <template #default="{ row }">{{ getAutomaVersion(row) }}</template>
        </el-table-column>
        <el-table-column label="浏览器" width="82" show-overflow-tooltip>
          <template #default="{ row }">{{ getBrowserName(row) }}</template>
        </el-table-column>
        <el-table-column label="浏览器版本" width="102" show-overflow-tooltip>
          <template #default="{ row }">{{ getBrowserVersion(row) }}</template>
        </el-table-column>
        <el-table-column label="是否拉黑" width="78" align="center">
          <template #default="{ row }">
            <el-tag :type="isBanned(row) ? 'danger' : 'success'" effect="plain">
              {{ isBanned(row) ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近心跳" width="166" class-name="nowrap-column">
          <template #default="{ row }">{{ formatDate(getLastActiveTime(row)) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="!isBanned(row)" link type="warning" :disabled="isClientActionLoading(row)"
              :loading="isClientActionLoading(row)" @click="handleOffline(row)">
              下线重连
            </el-button>
            <el-button v-if="!isBanned(row)" link type="danger" :disabled="isClientActionLoading(row)"
              :loading="isClientActionLoading(row)" @click="handleBan(row)">
              拉黑
            </el-button>
            <el-button v-else link type="success" :disabled="isClientActionLoading(row)"
              :loading="isClientActionLoading(row)" @click="handleUnban(row)">
              解除拉黑
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <AppPagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="pageSizes"
        :total="clients.length" />
    </section>

    <AppDialog v-model="detailVisible" title="客户端详情" width="min(1040px, calc(100vw - 32px))"
      class="client-detail-dialog" confirm-text="保存" cancel-text="关闭" :loading="detailSaving"
      :confirm-disabled="detailLoading" @confirm="handleSaveDetail">
      <div v-loading="detailLoading" class="detail-form server-detail-form">
        <div v-for="field in detailFields" :key="field.key" class="detail-field server-detail-field"
          :class="{ 'detail-field--wide': field.type === 'textarea' }">
          <span class="detail-label server-detail-label">{{ field.label }}</span>
          <div class="detail-control server-detail-control">
            <el-input v-if="field.editable" v-model="detailForm[field.key]" clearable />
            <div v-else class="detail-value" :class="{ 'detail-value--textarea': field.type === 'textarea' }">
              <span class="detail-text" :class="{ 'detail-text--empty': !field.value }">
                {{ field.value || '' }}
              </span>
              <el-tooltip v-if="field.copyable && field.value" content="复制" placement="top">
                <el-button class="detail-copy" text circle :icon="CopyDocument" :aria-label="`复制${field.label}`"
                  @click="copyDetailValue(field.value)" />
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
    </AppDialog>
  </section>
</template>
<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { CopyDocument, RefreshRight } from '@element-plus/icons-vue'
import AppDialog from '@/components/AppDialog.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { copyText } from '@/utils/browser'
import { formatDate as formatBaseDate, formatEmpty as formatBaseEmpty } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, getSafePage, normalizeList } from '@/utils/list'
import {
  batchBanClients,
  batchOfflineClients,
  banClient,
  getClientDetail,
  listClients,
  offlineClient,
  unbanClient,
  updateClient,
} from '@/services/client'

const clients = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = DEFAULT_PAGE_SIZES
const statusFilter = ref('')
const keywordFilter = ref('')
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailSaving = ref(false)
const detailClient = ref(null)
const detailForm = reactive(createDetailForm())
const clientTableRef = ref(null)
const batchOfflineLoading = ref(false)
const batchBanLoading = ref(false)
const clientActionLoadingIds = ref(new Set())

const pagedClients = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return clients.value.slice(start, start + pageSize.value)
})
const {
  selectedKeys: selectedClientIds,
  handleSelectionChange,
  restoreSelection: restoreClientSelection,
  retainSelectionByRows: retainClientSelectionByRows,
  resetSelection: resetClientSelection,
} = usePagedTableSelection({
  rows: pagedClients,
  getRowKey: getClientId,
})
const {
  run: scheduleFilterSearch,
  cancel: clearFilterSearchTimer,
} = useDebouncedAction(searchClientsNow, 200)

const detailFields = computed(() => {
  const fields = [
    { key: 'client_ip', label: '客户端 IP', value: formatEmpty(getClientIp(detailForm)), copyable: true },
    { key: 'client_id', label: '客户端 ID', value: formatEmpty(detailForm.client_id), hiddenWhenEmpty: true },
    { key: 'client_name', label: '原始客户端名称', value: formatEmpty(detailForm.client_name), hiddenWhenEmpty: true },
    { key: 'node_id', label: '执行节点 ID', value: formatEmpty(detailForm.node_id), copyable: true },
    { key: 'node_index', label: '执行节点序号', value: formatEmpty(detailForm.node_index) },
    { key: 'node_name', label: '执行节点名称', value: formatEmpty(detailForm.node_name), hiddenWhenEmpty: true },
    { key: 'machine_id', label: '机器 ID', value: formatEmpty(detailForm.machine_id), hiddenWhenEmpty: true },
    { key: 'machine_name', label: '机器名称', value: formatEmpty(detailForm.machine_name), hiddenWhenEmpty: true },
    { key: 'display_name', label: '客户端名称', value: formatEmpty(detailForm.display_name), editable: true },
    { key: 'hostname', label: '主机名', value: formatEmpty(detailForm.hostname) },
    { key: 'status', label: '客户端状态', value: getStatusText(detailForm) },
    { key: 'busy_status', label: '节点忙闲状态', value: getBusyStatusText(detailForm) },
    { key: 'current_execution_id', label: '当前执行 ID', value: formatEmpty(detailForm.current_execution_id) },
    { key: 'current_task_record_id', label: '当前任务记录 ID', value: formatEmpty(detailForm.current_task_record_id) },
    { key: 'plugin_status', label: 'Automa 状态', value: getAutomaStatusText(detailForm) },
    { key: 'automa_version', label: 'Automa 版本', value: formatEmpty(getAutomaVersion(detailForm)) },
    { key: 'worker_version', label: 'Worker 版本', value: formatEmpty(detailForm.worker_version) },
    { key: 'browser_name', label: '浏览器名称', value: formatEmpty(getBrowserName(detailForm)) },
    { key: 'browser_version', label: '浏览器版本', value: formatEmpty(getBrowserVersion(detailForm)) },
    { key: 'is_banned', label: '是否拉黑', value: isBanned(detailForm) ? '是' : '否' },
    { key: 'last_seen_at', label: '最近心跳/交互', value: formatDate(getLastActiveTime(detailForm)) },
    { key: 'id', label: '服务端 ID', value: formatEmpty(detailForm.id), copyable: true },
    { key: 'os_name', label: '操作系统', value: formatEmpty(detailForm.os_name) },
    { key: 'os_version', label: '系统版本', value: formatEmpty(detailForm.os_version) },
    { key: 'profile_dir', label: 'Profile 目录', value: formatEmpty(detailForm.profile_dir), type: 'textarea' },
    { key: 'extension_dir', label: '扩展目录', value: formatEmpty(detailForm.extension_dir), type: 'textarea' },
    {
      key: 'capabilities_json',
      label: '节点能力快照',
      value: formatDetailJson(detailForm.capabilities_json),
      type: 'textarea',
    },
    {
      key: 'user_agent',
      label: 'User-Agent',
      value: formatEmpty(detailForm.user_agent),
      type: 'textarea',
      copyable: true,
    },
    { key: 'ban_reason', label: '拉黑原因', value: formatEmpty(detailForm.ban_reason) },
    { key: 'first_seen_at', label: '首次连接时间', value: formatDate(detailForm.first_seen_at) },
    { key: 'connected_at', label: '连接时间', value: formatDate(detailForm.connected_at) },
    { key: 'disconnected_at', label: '断开时间', value: formatDate(detailForm.disconnected_at) },
    { key: 'last_command_at', label: '最近命令下发', value: formatDate(detailForm.last_command_at) },
    { key: 'last_workflow_sync_at', label: '最近工作流同步', value: formatDate(detailForm.last_workflow_sync_at) },
    { key: 'last_lock_renewed_at', label: '最近执行锁续期', value: formatDate(detailForm.last_lock_renewed_at) },
    { key: 'created_at', label: '创建时间', value: formatDate(detailForm.created_at) },
    { key: 'updated_at', label: '更新时间', value: formatDate(detailForm.updated_at) },
  ]

  return fields.filter((field) => !field.hiddenWhenEmpty || field.value)
})

onMounted(() => {
  loadClients()
})

watch(statusFilter, () => {
  clearFilterSearchTimer()
  currentPage.value = 1
  loadClients()
})

watch(keywordFilter, () => {
  scheduleFilterSearch()
})

watch([clients, pageSize], () => {
  currentPage.value = getSafePage({
    total: clients.value.length,
    page: currentPage.value,
    size: pageSize.value,
  })
})

watch(pagedClients, () => {
  restoreClientSelection(clientTableRef)
})

async function loadClients() {
  loading.value = true
  try {
    const data = await listClients({
      status: statusFilter.value,
      keyword: keywordFilter.value.trim(),
    })
    clients.value = normalizeList(data, 'clients')
    retainClientSelectionByRows(clients.value)
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  keywordFilter.value = ''
  statusFilter.value = ''
}

function searchClientsNow() {
  currentPage.value = 1
  loadClients()
}

async function openDetail(row) {
  const clientId = getClientId(row)
  detailClient.value = row
  setDetailForm(row)
  detailVisible.value = true
  if (!clientId) return

  detailLoading.value = true
  try {
    const data = await getClientDetail(clientId)
    detailClient.value = data.client || row
    setDetailForm(detailClient.value)
  } finally {
    detailLoading.value = false
  }
}

async function handleSaveDetail() {
  const clientId = getClientId(detailForm)
  if (!clientId) return

  detailSaving.value = true
  try {
    const displayName = detailForm.display_name.trim()
    const data = await updateClient(clientId, {
      display_name: displayName,
    })
    const updatedClient = data.client || {
      ...detailForm,
      display_name: displayName,
    }

    // Sync detail form after save 保存后同步详情表单
    detailClient.value = updatedClient
    setDetailForm(updatedClient)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已保存' })
    detailVisible.value = false
    await loadClients()
  } finally {
    detailSaving.value = false
  }
}

async function handleOffline(row) {
  const clientId = getClientId(row)
  if (!clientId || isClientActionLoading(row)) return

  setClientActionLoading(clientId, true)
  try {
    await offlineClient(clientId)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '已通知客户端下线重连' })
    await loadClients()
  } finally {
    setClientActionLoading(clientId, false)
  }
}

async function handleBatchOffline() {
  const ids = selectedClientIds.value.slice()
  if (ids.length === 0 || batchOfflineLoading.value) return

  batchOfflineLoading.value = true
  try {
    const result = await batchOfflineClients(ids)
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: `已通知 ${getBatchSuccessCount(result, ids)} 个客户端下线重连`,
    })
    await loadClients()
    resetClientSelection(clientTableRef)
  } finally {
    batchOfflineLoading.value = false
  }
}

async function handleBan(row) {
  const clientId = getClientId(row)
  if (!clientId || isClientActionLoading(row)) return

  setClientActionLoading(clientId, true)
  try {
    await banClient(clientId)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已拉黑' })
    await loadClients()
  } finally {
    setClientActionLoading(clientId, false)
  }
}

async function handleBatchBan() {
  const ids = selectedClientIds.value.slice()
  if (ids.length === 0 || batchBanLoading.value) return

  batchBanLoading.value = true
  try {
    const result = await batchBanClients(ids)
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: `已拉黑 ${getBatchSuccessCount(result, ids)} 个客户端`,
    })
    await loadClients()
    resetClientSelection(clientTableRef)
  } finally {
    batchBanLoading.value = false
  }
}

async function handleUnban(row) {
  const clientId = getClientId(row)
  if (!clientId || isClientActionLoading(row)) return

  setClientActionLoading(clientId, true)
  try {
    await unbanClient(clientId)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已解除拉黑' })
    await loadClients()
  } finally {
    setClientActionLoading(clientId, false)
  }
}

function getBatchSuccessCount(result, fallbackIds) {
  return Number(result?.success || 0) || fallbackIds.length
}

function isClientActionLoading(row) {
  const clientId = getClientId(row)
  return Boolean(clientId && clientActionLoadingIds.value.has(String(clientId)))
}

function setClientActionLoading(clientId, loading) {
  const key = String(clientId || '')
  if (!key) return

  const nextIds = new Set(clientActionLoadingIds.value)
  if (loading) {
    nextIds.add(key)
  } else {
    nextIds.delete(key)
  }
  clientActionLoadingIds.value = nextIds
}

function createDetailForm() {
  return {
    id: '',
    client_id: '',
    client_name: '',
    display_name: '',
    client_ip: '',
    machine_id: '',
    machine_name: '',
    node_id: '',
    node_index: '',
    node_name: '',
    worker_version: '',
    profile_dir: '',
    extension_dir: '',
    status: '',
    online: false,
    busy_status: '',
    current_execution_id: '',
    current_task_record_id: '',
    plugin_status: '',
    automa_status: '',
    automa_installed: undefined,
    automa_version: '',
    browser_name: '',
    browser_version: '',
    browser: '',
    os_name: '',
    os_version: '',
    hostname: '',
    capabilities_json: '',
    user_agent: '',
    banned: false,
    is_banned: false,
    ban_reason: '',
    first_seen_at: '',
    last_seen_at: '',
    last_seen: '',
    last_heartbeat_time: '',
    last_command_at: '',
    last_workflow_sync_at: '',
    last_lock_renewed_at: '',
    connected_at: '',
    disconnected_at: '',
    created_at: '',
    updated_at: '',
  }
}

function setDetailForm(row = {}) {
  Object.assign(detailForm, {
    ...createDetailForm(),
    id: row.id || '',
    client_id: row.client_id || row.clientId || '',
    client_name: row.client_name || row.clientName || '',
    display_name: row.display_name || row.displayName || '',
    client_ip: getClientIp(row),
    machine_id: row.machine_id || row.machineId || '',
    machine_name: row.machine_name || row.machineName || '',
    node_id: row.node_id || row.nodeId || '',
    node_index: row.node_index ?? row.nodeIndex ?? '',
    node_name: row.node_name || row.nodeName || '',
    worker_version: row.worker_version || row.workerVersion || '',
    profile_dir: row.profile_dir || row.profileDir || '',
    extension_dir: row.extension_dir || row.extensionDir || '',
    status: row.status || '',
    online: Boolean(row.online),
    busy_status: row.busy_status || row.busyStatus || '',
    current_execution_id: row.current_execution_id || row.currentExecutionId || '',
    current_task_record_id: row.current_task_record_id || row.currentTaskRecordId || '',
    plugin_status: row.plugin_status || row.pluginStatus || '',
    automa_status: row.automa_status || row.automaStatus || '',
    automa_installed: row.automa_installed ?? row.automaInstalled,
    automa_version: row.automa_version || row.automaVersion || '',
    browser_name: row.browser_name || row.browserName || '',
    browser_version: row.browser_version || row.browserVersion || '',
    browser: row.browser || '',
    os_name: row.os_name || row.osName || '',
    os_version: row.os_version || row.osVersion || '',
    hostname: row.hostname || '',
    capabilities_json: row.capabilities_json ?? row.capabilitiesJson ?? '',
    user_agent: row.user_agent || row.userAgent || '',
    banned: Boolean(row.banned),
    is_banned: Boolean(row.is_banned),
    ban_reason: row.ban_reason || row.banReason || '',
    first_seen_at: row.first_seen_at || row.firstSeenAt || '',
    last_seen_at: row.last_seen_at || row.lastSeenAt || '',
    last_seen: row.last_seen || row.lastSeen || '',
    last_heartbeat_time: row.last_heartbeat_time || row.lastHeartbeatTime || '',
    last_command_at: row.last_command_at || row.lastCommandAt || '',
    last_workflow_sync_at: row.last_workflow_sync_at || row.lastWorkflowSyncAt || '',
    last_lock_renewed_at: row.last_lock_renewed_at || row.lastLockRenewedAt || '',
    connected_at: row.connected_at || row.connectedAt || '',
    disconnected_at: row.disconnected_at || row.disconnectedAt || '',
    created_at: row.created_at || row.createdAt || '',
    updated_at: row.updated_at || row.updatedAt || '',
  })
}

function getClientId(row) {
  return row?.id || row?.client_id || row?.clientId || ''
}

function getClientIp(row) {
  // Client IP reads current IP first 客户端 IP 优先读取当前连接 IP
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function getClientName(row) {
  return row?.display_name || row?.displayName || row?.hostname || row?.client_name || row?.clientName || row?.name || ''
}

function getNodeId(row) {
  return row?.node_id || row?.nodeId || ''
}

function isBanned(row) {
  return Boolean(row?.banned || row?.is_banned || row?.status === 'banned')
}

function getStatusText(row) {
  // Client status keeps online state separate from ban flag 客户端状态与拉黑标记分开展示
  if (row?.online || row?.status === 'online') return '在线'
  if (row?.status === 'offline') return '离线'
  if (row?.status === 'banned') return '已拉黑'
  if (row?.status) return row.status
  return '离线'
}

function getStatusTagType(row) {
  const status = getStatusText(row)
  if (status === '在线') return 'success'
  if (status === '已拉黑') return 'danger'
  return 'info'
}

function getBusyStatusText(row) {
  const status = row?.busy_status || row?.busyStatus || ''
  if (status === 'idle') return '空闲'
  if (status === 'busy') return '执行中'
  if (status === 'unknown') return '未知'
  return status || '未知'
}

function getBusyStatusTagType(row) {
  const status = getBusyStatusText(row)
  if (status === '空闲') return 'success'
  if (status === '执行中') return 'warning'
  return 'info'
}

function getAutomaStatusText(row) {
  // Plugin status normalizes backend enum 插件状态兼容后端枚举值
  const status = row?.plugin_status || row?.automa_status || ''
  if (status === 'installed') return '已安装'
  if (status === 'not_installed') return '未安装'
  if (status === 'disabled') return '已禁用'
  if (status === 'error') return '异常'
  if (status === 'unknown') return '未知'
  if (row?.automa_installed === true) return '已安装'
  if (row?.automa_installed === false) return '未安装'
  return status || '未知'
}

function getAutomaTagType(row) {
  // Tag type mirrors readable Automa status 标签类型匹配可读状态
  const status = getAutomaStatusText(row)
  if (status === '已安装') return 'success'
  if (status === '未安装' || status === '已禁用' || status === '异常') return 'danger'
  return 'info'
}

function getAutomaVersion(row) {
  // Version supports snake and camel case 版本号兼容下划线和驼峰字段
  return row?.automa_version || row?.automaVersion || ''
}

function getBrowserName(row) {
  // Browser name falls back to legacy browser field 浏览器名称兼容旧 browser 字段
  return row?.browser_name || row?.browserName || row?.browser || ''
}

function getBrowserVersion(row) {
  // Browser version supports collected metadata 浏览器版本读取采集元数据
  return row?.browser_version || row?.browserVersion || ''
}

function getLastActiveTime(row) {
  // Last active time prefers heartbeat/interaction field 最近活跃时间优先使用心跳字段
  return (
    row?.last_seen_at ||
    row?.lastSeenAt ||
    row?.last_seen ||
    row?.last_heartbeat_time ||
    row?.lastHeartbeatTime ||
    row?.updated_at ||
    row?.updatedAt
  )
}

function formatEmpty(value, fallback = '') {
  return formatBaseEmpty(value, fallback, { treatDashAsEmpty: true })
}

function formatDetailJson(value) {
  if (!value) return ''
  if (typeof value === 'string') return value

  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function formatDate(value) {
  return formatBaseDate(value, { fallback: '' })
}

async function copyDetailValue(value) {
  await copyText(value)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已复制' })
}
</script>

<style scoped lang="scss">
.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item--status {
  width: 220px;
}

.filter-item--keyword {
  width: 320px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.client-table {
  width: 100%;
}

:deep(.client-table .cell),
:deep(.client-table .cell *) {
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  word-break: normal;
}

:deep(.client-table .el-button) {
  white-space: nowrap !important;
}

:deep(.client-table .el-button + .el-button) {
  margin-left: 4px;
}

:deep(.client-table th .cell),
:deep(.client-table th .cell *) {
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  word-break: keep-all;
}

.client-name {
  display: grid;
  gap: 2px;
  min-width: 0;

  span,
  small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    color: #909399;
  }
}

.detail-field {
  grid-template-columns: 112px minmax(0, 1fr);
}

.detail-form {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 20px;
}

.detail-control {
  grid-template-columns: minmax(0, 1fr);
}

.detail-field--wide {
  grid-column: 1 / -1;
}

.detail-value {
  display: flex;
  align-items: center;
  min-height: 32px;
  min-width: 0;
  padding: 5px 8px 5px 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #f8fafc;
  color: #303133;
  line-height: 20px;
}

.detail-text {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-text--empty {
  color: #a8abb2;
}

.detail-value--textarea {
  align-items: flex-start;
  min-height: 78px;
}

.detail-value--textarea .detail-text {
  display: -webkit-box;
  overflow: hidden;
  white-space: normal;
  word-break: break-all;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.detail-copy {
  flex: 0 0 auto;
  margin-left: 6px;
  color: #909399;
}

.detail-copy:hover {
  color: #409eff;
}

@media (max-width: 900px) {
  .detail-form {
    grid-template-columns: 1fr;
  }

  .detail-field--wide {
    grid-column: auto;
  }
}

@media (max-width: 640px) {
  .client-page {
    height: auto;
    overflow: visible;
  }

  .page-actions,
  .client-filters,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item--status,
  .filter-item--keyword {
    width: 100%;
  }

  .detail-field {
    grid-template-columns: 1fr;
  }

  .detail-label {
    text-align: left;
  }

  .detail-control {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
