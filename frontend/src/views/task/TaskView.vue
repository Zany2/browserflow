<template>
  <section class="task-page server-list-page">
    <header class="page-actions server-list-actions">
      <el-button :icon="RefreshRight" @click="loadTasks">刷新</el-button>
      <el-button type="primary" :icon="Plus" @click="handleCreateTask">新增任务</el-button>
      <el-button @click="resetTaskFilters">重置</el-button>
      <el-button type="danger" :disabled="selectedTaskIds.length === 0" @click="handleBatchDeleteTasks">
        删除选中
      </el-button>
    </header>

    <section class="task-panel server-list-panel">
      <div class="task-filters server-list-filters">
        <div class="task-filter-fields">
          <div class="filter-item filter-item--keyword">
            <span class="filter-label">关键词</span>
            <el-input v-model="taskFilters.keyword" clearable placeholder="任务名称、任务说明或自定义工作流名称" />
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
        <el-table-column prop="name" label="任务名称" min-width="100" show-overflow-tooltip />
        <el-table-column prop="description" label="任务说明" min-width="110" show-overflow-tooltip />
        <el-table-column label="自定义工作流名称" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.workflow_name || row.workflow_id || '' }}
          </template>
        </el-table-column>
        <el-table-column label="执行客户端 IP" min-width="105" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getTaskClientIp(row) }}
          </template>
        </el-table-column>
        <el-table-column label="执行节点 ID" min-width="105" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getTaskNodeId(row) }}
          </template>
        </el-table-column>
        <el-table-column label="执行计划" min-width="100" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getScheduleText(row) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="76" align="center" class-name="quick-edit-column">
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
        <el-table-column label="创建时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(row.created_at || row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="最近执行时间" width="168" class-name="nowrap-column">
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

      <div class="task-footer server-list-footer">
        <AppSelectionSummary :count="selectedTaskIds.length" unit="任务" />
        <AppPagination
          v-model:current-page="taskPage"
          v-model:page-size="taskPageSize"
          :page-sizes="pageSizes"
          :total="taskTotal"
        />
      </div>
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

        <el-form-item label="调度目标">
          <div class="client-config">
            <div class="client-select-row">
              <el-select
                v-model="taskForm.client_ip"
                class="client-select"
                clearable
                filterable
                placeholder="客户端 IP（可选，不选则自动匹配全部）"
                :loading="clientLoading"
                @change="handleClientIpChange"
                @clear="handleClientIpClear"
              >
                <el-option
                  v-for="option in clientIpOptions"
                  :key="option.ip"
                  :label="option.ip"
                  :value="option.ip"
                >
                  <div class="client-option client-ip-option">
                    <span>{{ option.ip }}</span>
                    <small>{{ option.total }} 个执行节点 / {{ option.online }} 个在线</small>
                  </div>
                </el-option>
              </el-select>
              <el-select
                v-model="taskForm.client_id"
                class="client-select"
                clearable
                filterable
                placeholder="执行节点（可选，不选则按客户端自动匹配）"
                :disabled="!taskForm.client_ip"
                :loading="clientLoading"
                @change="handleNodeChange"
                @clear="handleNodeClear"
              >
                <el-option
                  v-for="client in filteredNodeOptions"
                  :key="getClientId(client)"
                  :label="getClientOptionLabel(client)"
                  :value="getClientId(client)"
                >
                  <div class="client-option">
                    <span>{{ getClientNodeId(client) || '' }}</span>
                    <small>{{ getNodeSelectMeta(client) }}</small>
                  </div>
                </el-option>
              </el-select>
              <el-button :icon="RefreshRight" :loading="clientLoading" @click="loadClients">
                刷新客户端
              </el-button>
            </div>

            <div class="client-selection-summary">
              <el-tag effect="plain">IP：{{ taskForm.client_ip || '自动匹配' }}</el-tag>
              <el-tag effect="plain">节点：{{ taskForm.node_id || '自动匹配' }}</el-tag>
              <span class="client-selection-text">{{ selectedTargetText }}</span>
              <span v-if="clientWorkflowStatusText" :class="clientWorkflowStatusClass">
                {{ clientWorkflowStatusText }}
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
            <AppCronPicker v-model="taskForm.cron_expression" />
            <span class="form-help">
              不填写时不会自动调度，可在任务列表手动执行一次；填写后按定时任务处理。
            </span>
          </div>
        </el-form-item>

        <el-form-item label="繁忙策略">
          <div class="policy-editor">
            <el-radio-group v-model="taskForm.queue_policy">
              <el-radio-button label="queue">等待可用</el-radio-button>
              <el-radio-button label="fail">直接失败</el-radio-button>
              <el-radio-button label="skip">跳过本次</el-radio-button>
            </el-radio-group>
            <div class="policy-number-row">
              <el-form-item label="执行超时" label-width="80px">
                <el-input-number
                  v-model="taskForm.timeout_seconds"
                  :min="30"
                  :max="86400"
                  :step="30"
                  controls-position="right"
                />
                <span class="policy-unit">秒</span>
              </el-form-item>
              <el-form-item label="最大尝试" label-width="80px">
                <el-input-number
                  v-model="taskForm.max_attempts"
                  :min="1"
                  :max="20"
                  :step="1"
                  controls-position="right"
                />
                <span class="policy-unit">次</span>
              </el-form-item>
            </div>
            <div v-if="taskForm.queue_policy === 'queue'" class="policy-number-row">
              <el-form-item label="最大等待" label-width="80px">
                <el-input-number
                  v-model="taskForm.queue_wait_seconds"
                  :min="10"
                  :max="86400"
                  :step="30"
                  controls-position="right"
                />
                <span class="policy-unit">秒</span>
              </el-form-item>
              <el-form-item label="重试间隔" label-width="80px">
                <el-input-number
                  v-model="taskForm.queue_retry_interval_seconds"
                  :min="1"
                  :max="3600"
                  :step="1"
                  controls-position="right"
                />
                <span class="policy-unit">秒</span>
              </el-form-item>
            </div>
            <span class="form-help">
              执行超时是下发到节点后的运行等待时间；等待可用只在所有匹配节点繁忙时生效。
            </span>
            <span class="form-help">
              范围：执行超时 30-86400 秒；最大尝试 1-20 次；最大等待 10-86400 秒；重试间隔 1-3600 秒。
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
import AppCronPicker from '@/components/AppCronPicker.vue'
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
  created_time_range: [],
  enabled: '',
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
const clientIpOptions = computed(() => {
  const optionMap = new Map()
  for (const client of clientOptions.value) {
    const ip = getClientIp(client)
    if (!ip) continue
    const option = optionMap.get(ip) || { ip, total: 0, online: 0 }
    option.total += 1
    if (getClientStatus(client) === 'online') option.online += 1
    optionMap.set(ip, option)
  }
  return Array.from(optionMap.values()).sort((a, b) => a.ip.localeCompare(b.ip))
})
const filteredNodeOptions = computed(() => {
  const clientIP = taskForm.client_ip.trim()
  if (!clientIP) return []
  return clientOptions.value.filter((client) => getClientIp(client) === clientIP)
})
const selectedClient = computed(() => findClientById(taskForm.client_id))
const selectedTargetText = computed(() => {
  if (selectedClient.value) return `固定调度到该执行节点：${getClientDetailMeta(selectedClient.value)}`
  if (taskForm.client_ip) return '将在该客户端下自动匹配拥有此工作流的节点执行'
  return '自由调度：选择拥有此工作流的客户端节点执行'
})
const clientWorkflowStatusText = computed(() => {
  if (!taskForm.workflow_id || (!taskForm.client_ip && !taskForm.node_id)) return ''
  if (clientWorkflowChecking.value) return '正在检测当前调度目标是否拥有该工作流...'
  if (selectedClientHasWorkflow.value === true) {
    return taskForm.node_id ? '当前执行节点已拥有该工作流' : '该客户端存在拥有此工作流的节点'
  }
  if (selectedClientHasWorkflow.value === false) {
    return taskForm.node_id
      ? '当前执行节点没有上报该工作流，执行时可能失败'
      : '该客户端当前未发现拥有此工作流的节点，执行时可能失败'
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

watch(() => taskFilters.keyword, () => {
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
      start_time: startTime,
      end_time: endTime,
      enabled: taskFilters.enabled,
      page_num: taskPage.value,
      page_size: taskPageSize.value,
    })
    const list = normalizeList(data, 'tasks')
    tasks.value = list
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
    client_id: row.node_id || row.client_id || '',
    client_name: row.client_name || '',
    client_ip: row.client_ip || row.source_ip || '',
    machine_id: row.machine_id || '',
    node_id: row.node_id || '',
    node_name: row.node_name || '',
    dispatch_mode: row.dispatch_mode || '',
    queue_policy: row.queue_policy || 'queue',
    max_attempts: normalizeInteger(row.max_attempts, 3),
    timeout_seconds: normalizeInteger(row.timeout_seconds, 300),
    queue_wait_seconds: normalizeInteger(row.queue_wait_seconds, 60),
    queue_retry_interval_seconds: normalizeInteger(row.queue_retry_interval_seconds, 5),
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
  const targetNodeId = taskForm.node_id.trim()
  const clientIp = taskForm.client_ip.trim()
  if (!workflowId || (!targetNodeId && !clientIp)) return

  clientWorkflowChecking.value = true
  try {
    const data = await listAutomaSyncCandidatesByWorkflow(workflowId, {
      page_num: 1,
      page_size: clientWorkflowCheckPageSize,
      source_ip: clientIp,
      source_node_id: targetNodeId,
    })
    const candidates = normalizeList(data)
    if (seq !== clientWorkflowCheckSeq.value) return
    selectedClientHasWorkflow.value = candidates.some((item) => {
      if (!candidateHasClientWorkflow(item)) return false
      if (targetNodeId) return getClientNodeId(item) === targetNodeId
      return getClientIp(item) === clientIp
    })
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
    machine_id: row.machine_id || '',
    node_id: row.node_id || '',
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
      machine_id: row.machine_id || '',
      node_id: row.node_id || '',
      params: {
        ...row.params,
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

function handleClientIpChange(clientIP) {
  const nextClientIP = String(clientIP || '').trim()
  if (taskForm.client_ip !== nextClientIP) taskForm.client_ip = nextClientIP
  if (taskForm.node_id) {
    const selectedNode = findClientById(taskForm.client_id)
    if (!selectedNode || getClientIp(selectedNode) !== nextClientIP) {
      clearSelectedNode()
    }
  }
  checkSelectedClientWorkflow()
}

function handleClientIpClear() {
  taskForm.client_ip = ''
  clearSelectedNode()
  checkSelectedClientWorkflow()
}

function handleNodeChange(clientId) {
  const client = findClientById(clientId)
  taskForm.client_name = client ? getClientName(client) : ''
  if (client) taskForm.client_ip = getClientIp(client)
  taskForm.machine_id = client ? getClientMachineId(client) : ''
  taskForm.node_id = client ? getClientNodeId(client) : ''
  taskForm.node_name = client ? getClientNodeName(client) : ''
  checkSelectedClientWorkflow()
}

function handleNodeClear() {
  clearSelectedNode()
  checkSelectedClientWorkflow()
}

function clearSelectedNode() {
  taskForm.client_id = ''
  taskForm.client_name = ''
  taskForm.machine_id = ''
  taskForm.node_id = ''
  taskForm.node_name = ''
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
      message: 'Cron 表达式格式不正确，请填写 GoFrame gcron 支持的 5 段、6 段、@every 或预设表达式',
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
    machine_id: taskForm.machine_id.trim(),
    node_id: taskForm.node_id.trim(),
    node_name: taskForm.node_name.trim(),
    dispatch_mode: getTaskDispatchMode(taskForm),
    queue_policy: taskForm.queue_policy,
    max_attempts: normalizeInteger(taskForm.max_attempts, 3),
    timeout_seconds: normalizeInteger(taskForm.timeout_seconds, 300),
    queue_wait_seconds: normalizeInteger(taskForm.queue_wait_seconds, 60),
    queue_retry_interval_seconds: normalizeInteger(taskForm.queue_retry_interval_seconds, 5),
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
  if (/^@every\s+(\d+(ns|us|µs|ms|s|m|h))+$/i.test(cronExpression)) return true

  const parts = cronExpression.split(/\s+/)
  if (parts.length !== 5 && parts.length !== 6) return false

  const normalizedParts = parts.length === 5 ? ['0', ...parts] : parts
  const ranges = [
    { min: 0, max: 59, names: null, allowQuestion: false, allowHash: true },
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
  if (range.allowHash && part === '#') return true
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

function normalizeInteger(value, fallback) {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue) || numberValue <= 0) return fallback
  return Math.floor(numberValue)
}

function buildTaskPayloadFromRow(row, overrides = {}) {
  return {
    name: String(row?.name || '').trim(),
    description: String(row?.description || '').trim(),
    workflow_id: String(row?.workflow_id || row?.automa_id || '').trim(),
    workflow_name: String(row?.workflow_name || '').trim(),
    client_id: String(row?.client_id || '').trim(),
    client_name: String(row?.client_name || '').trim(),
    client_ip: getTaskClientIpValue(row),
    machine_id: String(row?.machine_id || '').trim(),
    node_id: String(row?.node_id || '').trim(),
    node_name: String(row?.node_name || '').trim(),
    dispatch_mode: String(row?.dispatch_mode || row?.dispatchMode || '').trim(),
    queue_policy: String(row?.queue_policy || row?.queuePolicy || 'queue').trim(),
    max_attempts: normalizeInteger(row?.max_attempts, 3),
    timeout_seconds: normalizeInteger(row?.timeout_seconds, 300),
    queue_wait_seconds: normalizeInteger(row?.queue_wait_seconds, 60),
    queue_retry_interval_seconds: normalizeInteger(row?.queue_retry_interval_seconds, 5),
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
  autoParamItems.value = []
  paramEntries.value = [createEmptyParamEntry()]
  selectedClientHasWorkflow.value = null
  clientWorkflowChecking.value = false
  clientWorkflowCheckSeq.value += 1
}

function resetTaskFilters() {
  taskFilters.keyword = ''
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
    ...payload,
    ...serverTask,
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
  return clientOptions.value.find((item) => getClientId(item) === clientId || String(item?.id || '') === String(clientId || ''))
}

function getWorkflowId(row) {
  return row?.automa_id || row?.workflow_id || row?.workflowId || row?.id || ''
}

function candidateHasClientWorkflow(row) {
  const status = String(row?.sync_status || row?.syncStatus || '').trim()
  if (status === 'client_missing') return false
  return Boolean(row?.automa_id || row?.workflow_id || row?.workflowId || row?.id)
}

function getClientId(row) {
  return getClientNodeId(row) || row?.client_id || row?.clientId || row?.id || ''
}

function getClientNodeId(row) {
  return row?.node_id || row?.nodeId || ''
}

function getClientNodeName(row) {
  return row?.node_name || row?.nodeName || ''
}

function getClientMachineId(row) {
  return row?.machine_id || row?.machineId || ''
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function getClientName(row) {
  return row?.client_name || row?.name || row?.hostname || getClientNodeName(row) || getClientId(row) || ''
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
  return [getClientNodeName(row) || getClientNodeId(row), getClientMachineId(row), getClientStatusText(row)].filter(Boolean).join(' / ')
}

function getNodeSelectMeta(row) {
  return [getClientIp(row), getClientNodeName(row), getClientStatusText(row)].filter(Boolean).join(' / ')
}

function getClientDetailMeta(row) {
  return [getClientNodeName(row), getClientStatusText(row)].filter(Boolean).join(' / ')
}

function getClientOptionLabel(row) {
  return [getClientNodeName(row) || getClientIp(row), getClientSelectMeta(row)].filter(Boolean).join(' / ')
}

function getTaskClientIp(row) {
  return getTaskClientIpValue(row) || '自动匹配'
}

function getTaskClientIpValue(row) {
  const client = row?.node_id ? findClientById(row?.node_id || row?.client_id) : null
  return String(row?.client_ip || (client ? getClientIp(client) : '') || '').trim()
}

function getTaskNodeId(row) {
  return String(row?.node_id || row?.nodeId || '').trim() || '自动匹配'
}

function getTaskDispatchMode(row) {
  if (row?.node_id) return 'node'
  if (row?.target_group_id) return 'group'
  if (row?.client_ip) return 'ip'
  return 'auto'
}

function getScheduleText(row) {
  const cronExpression = row?.cron_expression || row?.cron || ''
  return cronExpression || '手动执行'
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
    machine_id: '',
    node_id: '',
    node_name: '',
    dispatch_mode: 'auto',
    queue_policy: 'queue',
    max_attempts: 3,
    timeout_seconds: 300,
    queue_wait_seconds: 60,
    queue_retry_interval_seconds: 5,
    cron_expression: '',
    enabled: true,
  }
}
</script>

<style scoped lang="scss">
.task-filters {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 10px 20px;
}

.task-filter-fields {
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
}

.filter-item--keyword {
  width: 360px;
  min-width: 0;
}

.filter-item--created-time {
  width: 420px;
  min-width: 0;
}

.filter-item--enabled {
  width: 188px;
  min-width: 0;
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
.schedule-editor,
.policy-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.policy-number-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.policy-number-row :deep(.el-form-item) {
  margin-bottom: 0;
}

.policy-number-row :deep(.el-form-item__content) {
  flex-wrap: nowrap;
}

.policy-number-row :deep(.el-input-number) {
  width: 160px;
}

.policy-unit {
  margin-left: 8px;
  color: #909399;
  font-size: 13px;
}

.client-select-row {
  display: grid;
  grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.1fr) auto;
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
  .filter-item--keyword {
    width: 320px;
  }

  .filter-item--created-time {
    width: 380px;
  }

  .filter-item--enabled {
    width: 168px;
  }
}

@media (max-width: 640px) {
  .page-actions,
  .task-filters,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .task-filter-fields {
    width: 100%;
  }

  .filter-item--keyword,
  .filter-item--created-time,
  .filter-item--enabled {
    width: 100%;
  }
}
</style>
