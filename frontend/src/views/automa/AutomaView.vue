<template>
  <section class="automa-page">
    <header class="page-header">
      <div class="header-actions">
        <el-button @click="createDialogVisible = true">新增</el-button>
        <el-button @click="importDialogVisible = true">文件导入</el-button>
        <el-button type="primary" @click="syncDialogVisible = true">客户端同步</el-button>
        <el-button :icon="RefreshRight" @click="loadWorkflows">刷新</el-button>
      </div>
    </header>

    <section class="workflow-panel">
      <div class="workflow-filters">
        <div class="filter-item filter-item--keyword">
          <span class="filter-label">关键词</span>
          <el-input v-model="filters.keyword" clearable placeholder="工作流名称、工作流描述" />
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

        <el-button @click="resetFilters">重置</el-button>
        <el-button type="danger" :disabled="selectedWorkflowIds.length === 0" @click="handleBatchDelete">
          删除选中
        </el-button>
      </div>

      <el-table v-loading="loading" class="workflow-table adaptive-table" :data="pagedWorkflows" border height="100%"
        row-key="id" empty-text="暂无工作流" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" />

        <el-table-column label="工作流名称" min-width="180">
          <template #default="{ row }">
            <div class="workflow-name">
              <span>{{ row.name || '' }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="工作流描述" min-width="160">
          <template #default="{ row }">
            {{ row.description || '' }}
          </template>
        </el-table-column>

        <el-table-column label="来源" width="86">
          <template #default="{ row }">
            {{ formatSource(row.source) }}
          </template>
        </el-table-column>

        <el-table-column label="客户端 IP" width="120">
          <template #default="{ row }">
            {{ row.source_ip || '' }}
          </template>
        </el-table-column>

        <el-table-column label="是否可同步" width="92" align="center">
          <template #default="{ row }">
            <el-switch :model-value="!row.is_protected" :before-change="() => handleToggleSyncable(row)" />
          </template>
        </el-table-column>

        <el-table-column label="节点数" width="70" align="center">
          <template #default="{ row }">
            {{ row.node_count ?? '' }}
          </template>
        </el-table-column>

        <el-table-column label="连线数" width="70" align="center">
          <template #default="{ row }">
            {{ row.edge_count ?? '' }}
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

    <AppDialog v-model="detailVisible" title="工作流详情" width="720px" class="workflow-detail-dialog" confirm-text="保存"
      :loading="detailSaving" :confirm-disabled="detailLoading" @confirm="handleSaveDetail">
      <div v-loading="detailLoading" class="detail-form">
        <div v-for="field in detailFields" :key="field.key" class="detail-field">
          <span class="detail-label">{{ field.label }}</span>
          <div class="detail-control">
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
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { CopyDocument, RefreshRight } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import {
  batchDeleteAutomaWorkflows,
  createAutomaWorkflow,
  deleteAutomaWorkflow,
  getAutomaWorkflowDetail,
  importAutomaWorkflowFiles,
  listAutomaWorkflows,
  updateAutomaWorkflow,
  updateAutomaWorkflowProtected,
} from '@/services/automa'
import { listClients } from '@/services/client'
import AutomaImportDialog from './components/AutomaImportDialog.vue'
import AutomaJsonDialog from './components/AutomaJsonDialog.vue'
import AutomaSyncDialog from './components/AutomaSyncDialog.vue'

const workflows = ref([])
const loading = ref(false)
const saving = ref(false)
const importing = ref(false)
const detailLoading = ref(false)
const detailSaving = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = [10, 30, 60]
const filterSearchDelay = 200
let filterSearchTimer = null
const detailVisible = ref(false)
const detailWorkflow = ref(null)
const detailForm = reactive(createDetailForm())
const selectedWorkflowIds = ref([])
const clientIpLoading = ref(false)
const clientIpOptions = ref([])
const createDialogVisible = ref(false)
const importDialogVisible = ref(false)
const syncDialogVisible = ref(false)

const filters = reactive({
  keyword: '',
  source: '',
  source_ip: '',
})

const pagedWorkflows = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return workflows.value.slice(start, start + pageSize.value)
})

const detailFields = computed(() => [
  { key: 'id', label: '服务端 ID', value: formatEmpty(detailForm.id) },
  { key: 'automa_id', label: 'Automa ID', value: formatEmpty(detailForm.automa_id) },
  { key: 'name', label: '名称', editable: true },
  { key: 'description', label: '工作流描述', type: 'textarea', editable: true },
  { key: 'is_protected', label: '是否可同步', type: 'syncable-switch', editable: true },
  { key: 'source', label: '来源', value: formatSource(detailForm.source) },
  { key: 'source_ip', label: '客户端 IP', value: formatEmpty(detailForm.source_ip) },
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

onBeforeUnmount(() => {
  clearFilterSearchTimer()
})

watch(() => filters.source, () => {
  clearFilterSearchTimer()
  currentPage.value = 1
  loadWorkflows()
})

watch(() => filters.source_ip, () => {
  clearFilterSearchTimer()
  currentPage.value = 1
  loadWorkflows()
})

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

async function loadWorkflows() {
  loading.value = true
  try {
    const data = await listAutomaWorkflows({
      keyword: filters.keyword.trim(),
      source: filters.source,
      source_ip: normalizeFilterText(filters.source_ip),
      page_num: 1,
      page_size: 60,
    })
    workflows.value = normalizeList(data, 'workflows')
    selectedWorkflowIds.value = selectedWorkflowIds.value.filter((id) =>
      workflows.value.some((workflow) => getWorkflowId(workflow) === id),
    )
  } finally {
    loading.value = false
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

function handleClientIpClear() {
  filters.source_ip = ''
  searchFiltersNow()
}

function scheduleFilterSearch() {
  clearFilterSearchTimer()
  filterSearchTimer = window.setTimeout(() => {
    searchFiltersNow()
  }, filterSearchDelay)
}

function searchFiltersNow() {
  clearFilterSearchTimer()
  currentPage.value = 1
  loadWorkflows()
}

function clearFilterSearchTimer() {
  if (!filterSearchTimer) return

  window.clearTimeout(filterSearchTimer)
  filterSearchTimer = null
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
  await ElMessageBox.confirm('确认删除这个工作流吗？', '删除工作流', { type: 'warning' })
  await deleteAutomaWorkflow(getWorkflowId(row))
  showSuccessMessage('工作流已删除')
  await loadWorkflows()
}

async function handleBatchDelete() {
  const ids = selectedWorkflowIds.value.slice()
  if (ids.length === 0) return

  await ElMessageBox.confirm(`确认删除选中的 ${ids.length} 个工作流吗？`, '批量删除工作流', {
    type: 'warning',
  })
  await batchDeleteAutomaWorkflows(ids)
  showSuccessMessage('已删除选中工作流')
  await loadWorkflows()
}

function handleSelectionChange(selection) {
  selectedWorkflowIds.value = selection.map((item) => getWorkflowId(item)).filter(Boolean)
}

function resetFilters() {
  filters.keyword = ''
  filters.source = ''
  filters.source_ip = ''
}

function normalizeFilterText(value) {
  return String(value || '').trim()
}

function createDetailForm() {
  return {
    id: '',
    automa_id: '',
    name: '',
    description: '',
    source: 1,
    source_ip: '',
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

function normalizeList(data, fallbackKey) {
  const list = data?.list || data?.[fallbackKey] || []
  return Array.isArray(list) ? list : []
}

function getWorkflowId(row) {
  return row?.id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function getClientIp(row) {
  return row?.client_ip || row?.ip || row?.remote_ip || row?.last_ip || row?.source_ip || ''
}

function normalizeSourceValue(source, fallback = 1) {
  if (source === 1 || source === '1' || source === '导入' || source === '新增/导入') return 1
  if (source === 2 || source === '2' || source === '同步' || source === '客户端同步') return 2
  return Number(source) || fallback
}

function formatSource(source) {
  const sourceValue = normalizeSourceValue(source, 0)
  if (sourceValue === 1) return '新增/导入'
  if (sourceValue === 2) return '客户端同步'
  return ''
}

function formatEmpty(value, fallback = '') {
  if (value === undefined || value === null || value === '') return fallback
  return String(value)
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
  await navigator.clipboard.writeText(String(value || ''))
  showSuccessMessage('已复制')
}

function showSuccessMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message,
  })
}

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function formatDate(value) {
  if (!value) return ''
  const date = new Date(Number(value) || value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (num) => String(num).padStart(2, '0')
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
  ].join('-') + ` ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function formatListDate(value) {
  if (!value) return ''
  return formatDate(value)
}
</script>

<style scoped lang="scss">
.automa-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.page-header,
.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-header {
  justify-content: flex-end;
  flex-shrink: 0;
}

.header-actions {
  flex-wrap: wrap;
  justify-content: flex-end;
}

.workflow-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  padding: 14px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.workflow-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  flex-shrink: 0;
  gap: 12px;
  margin-bottom: 12px;
}

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
  width: 260px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.workflow-table {
  flex: 1;
  min-height: 0;
}

.workflow-name {
  display: grid;
  min-width: 0;

  span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.detail-form {
  display: grid;
  gap: 12px;
  max-height: 62vh;
  overflow: auto;
  padding-right: 4px;
}

.detail-field {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  align-items: start;
  gap: 6px;
  min-width: 0;
}

.detail-label {
  padding-top: 7px;
  color: #606266;
  font-size: 13px;
  text-align: right;
}

.detail-control {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 70px;
  gap: 6px;
  min-width: 0;
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
  .filter-item--ip {
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
