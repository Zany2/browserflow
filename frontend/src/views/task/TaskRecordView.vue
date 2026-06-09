<template>
  <section class="task-record-page server-list-page">
    <header class="page-actions server-list-actions">
      <el-button :icon="RefreshRight" @click="loadRecords">刷新</el-button>
      <el-button type="danger" :disabled="selectedRecordIds.length === 0" @click="handleBatchDeleteRecords">
        删除选中
      </el-button>
      <el-button @click="resetRecordFilters">重置</el-button>
    </header>

    <section class="record-panel server-list-panel">
      <div class="record-filters server-list-filters">
        <div class="record-filter-fields">
          <div class="record-filter-row">
            <div class="filter-item filter-item--task">
              <span class="filter-label">任务名称</span>
              <el-input v-model="recordFilters.task_name" clearable placeholder="任务名称" />
            </div>

            <div class="filter-item filter-item--workflow">
              <span class="filter-label">自定义工作流名称</span>
              <el-input v-model="recordFilters.workflow_name" clearable placeholder="自定义工作流名称" />
            </div>

            <div class="filter-item filter-item--execute-time">
              <span class="filter-label">执行时间</span>
              <AppTimeRangeFilter v-model="recordFilters.execute_time_range" />
            </div>
          </div>

          <div class="record-filter-row">
            <div class="filter-item filter-item--client">
              <span class="filter-label">客户端 IP</span>
              <el-select
                v-model="recordFilters.client_ip"
                clearable
                filterable
                placeholder="选择或检索客户端 IP"
                :loading="clientIpLoading"
                @visible-change="handleClientIpSelectVisible"
              >
                <el-option v-for="clientIp in clientIpOptions" :key="clientIp" :label="clientIp" :value="clientIp" />
              </el-select>
            </div>

            <div class="filter-item filter-item--node">
              <span class="filter-label">执行节点</span>
              <el-select
                v-model="recordFilters.node_ids"
                clearable
                filterable
                multiple
                collapse-tags
                collapse-tags-tooltip
                placeholder="请先选择客户端 IP"
                :disabled="!recordFilters.client_ip"
                :loading="clientIpLoading"
                @visible-change="handleClientIpSelectVisible"
              >
                <el-option
                  v-for="node in filteredNodeOptions"
                  :key="node.node_id"
                  :label="node.label"
                  :value="node.node_id"
                />
              </el-select>
            </div>

            <div class="filter-item filter-item--status">
              <span class="filter-label">状态</span>
              <el-select v-model="recordFilters.status" clearable placeholder="全部">
                <el-option label="全部" value="" />
                <el-option label="待执行" value="pending" />
                <el-option label="已下发" value="queued" />
                <el-option label="执行中" value="running" />
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
                <el-option label="已取消" value="cancelled" />
              </el-select>
            </div>
          </div>
        </div>
      </div>

      <el-table
        ref="recordTableRef"
        v-loading="loadingRecords"
        class="record-table server-list-table adaptive-table"
        :data="pagedRecords"
        border
        height="100%"
        :row-key="getRecordSelectionKey"
        empty-text="暂无执行记录"
        @selection-change="handleRecordSelectionChange"
      >
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column label="任务名称" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.task_name || '' }}</template>
        </el-table-column>
        <el-table-column label="自定义工作流名称" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.workflow_name || '' }}</template>
        </el-table-column>
        <el-table-column label="客户端 IP" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.client_ip || '' }}</template>
        </el-table-column>
        <el-table-column label="执行节点" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ getRecordNodeText(row) }}</template>
        </el-table-column>
        <el-table-column label="触发方式" width="90" align="center">
          <template #default="{ row }">{{ getTriggerText(row.trigger_type) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="getRecordStatusTag(row.status)" effect="plain">
              {{ row.status_text || getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="执行耗时" width="100" align="center">
          <template #default="{ row }">{{ formatDuration(row.duration_ms) }}</template>
        </el-table-column>
        <el-table-column label="执行时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">{{ formatDate(row.started_at || row.created_at || row.startedAt || row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.error_message || '' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openRecordDetail(row)">详情</el-button>
            <el-button link type="success" :disabled="!row.task_id" @click="handleRetryRecord(row)">重试</el-button>
            <el-button link type="danger" @click="handleDeleteRecord(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="record-footer server-list-footer">
        <AppSelectionSummary :count="selectedRecordIds.length" unit="记录" />
        <AppPagination
          v-model:current-page="recordPage"
          v-model:page-size="recordPageSize"
          :page-sizes="pageSizes"
          :total="recordTotal"
        />
      </div>
    </section>

    <AppDialog
      v-model="recordDetailVisible"
      title="执行记录详情"
      width="min(1040px, calc(100vw - 32px))"
      class="record-detail-dialog"
    >
      <div v-if="recordDetail" class="detail-form">
        <el-descriptions border :column="2" class="detail-descriptions">
          <el-descriptions-item label="记录 ID">{{ recordDetail.id || '' }}</el-descriptions-item>
          <el-descriptions-item label="任务">{{ recordDetail.task_name || recordDetail.task_id || '' }}</el-descriptions-item>
          <el-descriptions-item label="自定义工作流名称">{{ recordDetail.workflow_name || recordDetail.workflow_id || '' }}</el-descriptions-item>
          <el-descriptions-item label="客户端">{{ recordDetail.client_name || recordDetail.client_id || '' }}</el-descriptions-item>
          <el-descriptions-item label="客户端 IP">{{ recordDetail.client_ip || '' }}</el-descriptions-item>
          <el-descriptions-item label="执行节点">{{ getRecordNodeText(recordDetail) }}</el-descriptions-item>
          <el-descriptions-item label="机器 ID">{{ recordDetail.machine_id || '' }}</el-descriptions-item>
          <el-descriptions-item label="触发方式">{{ getTriggerText(recordDetail.trigger_type) }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ recordDetail.status_text || getStatusText(recordDetail.status) }}</el-descriptions-item>
          <el-descriptions-item label="执行耗时">{{ formatDuration(recordDetail.duration_ms) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(recordDetail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatDate(recordDetail.started_at) }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ formatDate(recordDetail.finished_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(recordDetail.updated_at) }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-block">
          <h3>错误信息</h3>
          <pre class="detail-json">{{ recordDetail.error_message || '' }}</pre>
        </div>

        <div class="detail-block">
          <h3>执行参数</h3>
          <pre class="detail-json">{{ formatJSON(recordDetail.params) }}</pre>
        </div>

        <div class="detail-block">
          <h3>执行结果</h3>
          <pre class="detail-json">{{ formatJSON(recordDetail.result) }}</pre>
        </div>

        <div class="detail-block">
          <h3>结果文件</h3>
          <el-table class="detail-files-table" :data="recordDetailFiles" border size="small" empty-text="暂无结果文件">
            <el-table-column prop="file_name" label="文件名" min-width="160" show-overflow-tooltip />
            <el-table-column prop="file_type" label="类型" width="110" />
            <el-table-column label="行数" width="90" align="right">
              <template #default="{ row }">{{ formatNumber(row.row_count) }}</template>
            </el-table-column>
            <el-table-column label="大小" width="100" align="right">
              <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ row }">
                <el-button link type="primary" :disabled="!row.id" @click="downloadRecordFile(row)">下载</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="detail-block">
          <h3>原始记录</h3>
          <pre class="detail-json">{{ formatJSON(recordDetail) }}</pre>
        </div>
      </div>

      <template #footer>
        <el-button @click="recordDetailVisible = false">关闭</el-button>
      </template>
    </AppDialog>
  </section>
</template>
<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { listClients } from '@/services/client'
import {
  deleteTaskRecords,
  executeTask,
  getTaskRecordDetail,
  getTaskRecordFileDownloadUrl,
  listTaskRecords,
} from '@/services/task'
import { formatDate as formatBaseDate, formatJSON } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, normalizeList } from '@/utils/list'

const records = ref([])
const clientIpOptions = ref([])
const nodeOptions = ref([])
const loadingRecords = ref(false)
const clientIpLoading = ref(false)
const recordTableRef = ref(null)
const recordPage = ref(1)
const recordPageSize = ref(10)
const recordTotal = ref(0)
const pageSizes = DEFAULT_PAGE_SIZES
const recordDetailVisible = ref(false)
const recordDetail = ref(null)
const recordDetailFiles = ref([])
const {
  run: scheduleFilterSearch,
  cancel: clearFilterSearchTimer,
} = useDebouncedAction(searchRecordsNow, 200)

const recordFilters = reactive({
  task_name: '',
  workflow_name: '',
  client_ip: '',
  node_ids: [],
  execute_time_range: [],
  status: '',
})

const pagedRecords = computed(() => records.value)
const filteredNodeOptions = computed(() => {
  const selectedClientIp = recordFilters.client_ip
  if (!selectedClientIp) return []
  return nodeOptions.value.filter((node) => node.client_ip === selectedClientIp)
})
const {
  selectedKeys: selectedRecordIds,
  handleSelectionChange: handleRecordSelectionChange,
  restoreSelection: restoreRecordSelection,
  retainSelectionByRows: retainRecordSelectionByRows,
  resetSelection: resetRecordSelection,
} = usePagedTableSelection({
  rows: pagedRecords,
  getRowKey: getRecordSelectionKey,
})

onMounted(() => {
  loadRecords()
  loadClientIpOptions()
})

watch(() => [recordFilters.task_name, recordFilters.workflow_name], () => {
  scheduleFilterSearch()
})

watch(() => [recordFilters.client_ip, recordFilters.node_ids, recordFilters.execute_time_range, recordFilters.status], () => {
  clearFilterSearchTimer()
  reloadFirstRecordPage()
}, { deep: true })

watch(() => recordFilters.client_ip, () => {
  const validNodeIds = new Set(filteredNodeOptions.value.map((node) => node.node_id))
  recordFilters.node_ids = recordFilters.node_ids.filter((nodeId) => validNodeIds.has(nodeId))
})

watch(recordPage, () => {
  loadRecords()
})

watch(recordPageSize, () => {
  reloadFirstRecordPage()
})

watch(pagedRecords, () => {
  restoreRecordSelection(recordTableRef)
})

async function loadRecords() {
  loadingRecords.value = true
  try {
    const [startTime, endTime] = getExecuteTimeRange()
    const data = await listTaskRecords({
      task_name: recordFilters.task_name.trim(),
      workflow_name: recordFilters.workflow_name.trim(),
      client_ip: recordFilters.client_ip.trim(),
      node_ids: recordFilters.node_ids,
      start_time: startTime,
      end_time: endTime,
      status: recordFilters.status.trim(),
      page_num: recordPage.value,
      page_size: recordPageSize.value,
    })
    const list = normalizeList(data, 'records')
    records.value = list
    recordTotal.value = Number(data?.total ?? list.length)
    retainRecordSelectionByRows(records.value)
  } finally {
    loadingRecords.value = false
  }
}

async function loadClientIpOptions() {
  clientIpLoading.value = true
  try {
    const data = await listClients()
    const clients = normalizeList(data, 'clients')
    clientIpOptions.value = clients
      .map(getClientIp)
      .filter(Boolean)
      .filter((clientIp, index, list) => list.indexOf(clientIp) === index)
    nodeOptions.value = clients
      .map((client) => {
        const clientIp = getClientIp(client)
        const nodeId = getClientNodeId(client)
        return {
          client_ip: clientIp,
          node_id: nodeId,
          label: [clientIp, getClientNodeName(client) || nodeId].filter(Boolean).join(' / '),
        }
      })
      .filter((node) => node.client_ip && node.node_id)
  } finally {
    clientIpLoading.value = false
  }
}

function handleClientIpSelectVisible(opened) {
  if (opened) loadClientIpOptions()
}

function searchRecordsNow() {
  reloadFirstRecordPage()
}

function reloadFirstRecordPage() {
  if (recordPage.value === 1) {
    loadRecords()
    return
  }
  recordPage.value = 1
}

function getRecordSelectionKey(row) {
  // Selection key 使用执行记录 ID 保持跨分页多选状态
  return String(row?.id || '').trim()
}

async function openRecordDetail(row) {
  recordDetail.value = row
  recordDetailFiles.value = []
  recordDetailVisible.value = true
  if (!row.id) return

  const data = await getTaskRecordDetail(row.id)
  recordDetail.value = data.record || row
  recordDetailFiles.value = Array.isArray(data.files) ? data.files : []
}

function downloadRecordFile(row) {
  if (!row?.id) return
  window.open(getTaskRecordFileDownloadUrl(row.id), '_blank', 'noopener')
}

async function handleRetryRecord(row) {
  if (!row.task_id) return

  await executeTask(row.task_id, {
    client_id: row.client_id || '',
    client_ip: row.client_ip || '',
    machine_id: row.machine_id || '',
    node_id: row.node_id || '',
    params: row.params || {},
  })
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '任务已重新下发' })
  await loadRecords()
}

async function handleBatchDeleteRecords() {
  const ids = selectedRecordIds.value.map((id) => Number(id)).filter((id) => id > 0)
  if (ids.length === 0) return

  const confirmed = await appConfirm({
    title: '批量删除执行记录',
    message: `确认删除选中的 ${ids.length} 条执行记录吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteTaskRecords(ids)
  resetRecordSelection(recordTableRef)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已删除选中执行记录' })
  await loadRecords()
}

async function handleDeleteRecord(row) {
  const id = Number(row?.id || 0)
  if (id <= 0) return

  const confirmed = await appConfirm({
    title: '删除执行记录',
    message: '确认删除这条执行记录吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteTaskRecords([id])
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '执行记录已删除' })
  await loadRecords()
}

function resetRecordFilters() {
  recordFilters.task_name = ''
  recordFilters.workflow_name = ''
  recordFilters.client_ip = ''
  recordFilters.node_ids = []
  recordFilters.execute_time_range = []
  recordFilters.status = ''
}

function getExecuteTimeRange() {
  const range = Array.isArray(recordFilters.execute_time_range)
    ? recordFilters.execute_time_range
    : []
  return [range[0] || '', range[1] || '']
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function getClientNodeId(row) {
  return row?.node_id || row?.nodeId || ''
}

function getClientNodeName(row) {
  return row?.node_name || row?.nodeName || ''
}

function getRecordNodeText(row) {
  const nodeName = String(row?.node_name || row?.nodeName || '').trim()
  const nodeId = String(row?.node_id || row?.nodeId || '').trim()
  if (nodeName && nodeName !== nodeId) return `${nodeName} / ${nodeId}`
  return nodeId || nodeName
}

function getRecordStatusTag(status) {
  if (status === 'success' || status === 'done') return 'success'
  if (status === 'failed' || status === 'error') return 'danger'
  if (status === 'running') return 'warning'
  if (status === 'queued') return 'primary'
  return 'info'
}

function getStatusText(status) {
  if (status === 'pending') return '待执行'
  if (status === 'queued') return '已下发'
  if (status === 'running') return '执行中'
  if (status === 'success' || status === 'done') return '成功'
  if (status === 'failed' || status === 'error') return '失败'
  if (status === 'cancelled') return '已取消'
  return status || ''
}

function getTriggerText(triggerType) {
  if (triggerType === 'cron') return '定时'
  if (triggerType === 'task_create') return '创建即执行'
  if (triggerType === 'skill') return 'Skill触发'
  if (triggerType === 'system') return '系统'
  return '手动'
}

function formatDuration(value) {
  const duration = Number(value) || 0
  if (duration <= 0) return ''
  if (duration < 1000) return `${duration}ms`
  return `${(duration / 1000).toFixed(duration >= 10000 ? 0 : 1)}s`
}

function formatFileSize(value) {
  const size = Number(value) || 0
  if (size <= 0) return ''
  if (size < 1024) return `${size}B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)}KB`
  return `${(size / 1024 / 1024).toFixed(1)}MB`
}

function formatNumber(value) {
  const numberValue = Number(value) || 0
  return numberValue > 0 ? String(numberValue) : ''
}

function formatDate(value) {
  return formatBaseDate(value, { fallback: '' })
}
</script>

<style scoped lang="scss">
.record-filters {
  display: flex;
  align-items: flex-start;
  flex-direction: row;
  gap: 10px 20px;
}

.record-filter-fields {
  display: flex;
  align-items: stretch;
  flex-direction: column;
  flex: 1 1 auto;
  gap: 10px;
  width: 100%;
  min-width: 0;
}

.record-filter-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px 18px;
  min-width: 0;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.filter-item--client :deep(.el-select),
.filter-item--node :deep(.el-select) {
  width: 100%;
}

.filter-item--task {
  flex: 0 0 260px;
}

.filter-item--workflow {
  flex: 0 0 330px;
}

.filter-item--execute-time {
  flex: 0 1 500px;
}

.filter-item--client {
  flex: 0 0 260px;
}

.filter-item--node {
  flex: 0 0 320px;
}

.filter-item--status {
  flex: 0 0 168px;
}

.filter-item--execute-time :deep(.app-time-range-filter) {
  grid-template-columns: minmax(130px, 1fr) auto minmax(130px, 1fr);
  width: 100%;
}

.filter-item--execute-time :deep(.el-date-editor) {
  width: 100%;
  min-width: 0;
}

.filter-item :deep(.el-input),
.filter-item :deep(.el-select) {
  flex: 1;
  min-width: 0;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.detail-form {
  display: grid;
  gap: 16px;
  width: 100%;
  min-width: 0;
  overflow: visible;
}

.detail-descriptions {
  margin-bottom: 0;
  max-width: 100%;
}

.detail-descriptions :deep(.el-descriptions__cell) {
  min-width: 0;
  word-break: break-word;
}

.detail-block {
  min-width: 0;
  margin-top: 0;
}

.detail-block h3 {
  margin: 0 0 8px;
  color: #303133;
  font-size: 14px;
}

.detail-json {
  box-sizing: border-box;
  max-width: 100%;
  padding: 12px;
  margin: 0;
  overflow: visible;
  color: #303133;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.detail-files-table {
  width: 100%;
}

.detail-files-table :deep(.el-table__inner-wrapper),
.detail-files-table :deep(.el-scrollbar),
.detail-files-table :deep(.el-scrollbar__wrap),
.detail-files-table :deep(.el-scrollbar__view) {
  min-width: 0;
}

.detail-files-table :deep(.el-table__header),
.detail-files-table :deep(.el-table__body) {
  width: 100% !important;
  table-layout: fixed;
}

.detail-files-table :deep(.cell) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1280px) {
  .filter-item--task {
    flex-basis: 240px;
  }

  .filter-item--workflow {
    flex-basis: 300px;
  }

  .filter-item--execute-time {
    flex-basis: 460px;
  }

  .filter-item--client {
    flex-basis: 240px;
  }

  .filter-item--node {
    flex-basis: 300px;
  }

  .filter-item--execute-time :deep(.app-time-range-filter) {
    grid-template-columns: minmax(120px, 1fr) auto minmax(120px, 1fr);
  }
}

@media (max-width: 640px) {

  .page-actions,
  .record-filters,
  .record-filter-fields,
  .record-filter-row,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .record-filter-fields {
    grid-template-columns: 1fr;
    width: 100%;
  }

  .filter-item--task,
  .filter-item--workflow,
  .filter-item--client,
  .filter-item--node,
  .filter-item--execute-time,
  .filter-item--status {
    flex: 1 1 auto;
    width: 100%;
  }

  .filter-item--execute-time :deep(.app-time-range-filter) {
    grid-template-columns: 1fr;
    width: 100%;
  }
}
</style>
