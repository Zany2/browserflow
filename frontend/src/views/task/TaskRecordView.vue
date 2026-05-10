<template>
  <section class="task-record-page">
    <header class="page-actions">
      <el-button type="primary" :icon="RefreshRight" @click="loadRecords">刷新</el-button>
    </header>

    <section class="record-panel">
      <div class="record-filters">
        <div class="filter-item filter-item--workflow">
          <span class="filter-label">工作流名称</span>
          <el-input v-model="recordFilters.workflow_name" clearable placeholder="模糊检索工作流名称" />
        </div>

        <div class="filter-item filter-item--client">
          <span class="filter-label">客户端IP</span>
          <el-select v-model="recordFilters.client_ip" clearable filterable placeholder="选择或检索客户端 IP"
            :loading="clientIpLoading" @visible-change="handleClientIpSelectVisible">
            <el-option label="全部" value="" />
            <el-option v-for="clientIp in clientIpOptions" :key="clientIp" :label="clientIp" :value="clientIp" />
          </el-select>
        </div>

        <div class="filter-item filter-item--execute-time">
          <span class="filter-label">执行时间</span>
          <AppTimeRangeFilter v-model="recordFilters.execute_time_range" />
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

        <el-button @click="resetRecordFilters">重置</el-button>
      </div>

      <el-table v-loading="loadingRecords" class="record-table adaptive-table" :data="pagedRecords" border height="100%"
        row-key="id" empty-text="暂无执行记录">
        <el-table-column label="任务名称" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.task_name || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="工作流名称" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.workflow_name || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="客户端IP" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.client_ip || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="触发方式" width="80" align="center">
          <template #default="{ row }">
            {{ getTriggerText(row.trigger_type) }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="getRecordStatusTag(row.status)" effect="plain">
              {{ row.status_text || getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="执行耗时" width="90" align="center">
          <template #default="{ row }">
            {{ formatDuration(row.duration_ms) }}
          </template>
        </el-table-column>

        <el-table-column label="执行时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(row.started_at || row.created_at || row.startedAt || row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="错误信息" min-width="160">
          <template #default="{ row }">
            {{ row.error_message || '' }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="110" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openRecordDetail(row)">详情</el-button>
            <el-button link type="success" :disabled="!row.task_id" @click="handleRetryRecord(row)">
              重试
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <AppPagination v-model:current-page="recordPage" v-model:page-size="recordPageSize" :page-sizes="pageSizes"
        :total="records.length" />
    </section>

    <AppDialog v-model="recordDetailVisible" title="执行记录详情" width="720px">
      <div v-if="recordDetail" class="detail-form">
        <el-descriptions border :column="2" class="detail-descriptions">
          <el-descriptions-item label="记录 ID">{{ recordDetail.id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="任务">{{ recordDetail.task_name || recordDetail.task_id || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="工作流">{{ recordDetail.workflow_name || recordDetail.workflow_id || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="客户端">{{ recordDetail.client_name || recordDetail.client_id || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="客户端 IP">{{ recordDetail.client_ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="触发方式">{{ getTriggerText(recordDetail.trigger_type) }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ recordDetail.status_text || getStatusText(recordDetail.status)
          }}</el-descriptions-item>
          <el-descriptions-item label="执行耗时">{{ formatDuration(recordDetail.duration_ms) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(recordDetail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatDate(recordDetail.started_at) }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ formatDate(recordDetail.finished_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(recordDetail.updated_at) }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-block">
          <h3>错误信息</h3>
          <pre class="detail-json">{{ recordDetail.error_message || '-' }}</pre>
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
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RefreshRight } from '@element-plus/icons-vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { listClients } from '@/services/client'
import { executeTask, getTaskRecordDetail, listTaskRecords } from '@/services/task'

const records = ref([])
const clientIpOptions = ref([])
const loadingRecords = ref(false)
const clientIpLoading = ref(false)
const recordPage = ref(1)
const recordPageSize = ref(10)
const pageSizes = [10, 30, 60]
const recordDetailVisible = ref(false)
const recordDetail = ref(null)
const filterSearchDelay = 200
let filterSearchTimer = null

const recordFilters = reactive({
  workflow_name: '',
  client_ip: '',
  execute_time_range: [],
  status: '',
})

const pagedRecords = computed(() => {
  const start = (recordPage.value - 1) * recordPageSize.value
  return records.value.slice(start, start + recordPageSize.value)
})

onMounted(() => {
  loadRecords()
  loadClientIpOptions()
})

onBeforeUnmount(() => {
  clearFilterSearchTimer()
})

watch(() => recordFilters.workflow_name, () => {
  scheduleFilterSearch()
})

watch(() => [recordFilters.client_ip, recordFilters.execute_time_range, recordFilters.status], () => {
  clearFilterSearchTimer()
  recordPage.value = 1
  loadRecords()
})

watch([records, recordPageSize], () => {
  recordPage.value = getSafePage({
    total: records.value.length,
    page: recordPage.value,
    size: recordPageSize.value,
  })
})

async function loadRecords() {
  loadingRecords.value = true
  try {
    const [startTime, endTime] = getExecuteTimeRange()
    const data = await listTaskRecords({
      workflow_name: recordFilters.workflow_name.trim(),
      client_ip: recordFilters.client_ip.trim(),
      start_time: startTime,
      end_time: endTime,
      status: recordFilters.status.trim(),
    })
    records.value = sortByTimeDesc(normalizeList(data, 'records'))
  } finally {
    loadingRecords.value = false
  }
}

async function loadClientIpOptions() {
  clientIpLoading.value = true
  try {
    const data = await listClients()
    clientIpOptions.value = normalizeList(data, 'clients')
      .map(getClientIp)
      .filter(Boolean)
      .filter((clientIp, index, list) => list.indexOf(clientIp) === index)
  } finally {
    clientIpLoading.value = false
  }
}

function handleClientIpSelectVisible(opened) {
  if (opened) loadClientIpOptions()
}

// Filter debounce 手动输入筛选条件 200ms 防抖
function scheduleFilterSearch() {
  clearFilterSearchTimer()
  filterSearchTimer = window.setTimeout(() => {
    filterSearchTimer = null
    recordPage.value = 1
    loadRecords()
  }, filterSearchDelay)
}

function clearFilterSearchTimer() {
  if (!filterSearchTimer) return

  window.clearTimeout(filterSearchTimer)
  filterSearchTimer = null
}

async function openRecordDetail(row) {
  recordDetail.value = row
  recordDetailVisible.value = true
  if (!row.id) return

  const data = await getTaskRecordDetail(row.id)
  recordDetail.value = data.record || row
}

async function handleRetryRecord(row) {
  if (!row.task_id) return

  await executeTask(row.task_id, {
    client_id: row.client_id || '',
    client_ip: row.client_ip || '',
    params: row.params || {},
  })
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '任务已重新下发' })
  await loadRecords()
}

function resetRecordFilters() {
  recordFilters.workflow_name = ''
  recordFilters.client_ip = ''
  recordFilters.execute_time_range = []
  recordFilters.status = ''
}

function getExecuteTimeRange() {
  const range = Array.isArray(recordFilters.execute_time_range)
    ? recordFilters.execute_time_range
    : []
  return [range[0] || '', range[1] || '']
}

function normalizeList(data, fallbackKey) {
  const list = data?.list || data?.[fallbackKey] || []
  return Array.isArray(list) ? list : []
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function sortByTimeDesc(data) {
  return data.slice().sort((a, b) => getTimeValue(b) - getTimeValue(a))
}

function getTimeValue(row) {
  const value =
    row?.started_at || row?.created_at || row?.updated_at || row?.startedAt || row?.createdAt
  return value ? new Date(value).getTime() || 0 : 0
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
  return status || '-'
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

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'

  const year = date.getFullYear()
  const month = padDatePart(date.getMonth() + 1)
  const day = padDatePart(date.getDate())
  const hour = padDatePart(date.getHours())
  const minute = padDatePart(date.getMinutes())
  const second = padDatePart(date.getSeconds())
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

function padDatePart(value) {
  return String(value).padStart(2, '0')
}

function formatJSON(value) {
  return JSON.stringify(value ?? {}, null, 2)
}
</script>

<style scoped lang="scss">
.task-record-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
}

.page-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  flex-shrink: 0;
}

.record-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 16px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.record-filters {
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

.filter-item--workflow {
  width: 280px;
}

.filter-item--client {
  width: 240px;
}

.filter-item--execute-time {
  width: 460px;
}

.filter-item--client :deep(.el-select),
.filter-item--execute-time :deep(.el-date-editor) {
  width: 100%;
}

.filter-item--status {
  width: 180px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.record-table {
  flex: 1;
  min-height: 0;
}

.detail-form {
  display: grid;
  gap: 16px;
  max-height: 70vh;
  overflow: auto;
  padding-right: 4px;
}

.detail-descriptions {
  margin-bottom: 0;
}

.detail-block {
  margin-top: 0;
}

.detail-block h3 {
  margin: 0 0 8px;
  color: #303133;
  font-size: 14px;
}

.detail-json {
  padding: 12px;
  margin: 0;
  overflow: auto;
  color: #303133;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 640px) {

  .page-actions,
  .record-filters,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item--workflow,
  .filter-item--client,
  .filter-item--execute-time,
  .filter-item--status {
    width: 100%;
  }
}
</style>
