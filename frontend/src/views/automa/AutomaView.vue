<template>
  <section class="automa-page server-list-page">
    <header class="page-header server-list-actions">
      <div class="header-actions server-list-actions__inner">
        <el-button @click="createDialogVisible = true">新增</el-button>
        <el-button @click="importDialogVisible = true">导入</el-button>
        <el-button type="primary" @click="syncDialogVisible = true">客户端同步</el-button>
        <el-button type="success" @click="maintenanceDialogVisible = true">客户端维护</el-button>
        <el-button :icon="Download" :loading="skillExporting" :disabled="workflows.length === 0"
          @click="handleExportSkill">
          导出 Skill
        </el-button>
        <el-button :icon="RefreshRight" @click="loadWorkflows">刷新</el-button>
      </div>
    </header>

    <section class="workflow-panel server-list-panel">
      <div class="workflow-filters server-list-filters">
        <div class="filter-item filter-item--keyword">
          <span class="filter-label">关键词</span>
          <el-input v-model="filters.keyword" clearable placeholder="检索自定义工作流名称、描述" />
        </div>

        <div class="filter-item filter-item--source">
          <span class="filter-label">来源</span>
          <el-select v-model="filters.source" clearable placeholder="全部">
            <el-option label="全部" value="" />
            <el-option label="新增/导入" :value="1" />
            <el-option label="客户端同步" :value="2" />
          </el-select>
        </div>

        <div class="filter-item filter-item--ip">
          <span class="filter-label">客户端 IP</span>
          <el-select v-model="filters.source_ip" clearable filterable placeholder="选择或检索客户端 IP"
            :loading="clientIpLoading" :value-on-clear="''" @clear="handleClientIpClear"
            @visible-change="handleClientIpSelectVisible">
            <el-option v-for="clientIp in clientIpOptions" :key="clientIp" :label="clientIp" :value="clientIp" />
          </el-select>
        </div>

        <div class="filter-item filter-item--node">
          <span class="filter-label">来源节点</span>
          <el-select v-model="filters.source_node_ids" clearable filterable multiple collapse-tags collapse-tags-tooltip
            placeholder="请先选择客户端 IP" :disabled="!filters.source_ip" :loading="clientIpLoading"
            @visible-change="handleClientIpSelectVisible">
            <el-option v-for="client in filteredClientNodeOptions" :key="client.source_node_id" :label="client.label"
              :value="client.source_node_id" />
          </el-select>
        </div>

        <el-button @click="resetFilters">重置</el-button>
        <el-button type="danger" :disabled="selectedWorkflowIds.length === 0" @click="handleBatchDelete">
          删除选中
        </el-button>
        <AppSelectionSummary :count="selectedWorkflowIds.length" unit="工作流" />
      </div>

      <el-table ref="workflowTableRef" v-loading="loading" class="workflow-table server-list-table adaptive-table"
        :data="pagedWorkflows" border height="100%" :row-key="getWorkflowId" empty-text="暂无工作流"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" reserve-selection />

        <el-table-column label="自定义工作流名称" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.name || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="自定义工作流描述" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.description || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="来源" width="86">
          <template #default="{ row }">
            {{ formatSource(row.source) }}
          </template>
        </el-table-column>

        <el-table-column label="客户端 IP" width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.source_ip || '' }}
          </template>
        </el-table-column>

        <el-table-column label="来源节点 ID" width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.source_node_id || '' }}
          </template>
        </el-table-column>

        <el-table-column label="是否可同步" width="104" align="center" class-name="quick-edit-column">
          <template #default="{ row }">
            <div class="quick-edit-cell">
              <el-switch class="quick-edit-switch" :model-value="!row.is_protected"
                :before-change="() => handleToggleSyncable(row)" />
            </div>
          </template>
        </el-table-column>

        <el-table-column label="节点数/连线数" width="110" align="center">
          <template #default="{ row }">
            {{ formatWorkflowGraphSize(row) }}
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatListDate(row.created_at) }}
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

      <AppPagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="pageSizes"
        :total="workflows.length" />
    </section>

    <AutomaJsonDialog v-model="createDialogVisible" :loading="saving" @submit="handleCreateWorkflows" />
    <AutomaImportDialog v-model="importDialogVisible" :loading="importing" @submit="handleImportFiles" />
    <AutomaSyncDialog v-model="syncDialogVisible" :workflows="workflows" @synced="loadWorkflows" />
    <AutomaClientMaintenanceDialog v-model="maintenanceDialogVisible" :workflows="workflows" @changed="loadWorkflows" />

    <AppDialog v-model="detailVisible" title="工作流详情" width="720px" class="workflow-detail-dialog" confirm-text="保存"
      :loading="detailSaving" :confirm-disabled="detailLoading" @confirm="handleSaveDetail">
      <div v-loading="detailLoading" class="detail-form server-detail-form">
        <div v-for="field in detailFields" :key="field.key" class="detail-field server-detail-field">
          <span class="detail-label server-detail-label">{{ field.label }}</span>
          <div class="detail-control server-detail-control">
            <el-input v-if="field.type === 'textarea'" v-model="detailForm[field.key]" clearable type="textarea"
              :rows="3" />
            <el-switch v-else-if="field.type === 'syncable-switch'" :model-value="!detailForm.is_protected"
              :before-change="() => handleToggleSyncable(detailForm)" />
            <el-switch v-else-if="field.type === 'switch'" v-model="detailForm[field.key]" />
            <el-input v-else-if="field.editable" v-model="detailForm[field.key]" clearable />
            <el-input v-else :model-value="field.value" disabled />
            <el-button v-if="!field.editable" :icon="CopyDocument" :disabled="!field.value"
              @click="copyDetailValue(field.value)">
              复制
            </el-button>
          </div>
        </div>
      </div>
    </AppDialog>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { CopyDocument, Download, RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { copyText } from '@/utils/browser'
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
import { listClients } from '@/services/client'
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

const filters = reactive({
  keyword: '',
  source: '',
  source_ip: '',
  source_node_ids: [],
})

const pagedWorkflows = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return workflows.value.slice(start, start + pageSize.value)
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
} = usePagedTableSelection({
  rows: pagedWorkflows,
  getRowKey: getWorkflowId,
})
const {
  run: scheduleFilterSearch,
  cancel: clearFilterSearchTimer,
} = useDebouncedAction(searchFiltersNow, 200)

const detailFields = computed(() => [
  { key: 'id', label: '服务端 ID', value: formatEmpty(detailForm.id) },
  { key: 'automa_id', label: 'Automa ID', value: formatEmpty(detailForm.automa_id) },
  { key: 'name', label: '自定义工作流名称', editable: true },
  { key: 'description', label: '自定义工作流描述', type: 'textarea', editable: true },
  { key: 'automa_name', label: 'Automa 工作流名称', value: formatEmpty(detailForm.automa_name) },
  { key: 'automa_description', label: 'Automa 工作流描述', value: formatEmpty(detailForm.automa_description) },
  { key: 'is_protected', label: '是否可同步', type: 'syncable-switch', editable: true },
  { key: 'source', label: '来源', value: formatSource(detailForm.source) },
  { key: 'source_ip', label: '客户端 IP', value: formatEmpty(detailForm.source_ip) },
  { key: 'source_node_id', label: '来源节点 ID', value: formatEmpty(detailForm.source_node_id) },
  { key: 'source_user_agent', label: 'User-Agent', value: formatEmpty(detailForm.source_user_agent) },
  { key: 'automa_version', label: 'Automa 版本', value: formatEmpty(detailForm.automa_version) },
  { key: 'ext_version', label: '扩展版本', value: formatEmpty(detailForm.ext_version) },
  { key: 'created_at_automa', label: 'Automa 创建时间', value: formatDate(detailForm.created_at_automa) },
  { key: 'updated_at_automa', label: 'Automa 更新时间', value: formatDate(detailForm.updated_at_automa) },
  { key: 'node_count', label: '节点数', value: formatEmpty(detailForm.node_count) },
  { key: 'edge_count', label: '连线数', value: formatEmpty(detailForm.edge_count) },
  { key: 'raw_json', label: '原始 JSON', value: formatEmpty(detailForm.raw_json) },
  { key: 'normalized_json', label: '规范化 JSON', value: formatEmpty(detailForm.normalized_json) },
  { key: 'content_hash', label: '内容 Hash', value: formatEmpty(detailForm.content_hash) },
  { key: 'revision', label: '版本号', value: formatEmpty(detailForm.revision) },
  { key: 'first_synced_at', label: '首次同步时间', value: formatDate(detailForm.first_synced_at) },
  { key: 'last_synced_at', label: '最近同步时间', value: formatDate(detailForm.last_synced_at) },
  { key: 'created_at', label: '创建时间', value: formatDate(detailForm.created_at) },
  { key: 'updated_at', label: '更新时间', value: formatDate(detailForm.updated_at) },
])

onMounted(() => {
  loadWorkflows()
  loadClientIpOptions()
})

watch(() => filters.source, () => {
  clearFilterSearchTimer()
  currentPage.value = 1
  loadWorkflows()
})

watch(() => filters.source_ip, () => {
  clearFilterSearchTimer()
  const validNodeIds = new Set(filteredClientNodeOptions.value.map((client) => client.source_node_id))
  filters.source_node_ids = filters.source_node_ids.filter((nodeId) => validNodeIds.has(nodeId))
  currentPage.value = 1
  loadWorkflows()
})

watch(() => filters.source_node_ids, () => {
  clearFilterSearchTimer()
  currentPage.value = 1
  loadWorkflows()
}, { deep: true })

watch(() => filters.keyword, () => {
  scheduleFilterSearch()
})

watch([workflows, pageSize], () => {
  currentPage.value = getSafePage({
    total: workflows.value.length,
    page: currentPage.value,
    size: pageSize.value,
  })
})

watch(pagedWorkflows, () => {
  restoreWorkflowSelection(workflowTableRef)
})

async function loadWorkflows() {
  loading.value = true
  try {
    const data = await listAutomaWorkflows({
      keyword: filters.keyword.trim(),
      source: filters.source,
      source_ip: normalizeText(filters.source_ip),
      source_node_ids: filters.source_node_ids,
      page_num: 1,
      page_size: 60,
    })
    workflows.value = normalizeList(data, 'workflows')
    retainWorkflowSelectionByRows(workflows.value)
  } finally {
    loading.value = false
  }
}

async function loadClientIpOptions() {
  clientIpLoading.value = true
  try {
    const data = await listClients()
    const seen = new Set()
    clientOptions.value = normalizeList(data, 'clients')
      .map((client) => {
        const clientIp = getClientIp(client)
        const nodeId = normalizeText(client?.node_id || client?.nodeId)
        const key = buildNodeIdentity(clientIp, nodeId)
        return {
          key,
          source_ip: clientIp,
          source_node_id: nodeId,
          label: nodeId ? `${clientIp} / ${nodeId}` : clientIp,
        }
      })
      .filter((client) => client.key)
      .filter((client) => {
        if (seen.has(client.key)) return false
        seen.add(client.key)
        return true
      })
    clientIpOptions.value = Array.from(new Set(clientOptions.value.map((client) => client.source_ip).filter(Boolean)))
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
  currentPage.value = 1
  loadWorkflows()
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
  showSuccessMessage('工作流已删除')
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

  await batchDeleteAutomaWorkflows(ids)
  showSuccessMessage('已删除选中工作流')
  await loadWorkflows()
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

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function buildNodeIdentity(clientIp, nodeId) {
  clientIp = normalizeText(clientIp)
  nodeId = normalizeText(nodeId)
  if (!clientIp) return nodeId
  if (!nodeId || nodeId === clientIp) return clientIp
  return `${clientIp}|${nodeId}`
}

function formatWorkflowGraphSize(row) {
  const nodeCount = row?.node_count ?? ''
  const edgeCount = row?.edge_count ?? ''
  return `${nodeCount} / ${edgeCount}`
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

function formatWorkflowMutationMessage(prefix, result, fallback) {
  const created = Number(result?.created || 0)
  const updated = Number(result?.updated || 0)
  const unchanged = Number(result?.unchanged || 0)
  const parts = []

  if (created > 0) parts.push(`新增 ${created} 个`)
  if (updated > 0) parts.push(`更新 ${updated} 个`)
  if (unchanged > 0) parts.push(`重复 ${unchanged} 个`)
  if (parts.length === 0) return fallback

  return `${prefix}：${parts.join('，')}`
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

function showSuccessMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message,
  })
}

function formatListDate(value) {
  return formatDate(value)
}
</script>

<style scoped lang="scss">
.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item--keyword {
  width: 320px;
}

.filter-item--source {
  width: 180px;
}

.filter-item--ip {
  width: 240px;
}

.filter-item--node {
  width: 260px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.field-value {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: #303133;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-field {
  grid-template-columns: 148px minmax(0, 1fr);
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
  .filter-item,
  .header-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item--keyword,
  .filter-item--source,
  .filter-item--ip,
  .filter-item--node {
    width: 100%;
  }

  .detail-field {
    grid-template-columns: 1fr;
  }

  .detail-label {
    text-align: left;
  }

  .detail-control {
    grid-template-columns: minmax(0, 1fr) 64px;
  }
}
</style>
