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

      <!-- Export scope 导出范围，决定使用选中、筛选还是全部工作流 -->
      <div class="filter-item filter-item--export-scope">
        <span class="filter-label">导出范围：</span>
        <el-select v-model="exportScope" placeholder="导出范围">
          <el-option
            v-for="item in exportScopeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>

      <el-button class="reset-button" @click="resetFilters">重置</el-button>
      <el-button
        class="export-skill-button"
        type="primary"
        :icon="Download"
        :loading="skillExporting"
        :disabled="exportTargetWorkflows.length === 0"
        @click="handleExportSkill"
      >
        导出 Skill
      </el-button>
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
      :row-key="getWorkflowId"
      empty-text="暂无工作流"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="40" align="center" reserve-selection />
      <el-table-column prop="name" label="工作流名称" min-width="220">
        <template #default="{ row }">
          <button class="workflow-name__link" type="button" @click="openWorkflow(getWorkflowId(row))">
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
          <div v-if="getTriggerParameters(row).length > 0" class="trigger-param-list">
            <div
              v-for="(param, index) in getTriggerParameters(row)"
              :key="getTriggerParamKey(param, index)"
              class="trigger-param-item"
              :title="getTriggerParamTitle(param)"
            >
              <span class="trigger-param-field trigger-param-field--name">
                <small>参数名</small>
                <strong>{{ formatTriggerParamName(param) }}</strong>
              </span>
              <span class="trigger-param-field trigger-param-field--meaning">
                <small>意义</small>
                <span>{{ formatTriggerParamMeaning(param) }}</span>
              </span>
              <span class="trigger-param-field trigger-param-field--default">
                <small>默认值</small>
                <span>{{ formatTriggerParamDefaultText(param?.defaultValue) }}</span>
              </span>
            </div>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="节点数" width="70" align="center">
        <template #default="{ row }">
          {{ getNodeCount(row) }}
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
          <el-button type="primary" link @click="executeWorkflow(row)">
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

    <AppDialog
      v-model="paramDialogVisible"
      title="执行参数"
      width="640px"
      :loading="paramSubmitting"
      @confirm="confirmExecuteWithParams"
      @closed="resetParamDialog"
    >
      <el-form label-width="120px" class="workflow-param-form">
        <el-form-item
          v-for="param in paramFormItems"
          :key="param.key"
          :label="param.name"
          :required="param.required"
        >
          <div class="workflow-param-control">
            <el-switch
              v-if="param.type === 'checkbox'"
              v-model="param.value"
              active-text="是"
              inactive-text="否"
            />
            <el-input
              v-else
              v-model="param.value"
              :type="param.type === 'json' ? 'textarea' : 'text'"
              :rows="param.type === 'json' ? 4 : undefined"
              :placeholder="param.placeholder"
            />
            <span v-if="param.description" class="workflow-param-help">
              {{ param.description }}
            </span>
            <span v-if="param.defaultText" class="workflow-param-help">
              默认值：{{ param.defaultText }}
            </span>
          </div>
        </el-form-item>
      </el-form>
    </AppDialog>
  </section>
</template>

<script setup>
import { Download, SortDown, SortUp } from '@element-plus/icons-vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import {
  exportAgentAutomaSkill,
  listAgentAutomaWorkflows,
  openAgentAutomaWorkflow,
  runAgentAutomaWorkflow,
} from '@/services/automa'

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
const exportScopeOptions = [
  { label: '已选工作流', value: 'selected' },
  { label: '当前筛选结果', value: 'filtered' },
  { label: '全部工作流', value: 'all' },
]

// Agent state 工作流页面只通过后端读取浏览器执行端，不直接桥接管理端 Automa
const error = ref('')
const loading = ref(false)
const skillExporting = ref(false)
const workflows = ref([])
const exportScope = ref('filtered')
const selectedWorkflows = ref([])
const agentBrowserId = ref('')
const paramDialogVisible = ref(false)
const paramSubmitting = ref(false)
const paramWorkflow = ref(null)
const paramFormItems = ref([])

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

const exportTargetWorkflows = computed(() => {
  if (exportScope.value === 'selected') return selectedWorkflows.value
  if (exportScope.value === 'all') return workflows.value
  return sortedWorkflows.value
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

async function loadWorkflows() {
  loading.value = true
  error.value = ''

  try {
    // Agent list 工作流列表来自当前启动浏览器中的 browser-agent
    const data = await listAgentAutomaWorkflows()
    workflows.value = normalizeAgentWorkflows(data)
    selectedWorkflows.value = []
    agentBrowserId.value = data?.browser_id || ''
  } catch (err) {
    workflows.value = []
    selectedWorkflows.value = []
    agentBrowserId.value = ''
    error.value = err.message
  } finally {
    loading.value = false
  }
}

async function openWorkflow(workflowId) {
  if (!workflowId) return
  error.value = ''

  try {
    // Agent open 打开命令经后端转发到对应浏览器执行端
    await openAgentAutomaWorkflow(workflowId, agentBrowserId.value)
  } catch (err) {
    error.value = err.message
  }
}

async function executeWorkflow(workflow) {
  const workflowId = getWorkflowId(workflow)
  if (!workflowId) return
  error.value = ''

  const params = getUniqueTriggerParameters(workflow)
  if (params.length > 0) {
    openParamDialog(workflow, params)
    return
  }

  try {
    // Agent run 执行命令经后端转发到对应浏览器执行端
    await runAgentAutomaWorkflow(workflowId, agentBrowserId.value, {})
  } catch (err) {
    error.value = err.message
  }
}

function openParamDialog(workflow, params) {
  // Param dialog 复用公共弹窗，在 BrowserFlow 侧完成参数填写与校验
  paramWorkflow.value = workflow
  paramFormItems.value = params.map((param, index) => createParamFormItem(param, index))
  paramDialogVisible.value = true
}

async function confirmExecuteWithParams() {
  const workflow = paramWorkflow.value
  const workflowId = getWorkflowId(workflow)
  if (!workflowId) return

  const variables = buildRunVariablesFromParamForm()
  if (!variables) return

  paramSubmitting.value = true
  error.value = ''
  try {
    // Agent run with variables 带参数执行，Automa 参数页由 BrowserFlow 跳过
    await runAgentAutomaWorkflow(workflowId, agentBrowserId.value, variables)
    paramDialogVisible.value = false
  } catch (err) {
    error.value = err.message
  } finally {
    paramSubmitting.value = false
  }
}

function resetParamDialog() {
  if (paramSubmitting.value) return
  paramWorkflow.value = null
  paramFormItems.value = []
}

function createParamFormItem(param, index) {
  const defaultValue = getTriggerParamDefaultRawValue(param)
  return {
    key: getTriggerParamKey(param, index),
    name: param?.name || `param_${index + 1}`,
    type: param?.type || 'string',
    description: param?.description || '',
    placeholder: param?.placeholder || '',
    required: isTriggerParamRequired(param),
    defaultText: formatTriggerParamDefaultValue(defaultValue),
    value: normalizeParamInputValue(param, defaultValue),
  }
}

function buildRunVariablesFromParamForm() {
  const variables = {}

  for (const item of paramFormItems.value) {
    if (isMissingRequiredParam(item)) {
      appMessage({ type: APP_MESSAGE_TYPE.warning, message: `请填写必填参数：${item.name}` })
      return null
    }

    const parsedValue = parseParamFormValue(item)
    if (parsedValue === undefined && item.type === 'json' && !isEmptyParamValue(item.value)) return null
    variables[item.name] = parsedValue
  }

  return variables
}

function parseParamFormValue(item) {
  if (item.type === 'number') {
    const value = Number(item.value)
    return Number.isNaN(value) ? 0 : value
  }
  if (item.type === 'json') {
    if (isEmptyParamValue(item.value)) return null
    try {
      return JSON.parse(item.value)
    } catch {
      appMessage({ type: APP_MESSAGE_TYPE.warning, message: `参数 ${item.name} 不是有效 JSON` })
      return undefined
    }
  }
  if (item.type === 'checkbox') return Boolean(item.value)
  return item.value
}

async function handleExportSkill() {
  if (exportTargetWorkflows.value.length === 0) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '暂无可导出的工作流' })
    return
  }

  skillExporting.value = true
  error.value = ''

  try {
    // Export scope 按用户选择决定导出范围，后端再按 scope 做最终解释
    const workflowIds = exportTargetWorkflows.value.map(getWorkflowId).filter(Boolean)
    const blob = await exportAgentAutomaSkill({
      browserId: agentBrowserId.value,
      scope: exportScope.value,
      workflowIds,
    })
    downloadBlob(blob, 'SKILL_AUTOMA.md')
    appMessage({ type: APP_MESSAGE_TYPE.success, message: 'Skill 已导出' })
  } catch (err) {
    error.value = err.message
    appMessage({ type: APP_MESSAGE_TYPE.error, message: err.message || '导出 Skill 失败' })
  } finally {
    skillExporting.value = false
  }
}

function downloadBlob(blob, filename) {
  // Download file 创建临时链接触发浏览器下载
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function handleSelectionChange(rows) {
  // Selected rows 保留当前已选工作流，供“已选工作流”导出范围使用
  selectedWorkflows.value = Array.isArray(rows) ? rows : []
}

function normalizeAgentWorkflows(data) {
  const list = data?.workflows || data?.list || data || []
  return Array.isArray(list) ? list : []
}

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
  if (key === 'mostUsed') return runCounts[getWorkflowId(workflow)] || 0
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

function getUniqueTriggerParameters(workflow) {
  // Unique params 与 Automa 参数面板保持一致，同名参数只保留第一个
  const seen = new Set()
  return getTriggerParameters(workflow).filter((param, index) => {
    const name = param?.name || `param_${index + 1}`
    if (seen.has(name)) return false
    seen.add(name)
    return true
  })
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

function formatTriggerParamName(param) {
  // Param name keeps Automa variable key visible 参数名直接展示 Automa 变量键
  return param?.name || '未命名'
}

function formatTriggerParamMeaning(param) {
  // Meaning prefers explicit description, then placeholder 参数意义优先使用说明，其次使用占位提示
  return param?.description || param?.placeholder || '-'
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

function formatTriggerParamDefaultText(value) {
  // Default text keeps empty value blank 默认值为空时保持空白
  return formatTriggerParamDefaultValue(value)
}

function getTriggerParamDefaultRawValue(param) {
  if (param?.defaultValue !== undefined) return param.defaultValue
  if (param?.default !== undefined) return param.default
  if (param?.value !== undefined) return param.value
  return ''
}

function isTriggerParamRequired(param) {
  return Boolean(param?.required || param?.data?.required)
}

function normalizeParamInputValue(param, value) {
  if (param?.type === 'checkbox') return Boolean(value)
  if (param?.type === 'json' && value && typeof value === 'object') return JSON.stringify(value, null, 2)
  return formatTriggerParamDefaultValue(value)
}

function isEmptyParamValue(value) {
  return value === undefined || value === null || value === ''
}

function isMissingRequiredParam(item) {
  if (!item.required) return false
  if (item.type === 'checkbox') return item.value !== true
  return isEmptyParamValue(item.value)
}

function getNodeCount(workflow) {
  return workflow.drawflow?.nodes?.length || 0
}

function getWorkflowId(workflow) {
  return workflow?.id || workflow?.workflowId || workflow?.workflow_id || workflow?.automaId || workflow?.automa_id || ''
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

.filter-item--export-scope {
  flex: 0 1 220px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
  white-space: nowrap;
}

.filter-item--name :deep(.el-input),
.filter-item--status :deep(.el-select),
.filter-item--export-scope :deep(.el-select),
.filter-item--time :deep(.el-date-editor) {
  width: 100%;
  min-width: 0;
}

.reset-button {
  margin-left: auto;
  flex-shrink: 0;
}

.export-skill-button {
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

.trigger-param-list {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.trigger-param-item {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding: 8px 10px;
  background: #f8fbff;
  border: 1px solid #dce8f5;
  border-radius: 6px;
}

.trigger-param-field {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  align-items: center;
  gap: 3px;
  min-width: 0;
}

.trigger-param-field small {
  color: #909399;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
}

.trigger-param-field span,
.trigger-param-field strong {
  min-width: 0;
  overflow: hidden;
  color: #303133;
  font-size: 13px;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trigger-param-field strong {
  font-weight: 700;
}

.workflow-error {
  margin-bottom: 4px;
}

.workflow-param-form {
  max-height: 58vh;
  overflow: auto;
  padding-right: 8px;
}

.workflow-param-control {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.workflow-param-help {
  color: #909399;
  font-size: 13px;
  line-height: 1.4;
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
  .filter-item--export-scope,
  .workflow-sort {
    width: 100%;
    max-width: none;
  }

  .reset-button,
  .export-skill-button {
    margin-left: 0;
  }

  .workflow-sort {
    grid-template-columns: 40px minmax(0, 1fr);
  }
}
</style>
