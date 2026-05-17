<template>
  <AppDialog
    v-model="visible"
    title="客户端同步"
    width="1280px"
    confirm-text="同步选中"
    :confirm-disabled="selectedIds.length === 0"
    :loading="syncing"
    @confirm="handleSync"
  >
    <div class="sync-dialog">
      <el-tabs v-model="activeMode" @tab-change="handleModeChange">
        <el-tab-pane label="按客户端" name="client" />
        <el-tab-pane label="按工作流" name="workflow" />
      </el-tabs>

      <div class="sync-toolbar">
        <template v-if="activeMode === 'client'">
          <el-select
            v-model="selectedIp"
            class="query-input"
            clearable
            filterable
            :loading="clientLoading"
            placeholder="选择在线客户端 IP"
            @visible-change="handleClientSelectVisible"
            @change="handleClientIpChange"
            @clear="handleClientIpClear"
          >
            <el-option
              v-for="clientIp in onlineClientIps"
              :key="clientIp"
              :label="clientIp"
              :value="clientIp"
            />
          </el-select>
        </template>
        <template v-else>
          <el-select
            v-model="selectedAutomaId"
            class="query-input"
            clearable
            filterable
            placeholder="选择工作流"
            @change="handleWorkflowChange"
            @clear="handleWorkflowClear"
          >
            <el-option
              v-for="workflow in workflows"
              :key="getWorkflowId(workflow)"
              :label="workflow.name || getWorkflowId(workflow)"
              :value="workflow.automa_id || getWorkflowId(workflow)"
            >
              <div class="workflow-option">
                <span>{{ workflow.name || '' }}</span>
                <small>{{ workflow.automa_id || getWorkflowId(workflow) }}</small>
              </div>
            </el-option>
          </el-select>
        </template>

        <el-input
          v-model="keyword"
          class="keyword-input"
          clearable
          :placeholder="keywordPlaceholder"
        />
      </div>

      <el-table
        ref="tableRef"
        v-loading="candidateLoading"
        class="candidate-table adaptive-table"
        :data="candidates"
        border
        height="420"
        row-key="row_key"
        empty-text="请选择查询条件后自动加载"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="40" :selectable="isSelectable" />

        <el-table-column v-if="activeMode === 'workflow'" label="客户端 IP" width="120">
          <template #default="{ row }">
            {{ row.source_ip || '' }}
          </template>
        </el-table-column>

        <el-table-column label="客户端工作流" min-width="200">
          <template #default="{ row }">
            <div class="workflow-name">
              <span>{{ row.name || '' }}</span>
              <small>{{ row.description || '' }}</small>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="数据库工作流" min-width="200">
          <template #default="{ row }">
            <div class="workflow-name">
              <span>{{ row.server_name || '' }}</span>
              <small>{{ row.server_description || '' }}</small>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="同步状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="getSyncTagType(row)" effect="plain">
              {{ getSyncText(row) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="工作流状态" width="112" align="center">
          <template #default="{ row }">
            <el-tag class="workflow-status-tag" :type="getWorkflowTagType(row)" effect="plain">
              {{ getWorkflowText(row) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="客户端更新时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatOptionalDate(row.updated_at_automa || row.updatedAt || row.updated_at) }}
          </template>
        </el-table-column>

        <el-table-column label="同步时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatOptionalDate(row.last_synced_at) }}
          </template>
        </el-table-column>

        <el-table-column label="服务端更新时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatOptionalDate(row.server_updated_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="sync-footer">
        <div class="sync-summary">
          <span>已选择 {{ selectedIds.length }} 个</span>
          <span>可同步 {{ selectableCount }} 个</span>
        </div>

        <AppPagination
          v-if="candidateTotal > 0"
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="pageSizes"
          :total="candidateTotal"
          layout="total, sizes, prev, pager, next"
        />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import { listClients } from '@/services/client'
import {
  listAutomaSyncCandidates,
  listAutomaSyncCandidatesByWorkflow,
  syncAutomaWorkflowsByIp,
} from '@/services/automa'

defineProps({
  workflows: {
    type: Array,
    default: () => [],
  },
})

const visible = defineModel({
  type: Boolean,
  default: false,
})

const emit = defineEmits(['synced'])

const tableRef = ref(null)
const activeMode = ref('client')
const selectedIp = ref('')
const selectedAutomaId = ref('')
const keyword = ref('')
const candidates = ref([])
const selectedRows = ref([])
const selectedIds = ref([])
const candidateLoading = ref(false)
const clientLoading = ref(false)
const onlineClientIps = ref([])
const syncing = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = [10, 30, 60]
const candidateTotal = ref(0)
const keywordSearchDelay = 200
let candidateRequestSeq = 0
let keywordSearchTimer = 0

const selectableCount = computed(() => candidates.value.filter(isSelectable).length)
const canLoad = computed(() => {
  return activeMode.value === 'client' ? Boolean(normalizeText(selectedIp.value)) : Boolean(selectedAutomaId.value)
})
const keywordPlaceholder = computed(() => {
  return activeMode.value === 'client' ? '检索客户端工作流' : '检索客户端 IP'
})

watch(visible, (nextVisible) => {
  if (nextVisible) {
    loadOnlineClientIps()
    return
  }

  clearKeywordSearchTimer()
  activeMode.value = 'client'
  selectedIp.value = ''
  selectedAutomaId.value = ''
  keyword.value = ''
  resetCandidatePage()
  resetCandidates()
})

watch(keyword, () => {
  scheduleKeywordSearch()
})

watch(currentPage, () => {
  if (!visible.value || !canLoad.value) return
  loadCandidates()
})

watch(pageSize, () => {
  if (!visible.value || !canLoad.value) return
  loadFirstCandidatePage()
})

onBeforeUnmount(() => {
  clearKeywordSearchTimer()
})

async function loadCandidates() {
  if (!canLoad.value) {
    showWarningMessage(activeMode.value === 'client' ? '请先选择在线客户端 IP' : '请先选择工作流')
    return
  }

  const requestSeq = ++candidateRequestSeq
  candidateLoading.value = true
  try {
    const params = {
      keyword: keyword.value.trim(),
      page_num: currentPage.value,
      page_size: pageSize.value,
    }
    const selectedClientIp = normalizeText(selectedIp.value)
    const data =
      activeMode.value === 'client'
        ? await listAutomaSyncCandidates(selectedClientIp, params)
        : await listAutomaSyncCandidatesByWorkflow(selectedAutomaId.value, params)
    if (requestSeq !== candidateRequestSeq) return

    const candidateList = normalizeList(data, 'workflows')
    candidateTotal.value = Number(data?.total ?? candidateList.length)
    candidates.value = candidateList.map((item) => ({
      ...item,
      row_key: `${normalizeText(item.source_ip || selectedClientIp)}_${getWorkflowId(item)}`,
    }))
    selectedRows.value = []
    selectedIds.value = []
    await nextTick()
    tableRef.value?.clearSelection()
  } finally {
    if (requestSeq === candidateRequestSeq) {
      candidateLoading.value = false
    }
  }
}

async function loadOnlineClientIps() {
  clientLoading.value = true
  try {
    const data = await listClients({
      status: 'online',
    })
    onlineClientIps.value = normalizeList(data, 'clients')
      .map(getClientIp)
      .filter(Boolean)
      .filter((clientIp, index, list) => list.indexOf(clientIp) === index)
  } finally {
    clientLoading.value = false
  }
}

function handleClientSelectVisible(opened) {
  if (opened) loadOnlineClientIps()
}

function handleClientIpChange(value) {
  selectedIp.value = normalizeText(value)
  resetCandidates()
  if (selectedIp.value) loadFirstCandidatePage()
}

function handleClientIpClear() {
  clearKeywordSearchTimer()
  selectedIp.value = ''
  resetCandidatePage()
  resetCandidates()
}

function handleWorkflowChange(value) {
  selectedAutomaId.value = value || ''
  resetCandidates()
  if (selectedAutomaId.value) loadFirstCandidatePage()
}

function handleWorkflowClear() {
  clearKeywordSearchTimer()
  selectedAutomaId.value = ''
  resetCandidatePage()
  resetCandidates()
}

function handleModeChange() {
  if (activeMode.value === 'client') {
    selectedAutomaId.value = ''
    loadOnlineClientIps()
  } else {
    selectedIp.value = ''
  }
  clearKeywordSearchTimer()
  keyword.value = ''
  resetCandidatePage()
  resetCandidates()
}

function scheduleKeywordSearch() {
  clearKeywordSearchTimer()
  if (!visible.value || !canLoad.value) return

  keywordSearchTimer = window.setTimeout(() => {
    keywordSearchTimer = 0
    loadFirstCandidatePage()
  }, keywordSearchDelay)
}

function clearKeywordSearchTimer() {
  if (!keywordSearchTimer) return

  window.clearTimeout(keywordSearchTimer)
  keywordSearchTimer = 0
}

async function handleSync() {
  const groups = groupSelectedRowsByIp()
  if (groups.size === 0) {
    showWarningMessage('请选择可同步的工作流')
    return
  }

  syncing.value = true
  try {
    for (const [sourceIp, workflowIds] of groups) {
      await syncAutomaWorkflowsByIp(sourceIp, workflowIds)
    }
    showSuccessMessage('同步完成')
    visible.value = false
    emit('synced')
  } finally {
    syncing.value = false
  }
}

function handleSelectionChange(selection) {
  selectedRows.value = selection
  selectedIds.value = selection.map((item) => `${normalizeText(item.source_ip || selectedIp.value)}_${getWorkflowId(item)}`)
}

function groupSelectedRowsByIp() {
  const groups = new Map()
  selectedRows.value.forEach((row) => {
    const sourceIp = normalizeText(row.source_ip || selectedIp.value)
    const workflowId = getWorkflowId(row)
    if (!sourceIp || !workflowId) return

    const current = groups.get(sourceIp) || []
    current.push(workflowId)
    groups.set(sourceIp, current)
  })
  return groups
}

function showSuccessMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message,
  })
}

function showWarningMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.warning,
    message,
  })
}

function resetCandidates() {
  candidateRequestSeq += 1
  candidateLoading.value = false
  candidates.value = []
  candidateTotal.value = 0
  selectedRows.value = []
  selectedIds.value = []
  tableRef.value?.clearSelection()
}

function resetCandidatePage() {
  currentPage.value = 1
}

function loadFirstCandidatePage() {
  if (currentPage.value === 1) {
    loadCandidates()
    return
  }

  currentPage.value = 1
}

function isSelectable(row) {
  return Boolean(row.has_update ?? row.hasUpdate ?? !row.synced)
}

function getSyncText(row) {
  if (row.has_update || row.hasUpdate) return '可同步'
  if (row.synced) return '已同步'
  return '未同步'
}

function getSyncTagType(row) {
  if (row.has_update || row.hasUpdate) return 'warning'
  if (row.synced) return 'success'
  return 'info'
}

function getWorkflowText(row) {
  if (row.sync_status === 'not_synced') return '数据库无记录'
  if (row.sync_status === 'client_newer') return '客户端较新'
  if (row.sync_status === 'server_newer') return '数据库较新'
  if (row.synced) return '内容一致'
  if (row.has_update || row.hasUpdate) return '内容有差异'
  return '待同步'
}

function getWorkflowTagType(row) {
  if (row.synced) return 'success'
  if (row.sync_status === 'server_newer') return 'danger'
  if (row.has_update || row.hasUpdate) return 'warning'
  return 'info'
}

function normalizeList(data, fallbackKey) {
  const list = data?.list || data?.[fallbackKey] || data?.candidates || []
  return Array.isArray(list) ? list : []
}

function normalizeText(value) {
  return String(value || '').trim()
}

function getWorkflowId(row) {
  return row?.id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function formatDate(value) {
  if (!value) return ''
  const date = new Date(Number(value) || value)
  if (Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = padDatePart(date.getMonth() + 1)
  const day = padDatePart(date.getDate())
  const hour = padDatePart(date.getHours())
  const minute = padDatePart(date.getMinutes())
  const second = padDatePart(date.getSeconds())
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

function formatOptionalDate(value) {
  if (!value) return ''
  return formatDate(value)
}

function padDatePart(value) {
  return String(value).padStart(2, '0')
}
</script>

<style scoped lang="scss">
.sync-dialog {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sync-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.query-input {
  width: 360px;
}

.keyword-input {
  width: 220px;
}

.candidate-table {
  width: 100%;
}

.candidate-table :deep(.nowrap-column .cell) {
  padding-inline: 8px;
  white-space: nowrap;
}

.workflow-status-tag {
  min-width: 84px;
  justify-content: center;
  white-space: nowrap;
}

.workflow-option {
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

.workflow-name {
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

.sync-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.sync-footer :deep(.app-pagination) {
  justify-content: flex-end;
  margin-top: 0;
}

.sync-summary {
  display: flex;
  flex: 0 0 auto;
  gap: 16px;
  color: #606266;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .sync-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .sync-footer {
    align-items: stretch;
    flex-direction: column;
  }

  .sync-footer :deep(.app-pagination) {
    justify-content: center;
  }

  .query-input,
  .keyword-input {
    width: 100%;
  }
}
</style>
