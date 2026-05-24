<template>
  <section class="client-page server-list-page">
    <header class="page-actions server-list-actions">
      <el-button type="primary" :icon="RefreshRight" @click="loadClients">刷新</el-button>
    </header>

    <section class="client-panel server-list-panel">
      <div class="client-filters server-list-filters">
        <div class="filter-item filter-item--keyword">
          <span class="filter-label">关键字</span>
          <el-input v-model="keywordFilter" clearable placeholder="客户端 IP、名称、节点、机器" />
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
        <el-button
          type="warning"
          :disabled="selectedClientIds.length === 0 || batchOfflineLoading"
          :loading="batchOfflineLoading"
          @click="handleBatchOffline"
        >
          下线重连
        </el-button>
        <el-button
          type="danger"
          :disabled="selectedClientIds.length === 0 || batchBanLoading"
          :loading="batchBanLoading"
          @click="handleBatchBan"
        >
          拉黑
        </el-button>
        <AppSelectionSummary :count="selectedClientIds.length" unit="客户端" />
      </div>

      <el-table
        ref="clientTableRef"
        v-loading="loading"
        class="client-table server-list-table adaptive-table"
        :data="pagedClients"
        border
        height="100%"
        :row-key="getClientId"
        empty-text="暂无客户端"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column label="客户端 IP" width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ getClientIp(row) || '' }}</template>
        </el-table-column>
        <el-table-column label="执行节点" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ getNodeDisplay(row) }}</template>
        </el-table-column>
        <el-table-column label="所属机器" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ getMachineDisplay(row) }}</template>
        </el-table-column>
        <el-table-column label="客户端名称" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ getClientName(row) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row)" effect="plain">{{ getStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="getAutomaTagType(row)" effect="plain">{{ getAutomaStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 版本" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ getAutomaVersion(row) }}</template>
        </el-table-column>
        <el-table-column label="浏览器" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ getBrowserName(row) }}</template>
        </el-table-column>
        <el-table-column label="浏览器版本" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ getBrowserVersion(row) }}</template>
        </el-table-column>
        <el-table-column label="是否拉黑" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="isBanned(row) ? 'danger' : 'success'" effect="plain">
              {{ isBanned(row) ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近心跳" width="160" class-name="nowrap-column">
          <template #default="{ row }">{{ formatDate(getLastActiveTime(row)) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="168" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="!isBanned(row)"
              link
              type="warning"
              :disabled="isClientActionLoading(row)"
              :loading="isClientActionLoading(row)"
              @click="handleOffline(row)"
            >
              下线重连
            </el-button>
            <el-button
              v-if="!isBanned(row)"
              link
              type="danger"
              :disabled="isClientActionLoading(row)"
              :loading="isClientActionLoading(row)"
              @click="handleBan(row)"
            >
              拉黑
            </el-button>
            <el-button
              v-else
              link
              type="success"
              :disabled="isClientActionLoading(row)"
              :loading="isClientActionLoading(row)"
              @click="handleUnban(row)"
            >
              解除拉黑
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <AppPagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="pageSizes"
        :total="clients.length"
      />
    </section>

    <AppDialog
      v-model="detailVisible"
      title="客户端详情"
      width="720px"
      class="client-detail-dialog"
      confirm-text="保存"
      cancel-text="关闭"
      :loading="detailSaving"
      :confirm-disabled="detailLoading"
      @confirm="handleSaveDetail"
    >
      <div v-loading="detailLoading" class="detail-form">
        <div v-for="field in detailFields" :key="field.key" class="detail-field">
          <span class="detail-label">{{ field.label }}</span>
          <div class="detail-control">
            <el-input v-if="field.type === 'textarea'" :model-value="field.value" disabled type="textarea" :rows="3" />
            <el-input v-else-if="field.editable" v-model="detailForm[field.key]" clearable />
            <el-input v-else :model-value="field.value" disabled />
            <el-button
              v-if="!field.editable"
              :icon="CopyDocument"
              :disabled="!field.value"
              @click="copyDetailValue(field.value)"
            >
              复制
            </el-button>
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
  return [
    { key: 'client_ip', label: '瀹㈡埛绔?IP', value: formatEmpty(getClientIp(detailForm)) },
    { key: 'node_id', label: '执行节点 ID', value: formatEmpty(detailForm.node_id) },
    { key: 'node_name', label: '执行节点名称', value: formatEmpty(detailForm.node_name) },
    { key: 'machine_id', label: '机器 ID', value: formatEmpty(detailForm.machine_id) },
    { key: 'machine_name', label: '机器名称', value: formatEmpty(detailForm.machine_name) },
    { key: 'client_name', label: '瀹㈡埛绔悕绉?, value: formatEmpty(detailForm.client_name), editable: true },
    { key: 'status', label: '瀹㈡埛绔姸鎬?, value: getStatusText(detailForm) },
    { key: 'plugin_status', label: 'Automa 鐘舵€?, value: getAutomaStatusText(detailForm) },
    { key: 'automa_version', label: 'Automa 鐗堟湰', value: formatEmpty(getAutomaVersion(detailForm)) },
    { key: 'browser_name', label: '娴忚鍣ㄥ悕绉?, value: formatEmpty(getBrowserName(detailForm)) },
    { key: 'browser_version', label: '娴忚鍣ㄧ増鏈?, value: formatEmpty(getBrowserVersion(detailForm)) },
    { key: 'is_banned', label: '鏄惁鎷夐粦', value: isBanned(detailForm) ? '鏄? : '鍚? },
    { key: 'last_seen_at', label: '鏈€杩戝績璺?浜や簰', value: formatDate(getLastActiveTime(detailForm)) },
    { key: 'id', label: '鏈嶅姟绔?ID', value: formatEmpty(detailForm.id) },
    { key: 'client_id', label: '瀹㈡埛绔爣璇?, value: formatEmpty(detailForm.client_id) },
    { key: 'os_name', label: '鎿嶄綔绯荤粺', value: formatEmpty(detailForm.os_name) },
    { key: 'os_version', label: '绯荤粺鐗堟湰', value: formatEmpty(detailForm.os_version) },
    { key: 'hostname', label: '涓绘満鍚?, value: formatEmpty(detailForm.hostname) },
    { key: 'user_agent', label: 'User-Agent', value: formatEmpty(detailForm.user_agent), type: 'textarea' },
    { key: 'ban_reason', label: '鎷夐粦鍘熷洜', value: formatEmpty(detailForm.ban_reason) },
    { key: 'first_seen_at', label: '棣栨杩炴帴鏃堕棿', value: formatDate(detailForm.first_seen_at) },
    { key: 'connected_at', label: '杩炴帴鏃堕棿', value: formatDate(detailForm.connected_at) },
    { key: 'disconnected_at', label: '鏂紑鏃堕棿', value: formatDate(detailForm.disconnected_at) },
    { key: 'created_at', label: '鍒涘缓鏃堕棿', value: formatDate(detailForm.created_at) },
    { key: 'updated_at', label: '鏇存柊鏃堕棿', value: formatDate(detailForm.updated_at) },
  ]
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
    const clientName = detailForm.client_name.trim()
    const data = await updateClient(clientId, {
      client_name: clientName,
    })
    const updatedClient = data.client || {
      ...detailForm,
      client_name: clientName,
    }

    // Sync detail form after save 淇濆瓨鍚庡悓姝ヨ鎯呰〃鍗?
    detailClient.value = updatedClient
    setDetailForm(updatedClient)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '瀹㈡埛绔凡淇濆瓨' })
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
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '宸查€氱煡瀹㈡埛绔笅绾块噸杩? })
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
      message: `宸查€氱煡 ${getBatchSuccessCount(result, ids)} 涓鎴风涓嬬嚎閲嶈繛`,
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
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '瀹㈡埛绔凡鎷夐粦' })
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
      message: `宸叉媺榛?${getBatchSuccessCount(result, ids)} 涓鎴风`,
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
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '瀹㈡埛绔凡瑙ｉ櫎鎷夐粦' })
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
    client_ip: '',
    machine_id: '',
    machine_name: '',
    node_id: '',
    node_name: '',
    status: '',
    online: false,
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
    user_agent: '',
    banned: false,
    is_banned: false,
    ban_reason: '',
    first_seen_at: '',
    last_seen_at: '',
    last_seen: '',
    last_heartbeat_time: '',
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
    client_ip: getClientIp(row),
    machine_id: row.machine_id || row.machineId || '',
    machine_name: row.machine_name || row.machineName || '',
    node_id: row.node_id || row.nodeId || '',
    node_name: row.node_name || row.nodeName || '',
    status: row.status || '',
    online: Boolean(row.online),
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
    user_agent: row.user_agent || row.userAgent || '',
    banned: Boolean(row.banned),
    is_banned: Boolean(row.is_banned),
    ban_reason: row.ban_reason || row.banReason || '',
    first_seen_at: row.first_seen_at || row.firstSeenAt || '',
    last_seen_at: row.last_seen_at || row.lastSeenAt || '',
    last_seen: row.last_seen || row.lastSeen || '',
    last_heartbeat_time: row.last_heartbeat_time || row.lastHeartbeatTime || '',
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
  // Client IP reads current IP first 瀹㈡埛绔?IP 浼樺厛璇诲彇褰撳墠杩炴帴 IP
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function getClientName(row) {
  return row?.client_name || row?.clientName || row?.name || row?.hostname || getClientId(row) || ''
}

function getNodeDisplay(row) {
  return [row?.node_name || row?.nodeName, row?.node_id || row?.nodeId].filter(Boolean).join(' / ')
}

function getMachineDisplay(row) {
  return [row?.machine_name || row?.machineName, row?.machine_id || row?.machineId].filter(Boolean).join(' / ')
}

function isBanned(row) {
  return Boolean(row?.banned || row?.is_banned || row?.status === 'banned')
}

function getStatusText(row) {
  // Client status keeps online state separate from ban flag 瀹㈡埛绔姸鎬佷笌鎷夐粦鏍囪鍒嗗紑灞曠ず
  if (row?.online || row?.status === 'online') return '鍦ㄧ嚎'
  if (row?.status === 'offline') return '绂荤嚎'
  if (row?.status === 'banned') return '宸叉媺榛?
  if (row?.status) return row.status
  return '绂荤嚎'
}

function getStatusTagType(row) {
  const status = getStatusText(row)
  if (status === '鍦ㄧ嚎') return 'success'
  if (status === '宸叉媺榛?) return 'danger'
  return 'info'
}

function getAutomaStatusText(row) {
  // Plugin status normalizes backend enum 鎻掍欢鐘舵€佸吋瀹瑰悗绔灇涓惧€?
  const status = row?.plugin_status || row?.automa_status || ''
  if (status === 'installed') return '宸插畨瑁?
  if (status === 'not_installed') return '鏈畨瑁?
  if (status === 'disabled') return '宸茬鐢?
  if (status === 'error') return '寮傚父'
  if (status === 'unknown') return '鏈煡'
  if (row?.automa_installed === true) return '宸插畨瑁?
  if (row?.automa_installed === false) return '鏈畨瑁?
  return status || '鏈煡'
}

function getAutomaTagType(row) {
  // Tag type mirrors readable Automa status 鏍囩绫诲瀷鍖归厤鍙鐘舵€?
  const status = getAutomaStatusText(row)
  if (status === '宸插畨瑁?) return 'success'
  if (status === '鏈畨瑁? || status === '宸茬鐢? || status === '寮傚父') return 'danger'
  return 'info'
}

function getAutomaVersion(row) {
  // Version supports snake and camel case 鐗堟湰鍙峰吋瀹逛笅鍒掔嚎鍜岄┘宄板瓧娈?
  return row?.automa_version || row?.automaVersion || ''
}

function getBrowserName(row) {
  // Browser name falls back to legacy browser field 娴忚鍣ㄥ悕绉板吋瀹规棫 browser 瀛楁
  return row?.browser_name || row?.browserName || row?.browser || ''
}

function getBrowserVersion(row) {
  // Browser version supports collected metadata 娴忚鍣ㄧ増鏈鍙栭噰闆嗗厓鏁版嵁
  return row?.browser_version || row?.browserVersion || ''
}

function getLastActiveTime(row) {
  // Last active time prefers heartbeat/interaction field 鏈€杩戞椿璺冩椂闂翠紭鍏堜娇鐢ㄥ績璺冲瓧娈?
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

function formatDate(value) {
  return formatBaseDate(value, { fallback: '' })
}

async function copyDetailValue(value) {
  await copyText(value)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '宸插鍒? })
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
  overflow: visible !important;
  text-overflow: clip !important;
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

.detail-form {
  display: grid;
  gap: 12px;
  max-height: 62vh;
  overflow: auto;
  padding-right: 4px;
}

.detail-field {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  align-items: start;
  gap: 6px;
  min-width: 0;
}

.detail-label {
  padding-top: 7px;
  color: #606266;
  font-size: 13px;
  text-align: right;
}

.detail-control {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 70px;
  gap: 6px;
  min-width: 0;
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
    grid-template-columns: minmax(0, 1fr) 64px;
  }
}
</style>
