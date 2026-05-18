<template>
  <section class="browser-page">
    <header class="page-toolbar">
      <div>
        <h1>浏览器</h1>
        <p>配置启动的浏览器，后续自动化操作会复用当前运行实例。</p>
      </div>
      <div class="status-card">
        <el-tag :type="status.running ? 'success' : 'info'">
          {{ status.running ? '运行中' : '未启动' }}
        </el-tag>
        <el-tag :type="currentAgentOnline ? 'success' : 'warning'">
          {{ currentAgentOnline ? 'Agent 在线' : 'Agent 离线' }}
        </el-tag>
        <span>{{ currentStatusText }}</span>
        <el-button
          class="executor-skill-button"
          type="primary"
          :icon="Download"
          :loading="executorSkillExporting"
          @click="handleExportExecutorSkill"
        >
          导出控制 Skill
        </el-button>
      </div>
    </header>

    <div class="browser-layout">
      <aside class="instance-panel">
        <div class="panel-title">
          <span>浏览器配置</span>
          <el-button type="primary" :icon="Plus" @click="handleNew">新建</el-button>
        </div>

        <div class="table-toolbar">
          <el-checkbox
            :model-value="isAllInstancesSelected"
            :indeterminate="isInstanceSelectionIndeterminate"
            :disabled="instances.length === 0"
            @change="handleToggleAllInstances"
          >
            全选
          </el-checkbox>
          <el-button
            link
            type="danger"
            :disabled="selectedInstanceIds.length === 0"
            @click="handleDeleteSelectedInstances"
          >
            删除选中
          </el-button>
        </div>

        <el-table
          class="adaptive-table"
          :data="pagedInstances"
          border
          highlight-current-row
          height="100%"
          :current-row-key="selectedId"
          row-key="id"
          @row-click="selectInstance"
        >
          <el-table-column width="40" align="center" class-name="action-column">
            <template #default="{ row }">
              <el-checkbox
                :model-value="selectedInstanceIds.includes(row.id)"
                @click.stop
                @change="(checked) => handleToggleInstance(row.id, checked)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="120">
            <template #default="{ row }">
              <div class="instance-name">
                <span class="instance-name__text" :title="row.name">{{ row.name }}</span>
                <el-tag v-if="row.is_current" size="small" type="primary">当前</el-tag>
                <el-tag v-if="row.is_active" size="small" type="success">运行</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="Automa" width="106" align="center" class-name="action-column">
            <template #default="{ row }">
              <div class="automa-status">
                <el-tag :type="getAutomaTagType(row)" effect="plain">
                  {{ getAutomaStatusText(row) }}
                </el-tag>
                <span v-if="getAutomaVersionText(row)" class="automa-version">
                  {{ getAutomaVersionText(row) }}
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="默认" width="60" align="center" class-name="action-column">
            <template #default="{ row }">
              <el-tag v-if="row.is_default" type="success" effect="plain">默认</el-tag>
              <span v-else></span>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="66">
            <template #default="{ row }">
              {{ row.type === 'remote' ? '远程' : '本地' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="138" class-name="operation-column">
            <template #default="{ row }">
              <el-button link type="primary" :disabled="row.is_active" @click.stop="handleStart(row)">启动</el-button>
              <el-button link type="danger" :disabled="!row.is_active" @click.stop="handleStop(row)">停止</el-button>
              <el-button link type="success" :disabled="!row.is_active || row.is_current" @click.stop="handleSwitch(row)">切换</el-button>
            </template>
          </el-table-column>
        </el-table>

        <AppPagination
          v-model:current-page="instanceCurrentPage"
          v-model:page-size="instancePageSize"
          class="instance-pagination"
          :page-sizes="instancePageSizes"
          :total="instances.length"
        />
      </aside>

      <main class="config-panel">
        <div class="panel-title">
          <span>{{ form.id ? '编辑配置' : '新建配置' }}</span>
          <div class="panel-actions">
            <el-button :icon="RefreshRight" @click="loadAll">刷新</el-button>
            <el-button type="primary" :icon="Check" @click="handleSave">{{ saveButtonText }}</el-button>
          </div>
        </div>

        <el-form class="browser-form" label-width="108px" :model="form">
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="名称">
                <el-input v-model="form.name" placeholder="例如：默认 Chrome" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="浏览器类型">
                <el-radio-group v-model="form.type">
                  <el-radio-button label="local">本地</el-radio-button>
                  <el-radio-button label="remote">远程</el-radio-button>
                </el-radio-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="说明">
            <el-input v-model="form.description" placeholder="用于区分不同浏览器配置" />
          </el-form-item>

          <template v-if="form.type === 'local'">
            <el-form-item label="浏览器路径">
              <el-input v-model="form.bin_path" placeholder="不填则使用 rod 自动查找的浏览器" />
            </el-form-item>
            <el-form-item label="用户目录">
              <el-input v-model="form.user_data_dir" placeholder="独立用户数据目录，可保持登录态" />
            </el-form-item>
          </template>

          <el-form-item v-else label="远程地址">
            <el-input v-model="form.control_url" placeholder="ws://127.0.0.1:9222/devtools/browser/..." />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="代理">
                <el-input v-model="form.proxy" placeholder="例如：http://127.0.0.1:7890" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="默认配置">
                <el-switch
                  v-model="form.is_default"
                  :loading="defaultSaving"
                  active-text="设为默认"
                  @change="handleDefaultChange"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="启动选项">
            <div class="switch-row">
              <el-switch v-model="form.headless" active-text="无头模式" />
              <el-switch v-model="form.no_sandbox" active-text="禁用沙箱" />
            </div>
          </el-form-item>

          <el-form-item label="启动参数">
            <el-input
              v-model="launchArgsText"
              type="textarea"
              :rows="5"
              placeholder="每行一个参数，例如：--disable-infobars"
            />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" :loading="saving" @click="handleSave">{{ saveButtonText }}</el-button>
            <el-button :disabled="!form.id" :loading="starting" @click="handleStart(form, true)">启动</el-button>
            <el-button :disabled="!form.id" @click="handleStop(form)">停止</el-button>
            <el-button :disabled="!form.id" @click="handleDelete">删除</el-button>
          </el-form-item>
        </el-form>

        <section class="runtime-panel">
          <h2>运行状态</h2>
          <el-descriptions border :column="2">
            <el-descriptions-item label="状态">
              {{ status.running ? '运行中' : '未启动' }}
            </el-descriptions-item>
            <el-descriptions-item label="当前配置">
              {{ status.instance?.name || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="运行时长">
              {{ uptimeText }}
            </el-descriptions-item>
            <el-descriptions-item label="调试地址">
              <div class="runtime-copy">
                <span class="runtime-value" :title="status.control_url || '-'">
                  {{ status.control_url || '-' }}
                </span>
                <el-button link type="primary" :disabled="!status.control_url" @click="copyRuntimeValue(status.control_url)">
                  复制
                </el-button>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="Agent 页面">
              <div class="runtime-copy">
                <span class="runtime-value" :title="status.agent_url || '-'">
                  {{ status.agent_url || '-' }}
                </span>
                <el-button link type="primary" :disabled="!status.agent_url" @click="copyRuntimeValue(status.agent_url)">
                  复制
                </el-button>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="Agent 状态">
              {{ currentAgentOnline ? '在线' : '离线' }}
            </el-descriptions-item>
          </el-descriptions>
        </section>
      </main>
    </div>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Check, Download, Plus, RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
import {
  createBrowserInstance,
  deleteBrowserInstance,
  getAgentStatus,
  getBrowserStatus,
  listBrowserInstances,
  startBrowserInstance,
  stopBrowserInstance,
  subscribeAgentStatus,
  subscribeBrowserStatus,
  switchBrowserInstance,
  updateBrowserInstance,
} from '@/services/browser'
import { exportBrowserExecutorSkill } from '@/services/browserExecutor'

const instances = ref([])
const selectedId = ref('')
const selectedInstanceIds = ref([])
const instanceCurrentPage = ref(1)
const instancePageSize = ref(10)
const instancePageSizes = [10, 30, 60]
const saving = ref(false)
const starting = ref(false)
const defaultSaving = ref(false)
const launchArgsText = ref('')
const agents = ref([])
const runtimeNow = ref(Date.now())
const runtimeClockTimer = ref(null)
const browserStatusRefreshing = ref(false)
const executorSkillExporting = ref(false)
const stopBrowserStatusSubscribe = ref(null)
const stopAgentStatusSubscribe = ref(null)
const status = ref({
  running: false,
  current_instance_id: '',
  uptime_seconds: 0,
})

const form = reactive(createEmptyForm())

const currentStatusText = computed(() => {
  if (!status.value.running) return '请选择或启动一个浏览器配置'
  return `当前：${status.value.instance?.name || status.value.current_instance_id}`
})
const currentAgentOnline = computed(() =>
  agents.value.some((agent) => agent.browser_id === status.value.current_instance_id && agent.online),
)
const saveButtonText = computed(() => (form.id ? '保存修改' : '新增配置'))
const pagedInstances = computed(() => {
  const start = (instanceCurrentPage.value - 1) * instancePageSize.value
  return instances.value.slice(start, start + instancePageSize.value)
})
const pagedInstanceIds = computed(() => pagedInstances.value.map((instance) => instance.id))
const selectedPagedInstanceIds = computed(() =>
  selectedInstanceIds.value.filter((id) => pagedInstanceIds.value.includes(id)),
)
const isAllInstancesSelected = computed(
  () => pagedInstanceIds.value.length > 0 && selectedPagedInstanceIds.value.length === pagedInstanceIds.value.length,
)
const isInstanceSelectionIndeterminate = computed(
  () => selectedPagedInstanceIds.value.length > 0 && selectedPagedInstanceIds.value.length < pagedInstanceIds.value.length,
)

const uptimeText = computed(() => {
  const seconds = getRuntimeSeconds()
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const rest = seconds % 60
  return `${hours}小时 ${minutes}分钟 ${rest}秒`
})

onMounted(async () => {
  await loadAll()
  startRuntimeRefresh()
})

onBeforeUnmount(() => {
  stopRuntimeRefresh()
})

watch([instances, instancePageSize], () => {
  instanceCurrentPage.value = getSafePage({
    total: instances.value.length,
    page: instanceCurrentPage.value,
    size: instancePageSize.value,
  })
})

async function loadAll() {
  const [instanceData, statusData, agentData] = await Promise.all([
    listBrowserInstances(),
    getBrowserStatus(),
    getAgentStatus(),
  ])
  instances.value = sortByCreatedDesc(instanceData.instances || [])
  selectedInstanceIds.value = selectedInstanceIds.value.filter((id) =>
    instances.value.some((instance) => instance.id === id),
  )
  status.value = statusData.status || status.value
  agents.value = agentData.agents || []
  markActiveInstances()
  instanceCurrentPage.value = getSafePage({
    total: instances.value.length,
    page: instanceCurrentPage.value,
    size: instancePageSize.value,
  })

  if (!selectedId.value && instances.value.length > 0) {
    selectInstance(instances.value[0])
  }
}

function startRuntimeRefresh() {
  stopRuntimeRefresh()
  // Runtime clock 运行时长本地计时，状态变化由管理端 WS 订阅推送
  runtimeClockTimer.value = window.setInterval(() => {
    runtimeNow.value = Date.now()
  }, 1000)
  stopBrowserStatusSubscribe.value = subscribeBrowserStatus(handleBrowserStatusChanged, (error) =>
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message }),
  )
  stopAgentStatusSubscribe.value = subscribeAgentStatus((nextAgents) => {
    agents.value = nextAgents || []
  }, (error) => appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message }))
}

async function handleBrowserStatusChanged(nextStatus) {
  status.value = nextStatus || status.value
  markActiveInstances()
  if (browserStatusRefreshing.value) return

  browserStatusRefreshing.value = true
  try {
    // Runtime list refresh 运行态变化后重新拉列表，确保非当前实例关闭也能同步状态
    await loadAll()
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
  } finally {
    browserStatusRefreshing.value = false
  }
}

function stopRuntimeRefresh() {
  if (runtimeClockTimer.value) {
    window.clearInterval(runtimeClockTimer.value)
    runtimeClockTimer.value = null
  }
  stopBrowserStatusSubscribe.value?.()
  stopBrowserStatusSubscribe.value = null
  stopAgentStatusSubscribe.value?.()
  stopAgentStatusSubscribe.value = null
}

function selectInstance(row) {
  selectedId.value = row.id
  Object.assign(form, normalizeForm(row))
  launchArgsText.value = (row.launch_args || []).join('\n')
}

function handleNew() {
  selectedId.value = ''
  Object.assign(form, createEmptyForm())
  launchArgsText.value = ''
}

async function handleSave() {
  saving.value = true
  try {
    const isCreate = !form.id
    const payload = buildPayload()
    const data = form.id
      ? await updateBrowserInstance(form.id, payload)
      : await createBrowserInstance(payload)
    appMessage({ type: APP_MESSAGE_TYPE.success, message: isCreate ? '新增成功' : '保存成功' })
    await loadAll()
    selectInstance(data.instance)
  } finally {
    saving.value = false
  }
}

async function handleDefaultChange(checked) {
  if (!form.id) return

  const nextDefault = Boolean(checked)
  const previousDefault = !nextDefault
  defaultSaving.value = true
  try {
    // Default update 默认配置变更，已有配置切换后立即保存并刷新列表标识
    await updateBrowserInstance(form.id, {
      ...buildPayload(),
      is_default: nextDefault,
    })
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: nextDefault ? '已设为默认配置' : '已取消默认配置',
    })
    await loadAll()
    const current = instances.value.find((instance) => instance.id === form.id)
    if (current) selectInstance(current)
  } catch (error) {
    form.is_default = previousDefault
    throw error
  } finally {
    defaultSaving.value = false
  }
}

async function handleStart(row, saveBeforeStart = false) {
  if (!row.id) {
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请先保存浏览器配置' })
    return
  }

  starting.value = true
  try {
    if (saveBeforeStart) {
      // Save before launch 编辑区启动前先保存表单，避免刷新后丢失未持久化配置
      const data = await updateBrowserInstance(form.id, buildPayload())
      if (data.instance) {
        selectInstance(data.instance)
      }
    }
    const data = await startBrowserInstance(row.id)
    status.value = data.status || status.value
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '浏览器已启动' })
    await loadAll()
  } finally {
    starting.value = false
  }
}

async function handleExportExecutorSkill() {
  executorSkillExporting.value = true
  try {
    // Export Skill downloads static browser-control instructions 导出静态浏览器控制 Skill
    const blob = await exportBrowserExecutorSkill()
    downloadBlob(blob, 'SKILL_BROWSER_EXECUTOR.md')
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '浏览器控制 Skill 已导出' })
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message || '导出浏览器控制 Skill 失败' })
  } finally {
    executorSkillExporting.value = false
  }
}

function downloadBlob(blob, filename) {
  // Download file 创建临时链接触发浏览器下载
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

async function handleStop(row) {
  if (!row.id) return
  const data = await stopBrowserInstance(row.id)
  status.value = data.status || status.value
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '浏览器已停止' })
  await loadAll()
}

async function handleSwitch(row) {
  const data = await switchBrowserInstance(row.id)
  status.value = data.status || status.value
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已切换当前浏览器' })
  await loadAll()
}

async function copyRuntimeValue(value) {
  if (!value) return

  await navigator.clipboard.writeText(value)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '已复制' })
}

async function handleDelete() {
  if (!form.id) return
  const confirmed = await appConfirm({
    title: '删除配置',
    message: '确认删除这个浏览器配置吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteBrowserInstance(form.id)
  removeInstancesFromState([form.id])
  handleNew()
  await loadAll()
}

function handleToggleInstance(instanceId, checked) {
  if (checked) {
    selectedInstanceIds.value = Array.from(new Set([...selectedInstanceIds.value, instanceId]))
    return
  }
  selectedInstanceIds.value = selectedInstanceIds.value.filter((id) => id !== instanceId)
}

function handleToggleAllInstances(checked) {
  const pageIds = pagedInstanceIds.value
  if (checked) {
    selectedInstanceIds.value = Array.from(new Set([...selectedInstanceIds.value, ...pageIds]))
    return
  }
  selectedInstanceIds.value = selectedInstanceIds.value.filter((id) => !pageIds.includes(id))
}

async function handleDeleteSelectedInstances() {
  const ids = selectedInstanceIds.value.slice()
  if (ids.length === 0) return

  const confirmed = await appConfirm({
    title: '批量删除配置',
    message: `确认删除选中的 ${ids.length} 个浏览器配置吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  // Batch delete 批量删除，复用现有单个删除接口
  await Promise.all(ids.map((id) => deleteBrowserInstance(id)))
  removeInstancesFromState(ids)
  if (ids.includes(form.id)) {
    handleNew()
  }
  await loadAll()
}

function removeInstancesFromState(instanceIds) {
  instances.value = instances.value.filter((instance) => !instanceIds.includes(instance.id))
  selectedInstanceIds.value = selectedInstanceIds.value.filter((id) => !instanceIds.includes(id))
}

function sortByCreatedDesc(data) {
  // Created order 新增时间倒序，保证列表展示最新配置在前
  return data.slice().sort((a, b) => getTimeValue(b.created_at) - getTimeValue(a.created_at))
}

function getTimeValue(value) {
  return value ? new Date(value).getTime() || 0 : 0
}

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function getRuntimeSeconds() {
  if (!status.value.running) return 0

  const startTime = getTimeValue(status.value.start_time)
  if (startTime > 0) {
    return Math.max(Math.floor((runtimeNow.value - startTime) / 1000), 0)
  }
  return status.value.uptime_seconds || 0
}

function getAgentByBrowserId(browserId) {
  return agents.value.find((agent) => agent.browser_id === browserId)
}

function getAutomaStatusText(instance) {
  const agent = getAgentByBrowserId(instance.id)
  if (!agent?.online) return '离线'
  if (!agent.automa_installed) return '不可用'
  return '已安装'
}

function getAutomaVersionText(instance) {
  const agent = getAgentByBrowserId(instance.id)
  if (!agent?.online || !agent.automa_installed) return ''
  return agent.automa_version || ''
}

function getAutomaTagType(instance) {
  const agent = getAgentByBrowserId(instance.id)
  if (!agent?.online) return 'info'
  return agent.automa_installed ? 'success' : 'warning'
}

function buildPayload() {
  return {
    ...form,
    launch_args: launchArgsText.value
      .split('\n')
      .map((item) => item.trim())
      .filter(Boolean),
  }
}

function markActiveInstances() {
  const currentId = status.value.current_instance_id
  const hasCurrent = Boolean(status.value.running && currentId)
  instances.value = instances.value.map((instance) => ({
    ...instance,
    // Active state 保留后端返回的运行状态，同时确保当前实例被标记为运行
    is_active: hasCurrent ? Boolean(instance.is_active || instance.id === currentId) : false,
    is_current: Boolean(hasCurrent && instance.id === currentId),
  }))
}

function normalizeForm(instance) {
  return {
    ...createEmptyForm(),
    ...instance,
    headless: Boolean(instance.headless),
    no_sandbox: Boolean(instance.no_sandbox),
  }
}

function createEmptyForm() {
  return {
    id: '',
    name: '',
    description: '',
    is_default: false,
    type: 'local',
    bin_path: '',
    user_data_dir: '',
    control_url: '',
    proxy: '',
    headless: false,
    no_sandbox: false,
    launch_args: [],
  }
}
</script>

<style scoped>
.browser-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  flex-shrink: 0;
}

.page-toolbar > div:first-child:not(.status-card) {
  display: none;
}

.status-card {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 10px;
  min-width: 260px;
  color: #606266;
}

.browser-layout {
  display: grid;
  grid-template-columns: 700px minmax(0, 1fr);
  flex: 1;
  gap: 16px;
  min-height: 0;
}

.instance-panel,
.config-panel {
  min-width: 0;
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.instance-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.instance-panel :deep(.el-table) {
  flex: 1;
  min-height: 0;
}

.instance-panel :deep(.el-table .cell) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  color: #303133;
  font-weight: 700;
  border-bottom: 1px solid #e4e7ed;
}

.panel-actions {
  display: flex;
  gap: 8px;
}

.instance-name {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.instance-name__text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instance-name :deep(.el-tag) {
  flex-shrink: 0;
}

.automa-status {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  min-width: 0;
  line-height: 1;
}

.automa-version {
  max-width: 88px;
  overflow: hidden;
  color: #909399;
  font-size: 11px;
  line-height: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instance-panel :deep(.el-button + .el-button) {
  margin-left: 6px;
}

.instance-panel :deep(.operation-column .cell) {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: visible;
  text-overflow: clip;
}

.instance-panel :deep(.operation-column .el-button) {
  flex-shrink: 0;
}

.table-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 16px;
  border-bottom: 1px solid #e4e7ed;
}

.instance-pagination {
  display: flex;
  justify-content: center;
  flex-shrink: 0;
  padding: 10px 12px;
  margin-top: 0;
  border-top: 1px solid #e4e7ed;
}

.instance-pagination :deep(.el-pagination) {
  justify-content: center;
  flex-wrap: nowrap;
}

.config-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.browser-form {
  flex-shrink: 0;
  min-width: 0;
  padding: 18px 18px 0;
}

.browser-form :deep(.el-form-item) {
  margin-bottom: 14px;
}

.switch-row {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
}

.runtime-panel {
  flex: 1;
  min-height: 0;
  padding: 0 18px 18px;
  overflow: hidden;
}

.runtime-panel h2 {
  margin: 8px 0 12px;
  color: #303133;
  font-size: 18px;
}

.runtime-panel :deep(.el-descriptions__table) {
  table-layout: fixed;
  width: 100%;
}

.runtime-panel :deep(.el-descriptions__label) {
  width: 112px;
  white-space: nowrap;
}

.runtime-value {
  display: block;
  flex: 1;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.runtime-copy {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.runtime-copy :deep(.el-button) {
  flex-shrink: 0;
}

@media (max-width: 1100px) {
  .browser-page {
    height: auto;
    overflow: visible;
  }

  .browser-layout {
    grid-template-columns: 1fr;
  }

  .page-toolbar {
    justify-content: flex-start;
  }
}
</style>
