<template>
  <section class="workflow-page">
    <div class="workflow-filters">
      <!-- Name filter 名称筛选，按工作流名称做模糊匹配 -->
      <div class="filter-item filter-item--name">
        <span class="filter-label">工作流名称：</span>
        <el-input
          v-model="searchKeywordInput"
          clearable
          placeholder="请输入工作流名称"
        />
      </div>

      <!-- Created filter 创建时间筛选，按起止时间过滤 -->
      <div class="filter-item filter-item--time">
        <span class="filter-label">创建时间：</span>
        <AppTimeRangeFilter v-model="createdTimeRange" value-format="x" />
      </div>

      <!-- Status filter 状态筛选，按启用状态过滤工作流 -->
      <div class="filter-item filter-item--status">
        <span class="filter-label">状态：</span>
        <el-select v-model="statusFilter" placeholder="全部">
          <el-option label="全部" value="" />
          <el-option label="启用中" value="enabled" />
          <el-option label="已禁用" value="disabled" />
        </el-select>
      </div>

      <!-- Sort controls 排序控件，对齐 Automa 的排序字段和默认规则 -->
      <div class="workflow-sort">
        <el-button class="sort-order-button" @click="toggleSortOrder">
          <el-icon>
            <component :is="sortOrder === 'asc' ? SortUp : SortDown" />
          </el-icon>
        </el-button>
        <el-select v-model="sortBy" class="sort-select" placeholder="排序方式">
          <el-option
            v-for="item in sortOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>

      <el-button class="reset-button" @click="resetFilters">重置</el-button>
    </div>

    <!-- Error message 错误提示，展示扩展未响应或调用失败原因 -->
    <el-alert
      v-if="error"
      class="workflow-error"
      type="error"
      :title="error"
      show-icon
      :closable="false"
    />

    <!-- Workflow table 工作流列表，展示常用元信息 -->
    <el-table
      v-loading="loading"
      class="workflow-table adaptive-table"
      :data="pagedWorkflows"
      border
      height="100%"
      row-key="id"
      empty-text="暂无工作流"
    >
      <el-table-column prop="name" label="工作流名称" min-width="220">
        <template #default="{ row }">
          <button class="workflow-name__link" type="button" @click="openWorkflow(row.id)">
            {{ row.name || '' }}
          </button>
        </template>
      </el-table-column>

      <el-table-column prop="description" label="工作流描述" min-width="220">
        <template #default="{ row }">
          <span class="workflow-description" :title="row.description || ''">
            {{ row.description || '' }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="状态" width="72" align="center">
        <template #default="{ row }">
          <el-tag :type="row.isDisabled ? 'info' : 'success'" effect="plain">
            {{ row.isDisabled ? '已禁用' : '启用中' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="触发器参数" min-width="220">
        <template #default="{ row }">
          <div v-if="getTriggerParameters(row).length > 0" class="trigger-params">
            <el-tag
              v-for="(param, index) in getTriggerParameters(row)"
              :key="getTriggerParamKey(param, index)"
              class="trigger-param-tag"
              effect="plain"
              :title="getTriggerParamTitle(param)"
            >
              {{ formatTriggerParam(param) }}
            </el-tag>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="节点数" width="70" align="center">
        <template #default="{ row }">
          {{ getNodeCount(row) }}
        </template>
      </el-table-column>

      <el-table-column label="数据列" width="70" align="center">
        <template #default="{ row }">
          {{ getDataColumnCount(row) }}
        </template>
      </el-table-column>

      <el-table-column label="创建时间" width="160" class-name="nowrap-column">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>

      <el-table-column label="更新时间" width="160" class-name="nowrap-column">
        <template #default="{ row }">
          {{ formatDate(row.updatedAt) }}
        </template>
      </el-table-column>

      <el-table-column label="操作" width="90" align="center">
        <template #default="{ row }">
          <el-button type="primary" link @click="executeWorkflow(row.id)">
            执行
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination 分页组件，控制当前列表的本地分页展示 -->
    <AppPagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="pageSizes"
      :total="sortedWorkflows.length"
    />
  </section>
</template>

<script setup>
import { SortDown, SortUp } from '@element-plus/icons-vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppPagination from '@/components/AppPagination.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { useAutoma } from '@/composables/useAutoma'

const PAGE_SIZE = 10
const SORT_STORAGE_KEY = 'workflow-sorts'

const savedSorts = JSON.parse(localStorage.getItem(SORT_STORAGE_KEY) || '{}')

const currentPage = ref(1)
const pageSize = ref(savedSorts.perPage || PAGE_SIZE)
const pageSizes = [10, 30, 60]
const searchKeywordInput = ref('')
const searchKeyword = ref('')
const createdTimeRange = ref([])
const statusFilter = ref('')
const sortBy = ref(savedSorts.sortBy || 'createdAt')
const sortOrder = ref(savedSorts.sortOrder || 'desc')
const filterSearchDelay = 200

let filterSearchTimer = 0

const sortOptions = [
  { label: '名称', value: 'name' },
  { label: '创建日期', value: 'createdAt' },
  { label: '上次更新', value: 'updatedAt' },
  { label: '最常用', value: 'mostUsed' },
]

// Automa state 扩展状态，页面只关心列表、加载态和执行方法
const {
  error,
  loading,
  workflows,
  loadWorkflows,
  openWorkflow,
  executeWorkflow,
} = useAutoma()

const filteredWorkflows = computed(() => {
  const keyword = searchKeyword.value.trim().toLocaleLowerCase()

  return workflows.value.filter((workflow) => {
    const name = workflow.name || ''
    const createdAt = Number(workflow.createdAt || 0)
    const [createdStart, createdEnd] = getCreatedTimeRange()
    const isNameMatched = !keyword || name.toLocaleLowerCase().includes(keyword)
    const isAfterStart = !createdStart || createdAt >= Number(createdStart)
    const isBeforeEnd = !createdEnd || createdAt <= Number(createdEnd)
    const isStatusMatched =
      !statusFilter.value ||
      (statusFilter.value === 'enabled' && !workflow.isDisabled) ||
      (statusFilter.value === 'disabled' && workflow.isDisabled)

    return isNameMatched && isAfterStart && isBeforeEnd && isStatusMatched
  })
})

const sortedWorkflows = computed(() => {
  return sortWorkflows({
    data: filteredWorkflows.value,
    key: sortBy.value,
    order: sortOrder.value,
  })
})

const pagedWorkflows = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return sortedWorkflows.value.slice(start, start + pageSize.value)
})

const toggleSortOrder = () => {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

const resetFilters = () => {
  clearFilterSearchTimer()
  searchKeywordInput.value = ''
  searchKeyword.value = ''
  createdTimeRange.value = []
  statusFilter.value = ''
  currentPage.value = 1
}

const saveSorts = () => {
  localStorage.setItem(
    SORT_STORAGE_KEY,
    JSON.stringify({
      sortBy: sortBy.value,
      sortOrder: sortOrder.value,
      perPage: pageSize.value,
    }),
  )
}

watch([workflows, pageSize, searchKeyword, createdTimeRange, statusFilter, sortBy, sortOrder], () => {
  currentPage.value = getSafePage({
    total: sortedWorkflows.value.length,
    page: currentPage.value,
    size: pageSize.value,
  })
  saveSorts()
})

watch(searchKeywordInput, () => {
  debounceSearchKeyword()
})

watch([searchKeyword, createdTimeRange, statusFilter, sortBy, sortOrder], () => {
  currentPage.value = 1
})

onMounted(() => {
  loadWorkflows()
})

onBeforeUnmount(() => {
  clearFilterSearchTimer()
})

function sortWorkflows({ data, key, order = 'asc' }) {
  const runCounts = getRunCounts()

  return data.slice().sort((a, b) => {
    let itemA = getSortValue(a, key, runCounts)
    let itemB = getSortValue(b, key, runCounts)

    if (typeof itemA === 'string') itemA = itemA.toLocaleLowerCase()
    if (typeof itemB === 'string') itemB = itemB.toLocaleLowerCase()

    let comparison = 0
    if (itemA > itemB) comparison = 1
    if (itemA < itemB) comparison = -1

    return order === 'desc' ? comparison * -1 : comparison
  })
}

function getSortValue(workflow, key, runCounts) {
  if (key === 'mostUsed') return runCounts[workflow.id] || 0
  return workflow[key] ?? ''
}

function getRunCounts() {
  try {
    return JSON.parse(localStorage.getItem('runCounts') || '{}') || {}
  } catch {
    return {}
  }
}

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function debounceSearchKeyword() {
  clearFilterSearchTimer()
  // Filter debounce 可输入筛选条件 200ms 防抖，保持与其他页面一致。
  filterSearchTimer = window.setTimeout(() => {
    searchKeyword.value = searchKeywordInput.value
    filterSearchTimer = 0
  }, filterSearchDelay)
}

function clearFilterSearchTimer() {
  if (!filterSearchTimer) return
  window.clearTimeout(filterSearchTimer)
  filterSearchTimer = 0
}

function getTriggerParameters(workflow) {
  // Trigger parameters 优先读取顶层触发器，兜底读取流程图 trigger 节点
  const triggerData = workflow?.trigger || getTriggerNode(workflow)?.data || {}
  return Array.isArray(triggerData.parameters) ? triggerData.parameters : []
}

function getTriggerNode(workflow) {
  // Trigger node Automa 新版流程图中触发器通常是 label 为 trigger 的节点
  if (workflow?.drawflow?.nodes) {
    return workflow.drawflow.nodes.find((node) => node?.label === 'trigger') || null
  }

  // Legacy drawflow 兼容 Automa 旧版 drawflow.Home.data 结构
  const legacyNodes = workflow?.drawflow?.drawflow?.Home?.data
  if (legacyNodes) {
    return Object.values(legacyNodes).find((node) => node?.name === 'trigger') || null
  }

  return null
}

function getTriggerParamKey(param, index) {
  // Param key 使用索引兜底，避免同名参数 key 冲突
  return `${param?.name || ''}_${param?.type || ''}_${param?.placeholder || ''}_${index}`
}

function getTriggerParamTitle(param) {
  const parts = [
    `名称：${param?.name || '-'}`,
    `类型：${formatTriggerParamType(param?.type)}`,
    `默认值：${formatTriggerParamDefaultValue(param?.defaultValue)}`,
  ]
  if (param?.placeholder) parts.push(`占位提示：${param.placeholder}`)
  if (param?.description) parts.push(`说明：${param.description}`)
  if (param?.data?.required) parts.push('必填：是')
  return parts.join('\n')
}

function formatTriggerParam(param) {
  const name = param?.name || '未命名'
  const type = formatTriggerParamType(param?.type)
  const defaultValue = formatTriggerParamDefaultValue(param?.defaultValue)
  return defaultValue ? `${name} / ${type} = ${defaultValue}` : `${name} / ${type}`
}

function formatTriggerParamType(type) {
  const typeMap = {
    string: '文本',
    number: '数字',
    json: 'JSON',
    checkbox: '勾选框',
  }
  return typeMap[type] || type || '未知'
}

function formatTriggerParamDefaultValue(value) {
  if (value === undefined || value === null || value === '') return ''
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function getNodeCount(workflow) {
  return workflow.drawflow?.nodes?.length || 0
}

function getDataColumnCount(workflow) {
  return workflow.dataColumns?.length || workflow.table?.length || 0
}

function getCreatedTimeRange() {
  const range = Array.isArray(createdTimeRange.value) ? createdTimeRange.value : []
  return [range[0] || '', range[1] || '']
}

function formatDate(value) {
  if (!value) return ''

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
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
</script>

<style scoped>
.workflow-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.workflow-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 16px;
}

.filter-item {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.filter-item--name {
  flex: 1 1 300px;
  max-width: 380px;
}

.filter-item--time {
  flex: 2 1 460px;
  max-width: 560px;
}

.filter-item--status {
  flex: 0 1 190px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
  white-space: nowrap;
}

.filter-item--name :deep(.el-input),
.filter-item--status :deep(.el-select),
.filter-item--time :deep(.el-date-editor) {
  width: 100%;
  min-width: 0;
}

.reset-button {
  margin-left: auto;
  flex-shrink: 0;
}

.workflow-sort {
  display: grid;
  grid-template-columns: 40px minmax(120px, 150px);
  align-items: center;
  flex-shrink: 0;
}

.sort-order-button {
  border-right: 0;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.sort-select {
  width: 100%;
}

.sort-select :deep(.el-select__wrapper) {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}

.workflow-table {
  flex: 1;
  min-height: 0;
  width: 100%;
}

.workflow-name__link {
  min-width: 0;
  padding: 0;
  overflow: hidden;
  color: #303133;
  font: inherit;
  font-weight: 700;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: transparent;
  border: 0;
  cursor: pointer;
}

.workflow-name__link:hover {
  color: #409eff;
}

.workflow-description {
  display: block;
  overflow: hidden;
  color: #909399;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trigger-params {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  min-width: 0;
}

.trigger-param-tag {
  max-width: 100%;
}

.trigger-param-tag :deep(.el-tag__content) {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-error {
  margin-bottom: 4px;
}

@media (max-width: 640px) {
  .workflow-page {
    height: auto;
    overflow: visible;
  }

  .workflow-filters {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item,
  .filter-item--name,
  .filter-item--time,
  .filter-item--status,
  .workflow-sort {
    width: 100%;
    max-width: none;
  }

  .reset-button {
    margin-left: 0;
  }

  .workflow-sort {
    grid-template-columns: 40px minmax(0, 1fr);
  }
}
</style>
