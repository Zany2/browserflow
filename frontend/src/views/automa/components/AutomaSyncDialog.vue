<template>
  <AppDialog v-model="visible" title="客户端同步" width="min(1760px, calc(100vw - 16px))" confirm-text="同步选中"
    :confirm-disabled="selectedIds.length === 0" :loading="syncing" @confirm="handleSync">
    <div class="sync-dialog">
      <el-tabs v-model="activeMode" @tab-change="handleModeChange">
        <el-tab-pane label="按客户端" name="client" />
        <el-tab-pane label="按工作流" name="workflow" />
      </el-tabs>

      <div class="sync-toolbar">
        <template v-if="activeMode === 'client'">
          <el-select v-model="selectedClientIp" class="client-ip-select" clearable filterable :loading="clientLoading"
            placeholder="选择客户端" @visible-change="handleClientSelectVisible" @change="handleClientIpChange"
            @clear="handleClientIpClear">
            <el-option v-for="clientIp in onlineClientIps" :key="clientIp" :label="clientIp" :value="clientIp" />
          </el-select>
          <el-select v-model="selectedNodeId" class="node-select" clearable filterable :disabled="!selectedClientIp"
            placeholder="全部执行节点" @change="handleNodeChange" @clear="handleNodeClear">
            <el-option label="全部执行节点" value="" />
            <el-option v-for="node in selectedClientNodes" :key="node.node_id" :label="node.node_id"
              :value="node.node_id" />
          </el-select>
        </template>
        <template v-else>
          <el-select v-model="selectedAutomaId" class="query-input" clearable filterable
            remote reserve-keyword :remote-method="handleWorkflowRemoteSearch" :loading="workflowOptionLoading"
            placeholder="选择工作流" @visible-change="handleWorkflowSelectVisible" @change="handleWorkflowChange"
            @clear="handleWorkflowClear">
            <el-option v-for="workflow in workflowSelectOptions" :key="getWorkflowId(workflow)"
              :label="workflow.name || workflow.automa_name || getWorkflowId(workflow)"
              :value="workflow.automa_id || getWorkflowId(workflow)">
              <div class="workflow-option">
                <span>
                  <em>自定义工作流名称</em>
                  {{ workflow.name || '' }}
                </span>
                <small>
                  <em>Automa 工作流名称</em>
                  {{ workflow.automa_name || workflow.automa_id || getWorkflowId(workflow) }}
                </small>
              </div>
            </el-option>
          </el-select>
          <el-select v-model="selectedClientIp" class="client-ip-select" clearable filterable :loading="clientLoading"
            placeholder="全部客户端" @visible-change="handleClientSelectVisible" @change="handleClientIpChange"
            @clear="handleClientIpClear">
            <el-option label="全部客户端" value="" />
            <el-option v-for="clientIp in onlineClientIps" :key="clientIp" :label="clientIp" :value="clientIp" />
          </el-select>
          <el-select v-model="selectedNodeId" class="node-select" clearable filterable :disabled="!selectedClientIp"
            placeholder="全部执行节点" @change="handleNodeChange" @clear="handleNodeClear">
            <el-option label="全部执行节点" value="" />
            <el-option v-for="node in selectedClientNodes" :key="node.node_id" :label="node.node_id"
              :value="node.node_id" />
          </el-select>
        </template>

        <el-select v-model="syncStatusFilter" class="status-select" clearable placeholder="全部同步状态"
          @change="handleCandidateFilterChange" @clear="handleCandidateFilterClear">
          <el-option label="全部同步状态" value="" />
          <el-option label="可同步" value="syncable" />
          <el-option label="已同步" value="synced" />
          <el-option label="未同步" value="not_synced" />
          <el-option label="客户端缺失" value="client_missing" />
          <el-option label="客户端较新" value="client_newer" />
          <el-option label="数据库较新" value="server_newer" />
        </el-select>
        <el-input v-model="keyword" class="keyword-input" clearable :placeholder="keywordPlaceholder" />
        <el-button @click="handleResetCurrentMode">重置</el-button>
      </div>

      <div class="sync-scope">
        {{ scopeText ? `当前展示：${scopeText}` : '' }}
      </div>

      <el-table ref="tableRef" v-loading="candidateLoading" class="candidate-table adaptive-table" :data="candidates"
        border height="100%" row-key="row_key" header-align="left" empty-text="请选择查询条件后自动加载"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" reserve-selection :selectable="isSelectable" />

        <el-table-column v-if="activeMode === 'workflow'" label="客户端/执行节点" width="164" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatClientNode(row) }}
          </template>
        </el-table-column>

        <el-table-column v-else label="执行节点" width="116" class-name="nowrap-column">
          <template #default="{ row }">
            {{ row.node_id || '' }}
          </template>
        </el-table-column>

        <el-table-column label="自定义工作流名称" min-width="104" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.server_name || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="自定义工作流描述" min-width="108" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="field-value">{{ row.server_description || '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="工作流名称" min-width="150">
          <template #default="{ row }">
            <span class="compare-line" :title="formatCompareText(row.automa_name || row.name, row.server_automa_name)">
              <span class="compare-item">
                <em>客户端</em>
                <span>{{ row.automa_name || row.name || '' }}</span>
              </span>
              <span class="compare-item">
                <em>数据库</em>
                <span>{{ row.server_automa_name || '' }}</span>
              </span>
            </span>
          </template>
        </el-table-column>

        <el-table-column label="工作流描述" min-width="150">
          <template #default="{ row }">
            <span class="compare-line"
              :title="formatCompareText(row.automa_description || row.description, row.server_automa_description)">
              <span class="compare-item">
                <em>客户端</em>
                <span>{{ row.automa_description || row.description || '' }}</span>
              </span>
              <span class="compare-item">
                <em>数据库</em>
                <span>{{ row.server_automa_description || '' }}</span>
              </span>
            </span>
          </template>
        </el-table-column>

        <el-table-column label="同步状态" width="104" header-align="center">
          <template #default="{ row }">
            <div class="center-cell">
              <el-tag :type="getSyncTagType(row)" effect="plain">
                {{ getSyncText(row) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="工作流状态" width="116" header-align="center">
          <template #default="{ row }">
            <div class="center-cell">
              <el-tag class="workflow-status-tag" :type="getWorkflowTagType(row)" effect="plain">
                {{ getWorkflowText(row) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="客户端更新时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            <span class="time-value"
              :title="formatOptionalDate(row.updated_at_automa || row.updatedAt || row.updated_at)">
              {{ formatOptionalDate(row.updated_at_automa || row.updatedAt || row.updated_at) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="同步时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            <span class="time-value" :title="formatOptionalDate(row.last_synced_at)">
              {{ formatOptionalDate(row.last_synced_at) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="数据库更新时间" width="168" class-name="nowrap-column">
          <template #default="{ row }">
            <span class="time-value" :title="formatOptionalDate(row.server_updated_at)">
              {{ formatOptionalDate(row.server_updated_at) }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <div class="sync-footer">
        <div class="sync-summary">
          <AppSelectionSummary :count="selectedIds.length" />
          <span>可同步 {{ selectableCount }} 个</span>
        </div>

        <AppPagination class="sync-pagination" :class="{ 'sync-pagination--hidden': candidateTotal <= 0 }"
          v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="pageSizes"
          :total="candidateTotal" layout="total, sizes, prev, pager, next" />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { listAllClients } from '@/services/client'
import {
  listAutomaWorkflows,
  listAutomaSyncCandidates,
  listAutomaSyncCandidatesByWorkflow,
  syncAutomaWorkflowsByIp,
} from '@/services/automa'
import { getClientIp } from '@/utils/clientNode'
import { formatDate } from '@/utils/format'
import { DEFAULT_PAGE_SIZES, normalizeList, normalizeText } from '@/utils/list'

defineProps({
  workflows: {
    type: Array,
    default: () => [],
  },
})

const visible = defineModel({
  type: Boolean,
  default: false,
})

const emit = defineEmits(['synced'])

const tableRef = ref(null)
const activeMode = ref('client')
const selectedClientIp = ref('')
const selectedNodeId = ref('')
const selectedAutomaId = ref('')
const keyword = ref('')
const syncStatusFilter = ref('')
const candidates = ref([])
const candidateLoading = ref(false)
const clientLoading = ref(false)
const workflowOptionLoading = ref(false)
const onlineClients = ref([])
const onlineWorkflowOptions = ref([])
const workflowOptionKeyword = ref('')
const syncing = ref(false)
const refreshCandidates = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = DEFAULT_PAGE_SIZES
const candidateTotal = ref(0)
const lastActiveMode = ref('client')
const restoringModeState = ref(false)
const modeStateCache = {
  client: createModeState(),
  workflow: createModeState(),
}
let candidateRequestSeq = 0
let workflowOptionRequestSeq = 0
const {
  selectedRows,
  selectedKeys: selectedIds,
  handleSelectionChange,
  restoreSelection,
  resetSelection,
} = usePagedTableSelection({
  rows: candidates,
  getRowKey: getSelectionKey,
  isSelectable,
})
const {
  run: runKeywordSearch,
  cancel: clearKeywordSearchTimer,
} = useDebouncedAction(loadFirstCandidatePage, 200)
const {
  run: runWorkflowOptionSearch,
  cancel: clearWorkflowOptionSearchTimer,
} = useDebouncedAction(loadOnlineWorkflowOptions, 200)

const selectableCount = computed(() => candidates.value.filter(isSelectable).length)
const workflowSelectOptions = computed(() => onlineWorkflowOptions.value)
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
const canLoad = computed(() => {
  return activeMode.value === 'client' ? Boolean(normalizeText(selectedClientIp.value)) : Boolean(selectedAutomaId.value)
})
const keywordPlaceholder = computed(() => {
  return activeMode.value === 'client' ? '客户端名称、描述、Automa ID等' : '客户端、执行节点等'
})
const scopeText = computed(() => {
  if (activeMode.value === 'client') {
    const clientIp = normalizeText(selectedClientIp.value)
    if (!clientIp) return ''
    const nodeId = normalizeText(selectedNodeId.value)
    return nodeId ? `${clientIp} / ${nodeId}` : `${clientIp} 的全部执行节点`
  }

  if (!selectedAutomaId.value) return ''
  const clientIp = normalizeText(selectedClientIp.value)
  const nodeId = normalizeText(selectedNodeId.value)
  if (!clientIp) return '全部客户端'
  return nodeId ? `${clientIp} / ${nodeId}` : `${clientIp} 的全部执行节点`
})

watch(visible, (nextVisible) => {
  if (nextVisible) {
    lastActiveMode.value = activeMode.value
    loadOnlineClientIps(true)
    return
  }

  clearKeywordSearchTimer()
  clearWorkflowOptionSearchTimer()
  resetModeStateCache()
  activeMode.value = 'client'
  lastActiveMode.value = 'client'
  selectedClientIp.value = ''
  selectedNodeId.value = ''
  selectedAutomaId.value = ''
  keyword.value = ''
  syncStatusFilter.value = ''
  refreshCandidates.value = false
  onlineWorkflowOptions.value = []
  workflowOptionKeyword.value = ''
  resetCandidatePage()
  resetCandidates()
})

watch(keyword, () => {
  if (restoringModeState.value) return
  resetSelection(tableRef)
  scheduleKeywordSearch()
})

watch(currentPage, () => {
  if (restoringModeState.value) return
  if (!visible.value || !canLoad.value) return
  loadCandidates()
})

watch(pageSize, () => {
  if (restoringModeState.value) return
  if (!visible.value || !canLoad.value) return
  loadFirstCandidatePage()
})

async function loadCandidates() {
  if (!canLoad.value) {
    showWarningMessage(activeMode.value === 'client' ? '请先选择在线客户端' : '请先选择工作流')
    return
  }

  const requestSeq = ++candidateRequestSeq
  const shouldRefresh = refreshCandidates.value
  refreshCandidates.value = false
  candidateLoading.value = true
  try {
    const selectedClient = getSelectedClient()
    const params = {
      keyword: keyword.value.trim(),
      sync_status: syncStatusFilter.value,
      source_node_id: selectedClient.node_id,
      page_num: currentPage.value,
      page_size: pageSize.value,
      refresh: shouldRefresh ? 1 : 0,
    }
    const data =
      activeMode.value === 'client'
        ? await listAutomaSyncCandidates(selectedClient.source_ip, params)
        : await listAutomaSyncCandidatesByWorkflow(selectedAutomaId.value, {
          ...params,
          source_ip: selectedClient.source_ip,
        })
    if (requestSeq !== candidateRequestSeq) return

    const candidateList = normalizeList(data, 'workflows')
    candidateTotal.value = Number(data?.total ?? candidateList.length)
    candidates.value = candidateList.map((item) => ({
      ...item,
      row_key: `${buildNodeIdentity(item.source_ip || selectedClient.source_ip, item.node_id || selectedClient.node_id)}_${getWorkflowId(item)}`,
    }))
    await restoreSelection(tableRef)
  } finally {
    if (requestSeq === candidateRequestSeq) {
      candidateLoading.value = false
    }
  }
}

async function loadOnlineClientIps(force = false) {
  if (!force && onlineClients.value.length > 0) return

  clientLoading.value = true
  try {
    const data = await listAllClients({
      status: 'online',
    })
    const seen = new Set()
    onlineClients.value = normalizeList(data, 'clients')
      .map((client) => {
        const clientIp = getClientIp(client)
        const nodeId = normalizeNodeId(clientIp, client?.node_id || client?.nodeId)
        const identity = buildNodeIdentity(clientIp, nodeId)
        return {
          key: identity,
          identity,
          source_ip: clientIp,
          node_id: nodeId,
          label: nodeId ? `${clientIp} / ${nodeId}` : clientIp,
        }
      })
      .filter((client) => client.identity)
      .filter((client) => {
        if (seen.has(client.identity)) return false
        seen.add(client.identity)
        return true
      })
    if (selectedClientIp.value && !onlineClientIps.value.includes(selectedClientIp.value)) {
      selectedClientIp.value = ''
      selectedNodeId.value = ''
      resetCandidates()
    }
    if (selectedNodeId.value && !selectedClientNodes.value.some((node) => node.node_id === selectedNodeId.value)) {
      selectedNodeId.value = ''
    }
  } finally {
    clientLoading.value = false
  }
}

function handleClientSelectVisible(opened) {
  if (opened) loadOnlineClientIps()
}

async function loadOnlineWorkflowOptions(force = false) {
  const keyword = workflowOptionKeyword.value.trim()
  if (!force && onlineWorkflowOptions.value.length > 0 && !keyword) return

  const requestSeq = ++workflowOptionRequestSeq
  workflowOptionLoading.value = true
  try {
    const data = await listAutomaWorkflows({
      keyword,
      page_num: 1,
      page_size: 30,
    })
    if (requestSeq !== workflowOptionRequestSeq) return

    const pageList = normalizeList(data, 'workflows')
    const workflowMap = new Map()

    onlineWorkflowOptions.value.forEach((workflow) => {
      const workflowId = workflow.automa_id || getWorkflowId(workflow)
      if (workflowId) workflowMap.set(workflowId, workflow)
    })
    pageList.forEach((workflow) => {
      const workflowId = workflow.automa_id || getWorkflowId(workflow)
      if (!workflowId) return

      workflowMap.set(workflowId, {
        ...workflow,
        automa_id: workflowId,
        name: workflow.name || '',
        automa_name: workflow.automa_name || workflow.automa_id || workflowId,
      })
    })

    onlineWorkflowOptions.value = Array.from(workflowMap.values())
    const options = Array.from(workflowMap.values())
    const selectedWorkflow = selectedAutomaId.value
      ? options.find((workflow) => (workflow.automa_id || getWorkflowId(workflow)) === selectedAutomaId.value)
      : null
    onlineWorkflowOptions.value = selectedWorkflow
      ? [selectedWorkflow, ...options.filter((workflow) => (workflow.automa_id || getWorkflowId(workflow)) !== selectedAutomaId.value)]
      : options
    if (selectedAutomaId.value && !selectedWorkflow && !keyword) {
      await loadSelectedWorkflowOption(selectedAutomaId.value)
    }
  } finally {
    if (requestSeq === workflowOptionRequestSeq) {
      workflowOptionLoading.value = false
    }
  }
}

function handleWorkflowSelectVisible(opened) {
  if (opened) loadOnlineWorkflowOptions()
}

function handleWorkflowRemoteSearch(value) {
  workflowOptionKeyword.value = normalizeText(value)
  clearWorkflowOptionSearchTimer()
  runWorkflowOptionSearch(true)
}

async function loadSelectedWorkflowOption(automaId) {
  automaId = normalizeText(automaId)
  if (!automaId) return

  const data = await listAutomaWorkflows({
    keyword: automaId,
    page_num: 1,
    page_size: 10,
  })
  const selectedWorkflow = normalizeList(data, 'workflows')
    .find((workflow) => (workflow.automa_id || getWorkflowId(workflow)) === automaId)
  if (!selectedWorkflow) return

  onlineWorkflowOptions.value = [
    selectedWorkflow,
    ...onlineWorkflowOptions.value.filter((workflow) => (workflow.automa_id || getWorkflowId(workflow)) !== automaId),
  ]
}

function handleClientIpChange(value) {
  selectedClientIp.value = normalizeText(value)
  selectedNodeId.value = ''
  refreshCandidates.value = true
  resetCandidates()
  if (activeMode.value === 'workflow' || selectedClientIp.value) loadFirstCandidatePage()
}

function handleClientIpClear() {
  clearKeywordSearchTimer()
  selectedClientIp.value = ''
  selectedNodeId.value = ''
  refreshCandidates.value = activeMode.value === 'workflow'
  resetCandidatePage()
  if (activeMode.value === 'workflow' && selectedAutomaId.value) {
    loadFirstCandidatePage()
    return
  }
  resetCandidates()
}

function handleNodeChange(value) {
  selectedNodeId.value = normalizeText(value)
  refreshCandidates.value = true
  resetCandidates()
  if (canLoad.value) loadFirstCandidatePage()
}

function handleNodeClear() {
  selectedNodeId.value = ''
  refreshCandidates.value = true
  resetCandidates()
  if (canLoad.value) loadFirstCandidatePage()
}

function handleCandidateFilterChange() {
  resetSelection(tableRef)
  if (canLoad.value) loadFirstCandidatePage()
}

function handleCandidateFilterClear() {
  handleCandidateFilterChange()
}

function handleWorkflowChange(value) {
  selectedAutomaId.value = value || ''
  refreshCandidates.value = true
  resetCandidates()
  if (selectedAutomaId.value) loadFirstCandidatePage()
}

function handleWorkflowClear() {
  clearKeywordSearchTimer()
  clearWorkflowOptionSearchTimer()
  selectedAutomaId.value = ''
  workflowOptionKeyword.value = ''
  refreshCandidates.value = false
  resetCandidatePage()
  resetCandidates()
}

function handleResetCurrentMode() {
  clearKeywordSearchTimer()
  clearWorkflowOptionSearchTimer()
  modeStateCache[activeMode.value] = createModeState()
  restoreModeState(activeMode.value)
  if (activeMode.value === 'workflow') {
    workflowOptionKeyword.value = ''
  }
  refreshCandidates.value = false
  resetCandidatePage()
  resetCandidates()
}

function handleModeChange() {
  const nextMode = activeMode.value
  const previousMode = lastActiveMode.value
  if (previousMode === nextMode) return

  saveModeState(previousMode)
  clearKeywordSearchTimer()
  clearWorkflowOptionSearchTimer()
  resetCandidates()
  restoreModeState(nextMode)
  lastActiveMode.value = nextMode

  if (activeMode.value === 'client') {
    loadOnlineClientIps()
  } else {
    loadOnlineWorkflowOptions()
    loadOnlineClientIps()
  }

  refreshCandidates.value = canLoad.value
  if (canLoad.value) loadCandidates()
}

function scheduleKeywordSearch() {
  clearKeywordSearchTimer()
  if (!visible.value || !canLoad.value) return

  runKeywordSearch()
}

async function handleSync() {
  const groups = groupSelectedRowsByNode()
  if (groups.size === 0) {
    showWarningMessage('请选择可同步的工作流')
    return
  }

  syncing.value = true
  try {
    for (const [, group] of groups) {
      await syncAutomaWorkflowsByIp(group.source_ip, group.workflow_ids, [], group.source_node_id)
    }
    showSuccessMessage('同步完成')
    visible.value = false
    emit('synced')
  } finally {
    syncing.value = false
  }
}

function groupSelectedRowsByNode() {
  const groups = new Map()
  selectedRows.value.forEach((row) => {
    const selectedClient = getSelectedClient()
    const sourceIp = normalizeText(row.source_ip || selectedClient.source_ip)
    const sourceNodeId = normalizeText(row.node_id || selectedClient.node_id)
    const identity = buildNodeIdentity(sourceIp, sourceNodeId)
    const workflowId = getWorkflowId(row)
    if (!identity || !sourceIp || !workflowId) return

    const current = groups.get(identity) || {
      source_ip: sourceIp,
      source_node_id: sourceNodeId,
      workflow_ids: [],
    }
    current.workflow_ids.push(workflowId)
    groups.set(identity, current)
  })
  return groups
}

function showSuccessMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.success,
    message,
  })
}

function showWarningMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.warning,
    message,
  })
}

function resetCandidates() {
  candidateRequestSeq += 1
  candidateLoading.value = false
  candidates.value = []
  candidateTotal.value = 0
  resetSelection(tableRef)
}

function resetCandidatePage() {
  currentPage.value = 1
}

function createModeState() {
  return {
    selectedClientIp: '',
    selectedNodeId: '',
    selectedAutomaId: '',
    keyword: '',
    syncStatusFilter: '',
    currentPage: 1,
    pageSize: 10,
  }
}

function saveModeState(mode = activeMode.value) {
  if (!modeStateCache[mode]) return

  modeStateCache[mode] = {
    selectedClientIp: selectedClientIp.value,
    selectedNodeId: selectedNodeId.value,
    selectedAutomaId: selectedAutomaId.value,
    keyword: keyword.value,
    syncStatusFilter: syncStatusFilter.value,
    currentPage: currentPage.value,
    pageSize: pageSize.value,
  }
}

function restoreModeState(mode = activeMode.value) {
  const state = modeStateCache[mode] || createModeState()
  restoringModeState.value = true
  selectedClientIp.value = state.selectedClientIp
  selectedNodeId.value = state.selectedNodeId
  selectedAutomaId.value = state.selectedAutomaId
  keyword.value = state.keyword
  syncStatusFilter.value = state.syncStatusFilter || ''
  currentPage.value = state.currentPage || 1
  pageSize.value = state.pageSize || 10
  restoringModeState.value = false
}

function resetModeStateCache() {
  modeStateCache.client = createModeState()
  modeStateCache.workflow = createModeState()
}

function loadFirstCandidatePage() {
  if (currentPage.value === 1) {
    loadCandidates()
    return
  }

  currentPage.value = 1
}

function isSelectable(row) {
  if (row.sync_status === 'client_missing') return false
  if (row.sync_status === 'server_newer') return false
  return Boolean(row.has_update ?? row.hasUpdate ?? !row.synced)
}

function getSyncText(row) {
  if (row.sync_status === 'client_missing') return '客户端缺失'
  if (row.sync_status === 'server_newer') return '不可同步'
  if (row.has_update || row.hasUpdate) return '可同步'
  if (row.synced) return '已同步'
  return '未同步'
}

function getSyncTagType(row) {
  if (row.sync_status === 'server_newer') return 'danger'
  if (row.has_update || row.hasUpdate) return 'warning'
  if (row.synced) return 'success'
  return 'info'
}

function getWorkflowText(row) {
  if (row.sync_status === 'client_missing') return '客户端无此工作流'
  if (row.sync_status === 'not_synced') return '数据库无记录'
  if (row.sync_status === 'client_newer') return '客户端较新'
  if (row.sync_status === 'server_newer') return '数据库较新'
  if (row.synced) return '内容一致'
  if (row.has_update || row.hasUpdate) return '内容有差异'
  return '待同步'
}

function getWorkflowTagType(row) {
  if (row.synced) return 'success'
  if (row.sync_status === 'client_missing') return 'info'
  if (row.sync_status === 'server_newer') return 'danger'
  if (row.has_update || row.hasUpdate) return 'warning'
  return 'info'
}

function getWorkflowId(row) {
  return row?.id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function getSelectionKey(row) {
  const selectedClient = getSelectedClient()
  return row?.row_key || `${buildNodeIdentity(row?.source_ip || selectedClient.source_ip, row?.node_id || selectedClient.node_id)}_${getWorkflowId(row)}`
}

function getSelectedClient() {
  const clientIp = normalizeText(selectedClientIp.value)
  const nodeId = normalizeNodeId(clientIp, selectedNodeId.value)
  return {
    key: buildNodeIdentity(clientIp, nodeId),
    identity: buildNodeIdentity(clientIp, nodeId),
    source_ip: clientIp,
    node_id: nodeId,
  }
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

function buildNodeIdentity(clientIp, nodeId) {
  clientIp = normalizeText(clientIp)
  nodeId = normalizeNodeId(clientIp, nodeId)
  if (!clientIp) return nodeId
  if (!nodeId || nodeId === clientIp) return clientIp
  return `${clientIp}|${nodeId}`
}

function formatOptionalDate(value) {
  if (!value) return ''
  return formatDate(value)
}

function formatClientNode(row) {
  const clientIp = normalizeText(row?.source_ip)
  const nodeId = normalizeText(row?.node_id)
  return [clientIp, nodeId].filter(Boolean).join(' / ')
}

function formatCompareText(clientValue, serverValue) {
  return `客户端：${clientValue || ''} 数据库：${serverValue || ''}`.trim()
}
</script>

<style scoped lang="scss">
.sync-dialog {
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: 608px;
  min-height: 0;
}

.sync-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.sync-scope {
  min-height: 20px;
  color: #606266;
  font-size: 13px;
  line-height: 20px;
}

.query-input {
  width: 360px;
}

.client-ip-select {
  width: 220px;
}

.node-select {
  width: 180px;
}

.status-select {
  width: 150px;
}

.keyword-input {
  width: 220px;
}

.candidate-table {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
}

.candidate-table :deep(.nowrap-column .cell) {
  padding-inline: 8px;
  white-space: nowrap;
}

.workflow-option {
  display: grid;
  gap: 2px;
  min-width: 0;

  span,
  small {
    display: flex;
    align-items: center;
    gap: 6px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  em {
    flex: 0 0 auto;
    color: #909399;
    font-style: normal;
  }

  small {
    color: #909399;
  }
}

.field-value {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: #303133;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-lines {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.field-line {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  align-items: center;
  gap: 6px;
  min-width: 0;

  em {
    color: #909399;
    font-style: normal;
  }

  span {
    min-width: 0;
    overflow: hidden;
    color: #303133;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.compare-line {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.compare-item {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  align-items: center;
  gap: 4px;
  min-width: 0;
  line-height: 18px;
  white-space: nowrap;

  em {
    color: #909399;
    font-style: normal;
  }

  span {
    min-width: 0;
    overflow: hidden;
    color: #303133;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.time-value {
  display: block;
  width: 100%;
  white-space: nowrap;
}

.center-cell {
  display: flex;
  justify-content: center;
  min-width: 0;
}

.center-cell :deep(.el-tag) {
  max-width: none;
  min-width: 72px;
  padding-inline: 8px;
  white-space: nowrap;
}

.sync-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 32px;
}

.sync-footer :deep(.app-pagination) {
  justify-content: flex-end;
  margin-top: 0;
}

.sync-pagination--hidden {
  visibility: hidden;
  pointer-events: none;
}

.sync-summary {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  gap: 16px;
  color: #606266;
  line-height: 24px;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .sync-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .sync-footer {
    align-items: stretch;
    flex-direction: column;
  }

  .sync-footer :deep(.app-pagination) {
    justify-content: center;
  }

  .query-input,
  .client-ip-select,
  .node-select,
  .status-select,
  .keyword-input {
    width: 100%;
  }
}
</style>
