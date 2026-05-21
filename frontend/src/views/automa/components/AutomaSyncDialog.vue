<template>
  <AppDialog
    v-model="visible"
    title="客户端同步"
    width="min(1520px, calc(100vw - 32px))"
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
            :loading="workflowOptionLoading"
            placeholder="选择工作流"
            @visible-change="handleWorkflowSelectVisible"
            @change="handleWorkflowChange"
            @clear="handleWorkflowClear"
          >
            <el-option
              v-for="workflow in workflowSelectOptions"
              :key="getWorkflowId(workflow)"
              :label="workflow.name || workflow.automa_name || getWorkflowId(workflow)"
              :value="workflow.automa_id || getWorkflowId(workflow)"
            >
              <div class="workflow-option">
                <span>
                  <em>数据库自定义名称</em>
                  {{ workflow.name || '' }}
                </span>
                <small>
                  <em>Automa 工作流名称</em>
                  {{ workflow.automa_name || workflow.automa_id || getWorkflowId(workflow) }}
                </small>
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
        header-align="left"
        empty-text="请选择查询条件后自动加载"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="40" reserve-selection :selectable="isSelectable" />

        <el-table-column v-if="activeMode === 'workflow'" label="客户端 IP" width="110" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.source_ip || '' }}
          </template>
        </el-table-column>

        <el-table-column label="自定义名称" min-width="125" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.server_name || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="自定义描述" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.server_description || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="工作流名称" min-width="190">
          <template #default="{ row }">
            <span class="compare-line" :title="formatCompareText(row.automa_name || row.name, row.server_automa_name)">
              <span class="compare-item">
                <em>客户端</em>
                <span>{{ row.automa_name || row.name || '' }}</span>
              </span>
              <span class="compare-item">
                <em>数据库</em>
                <span>{{ row.server_automa_name || '' }}</span>
              </span>
            </span>
          </template>
        </el-table-column>

        <el-table-column label="工作流描述" min-width="210">
          <template #default="{ row }">
            <span
              class="compare-line"
              :title="formatCompareText(row.automa_description || row.description, row.server_automa_description)"
            >
              <span class="compare-item">
                <em>客户端</em>
                <span>{{ row.automa_description || row.description || '' }}</span>
              </span>
              <span class="compare-item">
                <em>数据库</em>
                <span>{{ row.server_automa_description || '' }}</span>
              </span>
            </span>
          </template>
        </el-table-column>

        <el-table-column label="同步状态" width="96" header-align="center">
          <template #default="{ row }">
            <div class="center-cell">
              <el-tag :type="getSyncTagType(row)" effect="plain">
                {{ getSyncText(row) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="工作流状态" width="116" header-align="center">
          <template #default="{ row }">
            <div class="center-cell">
              <el-tag class="workflow-status-tag" :type="getWorkflowTagType(row)" effect="plain">
                {{ getWorkflowText(row) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="客户端更新时间" width="178" class-name="nowrap-column">
          <template #default="{ row }">
            <span class="time-value" :title="formatOptionalDate(row.updated_at_automa || row.updatedAt || row.updated_at)">
              {{ formatOptionalDate(row.updated_at_automa || row.updatedAt || row.updated_at) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="同步时间" width="178" class-name="nowrap-column">
          <template #default="{ row }">
            <span class="time-value" :title="formatOptionalDate(row.last_synced_at)">
              {{ formatOptionalDate(row.last_synced_at) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="数据库更新时间" width="178" class-name="nowrap-column">
          <template #default="{ row }">
            <span class="time-value" :title="formatOptionalDate(row.server_updated_at)">
              {{ formatOptionalDate(row.server_updated_at) }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <div class="sync-footer">
        <div class="sync-summary">
          <AppSelectionSummary :count="selectedIds.length" />
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
import { computed, ref, watch } from 'vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { listClients } from '@/services/client'
import {
  listAutomaSyncCandidates,
  listAutomaSyncCandidatesByWorkflow,
  syncAutomaWorkflowsByIp,
} from '@/services/automa'
import { formatDate } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, normalizeList, normalizeText } from '@/utils/list'

const props = defineProps({
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
const candidateLoading = ref(false)
const clientLoading = ref(false)
const workflowOptionLoading = ref(false)
const onlineClientIps = ref([])
const onlineWorkflowOptions = ref([])
const syncing = ref(false)
const refreshCandidates = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = DEFAULT_PAGE_SIZES
const candidateTotal = ref(0)
let candidateRequestSeq = 0
const {
  selectedRows,
  selectedKeys: selectedIds,
  handleSelectionChange,
  restoreSelection,
  resetSelection,
} = usePagedTableSelection({
  rows: candidates,
  getRowKey: getSelectionKey,
  isSelectable,
})
const {
  run: runKeywordSearch,
  cancel: clearKeywordSearchTimer,
} = useDebouncedAction(loadFirstCandidatePage, 200)

const selectableCount = computed(() => candidates.value.filter(isSelectable).length)
const workflowSelectOptions = computed(() => onlineWorkflowOptions.value)
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
  refreshCandidates.value = false
  onlineWorkflowOptions.value = []
  resetCandidatePage()
  resetCandidates()
})

watch(keyword, () => {
  resetSelection(tableRef)
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

async function loadCandidates() {
  if (!canLoad.value) {
    showWarningMessage(activeMode.value === 'client' ? '请先选择在线客户端 IP' : '请先选择工作流')
    return
  }

  const requestSeq = ++candidateRequestSeq
  const shouldRefresh = refreshCandidates.value
  refreshCandidates.value = false
  candidateLoading.value = true
  try {
    const params = {
      keyword: keyword.value.trim(),
      page_num: currentPage.value,
      page_size: pageSize.value,
      refresh: shouldRefresh ? 1 : 0,
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
    await restoreSelection(tableRef)
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
    if (selectedIp.value && !onlineClientIps.value.includes(selectedIp.value)) {
      selectedIp.value = ''
      resetCandidates()
    }
  } finally {
    clientLoading.value = false
  }
}

function handleClientSelectVisible(opened) {
  if (opened) loadOnlineClientIps()
}

async function loadOnlineWorkflowOptions() {
  workflowOptionLoading.value = true
  try {
    const data = await listAutomaSyncCandidatesByWorkflow('', {
      page_num: 1,
      page_size: 1000,
      refresh: 1,
    })
    const workflowMap = new Map()
    const dbWorkflowMap = new Map(
      normalizeList(props.workflows)
        .map((workflow) => [workflow.automa_id || getWorkflowId(workflow), workflow])
        .filter(([workflowId]) => Boolean(workflowId)),
    )

    normalizeList(data, 'workflows').forEach((item) => {
      const workflowId = item.automa_id || getWorkflowId(item)
      if (!workflowId || workflowMap.has(workflowId)) return

      const dbWorkflow = dbWorkflowMap.get(workflowId) || {}
      workflowMap.set(workflowId, {
        ...item,
        ...dbWorkflow,
        automa_id: workflowId,
        name: dbWorkflow.name || item.server_name || item.name || '',
        automa_name: dbWorkflow.automa_name || item.automa_name || item.name || '',
      })
    })

    onlineWorkflowOptions.value = Array.from(workflowMap.values())
    if (selectedAutomaId.value && !workflowMap.has(selectedAutomaId.value)) {
      selectedAutomaId.value = ''
      resetCandidates()
    }
  } finally {
    workflowOptionLoading.value = false
  }
}

function handleWorkflowSelectVisible(opened) {
  if (opened) loadOnlineWorkflowOptions()
}

function handleClientIpChange(value) {
  selectedIp.value = normalizeText(value)
  refreshCandidates.value = true
  resetCandidates()
  if (selectedIp.value) loadFirstCandidatePage()
}

function handleClientIpClear() {
  clearKeywordSearchTimer()
  selectedIp.value = ''
  refreshCandidates.value = false
  resetCandidatePage()
  resetCandidates()
}

function handleWorkflowChange(value) {
  selectedAutomaId.value = value || ''
  refreshCandidates.value = true
  resetCandidates()
  if (selectedAutomaId.value) loadFirstCandidatePage()
}

function handleWorkflowClear() {
  clearKeywordSearchTimer()
  selectedAutomaId.value = ''
  refreshCandidates.value = false
  resetCandidatePage()
  resetCandidates()
}

function handleModeChange() {
  if (activeMode.value === 'client') {
    selectedAutomaId.value = ''
    loadOnlineClientIps()
  } else {
    selectedIp.value = ''
    loadOnlineWorkflowOptions()
  }
  clearKeywordSearchTimer()
  keyword.value = ''
  refreshCandidates.value = false
  resetCandidatePage()
  resetCandidates()
}

function scheduleKeywordSearch() {
  clearKeywordSearchTimer()
  if (!visible.value || !canLoad.value) return

  runKeywordSearch()
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
  resetSelection(tableRef)
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
  if (row.sync_status === 'server_newer') return false
  return Boolean(row.has_update ?? row.hasUpdate ?? !row.synced)
}

function getSyncText(row) {
  if (row.sync_status === 'server_newer') return '不可同步'
  if (row.has_update || row.hasUpdate) return '可同步'
  if (row.synced) return '已同步'
  return '未同步'
}

function getSyncTagType(row) {
  if (row.sync_status === 'server_newer') return 'danger'
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

function getWorkflowId(row) {
  return row?.id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function getSelectionKey(row) {
  return row?.row_key || `${normalizeText(row?.source_ip || selectedIp.value)}_${getWorkflowId(row)}`
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function formatOptionalDate(value) {
  if (!value) return ''
  return formatDate(value)
}

function formatCompareText(clientValue, serverValue) {
  return `客户端：${clientValue || ''} 数据库：${serverValue || ''}`.trim()
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
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-option {
  display: grid;
  gap: 2px;
  min-width: 0;

  span,
  small {
    display: flex;
    align-items: center;
    gap: 6px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  em {
    flex: 0 0 auto;
    color: #909399;
    font-style: normal;
  }

  small {
    color: #909399;
  }
}

.field-value {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: #303133;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-lines {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.field-line {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  align-items: center;
  gap: 6px;
  min-width: 0;

  em {
    color: #909399;
    font-style: normal;
  }

  span {
    min-width: 0;
    overflow: hidden;
    color: #303133;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.compare-line {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.compare-item {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  align-items: center;
  gap: 4px;
  min-width: 0;
  line-height: 18px;
  white-space: nowrap;

  em {
    color: #909399;
    font-style: normal;
  }

  span {
    min-width: 0;
    overflow: hidden;
    color: #303133;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.time-value {
  display: block;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.center-cell {
  display: flex;
  justify-content: center;
  min-width: 0;
}

.center-cell :deep(.el-tag) {
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-status-tag {
  min-width: 0;
  justify-content: center;
  white-space: nowrap;
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
