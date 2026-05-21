<template>
  <AppDialog
    v-model="dialogVisible"
    title="同步后端工作流到本地 Automa"
    width="980px"
  >
    <div class="sync-dialog">
      <div class="sync-toolbar">
        <el-input
          v-model="syncKeyword"
          clearable
          placeholder="搜索后端工作流名称、ID、来源 IP"
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
        <el-table-column label="工作流名称" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.name || '' }}
          </template>
        </el-table-column>
        <el-table-column label="工作流描述" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '' }}
          </template>
        </el-table-column>
        <el-table-column label="本地状态" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="getLocalWorkflowTagType(row)" effect="plain">
              {{ getLocalWorkflowText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160" class-name="nowrap-column">
          <template #default="{ row }">
            {{ formatDate(getServerWorkflowUpdatedAt(row)) }}
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
    const data = await listAutomaWorkflows({
      keyword: syncKeyword.value.trim(),
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

function getLocalWorkflowText(row) {
  return getLocalWorkflowStatus(row).text
}

function getLocalWorkflowTagType(row) {
  return getLocalWorkflowStatus(row).type
}

function getLocalWorkflowStatus(row) {
  const localWorkflow = getMatchedLocalWorkflow(row)
  if (!localWorkflow) {
    return { text: '待同步', type: 'info' }
  }

  const serverHash = String(row?.serverContentHash || row?.content_hash || row?.contentHash || '').trim()
  const localHash = String(localWorkflow.contentHash || '').trim()
  if (serverHash && localHash && serverHash === localHash) {
    return { text: '已同步', type: 'success' }
  }
  if (serverHash && localHash) {
    return { text: '本地有差异', type: 'warning' }
  }

  return { text: '本地已存在', type: 'primary' }
}

function getMatchedLocalWorkflow(row) {
  return localWorkflowMap.value.get(getWorkflowDisplayId(row)) || null
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
