<template>
  <AppDialog
    v-model="visible"
    title="客户端维护"
    width="min(1560px, calc(100vw - 16px))"
  >
    <div class="maintenance-dialog">
      <el-tabs v-model="activeView" @tab-change="handleViewChange">
        <el-tab-pane label="服务端工作流" name="server" />
        <el-tab-pane label="客户端工作流" name="client" />
      </el-tabs>

      <div class="maintenance-toolbar">
        <el-select
          v-model="selectedClientIp"
          class="client-ip-select"
          clearable
          filterable
          :loading="clientLoading"
          placeholder="选择客户端 IP"
          @visible-change="handleClientSelectVisible"
          @change="handleClientIpChange"
          @clear="handleClientIpClear"
        >
          <el-option
            v-for="clientIp in onlineClientIps"
            :key="clientIp"
            :label="clientIp"
            :value="clientIp"
          />
        </el-select>
        <el-select
          v-model="selectedNodeIds"
          class="node-select"
          clearable
          filterable
          multiple
          collapse-tags
          collapse-tags-tooltip
          :disabled="!selectedClientIp"
          placeholder="选择执行节点"
          @change="handleNodeChange"
          @clear="handleNodeClear"
        >
          <el-option
            v-for="node in selectedClientNodes"
            :key="node.node_id"
            :label="node.node_id"
            :value="node.node_id"
          />
        </el-select>
        <el-input
          v-model="keyword"
          class="keyword-input"
          clearable
          :placeholder="keywordPlaceholder"
        />
        <el-button :disabled="!canQuery" :loading="refreshing" @click="refreshNodeWorkflows">
          刷新节点清单
        </el-button>
        <el-button @click="handleResetCurrentView">重置</el-button>
      </div>

      <div class="maintenance-actions">
        <AppSelectionSummary :count="selectedIds.length" unit="工作流" />
        <el-button
          v-if="activeView === 'server'"
          type="primary"
          :disabled="!canSubmit"
          :loading="submitting"
          @click="handleInstall"
        >
          安装到节点
        </el-button>
        <el-button
          v-if="activeView === 'server'"
          type="warning"
          :disabled="!canSubmit"
          :loading="submitting"
          @click="handleUpdate"
        >
          更新到节点
        </el-button>
        <el-button
          type="danger"
          :disabled="!canSubmit"
          :loading="submitting"
          @click="handleDelete"
        >
          从节点删除
        </el-button>
      </div>

      <el-table
        ref="tableRef"
        v-loading="loading || serverWorkflowLoading"
        class="maintenance-table adaptive-table"
        :data="pagedRows"
        border
        height="460"
        row-key="automa_id"
        empty-text="请选择客户端 IP 和执行节点"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column :label="nameColumnLabel" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.name || row.server_name || '' }}
          </template>
        </el-table-column>
        <el-table-column label="Automa 工作流名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.automa_name || row.automa_id || '' }}
          </template>
        </el-table-column>
        <el-table-column label="安装节点数" width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <span :title="getNodeCoverageTitle(row)">{{ getNodeCoverageText(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="数据来源" width="98" align="center">
          <template #default="{ row }">
            <el-tag :type="row.client_only ? 'warning' : 'success'" effect="plain">
              {{ getDataSourceText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="节点状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="getNodeStatusTagType(row)" effect="plain">
              {{ getNodeStatusText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="客户端更新时间" width="176" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatOptionalDate(row.client_updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="数据库更新时间" width="176" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatOptionalDate(row.server_updated_at || row.updated_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="maintenance-footer">
        <AppPagination
          v-if="filteredRows.length > 0"
          v-model:current-page="currentPage"
          v-model:page-size="currentPageSize"
          :page-sizes="pageSizes"
          :total="filteredRows.length"
        />
      </div>
    </div>

    <template #footer>
      <span class="maintenance-empty-footer" />
    </template>
  </AppDialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { listClients } from '@/services/client'
import {
  listAutomaWorkflows,
  listAutomaSyncCandidates,
  maintainClientAutomaWorkflows,
} from '@/services/automa'
import { formatDate } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, normalizeList, normalizeText } from '@/utils/list'

const props = defineProps({
  workflows: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['changed'])

const visible = defineModel({
  type: Boolean,
  default: false,
})

const tableRef = ref(null)
const activeView = ref('server')
const selectedClientIp = ref('')
const selectedNodeIds = ref([])
const keyword = ref('')
const onlineClients = ref([])
const serverWorkflowRows = ref([])
const candidateRows = ref([])
const clientLoading = ref(false)
const loading = ref(false)
const refreshing = ref(false)
const submitting = ref(false)
const serverWorkflowLoading = ref(false)
const fetchPageSize = 60
const currentPage = ref(1)
const currentPageSize = ref(10)
const pageSizes = DEFAULT_PAGE_SIZES
const viewStateCache = {
  server: createViewState(),
  client: createViewState(),
}

const onlineClientIps = computed(() => {
  const seen = new Set()
  const ips = []
  onlineClients.value.forEach((client) => {
    if (!client.source_ip || seen.has(client.source_ip)) return
    seen.add(client.source_ip)
    ips.push(client.source_ip)
  })
  return ips
})
const selectedClientNodes = computed(() => {
  const clientIp = normalizeText(selectedClientIp.value)
  if (!clientIp) return []
  return onlineClients.value.filter((client) => client.source_ip === clientIp && client.node_id)
})
const canQuery = computed(() => Boolean(normalizeText(selectedClientIp.value) && selectedNodeIds.value.length > 0))
const canSubmit = computed(() => canQuery.value && selectedIds.value.length > 0)
const selectedNodeCount = computed(() => selectedNodeIds.value.length)
const selectedServerIds = computed(() => selectedRows.value.filter((row) => !row.client_only).map(getWorkflowId).filter(Boolean))
const selectedClientOnlyCount = computed(() => selectedRows.value.filter((row) => row.client_only).length)
const serverWorkflows = computed(() => serverWorkflowRows.value.length > 0 ? serverWorkflowRows.value : props.workflows)
const keywordPlaceholder = computed(() => {
  return activeView.value === 'server' ? '检索服务端工作流名称 / ID' : '检索客户端工作流名称 / ID'
})
const nameColumnLabel = computed(() => activeView.value === 'server' ? '自定义工作流名称' : '工作流名称')
const filteredRows = computed(() => {
  const text = normalizeText(keyword.value).toLowerCase()
  if (!text) return candidateRows.value
  return candidateRows.value.filter((row) => {
    return [
      row.automa_id,
      row.name,
      row.description,
      row.automa_name,
      row.automa_description,
    ].some((value) => normalizeText(value).toLowerCase().includes(text))
  })
})
const pagedRows = computed(() => {
  const start = (currentPage.value - 1) * currentPageSize.value
  return filteredRows.value.slice(start, start + currentPageSize.value)
})
const {
  selectedRows,
  selectedKeys: selectedIds,
  handleSelectionChange,
  restoreSelection,
  resetSelection,
  retainSelectionByRows,
} = usePagedTableSelection({
  rows: pagedRows,
  getRowKey: getWorkflowId,
})

watch(visible, (nextVisible) => {
  if (nextVisible) {
    loadOnlineClients()
    loadServerWorkflows()
    return
  }
  resetDialog()
})

watch(keyword, () => {
  currentPage.value = 1
  retainSelectionByRows(filteredRows.value)
  restoreSelection(tableRef)
  saveViewState()
})

watch(currentPage, () => {
  restoreSelection(tableRef)
  saveViewState()
})

watch(currentPageSize, () => {
  currentPage.value = 1
  restoreSelection(tableRef)
  saveViewState()
})

watch(filteredRows, (rows) => {
  const totalPages = Math.max(Math.ceil(rows.length / currentPageSize.value), 1)
  if (currentPage.value > totalPages) {
    currentPage.value = totalPages
  }
  retainSelectionByRows(rows)
  restoreSelection(tableRef)
})

async function loadOnlineClients() {
  clientLoading.value = true
  try {
    const data = await listClients({ status: 'online' })
    const seen = new Set()
    onlineClients.value = normalizeList(data, 'clients')
      .map((client) => {
        const sourceIp = client?.client_ip || client?.ip || client?.remote_ip || client?.last_ip || client?.source_ip || ''
        const nodeId = normalizeNodeId(sourceIp, client?.node_id || client?.nodeId)
        const identity = buildNodeIdentity(sourceIp, nodeId)
        return {
          identity,
          source_ip: sourceIp,
          node_id: nodeId,
        }
      })
      .filter((client) => client.identity && client.source_ip && client.node_id)
      .filter((client) => {
        if (seen.has(client.identity)) return false
        seen.add(client.identity)
        return true
      })
  } finally {
    clientLoading.value = false
  }
}

async function loadServerWorkflows() {
  serverWorkflowLoading.value = true
  try {
    const rows = []
    let pageNum = 1
    let total = 0
    do {
      const data = await listAutomaWorkflows({
        page_num: pageNum,
        page_size: fetchPageSize,
      })
      const list = normalizeList(data, 'workflows')
      rows.push(...list)
      total = Number(data?.total || rows.length)
      pageNum += 1
      if (list.length === 0) break
    } while (rows.length < total)
    serverWorkflowRows.value = rows
  } finally {
    serverWorkflowLoading.value = false
  }
}

function handleClientSelectVisible(opened) {
  if (opened) loadOnlineClients()
}

function handleClientIpChange(value) {
  selectedClientIp.value = normalizeText(value)
  selectedNodeIds.value = []
  currentPage.value = 1
  resetCandidates()
  saveViewState()
}

function handleClientIpClear() {
  selectedClientIp.value = ''
  selectedNodeIds.value = []
  currentPage.value = 1
  resetCandidates()
  saveViewState()
}

function handleNodeChange(value) {
  selectedNodeIds.value = normalizeNodeIdList(value)
  currentPage.value = 1
  saveViewState()
  loadNodeCandidates()
}

function handleNodeClear() {
  selectedNodeIds.value = []
  currentPage.value = 1
  resetCandidates()
  saveViewState()
}

async function refreshNodeWorkflows() {
  refreshing.value = true
  try {
    await loadNodeCandidates(true)
    showSuccessMessage('节点清单已刷新')
  } finally {
    refreshing.value = false
  }
}

function handleViewChange() {
  const nextView = activeView.value
  const previousView = nextView === 'server' ? 'client' : 'server'
  saveViewState(previousView)
  resetCandidates()
  restoreViewState(nextView)
  currentPage.value = 1
  if (canQuery.value) loadNodeCandidates(false)
}

function handleResetCurrentView() {
  viewStateCache[activeView.value] = createViewState()
  restoreViewState(activeView.value)
  resetCandidates()
}

async function loadNodeCandidates(refresh = false) {
  if (!canQuery.value) {
    resetCandidates()
    return
  }
  loading.value = true
  try {
    if (serverWorkflowRows.value.length === 0 && !serverWorkflowLoading.value) {
      await loadServerWorkflows()
    }
    const candidates = await loadAllNodeCandidates(refresh)
    const candidateMap = new Map()
    candidates.forEach((item) => {
      const workflowId = item.automa_id || item.workflow_id || item.id
      if (!workflowId) return
      candidateMap.set(workflowId, mergeCandidate(candidateMap.get(workflowId), item))
    })
    candidateRows.value = activeView.value === 'server'
      ? buildServerRows(candidateMap)
      : buildClientRows(candidateMap)
    currentPage.value = 1
    clearSelection()
    await restoreSelection(tableRef)
  } finally {
    loading.value = false
  }
}

function buildServerRows(candidateMap) {
  return serverWorkflows.value.map((workflow) => {
    const workflowId = getWorkflowId(workflow)
    const candidate = candidateMap.get(workflowId) || {}
    return {
      ...workflow,
      automa_id: workflow.automa_id || workflowId,
      client_only: false,
      node_status: resolveNodeStatus(candidate),
      client_updated_at: candidate.updated_at_automa || candidate.updatedAt || candidate.updated_at || '',
      node_ids: candidate.node_ids || [candidate.node_id].filter(Boolean),
      server_updated_at: workflow.updated_at,
      sync_status: candidate.sync_status || 'client_missing',
    }
  })
}

function mergeCandidate(current = null, next = {}) {
  if (!current) {
    return {
      ...next,
      node_ids: [next.node_id].filter(Boolean),
    }
  }
  const nodeIds = new Set([...(current.node_ids || []), current.node_id, next.node_id].filter(Boolean))
  return {
    ...current,
    ...next,
    node_ids: Array.from(nodeIds),
    sync_status: mergeSyncStatus(current.sync_status, next.sync_status),
    synced: Boolean(current.synced && next.synced),
    updated_at_automa: Math.max(Number(current.updated_at_automa || 0), Number(next.updated_at_automa || 0)),
    updatedAt: Math.max(Number(current.updatedAt || 0), Number(next.updatedAt || 0)),
  }
}

function mergeSyncStatus(current = '', next = '') {
  const priority = ['client_missing', 'server_newer', 'client_newer', 'has_update', 'not_synced', 'synced']
  const currentIndex = priority.indexOf(current)
  const nextIndex = priority.indexOf(next)
  if (currentIndex < 0) return next || current
  if (nextIndex < 0) return current || next
  return priority[Math.min(currentIndex, nextIndex)]
}

function buildClientRows(candidateMap) {
  const rows = []
  candidateMap.forEach((candidate, workflowId) => {
    const clientOnly = candidate.sync_status === 'not_synced' || !candidate.server_id
    rows.push({
      id: candidate.server_id || '',
      automa_id: workflowId,
      name: candidate.server_name || candidate.name || candidate.automa_name || workflowId,
      description: candidate.server_description || candidate.description || candidate.automa_description || '',
      server_name: candidate.server_name || '',
      server_description: candidate.server_description || '',
      automa_name: candidate.automa_name || candidate.name || workflowId,
      automa_description: candidate.automa_description || candidate.description || '',
      client_only: clientOnly,
      node_status: clientOnly ? 'client_only' : resolveNodeStatus(candidate),
      client_updated_at: candidate.updated_at_automa || candidate.updatedAt || candidate.updated_at || '',
      node_ids: candidate.node_ids || [candidate.node_id].filter(Boolean),
      server_updated_at: candidate.server_updated_at || '',
      sync_status: candidate.sync_status || 'not_synced',
    })
  })
  return rows
}

async function loadAllNodeCandidates(refresh = false) {
  const rows = []
  for (const nodeId of selectedNodeIds.value) {
    let pageNum = 1
    let total = 0
    do {
      const data = await listAutomaSyncCandidates(selectedClientIp.value, {
        source_node_id: nodeId,
        page_num: pageNum,
        page_size: fetchPageSize,
        refresh: refresh && pageNum === 1 ? 1 : 0,
      })
      const list = normalizeList(data, 'workflows')
      rows.push(...list.map((item) => ({
        ...item,
        node_id: item.node_id || nodeId,
      })))
      total = Number(data?.total || rows.length)
      pageNum += 1
      if (list.length === 0) break
    } while (rows.filter((item) => item.node_id === nodeId).length < total)
  }
  return rows
}

async function handleInstall() {
  await submitMaintenance('install', '安装')
}

async function handleUpdate() {
  await submitMaintenance('update', '更新')
}

async function handleDelete() {
  const confirmed = await appConfirm({
    title: '从节点删除工作流',
    message: `确认从选中的 ${selectedNodeCount.value} 个执行节点删除选中的 ${selectedIds.value.length} 个工作流吗？服务端数据库不会删除。`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return
  await submitMaintenance('delete', '删除')
}

async function submitMaintenance(action, label) {
  if (!canSubmit.value) return
  if ((action === 'install' || action === 'update') && selectedClientOnlyCount.value > 0) {
    appMessage({
      type: APP_MESSAGE_TYPE.warning,
      message: '节点独有工作流没有服务端数据，不能安装或更新',
    })
    return
  }
  submitting.value = true
  try {
    const result = await maintainClientAutomaWorkflows({
      source_ip: selectedClientIp.value,
      source_node_id: selectedNodeIds.value[0] || '',
      source_node_ids: selectedNodeIds.value,
      action,
      workflow_ids: action === 'delete' ? selectedIds.value : selectedServerIds.value,
      refresh: true,
    })
    showSuccessMessage(result?.message || `${label}完成`)
    await loadNodeCandidates(false)
    emit('changed')
  } finally {
    submitting.value = false
  }
}

function clearSelection() {
  resetSelection(tableRef)
}

function resetCandidates() {
  candidateRows.value = []
  clearSelection()
}

function resetDialog() {
  resetViewStateCache()
  activeView.value = 'server'
  restoreViewState('server')
  serverWorkflowRows.value = []
  resetCandidates()
}

function createViewState() {
  return {
    selectedClientIp: '',
    selectedNodeIds: [],
    keyword: '',
    currentPage: 1,
    pageSize: 10,
  }
}

function saveViewState(view = activeView.value) {
  if (!viewStateCache[view]) return

  viewStateCache[view] = {
    selectedClientIp: selectedClientIp.value,
    selectedNodeIds: selectedNodeIds.value.slice(),
    keyword: keyword.value,
    currentPage: currentPage.value,
    pageSize: currentPageSize.value,
  }
}

function restoreViewState(view = activeView.value) {
  const state = viewStateCache[view] || createViewState()
  selectedClientIp.value = state.selectedClientIp
  selectedNodeIds.value = normalizeNodeIdList(state.selectedNodeIds)
  keyword.value = state.keyword
  currentPage.value = state.currentPage || 1
  currentPageSize.value = state.pageSize || 10
}

function resetViewStateCache() {
  viewStateCache.server = createViewState()
  viewStateCache.client = createViewState()
}

function resolveNodeStatus(candidate = {}) {
  if (candidate.sync_status === 'client_missing' || !candidate.automa_id) return 'missing'
  if (candidate.synced) return 'synced'
  return 'different'
}

function getNodeStatusText(row) {
  if (row.node_status === 'missing') return '未安装'
  if (row.node_status === 'synced') return '一致'
  if (row.node_status === 'client_only') return '节点独有'
  return '有差异'
}

function getDataSourceText(row) {
  if (activeView.value === 'client') return row.client_only ? '仅客户端' : '两端都有'
  return '服务端'
}

function getNodeStatusTagType(row) {
  if (row.node_status === 'missing') return 'info'
  if (row.node_status === 'synced') return 'success'
  if (row.node_status === 'client_only') return 'danger'
  return 'warning'
}

function getNodeCoverageText(row) {
  const nodeIds = getRowNodeIds(row)
  const total = selectedNodeCount.value
  return `${nodeIds.length} / ${total}`
}

function getNodeCoverageTitle(row) {
  const nodeIds = getRowNodeIds(row)
  if (nodeIds.length === 0) return row.node_status === 'missing' ? '所选节点均未安装' : ''
  return nodeIds.join('、')
}

function getWorkflowId(row) {
  return row?.automa_id || row?.workflow_id || row?.workflowId || row?.id || ''
}

function getRowNodeIds(row) {
  const values = Array.isArray(row?.node_ids) ? row.node_ids : [row?.node_id]
  const seen = new Set()
  const result = []
  values.forEach((nodeId) => {
    const normalized = normalizeNodeId(selectedClientIp.value, nodeId)
    if (!normalized || seen.has(normalized)) return
    seen.add(normalized)
    result.push(normalized)
  })
  return result
}

function buildNodeIdentity(clientIp, nodeId) {
  clientIp = normalizeText(clientIp)
  nodeId = normalizeNodeId(clientIp, nodeId)
  if (!clientIp) return nodeId
  if (!nodeId || nodeId === clientIp) return clientIp
  return `${clientIp}|${nodeId}`
}

function normalizeNodeId(clientIp, nodeId) {
  clientIp = normalizeText(clientIp)
  nodeId = normalizeText(nodeId)
  if (!clientIp || !nodeId) return nodeId

  const prefix = `${clientIp}|`
  while (nodeId.startsWith(prefix)) {
    nodeId = nodeId.slice(prefix.length)
  }
  return nodeId
}

function normalizeNodeIdList(values) {
  const nodeIds = Array.isArray(values) ? values : [values]
  const seen = new Set()
  const result = []
  nodeIds.forEach((nodeId) => {
    const normalized = normalizeNodeId(selectedClientIp.value, nodeId)
    if (!normalized || seen.has(normalized)) return
    seen.add(normalized)
    result.push(normalized)
  })
  return result
}

function formatOptionalDate(value) {
  if (!value) return ''
  return formatDate(value)
}

function showSuccessMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message,
  })
}
</script>

<style scoped lang="scss">
.maintenance-dialog {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.maintenance-toolbar,
.maintenance-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.client-ip-select {
  width: 220px;
}

.node-select {
  width: 260px;
}

.keyword-input {
  width: 260px;
}

.maintenance-actions {
  justify-content: flex-end;
}

.maintenance-table {
  width: 100%;
}

.maintenance-table :deep(.nowrap-column .cell) {
  white-space: nowrap;
}

.maintenance-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
}

.maintenance-footer :deep(.app-pagination) {
  justify-content: flex-end;
  margin-top: 0;
}

:global(.app-dialog:has(.maintenance-empty-footer) .el-dialog__footer) {
  display: none;
}

@media (max-width: 720px) {
  .maintenance-toolbar,
  .maintenance-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .client-ip-select,
  .node-select,
  .keyword-input {
    width: 100%;
  }

  .maintenance-footer,
  .maintenance-footer :deep(.app-pagination) {
    justify-content: center;
  }
}
</style>
