<template>
  <section class="automa-page server-list-page">
    <header class="page-header server-list-actions">
      <div class="header-actions server-list-actions__inner">
        <el-button @click="createDialogVisible = true">新增 JSON</el-button>
        <el-button @click="importDialogVisible = true">导入 ZIP</el-button>
        <el-button type="primary" @click="syncDialogVisible = true">客户端同步</el-button>
        <el-button type="success" @click="maintenanceDialogVisible = true">客户端维护</el-button>
        <el-button :icon="Download" :loading="skillExporting" :disabled="workflows.length === 0"
          @click="handleExportSkill">
          导出 Skill
        </el-button>
        <el-button type="danger" :disabled="selectedWorkflowIds.length === 0" @click="handleBatchDelete">
          删除选中
        </el-button>
        <el-button :icon="RefreshRight" @click="loadWorkflows">刷新</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </header>

    <section class="workflow-panel server-list-panel">
      <div class="workflow-filters server-list-filters">
        <div class="workflow-filter-fields server-filter-fields">
          <div class="server-filter-item server-filter-item--text filter-item--keyword">
            <span class="server-filter-label">关键词</span>
            <el-input v-model="filters.keyword" clearable placeholder="工作流名称、描述、Automa ID等" />
          </div>

          <div class="server-filter-item server-filter-item--short filter-item--source">
            <span class="server-filter-label">来源</span>
            <el-select v-model="filters.source" clearable placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="新增/导入" :value="1" />
              <el-option label="客户端同步" :value="2" />
            </el-select>
          </div>

          <div class="server-filter-item server-filter-item--short filter-item--syncable">
            <span class="server-filter-label">是否可同步</span>
            <el-select v-model="filters.syncable" clearable placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="是" :value="1" />
              <el-option label="否" :value="2" />
            </el-select>
          </div>

          <div class="server-filter-item server-filter-item--medium filter-item--ip">
            <span class="server-filter-label">客户端</span>
            <el-select v-model="filters.source_ip" clearable filterable placeholder="选择或检索客户端"
              :loading="clientIpLoading" :value-on-clear="''" @clear="handleClientIpClear"
              @visible-change="handleClientIpSelectVisible">
              <el-option v-for="clientIp in clientIpOptions" :key="clientIp" :label="clientIp" :value="clientIp" />
            </el-select>
          </div>

          <div class="server-filter-item server-filter-item--node filter-item--node">
            <span class="server-filter-label">来源节点</span>
            <el-select v-model="filters.source_node_ids" clearable filterable multiple collapse-tags
              collapse-tags-tooltip placeholder="请先选择客户端" :disabled="!filters.source_ip" :loading="clientIpLoading"
              @visible-change="handleClientIpSelectVisible">
              <el-option v-for="client in filteredClientNodeOptions" :key="client.source_node_id" :label="client.label"
                :value="client.source_node_id" />
            </el-select>
          </div>

          <div class="server-filter-item server-filter-item--time filter-item--created-time">
            <span class="server-filter-label">创建时间</span>
            <AppTimeRangeFilter v-model="filters.created_time_range" />
          </div>
        </div>
      </div>

      <el-table ref="workflowTableRef" v-loading="loading" class="workflow-table server-list-table adaptive-table"
        :data="pagedWorkflows" border height="100%" :row-key="getWorkflowId" empty-text="暂无工作流"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" reserve-selection />

        <el-table-column label="自定义工作流名称" min-width="112" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.name || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="自定义工作流描述" min-width="124" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.description || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="来源" width="112">
          <template #default="{ row }">
            {{ formatSource(row.source) }}
          </template>
        </el-table-column>

        <el-table-column label="客户端/来源节点" width="158" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatSourceNode(row) }}
          </template>
        </el-table-column>

        <el-table-column width="104" align="center" class-name="quick-edit-column">
          <template #header>
            <span class="table-label-with-help">
              是否可同步
              <el-tooltip :content="SYNCABLE_HELP" placement="top">
                <el-icon class="syncable-help-icon">
                  <InfoFilled />
                </el-icon>
              </el-tooltip>
            </span>
          </template>
          <template #default="{ row }">
            <div class="quick-edit-cell">
              <el-switch class="quick-edit-switch" :model-value="!row.is_protected"
                :before-change="() => handleToggleSyncable(row)" />
            </div>
          </template>
        </el-table-column>

        <el-table-column label="节点数/连线数" width="128" align="center">
          <template #default="{ row }">
            {{ formatWorkflowGraphSize(row) }}
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatListDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="最近同步到服务端" width="172" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatListDate(row.last_synced_at) }}
          </template>
        </el-table-column>

        <el-table-column label="更新时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatListDate(row.updated_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openWorkflowDetail(row)">详情</el-button>
            <el-button link type="danger" @click="handleDeleteWorkflow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="workflow-footer server-list-footer">
        <AppSelectionSummary :count="selectedWorkflowIds.length" unit="工作流" />
        <AppPagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="pageSizes"
          :total="workflowTotal" />
      </div>
    </section>

    <AutomaJsonDialog v-model="createDialogVisible" :loading="saving" @submit="handleCreateWorkflows" />
    <AutomaImportDialog v-model="importDialogVisible" :loading="importing" @submit="handleImportFiles" />
    <AutomaSyncDialog v-model="syncDialogVisible" :workflows="workflows" @synced="loadWorkflows" />
    <AutomaClientMaintenanceDialog v-model="maintenanceDialogVisible" :workflows="workflows" @changed="loadWorkflows" />

    <AppDialog v-model="detailVisible" title="工作流详情" width="min(1040px, calc(100vw - 32px))"
      class="workflow-detail-dialog" confirm-text="保存" :loading="detailSaving" :confirm-disabled="detailLoading"
      @confirm="handleSaveDetail">
      <ServerDetailGroups
        v-loading="detailLoading"
        :groups="detailGroups"
        :model="detailForm"
        :toggle-syncable="handleToggleSyncable"
        @copy="copyDetailValue"
        @expand="openLongTextDialog"
      />
    </AppDialog>

    <AppLongTextDialog
      v-model="longTextVisible"
      :title="longTextDialog.title"
      :value="longTextDialog.value"
      @copy="copyDetailValue"
    />
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Download, InfoFilled, RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppLongTextDialog from '@/components/AppLongTextDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import AppTimeRangeFilter from '@/components/AppTimeRangeFilter.vue'
import ServerDetailGroups from '@/components/ServerDetailGroups.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { copyText } from '@/utils/browser'
import { buildClientNodeOptions, getClientIp } from '@/utils/clientNode'
import { formatDate, formatEmpty } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, getSafePage, normalizeList, normalizeText } from '@/utils/list'
import {
  batchDeleteAutomaWorkflows,
  createAutomaWorkflow,
  deleteAutomaWorkflow,
  exportServerAutomaSkill,
  getAutomaWorkflowDetail,
  importAutomaWorkflowFiles,
  listAutomaWorkflows,
  updateAutomaWorkflow,
  updateAutomaWorkflowProtected,
} from '@/services/automa'
import { listAllClients } from '@/services/client'
import { downloadBlob } from '@/utils/browser'
import AutomaImportDialog from './components/AutomaImportDialog.vue'
import AutomaJsonDialog from './components/AutomaJsonDialog.vue'
import AutomaClientMaintenanceDialog from './components/AutomaClientMaintenanceDialog.vue'
import AutomaSyncDialog from './components/AutomaSyncDialog.vue'

const workflows = ref([])
const loading = ref(false)
const saving = ref(false)
const importing = ref(false)
const detailLoading = ref(false)
const detailSaving = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const workflowTotal = ref(0)
const pageSizes = DEFAULT_PAGE_SIZES
const detailVisible = ref(false)
const detailWorkflow = ref(null)
const detailForm = reactive(createDetailForm())
const workflowTableRef = ref(null)
const skillExporting = ref(false)
const clientIpLoading = ref(false)
const clientOptions = ref([])
const clientIpOptions = ref([])
const createDialogVisible = ref(false)
const importDialogVisible = ref(false)
const syncDialogVisible = ref(false)
const maintenanceDialogVisible = ref(false)
const longTextVisible = ref(false)
const longTextDialog = reactive({ title: '', value: '' })
const SYNCABLE_HELP = '开启后客户端可在工作流列表中看到并同步到本地；关闭后客户端不可见，也无法同步到本地。'

const filters = reactive({
  keyword: '',
  source: '',
  source_ip: '',
  source_node_ids: [],
  created_time_range: [],
  syncable: '',
})

const pagedWorkflows = computed(() => {
  return workflows.value
})
const filteredClientNodeOptions = computed(() => {
  const sourceIp = normalizeText(filters.source_ip)
  if (!sourceIp) return []
  return clientOptions.value.filter((client) => client.source_ip === sourceIp && client.source_node_id)
})
const {
  selectedKeys: selectedWorkflowIds,
  handleSelectionChange,
  restoreSelection: restoreWorkflowSelection,
  retainSelectionByRows: retainWorkflowSelectionByRows,
  resetSelection: resetWorkflowSelection,
} = usePagedTableSelection({
  rows: pagedWorkflows,
  getRowKey: getWorkflowId,
})
const {
  run: scheduleFilterSearch,
  cancel: clearFilterSearchTimer,
} = useDebouncedAction(searchFiltersNow, 200)

const detailGroups = computed(() => {
  const groups = [
    {
      key: 'base',
      title: '基础信息',
      fields: [
        { key: 'id', label: '服务端 ID', value: formatEmpty(detailForm.id), copyable: true },
        { key: 'automa_id', label: 'Automa ID', value: formatEmpty(detailForm.automa_id), copyable: true },
        { key: 'name', label: '自定义工作流名称', editable: true },
        { key: 'description', label: '自定义工作流描述', type: 'textarea', editable: true },
        { key: 'automa_name', label: 'Automa 工作流名称', value: formatEmpty(detailForm.automa_name) },
        { key: 'automa_description', label: 'Automa 工作流描述', value: formatEmpty(detailForm.automa_description) },
        { key: 'is_disabled', label: 'Automa 状态', value: formatAutomaStatus(detailForm) },
        { key: 'is_protected', label: '是否可同步', type: 'syncable-switch', editable: true, help: SYNCABLE_HELP },
      ],
    },
    {
      key: 'source',
      title: '来源信息',
      fields: [
        { key: 'source', label: '来源', value: formatSource(detailForm.source) },
        { key: 'source_ip', label: '客户端', value: formatEmpty(detailForm.source_ip), copyable: true },
        { key: 'source_node_id', label: '来源节点', value: formatEmpty(detailForm.source_node_id), copyable: true },
        {
          key: 'source_user_agent',
          label: 'User-Agent',
          value: formatEmpty(detailForm.source_user_agent),
          type: 'textarea',
          copyable: true,
          expandable: true,
        },
        { key: 'automa_version', label: 'Automa 版本', value: formatEmpty(detailForm.automa_version) },
        { key: 'ext_version', label: '扩展版本', value: formatEmpty(detailForm.ext_version) },
      ],
    },
    {
      key: 'graph',
      title: '结构信息',
      fields: [
        { key: 'node_count', label: '节点数', value: formatEmpty(detailForm.node_count) },
        { key: 'edge_count', label: '连线数', value: formatEmpty(detailForm.edge_count) },
        { key: 'content_hash', label: '内容 Hash', value: formatEmpty(detailForm.content_hash), copyable: true },
        { key: 'revision', label: '版本号', value: formatEmpty(detailForm.revision) },
      ],
    },
    {
      key: 'json',
      title: 'JSON 信息',
      fields: [
        {
          key: 'raw_json',
          label: '原始 JSON',
          value: formatDetailJson(detailForm.raw_json),
          type: 'textarea',
          copyable: true,
          expandable: true,
        },
        {
          key: 'normalized_json',
          label: '规范化 JSON',
          value: formatDetailJson(detailForm.normalized_json),
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
        { key: 'created_at_automa', label: 'Automa 创建时间', value: formatDate(detailForm.created_at_automa) },
        { key: 'updated_at_automa', label: 'Automa 更新时间', value: formatDate(detailForm.updated_at_automa) },
        { key: 'first_synced_at', label: '首次同步时间', value: formatDate(detailForm.first_synced_at) },
        { key: 'last_synced_at', label: '最近同步到服务端', value: formatDate(detailForm.last_synced_at) },
        { key: 'created_at', label: '创建时间', value: formatDate(detailForm.created_at) },
        { key: 'updated_at', label: '更新时间', value: formatDate(detailForm.updated_at) },
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
  loadWorkflows()
})

watch(() => [
  filters.source,
  filters.syncable,
  filters.created_time_range,
], () => {
  clearFilterSearchTimer()
  reloadFirstWorkflowPage()
})

watch(() => filters.source_ip, () => {
  clearFilterSearchTimer()
  const validNodeIds = new Set(filteredClientNodeOptions.value.map((client) => client.source_node_id))
  filters.source_node_ids = filters.source_node_ids.filter((nodeId) => validNodeIds.has(nodeId))
  reloadFirstWorkflowPage()
})

watch(() => filters.source_node_ids, () => {
  clearFilterSearchTimer()
  reloadFirstWorkflowPage()
}, { deep: true })

watch(() => filters.keyword, () => {
  scheduleFilterSearch()
})

watch(pagedWorkflows, () => {
  restoreWorkflowSelection(workflowTableRef)
})

watch(currentPage, () => {
  loadWorkflows()
})

watch(pageSize, () => {
  reloadFirstWorkflowPage()
})

async function loadWorkflows() {
  loading.value = true
  try {
    const [startTime, endTime] = getCreatedTimeRange()
    const data = await listAutomaWorkflows({
      keyword: filters.keyword.trim(),
      source: filters.source,
      syncable: filters.syncable,
      source_ip: normalizeText(filters.source_ip),
      source_node_ids: filters.source_node_ids,
      start_time: startTime,
      end_time: endTime,
      page_num: currentPage.value,
      page_size: pageSize.value,
    })
    const list = normalizeList(data, 'workflows')
    workflows.value = list
    workflowTotal.value = Number(data?.total ?? list.length)
    currentPage.value = getSafePage({
      total: workflowTotal.value,
      page: currentPage.value,
      size: pageSize.value,
    })
    retainWorkflowSelectionByRows(workflows.value)
  } finally {
    loading.value = false
  }
}

async function loadClientIpOptions(force = false) {
  if (!force && clientOptions.value.length > 0) return

  clientIpLoading.value = true
  try {
    const data = await listAllClients()
    const { ipOptions, nodeOptions } = buildClientNodeOptions(normalizeList(data, 'clients'), {
      label: ({ clientIp, nodeId }) => (nodeId ? `${clientIp} / ${nodeId}` : clientIp),
    })
    clientOptions.value = nodeOptions
    clientIpOptions.value = ipOptions
  } finally {
    clientIpLoading.value = false
  }
}

function handleClientIpSelectVisible(opened) {
  if (opened) loadClientIpOptions()
}

function handleClientIpClear() {
  filters.source_ip = ''
  filters.source_node_ids = []
  searchFiltersNow()
}

function searchFiltersNow() {
  clearFilterSearchTimer()
  reloadFirstWorkflowPage()
}

function reloadFirstWorkflowPage() {
  if (currentPage.value === 1) {
    loadWorkflows()
    return
  }
  currentPage.value = 1
}

async function handleCreateWorkflows(payload) {
  saving.value = true
  try {
    const workflowsToCreate = payload.map((workflow) => ({
      ...workflow,
      source: 1,
    }))

    const result = await createAutomaWorkflow(workflowsToCreate)
    showSuccessMessage(formatWorkflowMutationMessage('新增完成', result, `已新增 ${workflowsToCreate.length} 个工作流`))

    createDialogVisible.value = false
    await loadWorkflows()
  } finally {
    saving.value = false
  }
}

async function handleImportFiles(file) {
  importing.value = true
  try {
    const result = await importAutomaWorkflowFiles(file)
    showSuccessMessage(formatWorkflowMutationMessage('导入完成', result, '导入完成'))

    importDialogVisible.value = false
    await loadWorkflows()
  } finally {
    importing.value = false
  }
}

async function openWorkflowDetail(row) {
  const workflowId = getWorkflowId(row)
  setDetailForm(row)
  detailWorkflow.value = { ...row }
  detailVisible.value = true
  if (!workflowId) return

  detailLoading.value = true
  try {
    const data = await getAutomaWorkflowDetail(workflowId)
    detailWorkflow.value = data?.workflow || data || row
    setDetailForm(detailWorkflow.value)
  } finally {
    detailLoading.value = false
  }
}

async function handleSaveDetail() {
  if (!detailForm.id) return

  detailSaving.value = true
  try {
    await updateAutomaWorkflow(detailForm.id, {
      name: detailForm.name,
      description: detailForm.description,
      source: normalizeSourceValue(detailWorkflow.value?.source ?? detailForm.source, 1),
      is_protected: detailWorkflow.value?.is_protected ?? detailForm.is_protected,
      revision: detailForm.revision,
    })
    showSuccessMessage('工作流已保存')
    detailVisible.value = false
    await loadWorkflows()
  } finally {
    detailSaving.value = false
  }
}

async function handleDeleteWorkflow(row) {
  const confirmed = await appConfirm({
    title: '删除工作流',
    message: '确认删除这个工作流吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteAutomaWorkflow(getWorkflowId(row))
  showSuccessMessage('已删除 1 个工作流')
  await loadWorkflows()
}

async function handleBatchDelete() {
  const ids = selectedWorkflowIds.value.slice()
  if (ids.length === 0) return

  const confirmed = await appConfirm({
    title: '批量删除工作流',
    message: `确认删除选中的 ${ids.length} 个工作流吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  const result = await batchDeleteAutomaWorkflows(ids)
  showSuccessMessage(formatBatchDeleteMessage(result, ids))
  await loadWorkflows()
  resetWorkflowSelection(workflowTableRef)
}

async function handleExportSkill() {
  if (workflows.value.length === 0) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '暂无可导出的工作流' })
    return
  }

  skillExporting.value = true
  try {
    const workflowIds = selectedWorkflowIds.value.slice()
    const blob = await exportServerAutomaSkill({
      scope: workflowIds.length > 0 ? 'selected' : 'all',
      workflowIds,
    })
    downloadBlob(blob, 'SKILL.md')
    showSuccessMessage('Skill 已导出')
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '导出 Skill 失败' })
  } finally {
    skillExporting.value = false
  }
}

function resetFilters() {
  filters.keyword = ''
  filters.source = ''
  filters.source_ip = ''
  filters.source_node_ids = []
  filters.created_time_range = []
  filters.syncable = ''
}

function getCreatedTimeRange() {
  const range = Array.isArray(filters.created_time_range)
    ? filters.created_time_range
    : []
  return [range[0] || '', range[1] || '']
}

function createDetailForm() {
  return {
    id: '',
    automa_id: '',
    name: '',
    description: '',
    automa_name: '',
    automa_description: '',
    source: 1,
    source_ip: '',
    source_node_id: '',
    source_user_agent: '',
    automa_version: '',
    ext_version: '',
    created_at_automa: '',
    updated_at_automa: '',
    is_disabled: false,
    is_protected: false,
    node_count: 0,
    edge_count: 0,
    raw_json: '',
    normalized_json: '',
    content_hash: '',
    revision: 1,
    first_synced_at: '',
    last_synced_at: '',
    created_at: '',
    updated_at: '',
  }
}

function setDetailForm(row = {}) {
  Object.assign(detailForm, {
    ...createDetailForm(),
    ...row,
    source: normalizeSourceValue(row.source, 1),
    is_disabled: Boolean(row.is_disabled),
    is_protected: Boolean(row.is_protected),
    revision: row.revision || 1,
  })
}

function getWorkflowId(row) {
  return row?.id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function formatWorkflowGraphSize(row) {
  const nodeCount = row?.node_count ?? ''
  const edgeCount = row?.edge_count ?? ''
  return `${nodeCount} / ${edgeCount}`
}

function formatAutomaStatus(row) {
  return row?.is_disabled ? '禁用' : '启用'
}

function normalizeSourceValue(source, fallback = 1) {
  if (source === 1 || source === '1' || source === '导入' || source === '页面导入' || source === '新增/导入') return 1
  if (source === 2 || source === '2' || source === '同步' || source === '客户端同步') return 2
  return Number(source) || fallback
}

function formatSource(source) {
  const sourceValue = normalizeSourceValue(source, 0)
  if (sourceValue === 1) return '新增/导入'
  if (sourceValue === 2) return '客户端同步'
  return ''
}

function formatSourceNode(row) {
  const sourceIp = normalizeText(row?.source_ip)
  const sourceNodeId = normalizeText(row?.source_node_id)
  return [sourceIp, sourceNodeId].filter(Boolean).join(' / ')
}

function formatWorkflowMutationMessage(prefix, result, fallback) {
  const created = Number(result?.created || 0)
  const updated = Number(result?.updated || 0)
  const unchanged = Number(result?.unchanged || 0)
  const parts = []

  if (created > 0) parts.push(`新增 ${created} 个`)
  if (updated > 0) parts.push(`更新 ${updated} 个`)
  if (unchanged > 0) parts.push(`无变化 ${unchanged} 个`)
  if (parts.length === 0) return fallback

  return `${prefix}：${parts.join('，')}`
}

function formatBatchDeleteMessage(result, fallbackIds) {
  const success = Number(result?.success || 0) || fallbackIds.length
  const notFound = Number(result?.not_found || result?.notFound || 0)
  if (notFound > 0) {
    return `已删除 ${success} 个工作流（未找到 ${notFound} 个）`
  }
  return `已删除 ${success} 个工作流`
}

async function handleToggleSyncable(row) {
  if (!row?.id) return false

  const nextProtected = !row.is_protected
  const nextRevision = Number(row.revision || 1) + 1

  await updateAutomaWorkflowProtected(row.id, {
    is_protected: nextProtected,
    revision: row.revision || 1,
  })

  syncWorkflowProtectedState(row, nextProtected, nextRevision)
  showSuccessMessage('是否可同步已更新')
  return true
}

function syncWorkflowProtectedState(row, isProtected, revision) {
  const workflowId = getWorkflowId(row)
  const listWorkflow = workflows.value.find((workflow) => getWorkflowId(workflow) === workflowId)

  if (listWorkflow) {
    listWorkflow.is_protected = isProtected
    listWorkflow.revision = revision
  }

  if (detailForm.id && getWorkflowId(detailForm) === workflowId) {
    detailForm.is_protected = isProtected
    detailForm.revision = revision
  }

  if (detailWorkflow.value && getWorkflowId(detailWorkflow.value) === workflowId) {
    detailWorkflow.value = {
      ...detailWorkflow.value,
      is_protected: isProtected,
      revision,
    }
  }
}

async function copyDetailValue(value) {
  await copyText(value)
  showSuccessMessage('已复制')
}

function openLongTextDialog(field) {
  longTextDialog.title = field.label
  longTextDialog.value = field.value || ''
  longTextVisible.value = true
}

function showSuccessMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message,
  })
}

function formatListDate(value) {
  return formatDate(value)
}

function formatDetailJson(value) {
  if (!value) return ''
  if (typeof value !== 'string') {
    return JSON.stringify(value, null, 2)
  }

  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}
</script>

<style scoped lang="scss">
.workflow-filters {
  display: flex;
  align-items: flex-start;
  flex-direction: row;
  gap: 8px 14px;
}

.workflow-filter-fields {
  gap: 8px 14px;
}

.field-value {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: #303133;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-label-with-help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  white-space: nowrap;
}

.syncable-help-icon {
  flex: 0 0 auto;
  color: #909399;
  font-size: 14px;
  cursor: help;
}

@media (max-width: 1100px) {
  .automa-page {
    height: auto;
    overflow: visible;
  }

  .page-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

}

@media (max-width: 640px) {

  .workflow-filters,
  .header-actions {
    align-items: stretch;
    flex-direction: column;
  }

}
</style>
