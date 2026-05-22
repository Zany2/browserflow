<template>
  <section class="task-page server-list-page">
    <header class="page-actions server-list-actions">
      <el-button :icon="RefreshRight" @click="loadTasks">刷新</el-button>
      <el-button type="primary" :icon="Plus" @click="handleCreateTask">新增任务</el-button>
    </header>

    <section class="task-panel server-list-panel">
      <div class="task-filters server-list-filters">
        <div class="task-filter-fields">
          <div class="filter-item filter-item--keyword">
            <span class="filter-label">关键词</span>
            <el-input v-model="taskFilters.keyword" clearable placeholder="任务名称或任务说明" />
          </div>
          <div class="filter-item filter-item--workflow">
            <span class="filter-label">自定义工作流名称</span>
            <el-input v-model="taskFilters.workflow_name" clearable placeholder="自定义工作流名称" />
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
        </div>
        <div class="task-filter-actions">
          <el-button @click="resetTaskFilters">重置</el-button>
          <el-button type="danger" :disabled="selectedTaskIds.length === 0" @click="handleBatchDeleteTasks">
            删除选中
          </el-button>
          <AppSelectionSummary :count="selectedTaskIds.length" unit="任务" />
        </div>
      </div>

      <el-table
        ref="taskTableRef"
        v-loading="loadingTasks"
        class="task-table server-list-table adaptive-table"
        :data="pagedTasks"
        border
        height="100%"
        :row-key="getTaskSelectionKey"
        empty-text="暂无任务配置"
        @selection-change="handleTaskSelectionChange"
      >
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column prop="name" label="任务名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="description" label="任务说明" min-width="180" show-overflow-tooltip />
        <el-table-column label="自定义工作流名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.workflow_name || row.workflow_id || '' }}
          </template>
        </el-table-column>
        <el-table-column label="执行客户端 IP" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getTaskClientIp(row) }}
          </template>
        </el-table-column>
        <el-table-column label="执行计划" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getScheduleText(row) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="86" align="center" class-name="quick-edit-column">
          <template #default="{ row }">
            <div class="quick-edit-cell">
              <el-switch
                class="quick-edit-switch"
                :model-value="row.enabled !== false"
                :loading="isTaskStatusUpdating(row)"
                @change="(value) => handleQuickUpdateTaskStatus(row, value)"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(row.created_at || row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="最近执行时间" width="170" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(row.last_executed_at || row.lastExecutedAt) }}
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
        :total="taskTotal"
      />
    </section>

    <AppDialog
      v-model="taskDialogVisible"
      :title="taskDialogTitle"
      width="920px"
      @closed="handleDialogClosed"
    >
      <el-form class="task-config-form" label-width="136px" :model="taskForm">
        <el-form-item label="任务名称">
          <el-input v-model="taskForm.name" placeholder="请填写任务名称，不填写则随机生成" />
        </el-form-item>

        <el-form-item label="任务说明">
          <el-input
            v-model="taskForm.description"
            type="textarea"
            :rows="3"
            placeholder="请填写任务说明，不填写则随机生成"
          />
        </el-form-item>

        <el-form-item label="工作流" required>
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
                <span>{{ workflow.name || '' }}</span>
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
                placeholder="可选，搜索并选择客户端名称、ID、IP、浏览器"
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
                    <span>{{ getClientIp(client) || '' }}</span>
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
                <el-tag effect="plain">{{ getClientIp(selectedClient) || '' }}</el-tag>
                <span class="client-selection-text">{{ getClientSelectMeta(selectedClient) }}</span>
                <span v-if="clientWorkflowStatusText" :class="clientWorkflowStatusClass">
                  {{ clientWorkflowStatusText }}
                </span>
              </template>
              <span v-else class="client-selection-empty">
                未选择时，后端会匹配拥有该工作流的在线客户端执行
              </span>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="执行参数">
          <div class="params-editor">
            <div v-if="workflowParamLoading" class="param-loading">正在解析工作流参数...</div>
            <div v-else-if="autoParamItems.length > 0" class="auto-param-list">
              <div
                v-for="param in autoParamItems"
                :key="param.key"
                class="auto-param-row"
              >
                <div class="auto-param-meta">
                  <div class="auto-param-title">
                    <span>{{ param.name }}</span>
                    <el-tag v-if="param.required" size="small" type="danger" effect="plain">必填</el-tag>
                    <el-tag size="small" effect="plain">{{ formatParamTypeText(param.type) }}</el-tag>
                  </div>
                </div>
                <div class="auto-param-control">
                  <el-switch
                    v-if="param.type === 'checkbox'"
                    v-model="param.value"
                    active-text="是"
                    inactive-text="否"
                  />
                  <el-input
                    v-else
                    v-model="param.value"
                    :type="param.type === 'json' ? 'textarea' : param.type === 'number' ? 'number' : 'text'"
                    :rows="param.type === 'json' ? 4 : undefined"
                    :placeholder="param.placeholder"
                  />
                  <span v-if="getParamDescriptionText(param)" class="form-help">
                    {{ getParamDescriptionText(param) }}
                  </span>
                  <span v-if="param.defaultText" class="form-help">默认值：{{ param.defaultText }}</span>
                </div>
              </div>
            </div>
            <span v-else class="form-help">当前工作流未配置 Param 参数。</span>

            <div
              v-for="(item, index) in paramEntries"
              :key="`param_${index}`"
              class="param-row"
            >
              <el-input v-model="item.key" placeholder="key" />
              <el-select v-model="item.type" class="param-type-select" @change="handleManualParamTypeChange(item)">
                <el-option label="文本" value="string" />
                <el-option label="数字" value="number" />
                <el-option label="JSON" value="json" />
                <el-option label="勾选" value="checkbox" />
              </el-select>
              <el-switch
                v-if="item.type === 'checkbox'"
                v-model="item.value"
                active-text="是"
                inactive-text="否"
              />
              <el-input
                v-else
                v-model="item.value"
                :type="item.type === 'json' ? 'textarea' : item.type === 'number' ? 'number' : 'text'"
                :rows="item.type === 'json' ? 3 : undefined"
                placeholder="value"
              />
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
              工作流 Param 会自动渲染，手动新增的参数也会作为变量传入。
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
              不填写时不会自动调度，可在任务列表手动执行一次；填写后按定时任务处理。
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Plus, RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import {
  getAutomaWorkflowDetail,
  listAutomaSyncCandidatesByWorkflow,
  listAutomaWorkflows,
} from '@/services/automa'
import { listClients } from '@/services/client'
import { createTask, deleteTask, executeTask, listTasks, updateTask } from '@/services/task'
import { formatDate as formatBaseDate } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, normalizeList } from '@/utils/list'

const tasks = ref([])
const workflowOptions = ref([])
const clientOptions = ref([])
const loadingTasks = ref(false)
const taskTableRef = ref(null)
const workflowLoading = ref(false)
const workflowParamLoading = ref(false)
const clientLoading = ref(false)
const clientWorkflowChecking = ref(false)
const clientWorkflowCheckSeq = ref(0)
const saving = ref(false)
const taskDialogVisible = ref(false)
const executeParamDialogVisible = ref(false)
const executeParamSubmitting = ref(false)
const executeParamTask = ref(null)
const executeParamItems = ref([])
const taskStatusUpdatingIds = ref(new Set())
const taskPage = ref(1)
const taskPageSize = ref(10)
const taskTotal = ref(0)
const pageSizes = DEFAULT_PAGE_SIZES
const clientWorkflowCheckPageSize = DEFAULT_PAGE_SIZES[0]
const CRON_MONTH_NAMES = {
  jan: 1,
  feb: 2,
  mar: 3,
  apr: 4,
  may: 5,
  jun: 6,
  jul: 7,
  aug: 8,
  sep: 9,
  oct: 10,
  nov: 11,
  dec: 12,
}
const CRON_WEEK_NAMES = {
  sun: 0,
  mon: 1,
  tue: 2,
  wed: 3,
  thu: 4,
  fri: 5,
  sat: 6,
}
const CRON_PREDEFINED_PATTERNS = new Set([
  '@yearly',
  '@annually',
  '@monthly',
  '@weekly',
  '@daily',
  '@midnight',
  '@hourly',
])
const autoParamItems = ref([])
const paramEntries = ref([createEmptyParamEntry()])
const {
  run: scheduleFilterSearch,
  cancel: clearFilterSearchTimer,
} = useDebouncedAction(searchTasksNow, 200)

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
const pagedTasks = computed(() => tasks.value)
const {
  selectedKeys: selectedTaskIds,
  handleSelectionChange: handleTaskSelectionChange,
  restoreSelection: restoreTaskSelection,
  resetSelection: resetTaskSelection,
  retainSelectionByRows: retainTaskSelectionByRows,
  removeSelectionKeys: removeTaskSelectionKeys,
} = usePagedTableSelection({
  rows: pagedTasks,
  getRowKey: getTaskSelectionKey,
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
const clientWorkflowStatusText = computed(() => {
  if (!taskForm.workflow_id || !selectedClient.value) return ''
  if (clientWorkflowChecking.value) return '正在检测当前客户端是否拥有该工作流...'
  if (selectedClientHasWorkflow.value === true) return '当前客户端已拥有该工作流'
  if (selectedClientHasWorkflow.value === false) {
    return '当前客户端没有上报该工作流，保存后执行会直接失败'
  }
  return ''
})
const clientWorkflowStatusClass = computed(() => {
  if (selectedClientHasWorkflow.value === true) return 'form-success'
  if (selectedClientHasWorkflow.value === false) return 'form-warning'
  return 'form-help'
})
const selectedClientHasWorkflow = ref(null)

onMounted(() => {
  loadTasks()
  loadClients()
})

watch(() => [taskFilters.keyword, taskFilters.workflow_name], () => {
  scheduleFilterSearch()
})

watch(() => [taskFilters.created_time_range, taskFilters.enabled], () => {
  clearFilterSearchTimer()
  reloadFirstTaskPage()
})

watch(taskPage, () => {
  loadTasks()
})

watch(taskPageSize, () => {
  reloadFirstTaskPage()
})

watch(pagedTasks, () => {
  restoreTaskSelection(taskTableRef)
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
      page_num: taskPage.value,
      page_size: taskPageSize.value,
    })
    const list = normalizeList(data, 'tasks')
    tasks.value = sortByTimeDesc(list)
    taskTotal.value = Number(data?.total ?? list.length)
    retainTaskSelectionByRows(tasks.value)
  } finally {
    loadingTasks.value = false
  }
}

function searchTasksNow() {
  reloadFirstTaskPage()
}

function reloadFirstTaskPage() {
  if (taskPage.value === 1) {
    loadTasks()
    return
  }
  taskPage.value = 1
}

function getTaskSelectionKey(row) {
  // Selection key 使用任务 ID 保持跨分页多选状态
  return String(row?.id || '').trim()
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
  await loadWorkflowParamItems(taskForm.workflow_id, row.params || {})
  paramEntries.value = normalizeParamEntriesWithoutAuto(row.params)
  checkSelectedClientWorkflow()
  taskDialogVisible.value = true
}

async function loadWorkflowParamItems(workflowId, savedParams = {}) {
  autoParamItems.value = []
  workflowId = String(workflowId || '').trim()
  if (!workflowId) return

  workflowParamLoading.value = true
  try {
    const detail = await getAutomaWorkflowDetail(workflowId)
    const params = getUniqueTriggerParameters(parseWorkflowDetailPayload(detail))
    autoParamItems.value = params.map((param, index) => createExecuteParamItem(param, index, savedParams))
  } catch {
    autoParamItems.value = []
  } finally {
    workflowParamLoading.value = false
  }
}

async function checkSelectedClientWorkflow() {
  const seq = clientWorkflowCheckSeq.value + 1
  clientWorkflowCheckSeq.value = seq
  selectedClientHasWorkflow.value = null
  clientWorkflowChecking.value = false
  const workflowId = taskForm.workflow_id.trim()
  const clientIp = taskForm.client_ip.trim()
  if (!workflowId || !clientIp) return

  clientWorkflowChecking.value = true
  try {
    const data = await listAutomaSyncCandidatesByWorkflow(workflowId, {
      page_num: 1,
      page_size: clientWorkflowCheckPageSize,
      source_ip: clientIp,
    })
    const candidates = normalizeList(data)
    if (seq !== clientWorkflowCheckSeq.value) return
    selectedClientHasWorkflow.value = candidates.some((item) => getClientIp(item) === clientIp)
  } catch {
    if (seq !== clientWorkflowCheckSeq.value) return
    selectedClientHasWorkflow.value = null
  } finally {
    if (seq === clientWorkflowCheckSeq.value) clientWorkflowChecking.value = false
  }
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
  removeTaskSelectionKeys([row.id])
  tasks.value = tasks.value.filter((item) => item.id !== row.id)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '任务已删除' })
}

async function handleBatchDeleteTasks() {
  const ids = selectedTaskIds.value.slice()
  if (ids.length === 0) return

  const confirmed = await appConfirm({
    title: '批量删除任务',
    message: `确认删除选中的 ${ids.length} 个任务吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await Promise.all(ids.map((id) => deleteTask(id)))
  const deletedIds = new Set(ids)
  tasks.value = tasks.value.filter((item) => !deletedIds.has(getTaskSelectionKey(item)))
  resetTaskSelection(taskTableRef)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已删除选中任务' })
}

async function handleQuickUpdateTaskStatus(row, enabled) {
  const taskId = getTaskSelectionKey(row)
  if (!taskId || taskStatusUpdatingIds.value.has(taskId)) return

  const previousEnabled = row.enabled !== false
  row.enabled = enabled
  setTaskStatusUpdating(taskId, true)
  try {
    const payload = buildTaskPayloadFromRow(row, { enabled })
    const data = await updateTask(taskId, payload)
    upsertTask(mergeSavedTask(data?.task, payload, taskId))
    appMessage({ type: APP_MESSAGE_TYPE.success, message: enabled ? '任务已启用' : '任务已停用' })
  } catch (error) {
    row.enabled = previousEnabled
    appMessage({
      type: APP_MESSAGE_TYPE.error,
      message: error?.message || '任务状态修改失败',
    })
  } finally {
    setTaskStatusUpdating(taskId, false)
  }
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
      params: {
        ...(row.params || {}),
        ...params,
      },
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

async function handleWorkflowChange(workflowId) {
  const workflow = workflowOptions.value.find((item) => getWorkflowId(item) === workflowId)
  taskForm.workflow_name = workflow?.name || ''
  const savedParams = buildParamObject({ silent: true }) || {}
  const manualParams = buildManualParamObject({ silent: true }) || {}
  await loadWorkflowParamItems(workflowId, savedParams)
  paramEntries.value = normalizeParamEntriesWithoutAuto(manualParams)
  checkSelectedClientWorkflow()
}

function handleClientChange(clientId) {
  const client = findClientById(clientId)
  taskForm.client_name = client ? getClientName(client) : ''
  taskForm.client_ip = client ? getClientIp(client) : ''
  checkSelectedClientWorkflow()
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

function handleManualParamTypeChange(item) {
  if (item.type === 'checkbox') {
    item.value = item.value === true || item.value === 'true'
    return
  }
  item.value = formatParamDefaultValue(item.value)
}

function buildTaskPayload() {
  if (!taskForm.workflow_id.trim()) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请选择需要执行的工作流' })
    return null
  }

  const params = buildParamObject()
  if (params === null) return null

  const cronExpression = taskForm.cron_expression.trim()
  if (!isValidCronExpression(cronExpression)) {
    appMessage({
      type: APP_MESSAGE_TYPE.warning,
      message: 'Cron 表达式格式不正确，请填写 5 段或 6 段表达式，例如 */10 * * * * 或 0 */10 * * * *',
    })
    return null
  }

  return {
    name: taskForm.name.trim(),
    description: taskForm.description.trim(),
    workflow_id: taskForm.workflow_id.trim(),
    workflow_name: taskForm.workflow_name.trim(),
    client_id: taskForm.client_id.trim(),
    client_name: taskForm.client_name.trim(),
    client_ip: taskForm.client_ip.trim(),
    cron_expression: cronExpression,
    run_once_after_create: false,
    params,
    enabled: taskForm.enabled,
  }
}

function isValidCronExpression(expression) {
  const cronExpression = String(expression || '').trim()
  if (!cronExpression) return true
  if (CRON_PREDEFINED_PATTERNS.has(cronExpression.toLowerCase())) return true
  if (/^@every\s+\d+(ns|us|µs|ms|s|m|h)$/i.test(cronExpression)) return true

  const parts = cronExpression.split(/\s+/)
  if (parts.length !== 5 && parts.length !== 6) return false

  const normalizedParts = parts.length === 5 ? ['0', ...parts] : parts
  const ranges = [
    { min: 0, max: 59, names: null, allowQuestion: false },
    { min: 0, max: 59, names: null, allowQuestion: false },
    { min: 0, max: 23, names: null, allowQuestion: false },
    { min: 1, max: 31, names: null, allowQuestion: true },
    { min: 1, max: 12, names: CRON_MONTH_NAMES, allowQuestion: false },
    { min: 0, max: 6, names: CRON_WEEK_NAMES, allowQuestion: true },
  ]

  return normalizedParts.every((part, index) => isValidCronPart(part, ranges[index]))
}

function isValidCronPart(part, range) {
  if (!part) return false
  if (part === '*') return true
  if (range.allowQuestion && part === '?') return true

  return part.split(',').every((item) => isValidCronListItem(item, range))
}

function isValidCronListItem(item, range) {
  if (!item) return false
  const [base, step] = item.split('/')
  if (item.split('/').length > 2) return false
  if (step !== undefined && (!/^\d+$/.test(step) || Number(step) <= 0)) return false
  if (base === '*') return true
  if (range.allowQuestion && base === '?') return true
  if (base.includes('-')) {
    const [start, end] = base.split('-')
    if (!start || !end || base.split('-').length !== 2) return false
    const startValue = parseCronPartValue(start, range)
    const endValue = parseCronPartValue(end, range)
    return startValue !== null && endValue !== null && startValue <= endValue
  }

  return parseCronPartValue(base, range) !== null
}

function parseCronPartValue(value, range) {
  const text = String(value || '').trim()
  const namedValue = range.names?.[text.toLowerCase()]
  if (namedValue !== undefined) return namedValue
  if (!/^\d+$/.test(text)) return null

  const numberValue = Number(text)
  return numberValue >= range.min && numberValue <= range.max ? numberValue : null
}

function buildTaskPayloadFromRow(row, overrides = {}) {
  return {
    name: String(row?.name || '').trim(),
    description: String(row?.description || '').trim(),
    workflow_id: String(row?.workflow_id || row?.automa_id || '').trim(),
    workflow_name: String(row?.workflow_name || '').trim(),
    client_id: String(row?.client_id || '').trim(),
    client_name: String(row?.client_name || '').trim(),
    client_ip: String(getTaskClientIp(row)).trim(),
    cron_expression: String(row?.cron_expression || row?.cron || '').trim(),
    run_once_after_create: false,
    params: normalizeTaskParams(row?.params),
    enabled: row?.enabled !== false,
    ...overrides,
  }
}

function normalizeTaskParams(params) {
  if (!params) return {}
  if (typeof params === 'object') return params

  try {
    const parsed = JSON.parse(params)
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

function buildParamObject(options = {}) {
  const params = {}

  for (const item of autoParamItems.value) {
    if (isMissingRequiredParam(item)) {
      if (!options.silent) {
        appMessage({ type: APP_MESSAGE_TYPE.warning, message: `请填写必填参数：${item.name}` })
      }
      return null
    }

    const value = parseExecuteParamValue(item, options)
    if (value === undefined && item.type === 'json' && !isEmptyParamValue(item.value)) return null
    params[item.name] = value
  }

  const manualParams = buildManualParamObject(options)
  if (manualParams === null) return null
  return { ...params, ...manualParams }
}

function buildManualParamObject(options = {}) {
  const params = {}

  for (const item of paramEntries.value) {
    const key = item.key.trim()
    const value = item.type === 'checkbox' ? item.value : String(item.value || '').trim()

    if (!key && isEmptyParamValue(value)) continue
    if (!key) {
      if (!options.silent) appMessage({ type: APP_MESSAGE_TYPE.warning, message: '参数 key 不能为空' })
      return null
    }
    if (Object.prototype.hasOwnProperty.call(params, key)) {
      if (!options.silent) appMessage({ type: APP_MESSAGE_TYPE.warning, message: `参数 key 重复：${key}` })
      return null
    }
    if (autoParamItems.value.some((param) => param.name === key)) {
      if (!options.silent) appMessage({ type: APP_MESSAGE_TYPE.warning, message: `参数 key 已由工作流 Param 使用：${key}` })
      return null
    }

    const parsedValue = parseExecuteParamValue({
      name: key,
      type: item.type || 'string',
      value,
    }, options)
    if (parsedValue === undefined && item.type === 'json' && !isEmptyParamValue(value)) return null
    params[key] = parsedValue
  }

  return params
}

function resetTaskForm() {
  Object.assign(taskForm, createEmptyTaskForm())
  clientSelector.keyword = ''
  autoParamItems.value = []
  paramEntries.value = [createEmptyParamEntry()]
  selectedClientHasWorkflow.value = null
  clientWorkflowChecking.value = false
  clientWorkflowCheckSeq.value += 1
}

function resetTaskFilters() {
  taskFilters.keyword = ''
  taskFilters.workflow_name = ''
  taskFilters.created_time_range = []
  taskFilters.enabled = ''
}

function isTaskStatusUpdating(row) {
  return taskStatusUpdatingIds.value.has(getTaskSelectionKey(row))
}

function setTaskStatusUpdating(taskId, updating) {
  const nextIds = new Set(taskStatusUpdatingIds.value)
  if (updating) {
    nextIds.add(taskId)
  } else {
    nextIds.delete(taskId)
  }
  taskStatusUpdatingIds.value = nextIds
}

function upsertTask(task) {
  const index = tasks.value.findIndex((item) => item.id === task.id)
  if (index >= 0) {
    tasks.value.splice(index, 1, task)
    return
  }
  tasks.value = [task, ...tasks.value]
  taskTotal.value += 1
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

function normalizeParamEntries(params) {
  const entries = Object.entries(params || {}).map(([key, value]) => ({
    key,
    type: inferManualParamType(value),
    value: normalizeManualParamValue(value),
  }))
  return entries.length > 0 ? entries : [createEmptyParamEntry()]
}

function normalizeParamEntriesWithoutAuto(params) {
  const autoNames = new Set(autoParamItems.value.map((item) => item.name))
  const manualParams = Object.fromEntries(
    Object.entries(params || {}).filter(([key]) => !autoNames.has(key))
  )
  return normalizeParamEntries(manualParams)
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

function parseExecuteParamValue(item, options = {}) {
  if (item.type === 'number') {
    const value = Number(item.value)
    return Number.isNaN(value) ? 0 : value
  }
  if (item.type === 'json') {
    if (isEmptyParamValue(item.value)) return null
    try {
      return JSON.parse(item.value)
    } catch {
      if (!options.silent) {
        appMessage({ type: APP_MESSAGE_TYPE.warning, message: `参数 ${item.name} 不是有效 JSON` })
      }
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

function inferManualParamType(value) {
  if (typeof value === 'boolean') return 'checkbox'
  if (typeof value === 'number') return 'number'
  if (value && typeof value === 'object') return 'json'
  return 'string'
}

function normalizeManualParamValue(value) {
  if (typeof value === 'boolean') return value
  if (value && typeof value === 'object') return JSON.stringify(value, null, 2)
  return formatParamDefaultValue(value)
}

function isEmptyParamValue(value) {
  return value === undefined || value === null || value === ''
}

function isMissingRequiredParam(item) {
  if (!item.required) return false
  if (item.type === 'checkbox') return item.value !== true
  return isEmptyParamValue(item.value)
}

function formatParamTypeText(type) {
  const texts = {
    string: '文本',
    number: '数字',
    json: 'JSON',
    checkbox: '勾选',
  }
  return texts[type] || type || '文本'
}

function getParamDescriptionText(param) {
  const description = String(param?.description || '').trim()
  const placeholder = String(param?.placeholder || '').trim()
  return description && description !== placeholder ? description : ''
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
  return row?.client_name || row?.name || row?.hostname || getClientId(row) || ''
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
  return [getClientId(row), getClientStatusText(row)].filter(Boolean).join(' / ')
}

function getClientOptionLabel(row) {
  return [getClientIp(row), getClientSelectMeta(row)].filter(Boolean).join(' / ')
}

function getTaskClientIp(row) {
  const client = findClientById(row?.client_id)
  return row?.client_ip || (client ? getClientIp(client) : '') || '自动匹配'
}

function getScheduleText(row) {
  const cronExpression = row?.cron_expression || row?.cron || ''
  return cronExpression || '手动执行'
}

function sortByTimeDesc(data) {
  return data.slice().sort((a, b) => getTimeValue(b) - getTimeValue(a))
}

function getTimeValue(row) {
  const value = row?.updated_at || row?.created_at || row?.updatedAt || row?.createdAt
  return value ? new Date(value).getTime() || 0 : 0
}

function formatDate(value) {
  return formatBaseDate(value, { fallback: '' })
}

function createEmptyParamEntry() {
  return {
    key: '',
    type: 'string',
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
.task-filters {
  align-items: flex-start;
  justify-content: space-between;
}

.task-filter-fields {
  display: grid;
  flex: 1;
  grid-template-columns: minmax(240px, 1fr) minmax(300px, 1.1fr) minmax(380px, 1.4fr) minmax(140px, 0.6fr);
  gap: 12px 16px;
  min-width: 0;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item--created-time :deep(.el-date-editor) {
  width: 100%;
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

.task-filter-actions {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 12px;
  margin-left: auto;
}

.task-config-form :deep(.el-form-item__label) {
  flex: 0 0 136px;
  padding-right: 12px;
  white-space: nowrap;
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

.client-selection-summary .form-success,
.client-selection-summary .form-warning,
.client-selection-summary .form-help {
  margin-left: 4px;
}

.client-selection-text,
.client-selection-empty,
.form-help,
.form-success {
  color: #909399;
  font-size: 13px;
}

.form-success {
  color: #67c23a;
}

.form-warning {
  color: #f56c6c;
  font-size: 13px;
}

.param-loading {
  color: #909399;
  font-size: 13px;
}

.auto-param-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.auto-param-row {
  display: grid;
  grid-template-columns: minmax(180px, 240px) minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.auto-param-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.auto-param-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 32px;
}

.auto-param-title span:first-child {
  overflow-wrap: anywhere;
  font-weight: 500;
  color: #303133;
}

.auto-param-control {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.param-row {
  display: grid;
  grid-template-columns: minmax(120px, 180px) 112px minmax(0, 1fr) auto;
  gap: 12px;
  align-items: start;
}

.param-type-select {
  width: 112px;
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
  .auto-param-row,
  .param-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1280px) {
  .task-filter-fields {
    flex-basis: 100%;
    grid-template-columns: repeat(2, minmax(260px, 1fr));
  }

  .task-filter-actions {
    justify-content: flex-end;
    width: 100%;
  }
}

@media (max-width: 640px) {
  .page-actions,
  .task-filters,
  .task-filter-actions,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .task-filter-actions {
    margin-left: 0;
  }

  .task-filter-fields {
    grid-template-columns: 1fr;
    width: 100%;
  }
}
</style>
