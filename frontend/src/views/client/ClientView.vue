<template>
  <section class="client-page">
    <header class="page-actions">
      <el-button type="primary" :icon="RefreshRight" @click="loadClients">刷新</el-button>
    </header>

    <section class="client-panel">
      <div class="client-filters">
        <div class="filter-item filter-item--keyword">
          <span class="filter-label">关键词</span>
          <el-input v-model="keywordFilter" clearable placeholder="客户端 IP、客户端名称" />
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
        <el-button type="warning" :disabled="selectedClientIds.length === 0" @click="handleBatchOffline">
          下线重连
        </el-button>
        <el-button type="danger" :disabled="selectedClientIds.length === 0" @click="handleBatchBan">
          拉黑
        </el-button>
      </div>

      <el-table v-loading="loading" class="client-table adaptive-table" :data="pagedClients" border height="100%"
        row-key="id" empty-text="暂无客户端" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" />
        <el-table-column label="客户端 IP" width="118">
          <template #default="{ row }">
            {{ getClientIp(row) || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="客户端名称" min-width="120">
          <template #default="{ row }">
            {{ getClientName(row) }}
          </template>
        </el-table-column>
        <el-table-column label="客户端状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row)" effect="plain">
              {{ getStatusText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="getAutomaTagType(row)" effect="plain">
              {{ getAutomaStatusText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 版本" width="104">
          <template #default="{ row }">
            {{ getAutomaVersion(row) }}
          </template>
        </el-table-column>
        <el-table-column label="浏览器名称" width="100">
          <template #default="{ row }">
            {{ getBrowserName(row) }}
          </template>
        </el-table-column>
        <el-table-column label="浏览器版本" width="96">
          <template #default="{ row }">
            {{ getBrowserVersion(row) }}
          </template>
        </el-table-column>
        <el-table-column label="是否拉黑" width="86" align="center">
          <template #default="{ row }">
            <el-tag :type="isBanned(row) ? 'danger' : 'success'" effect="plain">
              {{ isBanned(row) ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近心跳/交互" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(getLastActiveTime(row)) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="168" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="!isBanned(row)" link type="warning" @click="handleOffline(row)">
              下线重连
            </el-button>
            <el-button v-if="!isBanned(row)" link type="danger" @click="handleBan(row)">
              拉黑
            </el-button>
            <el-button v-else link type="success" @click="handleUnban(row)">解除拉黑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <AppPagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="pageSizes"
        :total="clients.length" />
    </section>

    <AppDialog v-model="detailVisible" title="客户端详情" width="720px" class="client-detail-dialog" confirm-text="保存"
      cancel-text="关闭" :loading="detailSaving" :confirm-disabled="detailLoading" @confirm="handleSaveDetail">
      <div v-loading="detailLoading" class="detail-form">
        <div v-for="field in detailFields" :key="field.key" class="detail-field">
          <span class="detail-label">{{ field.label }}</span>
          <div class="detail-control">
            <el-input v-if="field.type === 'textarea'" :model-value="field.value" disabled type="textarea" :rows="3" />
            <el-input v-else-if="field.editable" v-model="detailForm[field.key]" clearable />
            <el-input v-else :model-value="field.value" disabled />
            <el-button v-if="!field.editable" :icon="CopyDocument" :disabled="!field.value"
              @click="copyDetailValue(field.value)">
              复制
            </el-button>
          </div>
        </div>
      </div>
    </AppDialog>

  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { CopyDocument, RefreshRight } from '@element-plus/icons-vue'
import AppDialog from '@/components/AppDialog.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
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
const pageSizes = [10, 30, 60]
const statusFilter = ref('')
const keywordFilter = ref('')
const filterSearchDelay = 200
let filterSearchTimer = null
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailSaving = ref(false)
const detailClient = ref(null)
const detailForm = reactive(createDetailForm())
const selectedClientIds = ref([])

const pagedClients = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return clients.value.slice(start, start + pageSize.value)
})

const detailFields = computed(() => {
  return [
    { key: 'client_ip', label: '客户端 IP', value: formatEmpty(getClientIp(detailForm)) },
    { key: 'client_name', label: '客户端名称', value: formatEmpty(detailForm.client_name), editable: true },
    { key: 'status', label: '客户端状态', value: getStatusText(detailForm) },
    { key: 'plugin_status', label: 'Automa 状态', value: getAutomaStatusText(detailForm) },
    { key: 'automa_version', label: 'Automa 版本', value: formatEmpty(getAutomaVersion(detailForm)) },
    { key: 'browser_name', label: '浏览器名称', value: formatEmpty(getBrowserName(detailForm)) },
    { key: 'browser_version', label: '浏览器版本', value: formatEmpty(getBrowserVersion(detailForm)) },
    { key: 'is_banned', label: '是否拉黑', value: isBanned(detailForm) ? '是' : '否' },
    { key: 'last_seen_at', label: '最近心跳/交互', value: formatDate(getLastActiveTime(detailForm)) },
    { key: 'id', label: '服务端 ID', value: formatEmpty(detailForm.id) },
    { key: 'client_id', label: '客户端标识', value: formatEmpty(detailForm.client_id) },
    { key: 'os_name', label: '操作系统', value: formatEmpty(detailForm.os_name) },
    { key: 'os_version', label: '系统版本', value: formatEmpty(detailForm.os_version) },
    { key: 'hostname', label: '主机名', value: formatEmpty(detailForm.hostname) },
    { key: 'user_agent', label: 'User-Agent', value: formatEmpty(detailForm.user_agent), type: 'textarea' },
    { key: 'ban_reason', label: '拉黑原因', value: formatEmpty(detailForm.ban_reason) },
    { key: 'first_seen_at', label: '首次连接时间', value: formatDate(detailForm.first_seen_at) },
    { key: 'connected_at', label: '连接时间', value: formatDate(detailForm.connected_at) },
    { key: 'disconnected_at', label: '断开时间', value: formatDate(detailForm.disconnected_at) },
    { key: 'created_at', label: '创建时间', value: formatDate(detailForm.created_at) },
    { key: 'updated_at', label: '更新时间', value: formatDate(detailForm.updated_at) },
  ]
})

onMounted(() => {
  loadClients()
})

onBeforeUnmount(() => {
  clearFilterSearchTimer()
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

async function loadClients() {
  loading.value = true
  try {
    const data = await listClients({
      status: statusFilter.value,
      keyword: keywordFilter.value.trim(),
    })
    clients.value = normalizeList(data, 'clients')
    selectedClientIds.value = selectedClientIds.value.filter((id) =>
      clients.value.some((client) => getClientId(client) === id),
    )
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  keywordFilter.value = ''
  statusFilter.value = ''
}

function scheduleFilterSearch() {
  clearFilterSearchTimer()
  filterSearchTimer = window.setTimeout(() => {
    filterSearchTimer = null
    currentPage.value = 1
    loadClients()
  }, filterSearchDelay)
}

function clearFilterSearchTimer() {
  if (!filterSearchTimer) return

  window.clearTimeout(filterSearchTimer)
  filterSearchTimer = null
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
  await offlineClient(getClientId(row))
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已通知客户端下线重连' })
  await loadClients()
}

async function handleBatchOffline() {
  const ids = selectedClientIds.value.slice()
  if (ids.length === 0) return

  const result = await batchOfflineClients(ids)
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message: `已通知 ${getBatchSuccessCount(result, ids)} 个客户端下线重连`,
  })
  await loadClients()
}

async function handleBan(row) {
  await banClient(getClientId(row))
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已拉黑' })
  await loadClients()
}

async function handleBatchBan() {
  const ids = selectedClientIds.value.slice()
  if (ids.length === 0) return

  const result = await batchBanClients(ids)
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message: `已拉黑 ${getBatchSuccessCount(result, ids)} 个客户端`,
  })
  await loadClients()
}

async function handleUnban(row) {
  const clientId = getClientId(row)
  if (!clientId) return

  await unbanClient(clientId)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已解除拉黑' })
  await loadClients()
}

function handleSelectionChange(selection) {
  selectedClientIds.value = selection.map((item) => getClientId(item)).filter(Boolean)
}

function getBatchSuccessCount(result, fallbackIds) {
  return Number(result?.success || 0) || fallbackIds.length
}

function createDetailForm() {
  return {
    id: '',
    client_id: '',
    client_name: '',
    client_ip: '',
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

function normalizeList(data, fallbackKey) {
  const list = data?.list || data?.[fallbackKey] || []
  return Array.isArray(list) ? list : []
}

function getClientId(row) {
  return row?.id || row?.client_id || row?.clientId || ''
}

function getClientIp(row) {
  // Client IP reads current IP first 客户端 IP 优先读取当前连接 IP
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function getClientName(row) {
  return row?.client_name || row?.clientName || row?.name || row?.hostname || getClientId(row) || '-'
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
  return row?.browser_name || row?.browserName || row?.browser || '-'
}

function getBrowserVersion(row) {
  // Browser version supports collected metadata 浏览器版本读取采集元数据
  return row?.browser_version || row?.browserVersion || '-'
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

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function formatEmpty(value, fallback = '') {
  if (value === undefined || value === null || value === '' || value === '-') return fallback
  return String(value)
}

function formatDate(value) {
  if (!value) return '-'
  const dateValue = typeof value === 'number' && value < 10000000000 ? value * 1000 : value
  const date = new Date(dateValue)
  if (Number.isNaN(date.getTime())) return '-'
  const pad = (num) => String(num).padStart(2, '0')
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
  ].join('-') + ` ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

async function copyDetailValue(value) {
  await navigator.clipboard.writeText(String(value || ''))
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已复制' })
}
</script>

<style scoped lang="scss">
.client-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.page-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.client-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 16px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.client-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 16px;
}

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
  flex: 1;
  min-height: 0;
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
