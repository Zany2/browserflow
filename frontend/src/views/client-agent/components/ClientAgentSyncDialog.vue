<template>
  <AppDialog
    v-model="dialogVisible"
    title="同步后端工作流到本地"
    width="1360px"
  >
    <div class="sync-dialog">
      <div class="sync-toolbar">
        <el-input
          v-model="syncKeyword"
          clearable
          placeholder="搜索后端自定义名称、描述"
          @clear="searchSyncWorkflowsNow"
          @keyup.enter="searchSyncWorkflowsNow"
        />
        <el-button :loading="syncListLoading" @click="searchSyncWorkflowsNow">刷新列表</el-button>
        <el-button :loading="localWorkflowLoading" @click="loadLocalWorkflows">刷新本地</el-button>
      </div>

      <el-table
        ref="syncTableRef"
        v-loading="syncListLoading"
        class="adaptive-table"
        :data="syncableWorkflows"
        border
        height="420"
        row-key="id"
        @selection-change="handleSyncSelectionChange"
      >
        <el-table-column type="selection" width="40" reserve-selection />
        <el-table-column label="自定义名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.name || '' }}
          </template>
        </el-table-column>
        <el-table-column label="自定义描述" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '' }}
          </template>
        </el-table-column>
        <el-table-column label="工作流名称" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getAutomaWorkflowName(row) }}
          </template>
        </el-table-column>
        <el-table-column label="工作流描述" min-width="190" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getAutomaWorkflowDescription(row) }}
          </template>
        </el-table-column>
        <el-table-column label="节点/连线" width="90" align="center">
          <template #default="{ row }">
            {{ formatWorkflowGraphSize(row) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(getServerWorkflowUpdatedAt(row)) }}
          </template>
        </el-table-column>
        <el-table-column label="同步状态" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="getSyncTagType(row)" effect="plain">
              {{ getSyncText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="工作流状态" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="getWorkflowTagType(row)" effect="plain">
              {{ getWorkflowText(row) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div class="sync-footer">
        <AppSelectionSummary class="sync-summary" :count="selectedSyncIds.length" unit="工作流" />
        <AppPagination
          v-model:current-page="syncPageNum"
          v-model:page-size="syncPageSize"
          :total="syncTotal"
          layout="total, sizes, prev, pager, next"
        />
      </div>
    </div>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button
        type="primary"
        :loading="syncing"
        :disabled="selectedSyncIds.length === 0 || syncing"
        @click="handleSyncSelectedWorkflows"
      >
        同步到本地 Automa
      </el-button>
    </template>
  </AppDialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import { useDebouncedAction } from '@/composables/useDebouncedAction'
import { usePagedTableSelection } from '@/composables/usePagedTableSelection'
import { getAutomaWorkflowDetail, listAutomaWorkflows } from '@/services/automa'
import { getAutomaWorkflows, importAutomaWorkflow } from '@/services/automaBridge'
import { formatDate } from '@/utils/format'
import { normalizeList } from '@/utils/list'
import { createWorkflowContentHash } from '@/utils/workflowHash'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue'])

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const syncKeyword = ref('')
const syncableWorkflows = ref([])
const syncTotal = ref(0)
const syncPageNum = ref(1)
const syncPageSize = ref(10)
const localWorkflowLoading = ref(false)
const syncListLoading = ref(false)
const syncing = ref(false)
const localWorkflowMap = ref(new Map())
const syncTableRef = ref(null)

const {
  selectedRows: selectedSyncRows,
  selectedKeys: selectedSyncIds,
  handleSelectionChange: handleSyncSelectionChange,
  restoreSelection: restoreSyncSelection,
  resetSelection: resetSyncSelection,
} = usePagedTableSelection({
  rows: syncableWorkflows,
  getRowKey: getServerWorkflowId,
})
const {
  run: scheduleSyncSearch,
  cancel: clearSyncSearchTimer,
} = useDebouncedAction(searchSyncWorkflowsNow, 200)

watch(dialogVisible, async (nextVisible) => {
  if (!nextVisible) {
    clearSyncSearchTimer()
    resetSyncSelection(syncTableRef)
    return
  }

  resetSyncDialogState()
  await Promise.all([loadSyncableWorkflows(), loadLocalWorkflows()])
})

watch(syncPageNum, () => {
  if (!dialogVisible.value) return
  loadSyncableWorkflows()
})

watch(syncPageSize, () => {
  if (!dialogVisible.value) return
  loadFirstSyncPage()
})

watch(syncKeyword, () => {
  if (!dialogVisible.value) return
  scheduleSyncSearch()
})

async function loadSyncableWorkflows() {
  syncListLoading.value = true

  try {
    const keyword = syncKeyword.value.trim()
    const data = await listAutomaWorkflows({
      keyword,
      custom_keyword: keyword,
      page_num: syncPageNum.value,
      page_size: syncPageSize.value,
      syncable: 1,
    })
    const workflowList = normalizeList(data, 'workflows').filter((item) => !item.is_deleted)
    syncableWorkflows.value = await enrichSyncableWorkflowHashes(workflowList)
    syncTotal.value = Number(data?.total || 0)
    await restoreSyncSelection(syncTableRef)
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '读取后端工作流失败' })
  } finally {
    syncListLoading.value = false
  }
}

async function loadLocalWorkflows() {
  localWorkflowLoading.value = true

  try {
    const workflows = await getAutomaWorkflows()
    localWorkflowMap.value = await buildLocalWorkflowMap(workflows)
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '读取本地工作流失败' })
  } finally {
    localWorkflowLoading.value = false
  }
}

function searchSyncWorkflowsNow() {
  clearSyncSearchTimer()
  resetSyncSelection(syncTableRef)
  loadFirstSyncPage()
}

function resetSyncDialogState() {
  clearSyncSearchTimer()
  syncKeyword.value = ''
  syncPageNum.value = 1
  syncableWorkflows.value = []
  syncTotal.value = 0
  resetSyncSelection(syncTableRef)
}

function loadFirstSyncPage() {
  if (syncPageNum.value === 1) {
    loadSyncableWorkflows()
    return
  }

  syncPageNum.value = 1
}

async function handleSyncSelectedWorkflows() {
  if (selectedSyncRows.value.length === 0 || syncing.value) return

  syncing.value = true

  try {
    let syncedCount = 0
    const rows = selectedSyncRows.value.slice()
    for (const row of rows) {
      const workflowId = getServerWorkflowId(row)
      const data = await getAutomaWorkflowDetail(workflowId)
      const workflow = data?.workflow || data
      await importAutomaWorkflow(extractWorkflowImportPayload(workflow))
      syncedCount += 1
    }

    await loadLocalWorkflows()
    resetSyncSelection(syncTableRef)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: `已同步 ${syncedCount} 个工作流` })
    dialogVisible.value = false
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '同步工作流失败' })
  } finally {
    syncing.value = false
  }
}

async function enrichSyncableWorkflowHashes(workflows) {
  return Promise.all(workflows.map(async (workflow) => {
    try {
      const detail = await getAutomaWorkflowDetail(getServerWorkflowId(workflow))
      const detailWorkflow = parseWorkflowPayload(detail?.raw_json || detail?.rawJson || detail?.normalized_json || detail?.normalizedJson)
      return {
        ...workflow,
        content_hash: workflow.content_hash || detail?.content_hash || detail?.contentHash || '',
        automa_name:
          workflow.automa_name || workflow.automaName || detail?.automa_name || detail?.automaName || detailWorkflow?.name || '',
        automa_description:
          workflow.automa_description ||
          workflow.automaDescription ||
          detail?.automa_description ||
          detail?.automaDescription ||
          detailWorkflow?.description ||
          '',
        serverContentHash: detailWorkflow ? await createWorkflowContentHash(detailWorkflow) : '',
        updated_at_automa:
          workflow.updated_at_automa || detail?.updated_at_automa || detail?.updatedAtAutoma || 0,
      }
    } catch {
      return workflow
    }
  }))
}

function extractWorkflowImportPayload(workflow) {
  if (!workflow) {
    throw new Error('工作流详情为空，无法同步')
  }

  const rawWorkflow =
    workflow.raw_json ||
    workflow.rawJson ||
    workflow.normalized_json ||
    workflow.normalizedJson ||
    workflow.data

  const parsedWorkflow = parseWorkflowPayload(rawWorkflow)
  const payload = parsedWorkflow || {
    id: workflow.automa_id || workflow.workflow_id || workflow.id,
    name: workflow.name,
    description: workflow.description,
    drawflow: workflow.drawflow_json || workflow.drawflow || {},
    settings: workflow.settings_json || workflow.settings || {},
    table: workflow.table_json || workflow.table || [],
    dataColumns: workflow.data_columns_json || workflow.dataColumns || [],
    trigger: workflow.trigger_json || workflow.trigger || null,
    globalData: workflow.global_data || '',
  }

  const automaId = workflow.automa_id || workflow.automaId || workflow.workflow_id || payload.id
  if (automaId) {
    payload.id = automaId
  }

  const createdAt = resolveAutomaTimestamp(
    workflow.created_at_automa,
    workflow.createdAtAutoma,
    payload.createdAt,
  )
  const updatedAt = resolveAutomaTimestamp(
    workflow.updated_at_automa,
    workflow.updatedAtAutoma,
    payload.updatedAt,
  )
  if (createdAt) {
    payload.createdAt = createdAt
  }
  if (updatedAt) {
    payload.updatedAt = updatedAt
  }
  if (!payload.table && payload.dataColumns) {
    payload.table = payload.dataColumns
  }

  return payload
}

function parseWorkflowPayload(value) {
  if (!value) return null
  if (typeof value === 'object') return { ...value }
  if (typeof value !== 'string') return null

  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' ? parsed : null
  } catch {
    return null
  }
}

function resolveAutomaTimestamp(...values) {
  for (const value of values) {
    if (!value) continue

    const numberValue = Number(value)
    if (Number.isFinite(numberValue) && numberValue > 0) return numberValue

    const dateValue = new Date(value).getTime()
    if (!Number.isNaN(dateValue)) return dateValue
  }

  return 0
}

function getServerWorkflowId(row) {
  return row?.id || row?.server_id || row?.automa_id || row?.workflow_id || row?.workflowId || ''
}

function getWorkflowDisplayId(row) {
  return row?.automa_id || row?.workflow_id || row?.workflowId || row?.id || ''
}

function getLocalWorkflowId(row) {
  return row?.id || row?.automaId || row?.automa_id || row?.workflowId || ''
}

function getSyncText(row) {
  const status = resolveLocalWorkflowStatus(row)
  if (status.key === 'local_newer') return '不可同步'
  if (status.key === 'synced') return '已同步'
  if (status.canSync) return '可同步'
  return '未同步'
}

function getSyncTagType(row) {
  const status = resolveLocalWorkflowStatus(row)
  if (status.key === 'local_newer') return 'danger'
  if (status.key === 'synced') return 'success'
  if (status.canSync) return 'warning'
  return 'info'
}

function getWorkflowText(row) {
  const status = resolveLocalWorkflowStatus(row)
  if (status.key === 'not_synced') return '本地无记录'
  if (status.key === 'local_newer') return '本地较新'
  if (status.key === 'server_newer') return '数据库较新'
  if (status.key === 'synced') return '内容一致'
  if (status.key === 'has_update') return '内容有差异'
  return '待同步'
}

function getWorkflowTagType(row) {
  const status = resolveLocalWorkflowStatus(row)
  if (status.key === 'synced') return 'success'
  if (status.key === 'local_newer') return 'danger'
  if (status.canSync) return 'warning'
  return 'info'
}

function resolveLocalWorkflowStatus(row) {
  const localWorkflow = getMatchedLocalWorkflow(row)
  if (!localWorkflow) {
    return { key: 'not_synced', canSync: true }
  }

  const serverHash = String(row?.serverContentHash || row?.content_hash || row?.contentHash || '').trim()
  const localHash = String(localWorkflow.contentHash || '').trim()
  if (serverHash && localHash && serverHash === localHash) {
    return { key: 'synced', canSync: false }
  }

  const serverUpdatedAt = resolveAutomaTimestamp(
    row?.updated_at_automa,
    row?.updatedAtAutoma,
    row?.updatedAt,
    row?.updated_at,
  )
  const localUpdatedAt = resolveAutomaTimestamp(
    localWorkflow.updatedAt,
    localWorkflow.updated_at,
    localWorkflow.updatedAtAutoma,
    localWorkflow.updated_at_automa,
  )
  if (serverUpdatedAt && localUpdatedAt && localUpdatedAt > serverUpdatedAt) {
    return { key: 'local_newer', canSync: false }
  }
  if (serverUpdatedAt && localUpdatedAt && localUpdatedAt < serverUpdatedAt) {
    return { key: 'server_newer', canSync: true }
  }
  if (serverHash && localHash) {
    return { key: 'has_update', canSync: true }
  }

  return { key: 'exists', canSync: true }
}

function getMatchedLocalWorkflow(row) {
  return localWorkflowMap.value.get(getWorkflowDisplayId(row)) || null
}

function formatWorkflowGraphSize(row) {
  const nodeCount = row?.node_count ?? row?.nodeCount ?? 0
  const edgeCount = row?.edge_count ?? row?.edgeCount ?? 0
  return `${nodeCount} / ${edgeCount}`
}

function getAutomaWorkflowName(row) {
  return row?.automa_name || row?.automaName || row?.workflow_name || row?.workflowName || ''
}

function getAutomaWorkflowDescription(row) {
  return row?.automa_description || row?.automaDescription || row?.workflow_description || row?.workflowDescription || ''
}

async function buildLocalWorkflowMap(workflows) {
  const workflowMap = new Map()
  await Promise.all(workflows.map(async (workflow) => {
    const workflowId = getLocalWorkflowId(workflow)
    if (workflowId) {
      workflowMap.set(workflowId, {
        ...workflow,
        contentHash: await createWorkflowContentHash(workflow),
      })
    }
  }))

  return workflowMap
}

function getServerWorkflowUpdatedAt(row) {
  return row?.updated_at_automa || row?.updatedAtAutoma || row?.updated_at || row?.updatedAt
}
</script>

<style scoped lang="scss">
.sync-dialog {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sync-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: 12px;
}

.sync-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.sync-footer :deep(.app-pagination) {
  justify-content: flex-end;
  margin-top: 0;
}

.sync-summary {
  color: #64748b;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .sync-toolbar {
    grid-template-columns: 1fr;
  }

  .sync-footer {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
