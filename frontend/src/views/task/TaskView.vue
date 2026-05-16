<template>
  <section class="task-page">
    <header class="page-actions">
      <el-button :icon="RefreshRight" @click="loadTasks">刷新</el-button>
      <el-button type="primary" :icon="Plus" @click="handleCreateTask">新增任务</el-button>
    </header>

    <section class="task-panel">
      <div class="task-filters">
        <div class="filter-item filter-item--keyword">
          <span class="filter-label">关键词</span>
          <el-input v-model="taskFilters.keyword" clearable placeholder="任务名称或任务说明" />
        </div>
        <div class="filter-item filter-item--workflow">
          <span class="filter-label">工作流名称</span>
          <el-input v-model="taskFilters.workflow_name" clearable placeholder="模糊检索工作流名称" />
        </div>
        <div class="filter-item filter-item--created-time">
          <span class="filter-label">创建时间</span>
          <AppTimeRangeFilter v-model="taskFilters.created_time_range" />
        </div>
        <div class="filter-item filter-item--enabled">
          <span class="filter-label">状态</span>
          <el-select v-model="taskFilters.enabled" clearable placeholder="全部">
            <el-option label="全部" value="" />
            <el-option label="启用" value="true" />
            <el-option label="停用" value="false" />
          </el-select>
        </div>
        <el-button @click="resetTaskFilters">重置</el-button>
      </div>

      <el-table
        v-loading="loadingTasks"
        class="task-table adaptive-table"
        :data="pagedTasks"
        border
        height="100%"
        row-key="id"
        empty-text="暂无任务配置"
      >
        <el-table-column prop="name" label="任务名称" min-width="140" />
        <el-table-column prop="description" label="任务说明" min-width="180" show-overflow-tooltip />
        <el-table-column label="工作流名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.workflow_name || row.workflow_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="执行客户端IP" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getTaskClientIp(row) }}
          </template>
        </el-table-column>
        <el-table-column label="执行计划" min-width="140">
          <template #default="{ row }">
            {{ getScheduleText(row) }}
          </template>
        </el-table-column>
        <el-table-column label="参数" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getTaskParamCount(row) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled === false ? 'info' : 'success'" effect="plain">
              {{ row.enabled === false ? '停用' : '启用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(row.created_at || row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="editTask(row)">编辑</el-button>
            <el-button link type="success" @click="handleExecuteTask(row)">执行</el-button>
            <el-button link type="danger" @click="handleDeleteTask(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <AppPagination
        v-model:current-page="taskPage"
        v-model:page-size="taskPageSize"
        :page-sizes="pageSizes"
        :total="tasks.length"
      />
    </section>

    <AppDialog
      v-model="taskDialogVisible"
      :title="taskDialogTitle"
      width="920px"
      @closed="handleDialogClosed"
    >
      <el-form label-width="120px" :model="taskForm">
        <el-form-item label="任务名称">
          <el-input v-model="taskForm.name" placeholder="请输入任务名称" />
        </el-form-item>

        <el-form-item label="任务说明">
          <el-input
            v-model="taskForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入任务说明"
          />
        </el-form-item>

        <el-form-item label="Automa 工作流">
          <el-select
            v-model="taskForm.workflow_id"
            class="full-width"
            filterable
            placeholder="请选择需要执行的工作流"
            :loading="workflowLoading"
            @change="handleWorkflowChange"
          >
            <el-option
              v-for="workflow in workflowOptions"
              :key="getWorkflowId(workflow)"
              :label="workflow.name || getWorkflowId(workflow)"
              :value="getWorkflowId(workflow)"
            >
              <div class="workflow-option">
                <span>{{ workflow.name || '未命名工作流' }}</span>
                <small>{{ getWorkflowId(workflow) }}</small>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="执行客户端">
          <div class="client-config">
            <div class="client-select-row">
              <el-select
                v-model="taskForm.client_id"
                class="client-select"
                clearable
                filterable
                placeholder="搜索并选择客户端名称、ID、IP、浏览器"
                :filter-method="handleClientKeywordFilter"
                :loading="clientLoading"
                @change="handleClientChange"
                @clear="handleClientClear"
                @visible-change="handleClientSelectVisibleChange"
              >
                <el-option
                  v-for="client in filteredClients"
                  :key="getClientId(client)"
                  :label="getClientOptionLabel(client)"
                  :value="getClientId(client)"
                >
                  <div class="client-option">
                    <span>{{ getClientIp(client) || '-' }}</span>
                    <small>{{ getClientSelectMeta(client) }}</small>
                  </div>
                </el-option>
              </el-select>
              <el-button :icon="RefreshRight" :loading="clientLoading" @click="loadClients">
                刷新客户端
              </el-button>
            </div>

            <div class="client-selection-summary">
              <template v-if="selectedClient">
                <el-tag effect="plain">{{ getClientIp(selectedClient) || '-' }}</el-tag>
                <span class="client-selection-text">{{ getClientSelectMeta(selectedClient) }}</span>
              </template>
              <span v-else class="client-selection-empty">未选择执行客户端</span>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="执行参数">
          <div class="params-editor">
            <div
              v-for="(item, index) in paramEntries"
              :key="`param_${index}`"
              class="param-row"
            >
              <el-input v-model="item.key" placeholder="key" />
              <el-input v-model="item.value" placeholder="value" />
              <el-button
                link
                type="danger"
                :disabled="paramEntries.length === 1"
                @click="removeParam(index)"
              >
                删除
              </el-button>
            </div>
            <el-button :icon="Plus" @click="addParam">新增参数</el-button>
            <span class="form-help">
              多个参数会在客户端执行 Automa 工作流时作为变量传入。
            </span>
          </div>
        </el-form-item>

        <el-form-item label="Cron 表达式">
          <div class="schedule-editor">
            <el-input
              v-model="taskForm.cron_expression"
              placeholder="可选，例如 0 */10 * * * *"
            />
            <span class="form-help">
              不填写时，后端按创建后执行一次处理；填写后按定时任务处理。
            </span>
          </div>
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="taskForm.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleDialogCancel">取消</el-button>
        <el-button @click="resetTaskForm">重置</el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveTask">
          {{ taskForm.id ? '保存修改' : '新增任务' }}
        </el-button>
      </template>
    </AppDialog>

    <AppDialog
      v-model="executeParamDialogVisible"
      title="执行参数"
      width="640px"
      :loading="executeParamSubmitting"
      @confirm="confirmExecuteTaskWithParams"
      @closed="resetExecuteParamDialog"
    >
      <el-form label-width="120px" class="execute-param-form">
        <el-form-item
          v-for="param in executeParamItems"
          :key="param.key"
          :label="param.name"
          :required="param.required"
        >
          <div class="execute-param-control">
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
            <span v-if="param.description" class="form-help">{{ param.description }}</span>
            <span v-if="param.defaultText" class="form-help">默认值：{{ param.defaultText }}</span>
          </div>
        </el-form-item>
      </el-form>
    </AppDialog>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Plus, RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { getAutomaWorkflowDetail, listAutomaWorkflows } from '@/services/automa'
import { listClients } from '@/services/client'
import { createTask, deleteTask, executeTask, listTasks, updateTask } from '@/services/task'

const tasks = ref([])
const workflowOptions = ref([])
const clientOptions = ref([])
const loadingTasks = ref(false)
const workflowLoading = ref(false)
const clientLoading = ref(false)
const saving = ref(false)
const taskDialogVisible = ref(false)
const executeParamDialogVisible = ref(false)
const executeParamSubmitting = ref(false)
const executeParamTask = ref(null)
const executeParamItems = ref([])
const taskPage = ref(1)
const taskPageSize = ref(10)
const pageSizes = [10, 30, 60]
const paramEntries = ref([createEmptyParamEntry()])
const filterSearchDelay = 200
let filterSearchTimer = null

const taskForm = reactive(createEmptyTaskForm())
const taskFilters = reactive({
  keyword: '',
  workflow_name: '',
  created_time_range: [],
  enabled: '',
})
const clientSelector = reactive({
  keyword: '',
})

const taskDialogTitle = computed(() => (taskForm.id ? '编辑任务' : '新增任务'))
const pagedTasks = computed(() => {
  const start = (taskPage.value - 1) * taskPageSize.value
  return tasks.value.slice(start, start + taskPageSize.value)
})
const filteredClients = computed(() => {
  const keyword = clientSelector.keyword.trim().toLowerCase()

  return clientOptions.value.filter((client) => {
    const keywordSource = [
      getClientId(client),
      getClientName(client),
      client.client_ip,
      client.ip,
      client.remote_ip,
      client.browser,
      client.browser_name,
      client.user_agent,
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()

    return !keyword || keywordSource.includes(keyword)
  })
})
const selectedClient = computed(() => findClientById(taskForm.client_id))

onMounted(() => {
  loadTasks()
  loadClients()
})

onBeforeUnmount(() => {
  clearFilterSearchTimer()
})

watch(() => [taskFilters.keyword, taskFilters.workflow_name], () => {
  scheduleFilterSearch()
})

watch(() => [taskFilters.created_time_range, taskFilters.enabled], () => {
  clearFilterSearchTimer()
  taskPage.value = 1
  loadTasks()
})

watch([tasks, taskPageSize], () => {
  taskPage.value = getSafePage({
    total: tasks.value.length,
    page: taskPage.value,
    size: taskPageSize.value,
  })
})

async function loadTasks() {
  loadingTasks.value = true
  try {
    const [startTime, endTime] = getCreatedTimeRange()
    const data = await listTasks({
      keyword: taskFilters.keyword.trim(),
      workflow_name: taskFilters.workflow_name.trim(),
      start_time: startTime,
      end_time: endTime,
      enabled: taskFilters.enabled,
    })
    tasks.value = sortByTimeDesc(normalizeList(data, 'tasks'))
  } finally {
    loadingTasks.value = false
  }
}

// Filter debounce 手动输入筛选条件 200ms 防抖
function scheduleFilterSearch() {
  clearFilterSearchTimer()
  filterSearchTimer = window.setTimeout(() => {
    filterSearchTimer = null
    taskPage.value = 1
    loadTasks()
  }, filterSearchDelay)
}

function clearFilterSearchTimer() {
  if (!filterSearchTimer) return

  window.clearTimeout(filterSearchTimer)
  filterSearchTimer = null
}

async function loadWorkflowOptions() {
  workflowLoading.value = true
  try {
    const data = await listAutomaWorkflows({ keyword: '' })
    workflowOptions.value = normalizeList(data, 'workflows').filter((item) => !item.is_deleted)
  } finally {
    workflowLoading.value = false
  }
}

async function loadClients() {
  clientLoading.value = true
  try {
    const data = await listClients({ status: '' })
    clientOptions.value = normalizeList(data, 'clients')
  } finally {
    clientLoading.value = false
  }
}

async function ensureDialogOptionsLoaded() {
  await Promise.all([
    workflowOptions.value.length > 0 ? Promise.resolve() : loadWorkflowOptions(),
    clientOptions.value.length > 0 ? Promise.resolve() : loadClients(),
  ])
}

async function handleCreateTask() {
  resetTaskForm()
  await ensureDialogOptionsLoaded()
  taskDialogVisible.value = true
}

async function editTask(row) {
  resetTaskForm()
  await ensureDialogOptionsLoaded()

  Object.assign(taskForm, {
    ...createEmptyTaskForm(),
    ...row,
    workflow_id: row.workflow_id || '',
    workflow_name: row.workflow_name || '',
    client_id: row.client_id || '',
    client_name: row.client_name || '',
    client_ip: row.client_ip || row.source_ip || '',
    cron_expression: row.cron_expression || row.cron || '',
    enabled: row.enabled !== false,
  })
  paramEntries.value = normalizeParamEntries(row.params)
  taskDialogVisible.value = true
}

async function handleSaveTask() {
  const payload = buildTaskPayload()
  if (!payload) return

  saving.value = true
  try {
    const data = taskForm.id
      ? await updateTask(taskForm.id, payload)
      : await createTask(payload)

    const savedTask = mergeSavedTask(data?.task, payload, taskForm.id)
    upsertTask(savedTask)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: taskForm.id ? '任务已保存' : '任务已新增' })
    taskDialogVisible.value = false
  } finally {
    saving.value = false
  }
}

function handleDialogCancel() {
  taskDialogVisible.value = false
}

function handleDialogClosed() {
  resetTaskForm()
}

async function handleDeleteTask(row) {
  const confirmed = await appConfirm({
    title: '删除任务',
    message: '确认删除这个任务吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteTask(row.id)
  tasks.value = tasks.value.filter((item) => item.id !== row.id)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '任务已删除' })
}

async function handleExecuteTask(row) {
  const params = await loadExecuteTriggerParameters(row)
  if (params.length > 0) {
    openExecuteParamDialog(row, params)
    return
  }

  await executeTask(row.id, {
    client_id: row.client_id || '',
    client_ip: row.client_ip || '',
    params: row.params || {},
  })
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '任务已下发' })
}

async function loadExecuteTriggerParameters(row) {
  const workflowId = row.workflow_id || ''
  if (!workflowId) return []

  try {
    const detail = await getAutomaWorkflowDetail(workflowId)
    return getUniqueTriggerParameters(parseWorkflowDetailPayload(detail))
  } catch {
    return []
  }
}

function openExecuteParamDialog(row, params) {
  // Execute params 合并任务已保存参数与工作流默认值，执行前允许用户调整
  executeParamTask.value = row
  executeParamItems.value = params.map((param, index) => createExecuteParamItem(param, index, row.params || {}))
  executeParamDialogVisible.value = true
}

async function confirmExecuteTaskWithParams() {
  const row = executeParamTask.value
  if (!row?.id) return

  const params = buildExecuteParamObject()
  if (!params) return

  executeParamSubmitting.value = true
  try {
    await executeTask(row.id, {
      client_id: row.client_id || '',
      client_ip: row.client_ip || '',
      params,
    })
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '任务已下发' })
    executeParamDialogVisible.value = false
  } finally {
    executeParamSubmitting.value = false
  }
}

function resetExecuteParamDialog() {
  if (executeParamSubmitting.value) return
  executeParamTask.value = null
  executeParamItems.value = []
}

function handleWorkflowChange(workflowId) {
  const workflow = workflowOptions.value.find((item) => getWorkflowId(item) === workflowId)
  taskForm.workflow_name = workflow?.name || ''
}

function handleClientChange(clientId) {
  const client = findClientById(clientId)
  taskForm.client_name = client ? getClientName(client) : ''
  taskForm.client_ip = client ? getClientIp(client) : ''
}

function handleClientClear() {
  clientSelector.keyword = ''
  handleClientChange('')
}

function handleClientKeywordFilter(keyword) {
  clientSelector.keyword = keyword
}

function handleClientSelectVisibleChange(visible) {
  if (!visible) clientSelector.keyword = ''
}

function addParam() {
  paramEntries.value.push(createEmptyParamEntry())
}

function removeParam(index) {
  if (paramEntries.value.length === 1) return
  paramEntries.value.splice(index, 1)
}

function buildTaskPayload() {
  if (!taskForm.name.trim()) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请输入任务名称' })
    return null
  }
  if (!taskForm.workflow_id.trim()) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请选择需要执行的工作流' })
    return null
  }
  if (!taskForm.client_id.trim()) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请选择需要执行的客户端' })
    return null
  }

  const params = buildParamObject()
  if (params === null) return null

  return {
    name: taskForm.name.trim(),
    description: taskForm.description.trim(),
    workflow_id: taskForm.workflow_id.trim(),
    workflow_name: taskForm.workflow_name.trim(),
    client_id: taskForm.client_id.trim(),
    client_name: taskForm.client_name.trim(),
    client_ip: taskForm.client_ip.trim(),
    cron_expression: taskForm.cron_expression.trim(),
    run_once_after_create: !taskForm.cron_expression.trim(),
    params,
    enabled: taskForm.enabled,
  }
}

function buildParamObject() {
  const params = {}

  for (const item of paramEntries.value) {
    const key = item.key.trim()
    const value = item.value.trim()

    if (!key && !value) continue
    if (!key) {
      appMessage({ type: APP_MESSAGE_TYPE.warning, message: '参数 key 不能为空' })
      return null
    }
    if (Object.prototype.hasOwnProperty.call(params, key)) {
      appMessage({ type: APP_MESSAGE_TYPE.warning, message: `参数 key 重复：${key}` })
      return null
    }

    params[key] = value
  }

  return params
}

function resetTaskForm() {
  Object.assign(taskForm, createEmptyTaskForm())
  clientSelector.keyword = ''
  paramEntries.value = [createEmptyParamEntry()]
}

function resetTaskFilters() {
  taskFilters.keyword = ''
  taskFilters.workflow_name = ''
  taskFilters.created_time_range = []
  taskFilters.enabled = ''
}

function upsertTask(task) {
  const index = tasks.value.findIndex((item) => item.id === task.id)
  if (index >= 0) {
    tasks.value.splice(index, 1, task)
    return
  }
  tasks.value = [task, ...tasks.value]
}

function mergeSavedTask(serverTask, payload, taskId) {
  return {
    ...serverTask,
    ...payload,
    id: serverTask?.id || taskId || `local_${Date.now()}`,
    updated_at: serverTask?.updated_at || new Date().toISOString(),
  }
}

function getCreatedTimeRange() {
  const range = Array.isArray(taskFilters.created_time_range)
    ? taskFilters.created_time_range
    : []
  return [range[0] || '', range[1] || '']
}

function normalizeList(data, fallbackKey) {
  const list = data?.list || data?.[fallbackKey] || []
  return Array.isArray(list) ? list : []
}

function normalizeParamEntries(params) {
  const entries = Object.entries(params || {}).map(([key, value]) => ({
    key,
    value: value == null ? '' : String(value),
  }))
  return entries.length > 0 ? entries : [createEmptyParamEntry()]
}

function parseWorkflowDetailPayload(detail) {
  const raw = detail?.normalized_json || detail?.normalizedJson || detail?.raw_json || detail?.rawJson || ''
  if (!raw) return detail || {}

  try {
    return JSON.parse(raw)
  } catch {
    return detail || {}
  }
}

function getUniqueTriggerParameters(workflow) {
  const seen = new Set()
  return getTriggerParameters(workflow).filter((param, index) => {
    const name = param?.name || `param_${index + 1}`
    if (seen.has(name)) return false
    seen.add(name)
    return true
  })
}

function getTriggerParameters(workflow) {
  const triggerData = workflow?.trigger || getTriggerNode(workflow)?.data || {}
  return Array.isArray(triggerData.parameters) ? triggerData.parameters : []
}

function getTriggerNode(workflow) {
  if (workflow?.drawflow?.nodes) {
    return workflow.drawflow.nodes.find((node) => node?.label === 'trigger') || null
  }

  const legacyNodes = workflow?.drawflow?.drawflow?.Home?.data
  if (legacyNodes) {
    return Object.values(legacyNodes).find((node) => node?.name === 'trigger') || null
  }

  return null
}

function createExecuteParamItem(param, index, savedParams) {
  const defaultValue = getTriggerParamDefaultRawValue(param)
  const name = param?.name || `param_${index + 1}`
  const value = Object.prototype.hasOwnProperty.call(savedParams, name) ? savedParams[name] : defaultValue
  return {
    key: `${name}_${param?.type || ''}_${index}`,
    name,
    type: param?.type || 'string',
    description: param?.description || '',
    placeholder: param?.placeholder || '',
    required: Boolean(param?.required || param?.data?.required),
    defaultText: formatParamDefaultValue(defaultValue),
    value: normalizeExecuteParamValue(param, value),
  }
}

function buildExecuteParamObject() {
  const params = {}

  for (const item of executeParamItems.value) {
    if (isMissingRequiredParam(item)) {
      appMessage({ type: APP_MESSAGE_TYPE.warning, message: `请填写必填参数：${item.name}` })
      return null
    }

    const value = parseExecuteParamValue(item)
    if (value === undefined && item.type === 'json' && !isEmptyParamValue(item.value)) return null
    params[item.name] = value
  }

  return params
}

function parseExecuteParamValue(item) {
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

function getTriggerParamDefaultRawValue(param) {
  if (param?.defaultValue !== undefined) return param.defaultValue
  if (param?.default !== undefined) return param.default
  if (param?.value !== undefined) return param.value
  return ''
}

function normalizeExecuteParamValue(param, value) {
  if (param?.type === 'checkbox') return Boolean(value)
  if (param?.type === 'json' && value && typeof value === 'object') return JSON.stringify(value, null, 2)
  return formatParamDefaultValue(value)
}

function formatParamDefaultValue(value) {
  if (value === undefined || value === null || value === '') return ''
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function isEmptyParamValue(value) {
  return value === undefined || value === null || value === ''
}

function isMissingRequiredParam(item) {
  if (!item.required) return false
  if (item.type === 'checkbox') return item.value !== true
  return isEmptyParamValue(item.value)
}

function findClientById(clientId) {
  return clientOptions.value.find((item) => getClientId(item) === clientId)
}

function getWorkflowId(row) {
  return row?.automa_id || row?.workflow_id || row?.workflowId || row?.id || ''
}

function getClientId(row) {
  return row?.client_id || row?.clientId || row?.id || ''
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function getClientName(row) {
  return row?.client_name || row?.name || row?.hostname || getClientId(row) || '-'
}

function getClientStatus(row) {
  if (row?.banned || row?.is_banned) return 'banned'
  if (row?.online || row?.status === 'online') return 'online'
  return 'offline'
}

function getClientStatusText(row) {
  const status = getClientStatus(row)
  if (status === 'online') return '在线'
  if (status === 'banned') return '已拉黑'
  return '离线'
}

function getClientSelectMeta(row) {
  return [getClientId(row) || '-', getClientStatusText(row)].join(' / ')
}

function getClientOptionLabel(row) {
  return `${getClientIp(row) || '-'} / ${getClientSelectMeta(row)}`
}

function getTaskClientIp(row) {
  const client = findClientById(row?.client_id)
  return row?.client_ip || (client ? getClientIp(client) : '') || '-'
}

function getTaskParamCount(row) {
  return Object.keys(row?.params || {}).length
}

function getScheduleText(row) {
  const cronExpression = row?.cron_expression || row?.cron || ''
  return cronExpression || '创建后执行一次'
}

function sortByTimeDesc(data) {
  return data.slice().sort((a, b) => getTimeValue(b) - getTimeValue(a))
}

function getTimeValue(row) {
  const value = row?.updated_at || row?.created_at || row?.updatedAt || row?.createdAt
  return value ? new Date(value).getTime() || 0 : 0
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

function createEmptyParamEntry() {
  return {
    key: '',
    value: '',
  }
}

function createEmptyTaskForm() {
  return {
    id: '',
    name: '',
    description: '',
    workflow_id: '',
    workflow_name: '',
    client_id: '',
    client_name: '',
    client_ip: '',
    cron_expression: '',
    enabled: true,
  }
}
</script>

<style scoped lang="scss">
.task-page {
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

.task-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 16px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.task-filters {
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

.filter-item--keyword {
  width: 300px;
}

.filter-item--workflow {
  width: 280px;
}

.filter-item--created-time {
  width: 460px;
}

.filter-item--created-time :deep(.el-date-editor) {
  width: 100%;
}

.filter-item--enabled {
  width: 190px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.task-table {
  flex: 1;
  min-height: 0;
}

.workflow-option,
.client-option {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.workflow-option span,
.workflow-option small,
.client-option span,
.client-option small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-option small,
.client-option small {
  color: #909399;
}

.full-width {
  width: 100%;
}

.client-config,
.params-editor,
.schedule-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.client-select-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
}

.client-select {
  min-width: 0;
}

.client-selection-summary {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 32px;
}

.client-selection-text,
.client-selection-empty,
.form-help {
  color: #909399;
  font-size: 13px;
}

.param-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  gap: 12px;
}

.execute-param-form {
  max-height: 58vh;
  overflow: auto;
  padding-right: 8px;
}

.execute-param-control {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

@media (max-width: 768px) {
  .client-select-row,
  .param-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .page-actions,
  .task-filters,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item--keyword,
  .filter-item--workflow,
  .filter-item--created-time,
  .filter-item--enabled {
    width: 100%;
  }
}
</style>
