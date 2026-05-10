<template>
  <section class="llm-config-page">
    <header class="page-header">
      <div>
        <h1>大模型配置</h1>
        <p>配置名称、提供商、模型名称和 API Key；Base URL 可留空，由后端按提供商自动选择。</p>
      </div>
      <el-button type="primary" :icon="RefreshRight" @click="loadAll">刷新</el-button>
    </header>

    <main class="config-layout">
      <section class="config-form-panel">
        <div class="panel-title">新增配置</div>
        <el-form label-width="96px" :model="configForm">
          <el-form-item label="配置名称">
            <el-input v-model="configForm.name" placeholder="例如：deepseek" />
          </el-form-item>
          <el-form-item label="提供商">
            <el-select v-model="configForm.provider" filterable @change="handleProviderChange">
              <el-option
                v-for="provider in providerCatalog"
                :key="provider.id"
                :label="provider.name"
                :value="provider.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模型名称">
            <el-input v-model="configForm.model" placeholder="请输入模型名称，例如：deepseek-chat" />
          </el-form-item>
          <el-form-item label="API Key">
            <el-input v-model="configForm.api_key" show-password placeholder="Ollama 可不填" />
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input v-model="configForm.base_url" placeholder="可不填，不填时后端使用提供商默认地址" />
          </el-form-item>
          <el-form-item>
            <el-checkbox v-model="configForm.is_default">设为默认</el-checkbox>
            <el-checkbox v-model="configForm.is_active">启用</el-checkbox>
          </el-form-item>
          <el-form-item>
            <el-button :loading="testing" @click="handleTestConfig">测试连接</el-button>
            <el-button type="primary" :loading="saving" @click="handleSaveConfig">保存配置</el-button>
          </el-form-item>
        </el-form>
      </section>

      <section class="config-list-panel">
        <div class="panel-title">配置列表</div>

        <div class="config-filters">
          <!-- Name filter 名称筛选，按配置名称模糊检索 -->
          <div class="filter-item filter-item--name">
            <span class="filter-label">名称：</span>
            <el-input v-model="searchKeyword" clearable placeholder="请输入配置名称" />
          </div>

          <!-- Provider filter 提供商筛选，单选过滤模型提供商 -->
          <div class="filter-item filter-item--provider">
            <span class="filter-label">模型提供商：</span>
            <el-select v-model="providerFilter" clearable filterable placeholder="全部提供商">
              <el-option label="全部" value="" />
              <el-option
                v-for="provider in providerCatalog"
                :key="provider.id"
                :label="provider.name"
                :value="provider.id"
              />
            </el-select>
          </div>

          <!-- Default filter 默认筛选，按是否默认过滤配置 -->
          <div class="filter-item filter-item--default">
            <span class="filter-label">是否默认：</span>
            <el-select v-model="defaultFilter" clearable placeholder="全部">
              <el-option label="全部" value="" />
              <el-option label="默认" value="default" />
              <el-option label="非默认" value="normal" />
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

        <div class="table-toolbar">
          <el-checkbox
            :model-value="isAllConfigsSelected"
            :indeterminate="isConfigSelectionIndeterminate"
            :disabled="filteredConfigs.length === 0"
            @change="handleToggleAllConfigs"
          >
            全选
          </el-checkbox>
          <el-button
            link
            type="danger"
            :disabled="selectedConfigIds.length === 0"
            @click="handleDeleteSelectedConfigs"
          >
            删除选中
          </el-button>
        </div>

        <el-table
          v-loading="loading"
          class="config-table adaptive-table"
          :data="pagedConfigs"
          border
          height="100%"
          empty-text="暂无配置"
        >
          <el-table-column width="40" align="center" class-name="action-column">
            <template #default="{ row }">
              <el-checkbox
                :model-value="selectedConfigIds.includes(row.id)"
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
          <el-table-column label="提供商" width="90" class-name="ellipsis-column" show-overflow-tooltip>
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
          <el-table-column label="默认" width="70" align="center" class-name="action-column">
            <template #default="{ row }">
              <el-tag v-if="row.is_default" type="success" effect="plain">默认</el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center" class-name="action-column">
            <template #default="{ row }">
              <el-switch
                :model-value="row.is_active"
                :loading="isStatusUpdating(row.id)"
                active-text="启用"
                inactive-text="停用"
                inline-prompt
                @change="(checked) => handleToggleConfigStatus(row, checked)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="70" align="center" class-name="action-column">
            <template #default="{ row }">
              <el-button link type="danger" @click="handleDeleteConfig(row.id)">删除</el-button>
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
import { RefreshRight } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
import {
  createLLMConfig,
  deleteLLMConfig,
  deleteLLMConfigs,
  listLLMConfigs,
  listLLMProviders,
  testLLMConfig,
  updateLLMConfig,
} from '@/services/llmChat'

const configs = ref([])
const providerCatalog = ref([])
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const pageSizes = [10, 30, 60]
const searchKeyword = ref('')
const providerFilter = ref('')
const defaultFilter = ref('')
const statusFilter = ref('')
const selectedConfigIds = ref([])
const statusUpdatingIds = ref([])

const configForm = reactive(createEmptyForm())

const filteredConfigs = computed(() => {
  const keyword = searchKeyword.value.trim().toLocaleLowerCase()
  return configs.value.filter((config) => {
    const name = config.name || ''
    const isNameMatched = !keyword || name.toLocaleLowerCase().includes(keyword)
    const isProviderMatched = !providerFilter.value || config.provider === providerFilter.value
    const isDefaultMatched =
      !defaultFilter.value ||
      (defaultFilter.value === 'default' && config.is_default) ||
      (defaultFilter.value === 'normal' && !config.is_default)
    const isStatusMatched =
      !statusFilter.value ||
      (statusFilter.value === 'active' && config.is_active) ||
      (statusFilter.value === 'inactive' && !config.is_active)
    return isNameMatched && isProviderMatched && isDefaultMatched && isStatusMatched
  })
})

const pagedConfigs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredConfigs.value.slice(start, start + pageSize.value)
})
const isAllConfigsSelected = computed(
  () => filteredConfigs.value.length > 0 && selectedConfigIds.value.length === filteredConfigs.value.length,
)
const isConfigSelectionIndeterminate = computed(
  () => selectedConfigIds.value.length > 0 && selectedConfigIds.value.length < filteredConfigs.value.length,
)

onMounted(() => {
  loadAll()
})

watch([filteredConfigs, pageSize], () => {
  currentPage.value = getSafePage({
    total: filteredConfigs.value.length,
    page: currentPage.value,
    size: pageSize.value,
  })
})

watch([searchKeyword, providerFilter, defaultFilter, statusFilter], () => {
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
    configs.value.some((config) => config.id === id),
  )
}

function handleProviderChange() {
  const provider = providerCatalog.value.find((item) => item.id === configForm.provider)
  if (provider && !configForm.name) {
    configForm.name = provider.name
  }
}

async function handleSaveConfig() {
  saving.value = true
  try {
    await createLLMConfig({ ...configForm })
    appMessage({ type: APP_MESSAGE_TYPE.success, message: '保存成功' })
    resetConfigForm()
    await loadConfigs()
  } finally {
    saving.value = false
  }
}

async function handleTestConfig() {
  testing.value = true
  try {
    const data = await testLLMConfig({ ...configForm })
    if (data.success === false) {
      appMessage({ type: APP_MESSAGE_TYPE.error, message: data.message || '连接失败' })
      return
    }
    appMessage({ type: APP_MESSAGE_TYPE.success, message: data.message || '连接成功' })
  } finally {
    testing.value = false
  }
}

async function handleDeleteConfig(configId) {
  const confirmed = await appConfirm({
    title: '删除配置',
    message: '确认删除这个模型配置吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteLLMConfig(configId)
  removeConfigsFromState([configId])
  await loadConfigs()
}

async function handleToggleConfigStatus(config, checked) {
  const nextActive = Boolean(checked)
  const previousActive = config.is_active
  statusUpdatingIds.value = Array.from(new Set([...statusUpdatingIds.value, config.id]))
  config.is_active = nextActive

  try {
    // Status update 状态更新，复用完整配置更新接口避免新增后端路由
    await updateLLMConfig(config.id, { ...config, is_active: nextActive })
    appMessage({ type: APP_MESSAGE_TYPE.success, message: nextActive ? '已启用' : '已停用' })
  } catch (error) {
    config.is_active = previousActive
    throw error
  } finally {
    statusUpdatingIds.value = statusUpdatingIds.value.filter((id) => id !== config.id)
  }
}

function handleToggleConfig(configId, checked) {
  if (checked) {
    selectedConfigIds.value = Array.from(new Set([...selectedConfigIds.value, configId]))
    return
  }
  selectedConfigIds.value = selectedConfigIds.value.filter((id) => id !== configId)
}

function handleToggleAllConfigs(checked) {
  selectedConfigIds.value = checked ? filteredConfigs.value.map((config) => config.id) : []
}

async function handleDeleteSelectedConfigs() {
  const ids = selectedConfigIds.value.slice()
  if (ids.length === 0) return

  const confirmed = await appConfirm({
    title: '批量删除配置',
    message: `确认删除选中的 ${ids.length} 个模型配置吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

  await deleteLLMConfigs(ids)
  removeConfigsFromState(ids)
  await loadConfigs()
}

function resetFilters() {
  searchKeyword.value = ''
  providerFilter.value = ''
  defaultFilter.value = ''
  statusFilter.value = ''
  currentPage.value = 1
}

function removeConfigsFromState(configIds) {
  configs.value = configs.value.filter((config) => !configIds.includes(config.id))
  selectedConfigIds.value = selectedConfigIds.value.filter((id) => !configIds.includes(id))
}

function sortByCreatedDesc(data) {
  // Created order 新增时间倒序，保证列表展示最新配置在前
  return data.slice().sort((a, b) => getTimeValue(b.created_at) - getTimeValue(a.created_at))
}

function getTimeValue(value) {
  return value ? new Date(value).getTime() || 0 : 0
}

function resetConfigForm() {
  Object.assign(configForm, createEmptyForm())
}

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function getProviderName(providerId) {
  return providerCatalog.value.find((provider) => provider.id === providerId)?.name || providerId || '-'
}

function isStatusUpdating(configId) {
  return statusUpdatingIds.value.includes(configId)
}

function createEmptyForm() {
  return {
    name: '',
    provider: 'openai',
    api_key: '',
    model: '',
    base_url: '',
    is_default: false,
    is_active: true,
  }
}
</script>

<style scoped>
.llm-config-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: min(100%, 1600px);
  height: 100%;
  margin: 0 auto;
  min-height: 0;
  overflow: hidden;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
}

.page-header > div:first-child {
  display: none;
}

.config-layout {
  display: grid;
  grid-template-columns: 420px minmax(0, 1fr);
  flex: 1;
  gap: 16px;
  min-height: 0;
}

.config-form-panel,
.config-list-panel {
  min-width: 0;
  padding: 16px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.config-form-panel {
  overflow-y: auto;
}

.config-list-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
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

.panel-title {
  margin-bottom: 16px;
  color: #303133;
  font-size: 18px;
  font-weight: 700;
}

.config-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 16px;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item--name {
  width: 320px;
}

.filter-item--provider {
  width: 300px;
}

.filter-item--default {
  width: 220px;
}

.filter-item--status {
  width: 180px;
}

.filter-label {
  flex-shrink: 0;
  color: #606266;
}

.reset-button {
  margin-left: auto;
}

.table-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
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
  .config-filters,
  .filter-item {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-item--name,
  .filter-item--provider,
  .filter-item--default,
  .filter-item--status {
    width: 100%;
  }

  .reset-button {
    margin-left: 0;
  }
}
</style>
