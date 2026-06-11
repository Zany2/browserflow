<template>
  <AppDialog v-model="visible" title="客户端维护" width="min(1760px, calc(100vw - 16px))">
    <div class="maintenance-dialog">
      <el-tabs v-model="activeView" @tab-change="handleViewChange">
        <el-tab-pane label="服务端工作流" name="server" />
        <el-tab-pane label="客户端工作流" name="client" />
      </el-tabs>

      <div class="maintenance-toolbar">
        <el-select v-model="selectedClientIp" class="client-ip-select" clearable filterable :loading="clientLoading"
          placeholder="选择客户端" @visible-change="handleClientSelectVisible" @change="handleClientIpChange"
          @clear="handleClientIpClear">
          <el-option v-for="clientIp in onlineClientIps" :key="clientIp" :label="clientIp" :value="clientIp" />
        </el-select>
        <el-select v-model="selectedNodeIds" class="node-select" clearable filterable multiple collapse-tags
          collapse-tags-tooltip :disabled="!selectedClientIp" placeholder="选择执行节点" @change="handleNodeChange"
          @clear="handleNodeClear">
          <el-option v-for="node in selectedClientNodes" :key="node.node_id" :label="node.node_id"
            :value="node.node_id" />
        </el-select>
        <el-input v-model="keyword" class="keyword-input" clearable :placeholder="keywordPlaceholder" />
        <el-button :disabled="!canQuery" :loading="refreshing" @click="refreshNodeWorkflows">
          刷新节点清单
        </el-button>
        <el-button @click="handleResetCurrentView">重置</el-button>
      </div>

      <div class="maintenance-actions">
        <AppSelectionSummary :count="selectedIds.length" unit="工作流" />
        <el-button v-if="activeView === 'server'" type="primary" :disabled="!canInstallOrUpdate" :loading="submitting"
          @click="handleInstall">
          安装到节点
        </el-button>
        <el-button v-if="activeView === 'server'" type="warning" :disabled="!canInstallOrUpdate" :loading="submitting"
          @click="handleUpdate">
          更新到节点
        </el-button>
        <el-button type="danger" :disabled="!canSubmit" :loading="submitting" @click="handleDelete">
          从节点删除
        </el-button>
      </div>

      <el-table ref="tableRef" v-loading="loading || serverWorkflowLoading" class="maintenance-table adaptive-table"
        :data="pagedRows" border height="100%" row-key="automa_id" empty-text="请选择客户端和执行节点"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column type="expand" width="40">
          <template #default="{ row }">
            <div class="node-detail-panel">
              <div class="node-detail-grid node-detail-grid--header">
                <span>执行节点</span>
                <span>安装状态</span>
                <span>节点状态</span>
                <span>同步状态</span>
                <span>客户端更新时间</span>
                <span>数据库更新时间</span>
              </div>
              <div v-for="detail in getNodeDetailRows(row)" :key="detail.node_id" class="node-detail-grid">
                <span class="node-detail-text" :title="detail.node_id">{{ detail.node_id }}</span>
                <span>
                  <el-tag :type="detail.installed ? 'success' : 'info'" effect="plain">
                    {{ detail.installed ? '已安装' : '未安装' }}
                  </el-tag>
                </span>
                <span>
                  <el-tag :type="getNodeDetailStatusTagType(detail)" effect="plain">
                    {{ getNodeDetailStatusText(detail) }}
                  </el-tag>
                </span>
                <span>
                  <el-tag :type="getSyncStatusTagType(detail.sync_status)" effect="plain">
                    {{ getSyncStatusText(detail.sync_status) }}
                  </el-tag>
                </span>
                <span class="node-detail-text">{{ formatOptionalDate(detail.client_updated_at) }}</span>
                <span class="node-detail-text">{{ formatOptionalDate(detail.server_updated_at) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
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
        <el-table-column label="节点覆盖" min-width="190">
          <template #default="{ row }">
            <el-tooltip placement="top" effect="dark">
              <template #content>
                <div class="node-coverage-tooltip">{{ getNodeCoverageTitle(row) }}</div>
              </template>
              <div class="node-coverage">
                <div>
                  <em>安装节点</em>
                  <span>{{ formatNodeList(getInstalledNodeIds(row)) }}</span>
                </div>
                <div>
                  <em>未安装节点</em>
                  <span>{{ formatNodeList(getMissingNodeIds(row)) }}</span>
                </div>
              </div>
            </el-tooltip>
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
        <AppPagination class="maintenance-pagination"
          :class="{ 'maintenance-pagination--hidden': filteredRows.length <= 0 }" v-model:current-page="currentPage"
          v-model:page-size="currentPageSize" :page-sizes="pageSizes" :total="filteredRows.length" />
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
import { listAllClients } from '@/services/client'
import {
  listAutomaWorkflows,
  listAutomaSyncCandidates,
  maintainClientAutomaWorkflows,
} from '@/services/automa'
import { getClientIp } from '@/utils/clientNode'
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
const selectedServerIds = computed(() => selectedRows.value.filter((row) => !hasClientOnlyNode(row)).map(getWorkflowId).filter(Boolean))
const selectedClientOnlyCount = computed(() => selectedRows.value.filter(hasClientOnlyNode).length)
const canInstallOrUpdate = computed(() => canSubmit.value && selectedClientOnlyCount.value === 0)
const serverWorkflows = computed(() => serverWorkflowRows.value.length > 0 ? serverWorkflowRows.value : props.workflows)
const keywordPlaceholder = computed(() => {
  return activeView.value === 'server' ? '服务端工作流名称、ID等' : '客户端工作流名称、ID等'
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
    loadOnlineClients(true)
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

async function loadOnlineClients(force = false) {
  if (!force && onlineClients.value.length > 0) return

  clientLoading.value = true
  try {
    const data = await listAllClients({ status: 'online' })
    const seen = new Set()
    onlineClients.value = normalizeList(data, 'clients')
      .map((client) => {
        const sourceIp = getClientIp(client)
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
  if (serverWorkflowRows.value.length > 0) return

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
    if (activeView.value === 'server' && serverWorkflowRows.value.length === 0 && !serverWorkflowLoading.value) {
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
    const nodeDetails = buildNodeDetails(candidate, workflow)
    return {
      ...workflow,
      automa_id: workflow.automa_id || workflowId,
      client_only: false,
      node_status: resolveNodeSummaryStatus(nodeDetails),
      client_updated_at: candidate.updated_at_automa || candidate.updatedAt || candidate.updated_at || '',
      node_ids: candidate.node_ids || [candidate.node_id].filter(Boolean),
      node_details: nodeDetails,
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
      node_items: [next],
    }
  }
  const nodeIds = new Set([...(current.node_ids || []), current.node_id, next.node_id].filter(Boolean))
  return {
    ...current,
    ...next,
    node_ids: Array.from(nodeIds),
    node_items: [...(current.node_items || []), next],
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
    const nodeDetails = buildNodeDetails(candidate, candidate)
    const clientOnly = nodeDetails.some((detail) => detail.client_only)
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
      node_status: clientOnly ? 'client_only' : resolveNodeSummaryStatus(nodeDetails),
      client_updated_at: candidate.updated_at_automa || candidate.updatedAt || candidate.updated_at || '',
      node_ids: candidate.node_ids || [candidate.node_id].filter(Boolean),
      node_details: nodeDetails,
      server_updated_at: candidate.server_updated_at || '',
      sync_status: candidate.sync_status || 'not_synced',
    })
  })
  return rows
}

async function loadAllNodeCandidates(refresh = false) {
  const rows = []
  let pageNum = 1
  let total = 0
  do {
    const data = await listAutomaSyncCandidates(selectedClientIp.value, {
      source_node_ids: selectedNodeIds.value,
      page_num: pageNum,
      page_size: fetchPageSize,
      refresh: refresh && pageNum === 1 ? 1 : 0,
    })
    const list = normalizeList(data, 'workflows')
    rows.push(...list)
    total = Number(data?.total || rows.length)
    pageNum += 1
    if (list.length === 0) break
  } while (rows.length < total)

  if (rows.length > 0 || selectedNodeIds.value.length <= 1) {
    return rows
  }

  for (const nodeId of selectedNodeIds.value) {
    let fallbackPageNum = 1
    let fallbackTotal = 0
    do {
      const data = await listAutomaSyncCandidates(selectedClientIp.value, {
        source_node_id: nodeId,
        page_num: fallbackPageNum,
        page_size: fetchPageSize,
        refresh: refresh && fallbackPageNum === 1 ? 1 : 0,
      })
      const list = normalizeList(data, 'workflows')
      rows.push(...list.map((item) => ({
        ...item,
        node_id: item.node_id || nodeId,
      })))
      fallbackTotal = Number(data?.total || rows.length)
      fallbackPageNum += 1
      if (list.length === 0) break
    } while (rows.filter((item) => item.node_id === nodeId).length < fallbackTotal)
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

function buildNodeDetails(candidate = {}, workflow = {}) {
  const nodeItems = Array.isArray(candidate.node_items) ? candidate.node_items : []
  const nodeMap = new Map()
  nodeItems.forEach((item) => {
    const nodeId = normalizeNodeId(selectedClientIp.value, item.node_id)
    if (!nodeId) return
    nodeMap.set(nodeId, item)
  })

  return selectedNodeIds.value.map((nodeId) => {
    const item = nodeMap.get(nodeId)
    const syncStatus = item?.sync_status || 'client_missing'
    return {
      node_id: nodeId,
      installed: Boolean(item && syncStatus !== 'client_missing'),
      sync_status: syncStatus,
      synced: Boolean(item?.synced),
      client_only: Boolean(item && (syncStatus === 'not_synced' || !item.server_id)),
      client_updated_at: item?.updated_at_automa || item?.updatedAt || item?.updated_at || '',
      server_updated_at: item?.server_updated_at || workflow?.server_updated_at || workflow?.updated_at || '',
    }
  })
}

function resolveNodeSummaryStatus(details = []) {
  if (details.length === 0) return 'missing'

  const installedCount = details.filter((detail) => detail.installed).length
  if (installedCount === 0) return 'missing'
  if (details.some((detail) => detail.client_only)) return 'client_only'
  if (installedCount < details.length) return 'partial_missing'
  if (details.every((detail) => detail.synced)) return 'synced'
  return 'different'
}

function getNodeStatusText(row) {
  if (row.node_status === 'missing') return '未安装'
  if (row.node_status === 'partial_missing') return '部分未安装'
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
  if (row.node_status === 'partial_missing') return 'warning'
  if (row.node_status === 'synced') return 'success'
  if (row.node_status === 'client_only') return 'danger'
  return 'warning'
}

function getNodeDetailRows(row) {
  return Array.isArray(row?.node_details) ? row.node_details : buildNodeDetails({}, row)
}

function hasClientOnlyNode(row) {
  return getNodeDetailRows(row).some((detail) => detail.client_only)
}

function getNodeDetailStatusText(detail) {
  if (!detail.installed) return '未安装'
  if (detail.client_only) return '节点独有'
  if (detail.synced) return '一致'
  return '有差异'
}

function getNodeDetailStatusTagType(detail) {
  if (!detail.installed) return 'info'
  if (detail.client_only) return 'danger'
  if (detail.synced) return 'success'
  return 'warning'
}

function getSyncStatusText(status) {
  if (status === 'client_missing') return '客户端缺失'
  if (status === 'not_synced') return '未同步'
  if (status === 'client_newer') return '客户端较新'
  if (status === 'server_newer') return '数据库较新'
  if (status === 'synced') return '已同步'
  if (status === 'has_update') return '有更新'
  return ''
}

function getSyncStatusTagType(status) {
  if (status === 'synced') return 'success'
  if (status === 'server_newer') return 'danger'
  if (status === 'client_newer' || status === 'has_update') return 'warning'
  return 'info'
}

function getNodeCoverageTitle(row) {
  const installed = getInstalledNodeIds(row)
  const missing = getMissingNodeIds(row)
  return `安装节点：${formatNodeList(installed)}\n未安装节点：${formatNodeList(missing)}`
}

function getWorkflowId(row) {
  return row?.automa_id || row?.workflow_id || row?.workflowId || row?.id || ''
}

function getInstalledNodeIds(row) {
  return getNodeDetailRows(row)
    .filter((detail) => detail.installed)
    .map((detail) => detail.node_id)
}

function getMissingNodeIds(row) {
  return getNodeDetailRows(row)
    .filter((detail) => !detail.installed)
    .map((detail) => detail.node_id)
}

function formatNodeList(nodeIds) {
  return nodeIds.length > 0 ? nodeIds.join('、') : ''
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
  height: 608px;
  min-height: 0;
  gap: 14px;
}

.maintenance-toolbar,
.maintenance-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.client-ip-select {
  width: 220px;
}

.node-select {
  width: 180px;
}

.keyword-input {
  width: 220px;
}

.maintenance-actions {
  justify-content: flex-end;
  min-height: 32px;
}

.maintenance-table {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
}

.maintenance-table :deep(.nowrap-column .cell) {
  padding-inline: 8px;
  white-space: nowrap;
}

.node-detail-panel {
  display: grid;
  gap: 6px;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid #ebeef5;
}

.node-detail-grid {
  display: grid;
  grid-template-columns: minmax(120px, 1.1fr) 88px 96px 96px 150px 150px;
  align-items: center;
  gap: 10px;
  min-width: 0;
  color: #303133;
  font-size: 13px;
  line-height: 24px;
}

.node-detail-grid--header {
  color: #909399;
  font-weight: 600;
}

.node-detail-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-coverage {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.node-coverage div {
  display: grid;
  grid-template-columns: 70px minmax(0, 1fr);
  align-items: center;
  gap: 4px;
  min-width: 0;
  line-height: 18px;
}

.node-coverage em {
  color: #909399;
  font-style: normal;
  white-space: nowrap;
}

.node-coverage span {
  min-width: 0;
  overflow: hidden;
  color: #303133;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-coverage-tooltip {
  white-space: pre-line;
}

.maintenance-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  min-height: 32px;
}

.maintenance-footer :deep(.app-pagination) {
  justify-content: flex-end;
  margin-top: 0;
}

.maintenance-pagination--hidden {
  visibility: hidden;
  pointer-events: none;
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
