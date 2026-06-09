<template>
  <section class="client-page server-list-page">
    <header class="page-actions server-list-actions">
      <el-button :icon="RefreshRight" @click="loadClients">刷新</el-button>
      <el-button @click="resetFilters">重置</el-button>
      <el-button type="warning" :disabled="selectedClientIds.length === 0 || batchOfflineLoading"
        :loading="batchOfflineLoading" @click="handleBatchOffline">
        下线重连
      </el-button>
      <el-button type="danger" :disabled="selectedUnbannedClientIds.length === 0 || batchBanLoading"
        :loading="batchBanLoading" @click="handleBatchBan">
        拉黑
      </el-button>
      <el-button type="success" :disabled="selectedBannedClientIds.length === 0 || batchUnbanLoading"
        :loading="batchUnbanLoading" @click="handleBatchUnban">
        解除拉黑
      </el-button>
    </header>

    <section class="client-panel server-list-panel">
      <div class="client-filters server-list-filters">
        <div class="client-filter-fields">
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
            </el-select>
          </div>

          <div class="filter-item filter-item--busy">
            <span class="filter-label">忙闲状态</span>
            <el-select v-model="busyStatusFilter" clearable placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="空闲" value="idle" />
              <el-option label="执行中" value="busy" />
              <el-option label="未知" value="unknown" />
            </el-select>
          </div>

          <div class="filter-item filter-item--banned">
            <span class="filter-label">是否拉黑</span>
            <el-select v-model="bannedFilter" clearable placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="是" value="true" />
              <el-option label="否" value="false" />
            </el-select>
          </div>

          <div class="filter-item filter-item--heartbeat">
            <span class="filter-label">最近心跳</span>
            <AppTimeRangeFilter v-model="lastSeenTimeRange" start-placeholder="开始时间" end-placeholder="结束时间" />
          </div>

        </div>
      </div>

      <el-table ref="clientTableRef" v-loading="loading" class="client-table server-list-table"
        :data="pagedClients" border height="100%" :row-key="getClientId" empty-text="暂无客户端"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column label="客户端 IP" width="124">
          <template #default="{ row }">{{ getClientIp(row) || '' }}</template>
        </el-table-column>
        <el-table-column label="执行节点 ID" min-width="126">
          <template #default="{ row }">{{ getNodeId(row) }}</template>
        </el-table-column>
        <el-table-column label="客户端名称" min-width="150">
          <template #default="{ row }">{{ getClientName(row) }}</template>
        </el-table-column>
        <el-table-column width="74" align="center">
          <template #header>
            <span class="table-label-with-help">
              状态
              <el-tooltip :content="CLIENT_STATUS_HELP" placement="top">
                <el-icon class="status-help-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>
          <template #default="{ row }">
            <el-tag v-if="getStatusText(row)" :type="getStatusTagType(row)" effect="plain">{{ getStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column width="96" align="center">
          <template #header>
            <span class="table-label-with-help">
              忙闲状态
              <el-tooltip :content="BUSY_STATUS_HELP" placement="top">
                <el-icon class="status-help-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>
          <template #default="{ row }">
            <el-tag v-if="getBusyStatusText(row)" :type="getBusyStatusTagType(row)" effect="plain">{{ getBusyStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 状态" width="116" align="center">
          <template #default="{ row }">
            <el-tag v-if="getAutomaStatusText(row)" :type="getAutomaTagType(row)" effect="plain">{{ getAutomaStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Automa 版本" width="128">
          <template #default="{ row }">{{ getAutomaVersion(row) }}</template>
        </el-table-column>
        <el-table-column label="浏览器" width="90">
          <template #default="{ row }">{{ getBrowserName(row) }}</template>
        </el-table-column>
        <el-table-column label="浏览器版本" width="116">
          <template #default="{ row }">{{ getBrowserVersion(row) }}</template>
        </el-table-column>
        <el-table-column label="是否拉黑" width="88" align="center">
          <template #default="{ row }">
            <el-tag :type="isBanned(row) ? 'danger' : 'success'" effect="plain">
              {{ isBanned(row) ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近心跳" width="176" class-name="nowrap-column">
          <template #default="{ row }">{{ formatDate(getLastActiveTime(row)) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="174" align="center">
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

      <div class="client-footer server-list-footer">
        <AppSelectionSummary :count="selectedClientIds.length" unit="客户端" />
        <AppPagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="pageSizes"
          :total="clientTotal" />
      </div>
    </section>

    <AppDialog v-model="detailVisible" title="客户端详情" width="min(1040px, calc(100vw - 32px))"
      class="client-detail-dialog" confirm-text="保存" cancel-text="关闭" :loading="detailSaving"
      :confirm-disabled="detailLoading" @confirm="handleSaveDetail">
      <div v-loading="detailLoading" class="detail-groups">
        <section v-for="group in detailGroups" :key="group.key" class="detail-group">
          <div class="detail-group-title">{{ group.title }}</div>
          <div class="detail-form server-detail-form">
            <div v-for="field in group.fields" :key="field.key" class="detail-field server-detail-field"
              :class="{ 'detail-field--wide': field.type === 'textarea' }">
              <span class="detail-label server-detail-label">
                <span>{{ field.label }}</span>
                <el-tooltip v-if="field.help" :content="field.help" placement="top">
                  <el-icon class="status-help-icon"><InfoFilled /></el-icon>
                </el-tooltip>
              </span>
              <div class="detail-control server-detail-control">
                <el-input v-if="field.editable" v-model="detailForm[field.key]" clearable
                  :maxlength="field.maxlength" />
                <div v-else class="detail-value" :class="{ 'detail-value--textarea': field.type === 'textarea' }">
                  <span class="detail-text" :class="{ 'detail-text--empty': !field.value }">
                    {{ field.value || '' }}
                  </span>
                  <el-tooltip v-if="field.copyable && field.value" content="复制" placement="top">
                    <el-button class="detail-copy" text circle :icon="CopyDocument" :aria-label="`复制${field.label}`"
                      @click="copyDetailValue(field.value)" />
                  </el-tooltip>
                  <el-tooltip v-if="field.expandable && field.value" content="查看完整内容" placement="top">
                    <el-button class="detail-copy" text circle :icon="View" :aria-label="`查看${field.label}`"
                      @click="openLongTextDialog(field)" />
                  </el-tooltip>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </AppDialog>

    <AppDialog v-model="banReasonVisible" :title="banReasonTitle" width="520px" confirm-text="拉黑" cancel-text="取消"
      :loading="banReasonSubmitting" @confirm="confirmBanWithReason" @closed="resetBanReasonDialog">
      <el-form label-width="84px" class="ban-reason-form">
        <el-form-item label="拉黑原因">
          <el-input v-model="banReasonForm.reason" type="textarea" :rows="4" maxlength="256" clearable
            placeholder="可为空，填写后会保存到客户端详情" />
        </el-form-item>
      </el-form>
    </AppDialog>

    <AppDialog v-model="longTextVisible" :title="longTextDialog.title" width="min(760px, calc(100vw - 32px))"
      confirm-text="复制" cancel-text="关闭" @confirm="copyDetailValue(longTextDialog.value)">
      <pre class="long-text-content">{{ longTextDialog.value }}</pre>
    </AppDialog>
  </section>
</template>
<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { CopyDocument, InfoFilled, RefreshRight, View } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import AppDialog from '@/components/AppDialog.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { copyText } from '@/utils/browser'
import { formatDate as formatBaseDate, formatEmpty as formatBaseEmpty } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, getSafePage, normalizeList } from '@/utils/list'
import {
  batchBanClients,
  batchOfflineClients,
  batchUnbanClients,
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
const clientTotal = ref(0)
const pageSizes = DEFAULT_PAGE_SIZES
const statusFilter = ref('')
const busyStatusFilter = ref('')
const bannedFilter = ref('')
const keywordFilter = ref('')
const lastSeenTimeRange = ref([])
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailSaving = ref(false)
const detailClient = ref(null)
const detailForm = reactive(createDetailForm())
const clientTableRef = ref(null)
const batchOfflineLoading = ref(false)
const batchBanLoading = ref(false)
const batchUnbanLoading = ref(false)
const clientActionLoadingIds = ref(new Set())
const CLIENT_DISPLAY_NAME_MAX_LENGTH = 128
const CLIENT_BAN_REASON_MAX_LENGTH = 256
const banReasonVisible = ref(false)
const banReasonSubmitting = ref(false)
const banReasonTarget = ref({ type: '', row: null, ids: [] })
const banReasonForm = reactive({ reason: '' })
const banReasonTitle = computed(() => (banReasonTarget.value.type === 'batch' ? '批量拉黑客户端' : '拉黑客户端'))
const longTextVisible = ref(false)
const longTextDialog = reactive({ title: '', value: '' })
const CLIENT_STATUS_HELP = '表示客户端与服务端的连接状态，在线表示 WebSocket 当前可用，离线表示连接已断开或未建立。'
const BUSY_STATUS_HELP = '表示执行节点当前是否正在运行任务，空闲可接收任务，执行中表示已有任务占用，未知表示暂未上报。'

const pagedClients = computed(() => {
  return clients.value
})
const selectedBannedClientIds = computed(() => {
  const selectedIdSet = new Set(selectedClientIds.value.map((id) => String(id)))
  return clients.value
    .filter((client) => selectedIdSet.has(String(getClientId(client))) && isBanned(client))
    .map(getClientId)
    .filter(Boolean)
})
const selectedUnbannedClientIds = computed(() => {
  const selectedIdSet = new Set(selectedClientIds.value.map((id) => String(id)))
  return clients.value
    .filter((client) => selectedIdSet.has(String(getClientId(client))) && !isBanned(client))
    .map(getClientId)
    .filter(Boolean)
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

const detailGroups = computed(() => {
  const groups = [
    {
      key: 'base',
      title: '基础信息',
      fields: [
        { key: 'id', label: '服务端 ID', value: formatEmpty(detailForm.id), copyable: true },
        {
          key: 'display_name',
          label: '客户端名称',
          value: formatEmpty(detailForm.display_name),
          editable: true,
          maxlength: CLIENT_DISPLAY_NAME_MAX_LENGTH,
        },
        { key: 'client_ip', label: '客户端 IP', value: formatEmpty(getClientIp(detailForm)), copyable: true },
        { key: 'node_id', label: '执行节点 ID', value: formatEmpty(detailForm.node_id), copyable: true },
        { key: 'node_index', label: '执行节点序号', value: formatEmpty(detailForm.node_index) },
        { key: 'hostname', label: '主机名', value: formatEmpty(detailForm.hostname) },
      ],
    },
    {
      key: 'status',
      title: '状态信息',
      fields: [
        { key: 'status', label: '客户端状态', value: getStatusText(detailForm), help: CLIENT_STATUS_HELP },
        { key: 'busy_status', label: '节点忙闲状态', value: getBusyStatusText(detailForm), help: BUSY_STATUS_HELP },
        { key: 'current_execution_id', label: '当前执行 ID', value: formatEmpty(detailForm.current_execution_id) },
        { key: 'current_task_record_id', label: '当前任务记录 ID', value: formatEmpty(detailForm.current_task_record_id) },
        { key: 'plugin_status', label: 'Automa 状态', value: getAutomaStatusText(detailForm) },
        { key: 'automa_version', label: 'Automa 版本', value: formatEmpty(getAutomaVersion(detailForm)) },
        { key: 'worker_version', label: 'Worker 版本', value: formatEmpty(detailForm.worker_version) },
        { key: 'is_banned', label: '是否拉黑', value: isBanned(detailForm) ? '是' : '否' },
        {
          key: 'ban_reason',
          label: '拉黑原因',
          value: formatEmpty(detailForm.ban_reason),
          editable: true,
          maxlength: CLIENT_BAN_REASON_MAX_LENGTH,
        },
      ],
    },
    {
      key: 'environment',
      title: '环境信息',
      fields: [
        { key: 'browser_name', label: '浏览器名称', value: formatEmpty(getBrowserName(detailForm)) },
        { key: 'browser_version', label: '浏览器版本', value: formatEmpty(getBrowserVersion(detailForm)) },
        { key: 'os_name', label: '操作系统', value: formatEmpty(detailForm.os_name) },
        { key: 'os_version', label: '系统版本', value: formatEmpty(detailForm.os_version) },
        {
          key: 'profile_dir',
          label: 'Profile 目录',
          value: formatEmpty(detailForm.profile_dir),
          type: 'textarea',
          copyable: true,
          expandable: true,
        },
        {
          key: 'extension_dir',
          label: '扩展目录',
          value: formatEmpty(detailForm.extension_dir),
          type: 'textarea',
          copyable: true,
          expandable: true,
        },
        {
          key: 'user_agent',
          label: 'User-Agent',
          value: formatEmpty(detailForm.user_agent),
          type: 'textarea',
          copyable: true,
          expandable: true,
        },
        {
          key: 'capabilities_json',
          label: '节点能力快照',
          value: formatDetailJson(detailForm.capabilities_json),
          type: 'textarea',
          copyable: true,
          expandable: true,
        },
      ],
    },
    {
      key: 'time',
      title: '时间信息',
      fields: [
        { key: 'first_seen_at', label: '首次连接时间', value: formatDate(detailForm.first_seen_at) },
        { key: 'last_seen_at', label: '最近心跳/交互', value: formatDate(getLastActiveTime(detailForm)) },
        { key: 'connected_at', label: '连接时间', value: formatDate(detailForm.connected_at) },
        { key: 'disconnected_at', label: '断开时间', value: formatDate(detailForm.disconnected_at) },
        { key: 'last_command_at', label: '最近命令下发', value: formatDate(detailForm.last_command_at) },
        { key: 'last_workflow_sync_at', label: '最近工作流同步', value: formatDate(detailForm.last_workflow_sync_at) },
        { key: 'last_lock_renewed_at', label: '最近执行锁续期', value: formatDate(detailForm.last_lock_renewed_at) },
        { key: 'created_at', label: '创建时间', value: formatDate(detailForm.created_at) },
        { key: 'updated_at', label: '更新时间', value: formatDate(detailForm.updated_at) },
        { key: 'deleted_at', label: '删除时间', value: formatDate(detailForm.deleted_at) },
      ],
    },
    {
      key: 'compat',
      title: '兼容信息',
      fields: [
        { key: 'client_id', label: '客户端 ID', value: formatEmpty(detailForm.client_id), hiddenWhenEmpty: true },
        { key: 'client_name', label: '原始客户端名称', value: formatEmpty(detailForm.client_name), hiddenWhenEmpty: true },
        { key: 'node_name', label: '执行节点名称', value: formatEmpty(detailForm.node_name), hiddenWhenEmpty: true },
        { key: 'machine_id', label: '机器 ID', value: formatEmpty(detailForm.machine_id), hiddenWhenEmpty: true },
        { key: 'machine_name', label: '机器名称', value: formatEmpty(detailForm.machine_name), hiddenWhenEmpty: true },
      ],
    },
  ]

  return groups
    .map((group) => ({
      ...group,
      fields: group.fields.filter((field) => !field.hiddenWhenEmpty || field.value),
    }))
    .filter((group) => group.fields.length > 0)
})

onMounted(() => {
  loadClients()
})

watch(statusFilter, () => {
  clearFilterSearchTimer()
  reloadFirstClientPage()
})

watch(busyStatusFilter, () => {
  clearFilterSearchTimer()
  reloadFirstClientPage()
})

watch(bannedFilter, () => {
  clearFilterSearchTimer()
  reloadFirstClientPage()
})

watch(lastSeenTimeRange, () => {
  clearFilterSearchTimer()
  reloadFirstClientPage()
})

watch(keywordFilter, () => {
  scheduleFilterSearch()
})

watch(pagedClients, () => {
  restoreClientSelection(clientTableRef)
})

watch(currentPage, () => {
  loadClients()
})

watch(pageSize, () => {
  reloadFirstClientPage()
})

async function loadClients() {
  loading.value = true
  try {
    const [lastSeenStartTime, lastSeenEndTime] = getLastSeenTimeRange()
    const data = await listClients({
      status: statusFilter.value,
      busy_status: busyStatusFilter.value,
      is_banned: bannedFilter.value,
      keyword: keywordFilter.value.trim(),
      last_seen_start_time: lastSeenStartTime,
      last_seen_end_time: lastSeenEndTime,
      page_num: currentPage.value,
      page_size: pageSize.value,
    })
    const list = normalizeList(data, 'clients')
    clients.value = list
    clientTotal.value = Number(data?.total ?? list.length)
    currentPage.value = getSafePage({
      total: clientTotal.value,
      page: currentPage.value,
      size: pageSize.value,
    })
    retainClientSelectionByRows(clients.value)
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  keywordFilter.value = ''
  statusFilter.value = ''
  busyStatusFilter.value = ''
  bannedFilter.value = ''
  lastSeenTimeRange.value = []
}

function searchClientsNow() {
  reloadFirstClientPage()
}

function reloadFirstClientPage() {
  if (currentPage.value === 1) {
    loadClients()
    return
  }
  currentPage.value = 1
}

function getLastSeenTimeRange() {
  const [startTime, endTime] = lastSeenTimeRange.value || []
  return [startTime || '', endTime || '']
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
      ban_reason: detailForm.ban_reason.trim(),
    })
    const updatedClient = data.client || {
      ...detailForm,
      display_name: displayName,
      ban_reason: detailForm.ban_reason.trim(),
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

  const confirmed = await appConfirm({
    title: '批量下线重连',
    message: `确认通知选中的 ${ids.length} 个客户端下线重连吗？`,
    type: APP_CONFIRM_TYPE.warning,
    confirmText: '下线重连',
  })
  if (!confirmed) return

  batchOfflineLoading.value = true
  try {
    const result = await batchOfflineClients(ids)
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: formatBatchActionMessage('下线重连', '个客户端', result, ids),
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

  banReasonTarget.value = { type: 'single', row, ids: [clientId] }
  banReasonForm.reason = ''
  banReasonVisible.value = true
}

async function handleBatchBan() {
  const ids = selectedUnbannedClientIds.value.slice()
  if (ids.length === 0 || batchBanLoading.value) return

  banReasonTarget.value = { type: 'batch', row: null, ids }
  banReasonForm.reason = ''
  banReasonVisible.value = true
}

async function confirmBanWithReason() {
  const target = banReasonTarget.value || {}
  const reason = banReasonForm.reason.trim()
  if (target.type === 'batch') {
    await confirmBatchBan(target.ids || [], reason)
    return
  }
  await confirmSingleBan(target.row, reason)
}

async function confirmSingleBan(row, reason) {
  const clientId = getClientId(row)
  if (!clientId || isClientActionLoading(row) || banReasonSubmitting.value) return

  banReasonSubmitting.value = true
  setClientActionLoading(clientId, true)
  try {
    await banClient(clientId, reason)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已拉黑' })
    banReasonVisible.value = false
    await loadClients()
  } finally {
    setClientActionLoading(clientId, false)
    banReasonSubmitting.value = false
  }
}

async function confirmBatchBan(ids, reason) {
  if (ids.length === 0 || batchBanLoading.value || banReasonSubmitting.value) return

  banReasonSubmitting.value = true
  batchBanLoading.value = true
  try {
    const result = await batchBanClients(ids, reason)
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: formatBatchActionMessage('拉黑', '个客户端', result, ids),
    })
    banReasonVisible.value = false
    await loadClients()
    resetClientSelection(clientTableRef)
  } finally {
    batchBanLoading.value = false
    banReasonSubmitting.value = false
  }
}

async function handleBatchUnban() {
  const ids = selectedBannedClientIds.value.slice()
  if (ids.length === 0 || batchUnbanLoading.value) return

  const confirmed = await appConfirm({
    title: '批量解除拉黑',
    message: `确认解除选中的 ${ids.length} 个客户端拉黑吗？`,
    type: APP_CONFIRM_TYPE.warning,
    confirmText: '解除拉黑',
  })
  if (!confirmed) return

  batchUnbanLoading.value = true
  try {
    const result = await batchUnbanClients(ids)
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: formatBatchActionMessage('解除', '个客户端拉黑', result, ids),
    })
    await loadClients()
    resetClientSelection(clientTableRef)
  } finally {
    batchUnbanLoading.value = false
  }
}

async function handleUnban(row) {
  const clientId = getClientId(row)
  if (!clientId || isClientActionLoading(row)) return

  const confirmed = await appConfirm({
    title: '解除拉黑',
    message: '确认解除这个客户端的拉黑状态吗？',
    type: APP_CONFIRM_TYPE.warning,
    confirmText: '解除拉黑',
  })
  if (!confirmed) return

  setClientActionLoading(clientId, true)
  try {
    await unbanClient(clientId)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端已解除拉黑' })
    await loadClients()
  } finally {
    setClientActionLoading(clientId, false)
  }
}

function resetBanReasonDialog() {
  banReasonTarget.value = { type: '', row: null, ids: [] }
  banReasonForm.reason = ''
  banReasonSubmitting.value = false
}

function getBatchSuccessCount(result, fallbackIds) {
  return Number(result?.success || 0) || fallbackIds.length
}

function formatBatchActionMessage(actionText, unitText, result, fallbackIds) {
  const success = getBatchSuccessCount(result, fallbackIds)
  const details = []
  const notFound = Number(result?.not_found || result?.notFound || 0)
  const notified = Number(result?.notified || 0)
  if (notFound > 0) details.push(`未找到 ${notFound} 个`)
  if (notified > 0) details.push(`通知 ${notified} 个`)
  return details.length > 0
    ? `已${actionText} ${success} ${unitText}（${details.join('，')}）`
    : `已${actionText} ${success} ${unitText}`
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
    deleted_at: '',
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
    deleted_at: row.deleted_at || row.deletedAt || '',
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
  return ''
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
  return status || ''
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
  return status || ''
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

function openLongTextDialog(field) {
  longTextDialog.title = field.label
  longTextDialog.value = field.value || ''
  longTextVisible.value = true
}
</script>

<style scoped lang="scss">
.client-filters {
  display: flex;
  align-items: flex-start;
  flex-direction: row;
  gap: 10px 20px;
}

.client-filter-fields {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  flex: 1 1 auto;
  gap: 10px 20px;
  width: 100%;
  min-width: 0;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.filter-item--status {
  min-width: 0;
}

.filter-item--busy {
  min-width: 0;
}

.filter-item--banned {
  min-width: 0;
}

.filter-item--heartbeat {
  min-width: 0;
}

.filter-item--keyword {
  width: 360px;
  min-width: 0;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.filter-item :deep(.el-input),
.filter-item :deep(.el-select) {
  flex: 1;
  min-width: 0;
}

.filter-item--status :deep(.el-select),
.filter-item--busy :deep(.el-select),
.filter-item--banned :deep(.el-select) {
  width: 144px;
  flex: 0 0 144px;
}

.filter-item--heartbeat :deep(.app-time-range-filter) {
  grid-template-columns: 200px auto 200px;
  justify-content: start;
  width: auto;
}

.client-table {
  width: 100%;
}

.table-label-with-help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  white-space: nowrap;
}

.status-help-icon {
  flex: 0 0 auto;
  color: #909399;
  font-size: 14px;
  cursor: help;
}

:deep(.client-table .cell) {
  white-space: nowrap;
  word-break: keep-all;
}

:deep(.client-table .el-button + .el-button) {
  margin-left: 4px;
}

:deep(.client-table .el-button) {
  white-space: nowrap;
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

.detail-groups {
  display: grid;
  gap: 16px;
}

.detail-group {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.detail-group-title {
  display: flex;
  align-items: center;
  min-height: 24px;
  color: #303133;
  font-size: 14px;
  font-weight: 600;
}

.detail-group + .detail-group {
  padding-top: 2px;
  border-top: 1px solid #ebeef5;
}

.detail-field {
  grid-template-columns: 116px minmax(0, 1fr);
}

.detail-form {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 20px;
}

.detail-control {
  grid-template-columns: minmax(0, 1fr);
}

.detail-label {
  display: inline-flex;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 4px;
}

.detail-label .status-help-icon {
  margin-top: 1px;
}

.detail-control :deep(.el-input__wrapper) {
  min-height: 32px;
}

.detail-field--wide {
  grid-column: 1 / -1;
}

.detail-value {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  height: 32px;
  min-width: 0;
  padding: 4px 8px 4px 10px;
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
  height: auto;
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
  width: 24px;
  height: 24px;
  margin-left: 6px;
  padding: 0;
  color: #909399;
}

.detail-copy :deep(.el-icon) {
  font-size: 14px;
}

.detail-copy:hover {
  color: #409eff;
}

.long-text-content {
  box-sizing: border-box;
  max-height: min(520px, calc(100vh - 240px));
  min-height: 180px;
  margin: 0;
  padding: 12px;
  overflow: auto;
  color: #303133;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  background: #f8fafc;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

@media (max-width: 1280px) {
  .filter-item--keyword {
    width: 320px;
  }

  .filter-item--heartbeat :deep(.app-time-range-filter) {
    grid-template-columns: 180px auto 180px;
  }
}

@media (max-width: 1100px) {
  .filter-item--keyword {
    width: 280px;
  }

  .filter-item--status :deep(.el-select),
  .filter-item--busy :deep(.el-select),
  .filter-item--banned :deep(.el-select) {
    width: 118px;
    flex-basis: 118px;
  }
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
  .client-filter-fields,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .client-filter-fields {
    width: 100%;
  }

  .filter-item--keyword,
  .filter-item--status,
  .filter-item--busy,
  .filter-item--banned {
    width: auto;
    max-width: none;
  }

  .filter-item--status :deep(.el-select),
  .filter-item--busy :deep(.el-select),
  .filter-item--banned :deep(.el-select) {
    width: 100%;
    flex: 1 1 auto;
  }

  .filter-item--heartbeat :deep(.app-time-range-filter) {
    grid-template-columns: 1fr;
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
