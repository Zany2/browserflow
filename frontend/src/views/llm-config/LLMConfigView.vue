<template>
  <section class="llm-config-page windows-workspace-page">
    <header class="page-header windows-workspace-actions">
      <div>
        <h1>大模型配置</h1>
        <p>
          配置名称、提供商、模型名称和 API Key；Base URL
          可留空，由后端按提供商自动选择。
        </p>
      </div>
      <el-button type="primary" :icon="RefreshRight" @click="loadAll"
        >刷新</el-button
      >
    </header>

    <main class="config-layout windows-workspace-layout">
      <section class="config-form-panel windows-workspace-panel">
        <div class="panel-title windows-workspace-panel__header">
          <span>{{ configForm.id ? '编辑配置' : '新增配置' }}</span>
          <el-button type="primary" :icon="Plus" @click="handleNewConfig"
            >新建</el-button
          >
        </div>
        <el-form label-width="96px" :model="configForm">
          <el-form-item label="配置名称">
            <el-input v-model="configForm.name" placeholder="例如：deepseek" />
          </el-form-item>
          <el-form-item label="提供商">
            <el-select
              v-model="configForm.provider"
              filterable
              @change="handleProviderChange"
            >
              <el-option
                v-for="provider in providerCatalog"
                :key="provider.id"
                :label="provider.name"
                :value="provider.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模型名称">
            <el-input
              v-model="configForm.model"
              placeholder="请输入模型名称，例如：deepseek-chat"
            />
          </el-form-item>
          <el-form-item label="API Key">
            <el-input
              v-model="configForm.api_key"
              type="password"
              show-password
              autocomplete="new-password"
              placeholder="Ollama 可不填"
            />
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input
              v-model="configForm.base_url"
              placeholder="可不填，不填时后端使用提供商默认地址"
            />
          </el-form-item>
          <el-form-item>
            <el-checkbox v-model="configForm.is_default">设为默认</el-checkbox>
            <el-checkbox v-model="configForm.is_active">启用</el-checkbox>
          </el-form-item>
          <el-form-item>
            <el-button :loading="testing" @click="handleTestConfig"
              >测试连接</el-button
            >
            <el-button
              type="primary"
              :loading="saving"
              @click="handleSaveConfig"
              >{{ saveButtonText }}</el-button
            >
          </el-form-item>
        </el-form>
      </section>

      <section
        class="config-list-panel windows-workspace-panel windows-workspace-panel--stack"
      >
        <div class="panel-title windows-workspace-panel__header">配置列表</div>

        <div class="config-filters windows-workspace-filters">
          <!-- Name filter 名称筛选，按配置名称模糊检索 -->
          <div class="filter-item filter-item--name">
            <span class="filter-label">名称：</span>
            <el-input
              v-model="searchKeyword"
              clearable
              placeholder="请输入配置名称"
            />
          </div>

          <!-- Provider filter 提供商筛选，单选过滤模型提供商 -->
          <div class="filter-item filter-item--provider">
            <span class="filter-label">模型提供商：</span>
            <el-select
              v-model="providerFilter"
              clearable
              filterable
              placeholder="全部提供商"
            >
              <el-option label="全部" value="" />
              <el-option
                v-for="provider in providerCatalog"
                :key="provider.id"
                :label="provider.name"
                :value="provider.id"
              />
            </el-select>
          </div>

          <!-- Status filter 状态筛选，按启用状态过滤配置 -->
          <div class="filter-item filter-item--status">
            <span class="filter-label">状态：</span>
            <el-select v-model="statusFilter" clearable placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="启用" value="active" />
              <el-option label="停用" value="inactive" />
            </el-select>
          </div>

          <el-button class="reset-button" @click="resetFilters">重置</el-button>
        </div>

        <div class="table-toolbar windows-workspace-selection">
          <el-button
            link
            type="danger"
            :disabled="selectedConfigIds.length === 0"
            @click="handleDeleteSelectedConfigs"
          >
            删除选中
          </el-button>
          <AppSelectionSummary :count="selectedConfigIds.length" unit="配置" />
        </div>

        <el-table
          v-loading="loading"
          class="config-table windows-workspace-table adaptive-table"
          :data="pagedConfigs"
          border
          highlight-current-row
          height="100%"
          :current-row-key="selectedConfigId"
          empty-text="暂无配置"
          row-key="id"
          @row-click="selectConfig"
        >
          <el-table-column width="40" align="center" class-name="action-column">
            <template #header>
              <el-checkbox
                :model-value="isAllConfigsSelected"
                :indeterminate="isConfigSelectionIndeterminate"
                :disabled="filteredConfigs.length === 0"
                @change="handleToggleAllConfigs"
              />
            </template>
            <template #default="{ row }">
              <el-checkbox
                :model-value="selectedConfigIds.includes(row.id)"
                @click.stop
                @change="(checked) => handleToggleConfig(row.id, checked)"
              />
            </template>
          </el-table-column>
          <el-table-column
            prop="name"
            label="名称"
            min-width="110"
            class-name="ellipsis-column"
            show-overflow-tooltip
          />
          <el-table-column
            label="提供商"
            width="90"
            class-name="ellipsis-column"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              {{ getProviderName(row.provider) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="model"
            label="模型名称"
            min-width="110"
            class-name="ellipsis-column"
            show-overflow-tooltip
          />
          <el-table-column
            prop="base_url"
            label="Base URL"
            min-width="140"
            class-name="ellipsis-column"
            show-overflow-tooltip
          />
          <el-table-column
            label="默认"
            width="96"
            align="center"
            class-name="action-column default-column"
          >
            <template #default="{ row }">
              <div
                :key="`${row.id}-${Boolean(row.is_default)}-${isDefaultUpdating(row.id)}`"
                class="quick-edit-cell"
              >
                <span
                  v-if="row.is_default"
                  class="quick-edit-action quick-edit-badge"
                  >默认</span
                >
                <button
                  v-else
                  class="quick-edit-action quick-edit-button"
                  type="button"
                  :disabled="isDefaultUpdating(row.id)"
                  @click.stop="handleSetDefaultConfig(row)"
                >
                  {{ isDefaultUpdating(row.id) ? '设置中' : '设为默认' }}
                </button>
              </div>
            </template>
          </el-table-column>
          <el-table-column
            label="状态"
            width="86"
            align="center"
            class-name="action-column quick-edit-column"
          >
            <template #default="{ row }">
              <el-switch
                class="quick-edit-switch"
                :model-value="row.is_active"
                :loading="isStatusUpdating(row.id)"
                @click.stop
                @change="(checked) => handleToggleConfigStatus(row, checked)"
              />
            </template>
          </el-table-column>
          <el-table-column
            label="操作"
            width="70"
            align="center"
            class-name="action-column"
          >
            <template #default="{ row }">
              <el-button
                link
                type="danger"
                @click.stop="handleDeleteConfig(row.id)"
                >删除</el-button
              >
            </template>
          </el-table-column>
        </el-table>

        <AppPagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="pageSizes"
          :total="filteredConfigs.length"
        />
      </section>
    </main>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Plus, RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { DEFAULT_PAGE_SIZES, getSafePage } from '@/utils/list'
import {
  createLLMConfig,
  deleteLLMConfig,
  deleteLLMConfigs,
  listLLMConfigs,
  listLLMProviders,
  testLLMConfig,
  updateLLMConfig
} from '@/services/llmChat'

const configs = ref([])
const providerCatalog = ref([])
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = DEFAULT_PAGE_SIZES
const searchKeyword = ref('')
const providerFilter = ref('')
const statusFilter = ref('')
const selectedConfigId = ref('')
const selectedConfigIds = ref([])
const defaultUpdatingIds = ref([])
const statusUpdatingIds = ref([])

const configForm = reactive(createEmptyForm())
const saveButtonText = computed(() => (configForm.id ? '保存修改' : '新增配置'))

const filteredConfigs = computed(() => {
  const keyword = searchKeyword.value.trim().toLocaleLowerCase()
  return configs.value.filter((config) => {
    const name = config.name || ''
    const isNameMatched = !keyword || name.toLocaleLowerCase().includes(keyword)
    const isProviderMatched =
      !providerFilter.value || config.provider === providerFilter.value
    const isStatusMatched =
      !statusFilter.value ||
      (statusFilter.value === 'active' && config.is_active) ||
      (statusFilter.value === 'inactive' && !config.is_active)
    return isNameMatched && isProviderMatched && isStatusMatched
  })
})

const pagedConfigs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredConfigs.value.slice(start, start + pageSize.value)
})
const isAllConfigsSelected = computed(
  () =>
    filteredConfigs.value.length > 0 &&
    selectedConfigIds.value.length === filteredConfigs.value.length
)
const isConfigSelectionIndeterminate = computed(
  () =>
    selectedConfigIds.value.length > 0 &&
    selectedConfigIds.value.length < filteredConfigs.value.length
)

onMounted(() => {
  loadAll()
})

watch([filteredConfigs, pageSize], () => {
  currentPage.value = getSafePage({
    total: filteredConfigs.value.length,
    page: currentPage.value,
    size: pageSize.value
  })
})

watch([searchKeyword, providerFilter, statusFilter], () => {
  currentPage.value = 1
})

async function loadAll() {
  loading.value = true
  try {
    await Promise.all([loadProviders(), loadConfigs()])
  } finally {
    loading.value = false
  }
}

async function loadProviders() {
  const data = await listLLMProviders()
  providerCatalog.value = data.providers || []
}

async function loadConfigs() {
  const data = await listLLMConfigs()
  configs.value = sortByCreatedDesc(data.configs || [])
  selectedConfigIds.value = selectedConfigIds.value.filter((id) =>
    configs.value.some((config) => config.id === id)
  )
  if (
    selectedConfigId.value &&
    !configs.value.some((config) => config.id === selectedConfigId.value)
  ) {
    resetConfigForm()
  }
}

function handleProviderChange() {
  const provider = providerCatalog.value.find(
    (item) => item.id === configForm.provider
  )
  if (provider && !configForm.name) {
    configForm.name = provider.name
  }
}

async function handleSaveConfig() {
  saving.value = true
  try {
    const isCreate = !configForm.id
    const data = isCreate
      ? await createLLMConfig(buildPayload())
      : await updateLLMConfig(configForm.id, buildPayload())
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: isCreate ? '新增成功' : '保存成功'
    })
    await loadConfigs()
    selectSavedConfig(data?.config || configForm)
  } finally {
    saving.value = false
  }
}

async function handleTestConfig() {
  testing.value = true
  try {
    const data = await testLLMConfig(buildPayload())
    if (data.success === false) {
      appMessage({
        type: APP_MESSAGE_TYPE.error,
        message: data.message || '连接失败'
      })
      return
    }
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: data.message || '连接成功'
    })
  } finally {
    testing.value = false
  }
}

async function handleDeleteConfig(configId) {
  const confirmed = await appConfirm({
    title: '删除配置',
    message: '确认删除这个模型配置吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除'
  })
  if (!confirmed) return

  await deleteLLMConfig(configId)
  removeConfigsFromState([configId])
  if (configForm.id === configId) {
    resetConfigForm()
  }
  await loadConfigs()
}

async function handleToggleConfigStatus(config, checked) {
  const nextActive = Boolean(checked)
  const previousActive = config.is_active
  statusUpdatingIds.value = Array.from(
    new Set([...statusUpdatingIds.value, config.id])
  )
  config.is_active = nextActive

  try {
    // Status update 状态更新，复用完整配置更新接口避免新增后端路由
    await updateLLMConfig(
      config.id,
      buildConfigPayload(config, { is_active: nextActive })
    )
    if (configForm.id === config.id) {
      configForm.is_active = nextActive
    }
    appMessage({
      type: APP_MESSAGE_TYPE.success,
      message: nextActive ? '已启用' : '已停用'
    })
  } catch (error) {
    config.is_active = previousActive
    if (configForm.id === config.id) {
      configForm.is_active = previousActive
    }
    throw error
  } finally {
    statusUpdatingIds.value = statusUpdatingIds.value.filter(
      (id) => id !== config.id
    )
  }
}

async function handleSetDefaultConfig(config) {
  if (config.is_default) return

  defaultUpdatingIds.value = Array.from(
    new Set([...defaultUpdatingIds.value, config.id])
  )
  try {
    // Default update 默认配置只允许一个，后端保存时会清理其他默认项
    await updateLLMConfig(
      config.id,
      buildConfigPayload(config, { is_default: true })
    )
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '已设为默认模型' })
    await loadConfigs()
    syncSelectedConfigFlags()
  } finally {
    defaultUpdatingIds.value = defaultUpdatingIds.value.filter(
      (id) => id !== config.id
    )
  }
}

function handleToggleConfig(configId, checked) {
  if (checked) {
    selectedConfigIds.value = Array.from(
      new Set([...selectedConfigIds.value, configId])
    )
    return
  }
  selectedConfigIds.value = selectedConfigIds.value.filter(
    (id) => id !== configId
  )
}

function handleToggleAllConfigs(checked) {
  selectedConfigIds.value = checked
    ? filteredConfigs.value.map((config) => config.id)
    : []
}

async function handleDeleteSelectedConfigs() {
  const ids = selectedConfigIds.value.slice()
  if (ids.length === 0) return

  const confirmed = await appConfirm({
    title: '批量删除配置',
    message: `确认删除选中的 ${ids.length} 个模型配置吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除'
  })
  if (!confirmed) return

  await deleteLLMConfigs(ids)
  removeConfigsFromState(ids)
  if (ids.includes(configForm.id)) {
    resetConfigForm()
  }
  await loadConfigs()
}

function resetFilters() {
  searchKeyword.value = ''
  providerFilter.value = ''
  statusFilter.value = ''
  currentPage.value = 1
}

function removeConfigsFromState(configIds) {
  configs.value = configs.value.filter(
    (config) => !configIds.includes(config.id)
  )
  selectedConfigIds.value = selectedConfigIds.value.filter(
    (id) => !configIds.includes(id)
  )
}

function sortByCreatedDesc(data) {
  // Created order 新增时间倒序，保证列表展示最新配置在前
  return data
    .slice()
    .sort((a, b) => getTimeValue(b.created_at) - getTimeValue(a.created_at))
}

function getTimeValue(value) {
  return value ? new Date(value).getTime() || 0 : 0
}

function selectConfig(row) {
  selectedConfigId.value = row.id
  Object.assign(configForm, normalizeConfigForm(row))
}

function selectSavedConfig(config) {
  // Saved selection 保存后回到已持久化配置，保持表格高亮与表单一致
  if (!config?.id) return
  const current = configs.value.find((item) => item.id === config.id) || config
  if (current?.id) {
    selectConfig(current)
  }
}

function handleNewConfig() {
  resetConfigForm()
}

function resetConfigForm() {
  selectedConfigId.value = ''
  Object.assign(configForm, createEmptyForm())
}

function getProviderName(providerId) {
  return (
    providerCatalog.value.find((provider) => provider.id === providerId)
      ?.name ||
    providerId ||
    ''
  )
}

function isStatusUpdating(configId) {
  return statusUpdatingIds.value.includes(configId)
}

function isDefaultUpdating(configId) {
  return defaultUpdatingIds.value.includes(configId)
}

function buildPayload() {
  // Payload whitelist 表单入参白名单，避免列表时间字段回传
  return buildConfigPayload(configForm)
}

function buildConfigPayload(config, overrides = {}) {
  const nextConfig = { ...config, ...overrides }
  return {
    id: nextConfig.id,
    name: nextConfig.name,
    provider: nextConfig.provider,
    api_key: nextConfig.api_key,
    model: nextConfig.model,
    base_url: nextConfig.base_url,
    is_default: Boolean(nextConfig.is_default),
    is_active: Boolean(nextConfig.is_active)
  }
}

function syncSelectedConfigFlags() {
  if (!configForm.id) return

  const current = configs.value.find((config) => config.id === configForm.id)
  if (!current) return

  configForm.is_default = Boolean(current.is_default)
  configForm.is_active = Boolean(current.is_active)
}

function normalizeConfigForm(config) {
  return {
    ...createEmptyForm(),
    ...config,
    is_default: Boolean(config.is_default),
    is_active: Boolean(config.is_active)
  }
}

function createEmptyForm() {
  return {
    id: '',
    name: '',
    provider: 'openai',
    api_key: '',
    model: '',
    base_url: '',
    is_default: false,
    is_active: true
  }
}
</script>

<style scoped>
.page-header {
  gap: 16px;
}

.page-header > div:first-child {
  display: none;
}

.config-layout {
  display: grid;
  grid-template-columns: 420px minmax(0, 1fr);
}

.config-form-panel,
.config-list-panel {
}

.config-form-panel {
  padding: 16px;
  overflow-y: auto;
}

.config-list-panel {
  padding: 16px;
}

.config-table {
  flex: 1;
  min-height: 0;
  width: 100%;
}

.config-table :deep(.ellipsis-column .cell) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.config-table :deep(.action-column .cell) {
  overflow: visible;
  text-overflow: clip;
  white-space: nowrap;
}

.config-table :deep(.default-column .cell) {
  display: flex;
  justify-content: center;
  padding: 0 8px;
}

.panel-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 16px;
}

.config-table :deep(.el-table__row) {
  cursor: pointer;
}

.config-filters {
  align-items: center;
  gap: 12px;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item--name {
  width: clamp(260px, 32%, 320px);
}

.filter-item--provider {
  width: clamp(260px, 28%, 300px);
}

.filter-item--status {
  width: 184px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
  white-space: nowrap;
}

.filter-item :deep(.el-input),
.filter-item :deep(.el-select) {
  flex: 1;
  min-width: 0;
}

.reset-button {
  flex-shrink: 0;
  margin-left: 0;
}

.table-toolbar {
  margin-bottom: 12px;
}

.config-list-panel > :deep(.app-pagination) {
  flex-shrink: 0;
}

@media (max-width: 1100px) {
  .llm-config-page {
    height: auto;
    overflow: visible;
  }

  .config-layout {
    grid-template-columns: 1fr;
  }

  .page-header {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .config-filters {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item {
    align-items: center;
  }

  .filter-item--name,
  .filter-item--provider,
  .filter-item--status {
    width: 100%;
  }

  .reset-button {
    width: 100%;
  }
}
</style>
