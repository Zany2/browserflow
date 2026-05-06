<template>
  <section class="llm-chat-page">
    <aside class="session-panel">
      <div class="panel-header">
        <h1>大模型对话</h1>
        <el-button type="primary" :icon="Plus" @click="handleCreateSession">新建</el-button>
      </div>

      <div class="config-bar">
        <el-select v-model="selectedConfigId" placeholder="选择模型" filterable>
          <el-option
            v-for="config in activeConfigs"
            :key="config.id"
            :label="getConfigLabel(config)"
            :value="config.id"
          />
        </el-select>
        <el-button :icon="Setting" @click="configDialogVisible = true" />
      </div>

      <div class="session-toolbar">
        <el-checkbox
          :model-value="isAllSessionsSelected"
          :indeterminate="isSessionSelectionIndeterminate"
          :disabled="sessions.length === 0"
          @change="handleToggleAllSessions"
        >
          全选
        </el-checkbox>
        <el-button
          link
          type="danger"
          :disabled="selectedSessionIds.length === 0"
          @click="handleDeleteSelectedSessions"
        >
          删除选中
        </el-button>
      </div>

      <div class="session-list">
        <button
          v-for="session in pagedSessions"
          :key="session.id"
          class="session-item"
          :class="{ 'is-active': currentSession?.id === session.id }"
          type="button"
          @click="currentSession = session"
        >
          <el-checkbox
            class="session-checkbox"
            :model-value="selectedSessionIds.includes(session.id)"
            @click.stop
            @change="(checked) => handleToggleSession(session.id, checked)"
          />
          <span class="session-title">{{ getSessionTitle(session) }}</span>
          <span class="session-meta">{{ session.messages?.length || 0 }} 条消息</span>
          <el-button
            class="session-delete"
            link
            type="danger"
            :icon="Delete"
            @click.stop="handleDeleteSession(session.id)"
          />
        </button>
      </div>

      <AppPagination
        v-model:current-page="sessionCurrentPage"
        v-model:page-size="sessionPageSize"
        class="session-pagination"
        :page-sizes="sessionPageSizes"
        :total="sessions.length"
        layout="prev, pager, next, sizes"
      />
    </aside>

    <main class="chat-panel">
      <header class="chat-header">
        <div>
          <h2>{{ currentModelName }}</h2>
          <p>仿照 BrowserWing 的会话体验，支持模型配置、会话持久化和流式输出。</p>
        </div>
        <el-button :icon="Setting" @click="configDialogVisible = true">模型配置</el-button>
      </header>

      <div ref="messageListRef" class="message-list">
        <el-empty v-if="!currentSession" description="请选择或新建一个会话" />
        <template v-else>
          <div
            v-for="message in currentSession.messages"
            :key="message.id"
            class="message-row"
            :class="`is-${message.role}`"
          >
            <div class="message-bubble">
              <div class="message-role">{{ message.role === 'user' ? '我' : 'AI' }}</div>
              <div class="message-content">{{ message.content }}</div>
              <div class="message-time">{{ formatTime(message.timestamp) }}</div>
            </div>
          </div>
          <div v-if="streaming" class="message-row is-assistant">
            <div class="message-bubble">
              <div class="typing-dot"></div>
              <span>正在生成...</span>
            </div>
          </div>
        </template>
      </div>

      <footer class="chat-input-bar">
        <el-input
          v-model="inputMessage"
          class="chat-input"
          type="textarea"
          :rows="3"
          resize="none"
          placeholder="输入要对话的内容，Enter 发送，Shift+Enter 换行"
          :disabled="!currentSession || streaming"
          @keydown.enter="handleInputEnter"
        />
        <el-button
          class="send-button"
          type="primary"
          :icon="Promotion"
          :loading="streaming"
          :disabled="!canSend"
          @click="handleSendMessage"
        >
          发送
        </el-button>
      </footer>
    </main>

    <AppDialog v-model="configDialogVisible" title="大模型配置" width="720px">
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
          <el-input v-model="configForm.base_url" placeholder="可不填，不填时后端根据提供商使用默认 Base URL" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="configForm.is_default">设为默认</el-checkbox>
          <el-checkbox v-model="configForm.is_active">启用</el-checkbox>
        </el-form-item>
      </el-form>

      <el-table class="config-table adaptive-table" :data="configs" border>
        <el-table-column prop="name" label="名称" min-width="110" />
        <el-table-column prop="provider" label="提供商" width="90" />
        <el-table-column prop="model" label="模型" min-width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'">
              {{ row.is_active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70">
          <template #default="{ row }">
            <el-button link type="danger" @click="handleDeleteConfig(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="configDialogVisible = false">关闭</el-button>
        <el-button @click="handleTestConfig">测试连接</el-button>
        <el-button type="primary" @click="handleSaveConfig">保存配置</el-button>
      </template>
    </AppDialog>
  </section>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { Delete, Plus, Promotion, Setting } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppDialog from '@/components/AppDialog.vue'
import AppPagination from '@/components/AppPagination.vue'
import {
  createChatSession,
  createLLMConfig,
  deleteChatSession,
  deleteChatSessions,
  deleteLLMConfig,
  listChatSessions,
  listLLMConfigs,
  listLLMProviders,
  streamChatMessage,
  testLLMConfig,
} from '@/services/llmChat'

const configs = ref([])
const providerCatalog = ref([])
const sessions = ref([])
const currentSession = ref(null)
const selectedConfigId = ref('')
const selectedSessionIds = ref([])
const sessionCurrentPage = ref(1)
const sessionPageSize = ref(10)
const sessionPageSizes = [10, 30, 60]
const inputMessage = ref('')
const streaming = ref(false)
const configDialogVisible = ref(false)
const messageListRef = ref(null)

const configForm = reactive({
  name: '',
  provider: 'openai',
  api_key: '',
  model: '',
  base_url: '',
  is_default: false,
  is_active: true,
})

const activeConfigs = computed(() => configs.value.filter((config) => config.is_active))
const canSend = computed(() => Boolean(currentSession.value && inputMessage.value.trim() && !streaming.value))
const pagedSessions = computed(() => {
  const start = (sessionCurrentPage.value - 1) * sessionPageSize.value
  return sessions.value.slice(start, start + sessionPageSize.value)
})
const pagedSessionIds = computed(() => pagedSessions.value.map((session) => session.id))
const selectedPagedSessionIds = computed(() =>
  selectedSessionIds.value.filter((id) => pagedSessionIds.value.includes(id)),
)
const isAllSessionsSelected = computed(
  () => pagedSessionIds.value.length > 0 && selectedPagedSessionIds.value.length === pagedSessionIds.value.length,
)
const isSessionSelectionIndeterminate = computed(
  () => selectedPagedSessionIds.value.length > 0 && selectedPagedSessionIds.value.length < pagedSessionIds.value.length,
)
const currentProvider = computed(() =>
  providerCatalog.value.find((provider) => provider.id === configForm.provider),
)
const currentModelName = computed(() => {
  const config = configs.value.find((item) => item.id === currentSession.value?.llm_config_id)
  return config ? getConfigLabel(config) : '未选择模型'
})

watch(
  () => currentSession.value?.messages?.length,
  () => scrollToBottom(),
)

watch([sessions, sessionPageSize], () => {
  sessionCurrentPage.value = getSafePage({
    total: sessions.value.length,
    page: sessionCurrentPage.value,
    size: sessionPageSize.value,
  })
})

onMounted(async () => {
  await Promise.all([loadProviders(), loadConfigs(), loadSessions()])
})

async function loadProviders() {
  const data = await listLLMProviders()
  providerCatalog.value = data.providers || []
}

async function loadConfigs() {
  const data = await listLLMConfigs()
  configs.value = data.configs || []
  selectedConfigId.value =
    configs.value.find((config) => config.is_default && config.is_active)?.id ||
    activeConfigs.value[0]?.id ||
    ''
}

async function loadSessions(preferredSessionId = currentSession.value?.id) {
  const data = await listChatSessions()
  sessions.value = sortSessionsByUpdatedDesc(data.sessions || [])
  currentSession.value =
    sessions.value.find((session) => session.id === preferredSessionId) ||
    sessions.value[0] ||
    null
  selectedSessionIds.value = selectedSessionIds.value.filter((id) =>
    sessions.value.some((session) => session.id === id),
  )
}

async function handleCreateSession() {
  if (!selectedConfigId.value) {
    configDialogVisible.value = true
    ElMessage.warning('请先配置并启用大模型')
    return
  }

  const data = await createChatSession(selectedConfigId.value)
  sessions.value = sortSessionsByUpdatedDesc([data.session, ...sessions.value])
  sessionCurrentPage.value = 1
  currentSession.value = data.session
}

async function handleDeleteSession(sessionId) {
  await ElMessageBox.confirm('确认删除这个会话吗？', '删除会话', { type: 'warning' })
  await deleteChatSession(sessionId)
  removeSessionsFromState([sessionId])
}

function handleToggleSession(sessionId, checked) {
  if (checked) {
    selectedSessionIds.value = Array.from(new Set([...selectedSessionIds.value, sessionId]))
    return
  }
  selectedSessionIds.value = selectedSessionIds.value.filter((id) => id !== sessionId)
}

function handleToggleAllSessions(checked) {
  const pageIds = pagedSessionIds.value
  if (checked) {
    selectedSessionIds.value = Array.from(new Set([...selectedSessionIds.value, ...pageIds]))
    return
  }
  selectedSessionIds.value = selectedSessionIds.value.filter((id) => !pageIds.includes(id))
}

async function handleDeleteSelectedSessions() {
  const ids = selectedSessionIds.value.slice()
  if (ids.length === 0) return

  await ElMessageBox.confirm(`确认删除选中的 ${ids.length} 个会话吗？`, '批量删除会话', {
    type: 'warning',
  })
  await deleteChatSessions(ids)
  removeSessionsFromState(ids)
}

function removeSessionsFromState(sessionIds) {
  sessions.value = sessions.value.filter((session) => !sessionIds.includes(session.id))
  selectedSessionIds.value = selectedSessionIds.value.filter((id) => !sessionIds.includes(id))
  if (currentSession.value && sessionIds.includes(currentSession.value.id)) {
    currentSession.value = sessions.value[0] || null
  }
}

function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(total / size), 1)
  return Math.min(page, maxPage)
}

function sortSessionsByUpdatedDesc(data) {
  // Session order 会话排序，跟随后端 updated_at，兜底使用 created_at。
  return data.slice().sort((prev, next) => getSessionTime(next) - getSessionTime(prev))
}

function getSessionTime(session) {
  const value = session?.updated_at || session?.created_at
  return value ? new Date(value).getTime() || 0 : 0
}

async function handleSendMessage() {
  if (!currentSession.value || streaming.value) return
  if (!inputMessage.value.trim()) {
    ElMessage.warning('请输入对话内容')
    return
  }

  const messageText = inputMessage.value.trim()
  inputMessage.value = ''
  streaming.value = true

  const userMessage = {
    id: `local_user_${Date.now()}`,
    role: 'user',
    content: messageText,
    timestamp: new Date().toISOString(),
  }
  const assistantMessage = {
    id: `local_assistant_${Date.now()}`,
    role: 'assistant',
    content: '',
    timestamp: new Date().toISOString(),
  }

  currentSession.value.messages.push(userMessage, assistantMessage)
  scrollToBottom()

  const sessionId = currentSession.value.id
  try {
    await streamChatMessage(sessionId, messageText, (chunk) => {
      if (chunk.type === 'message') {
        assistantMessage.id = chunk.message_id || assistantMessage.id
        assistantMessage.content += chunk.content || ''
      }
      if (chunk.type === 'error') {
        throw new Error(chunk.error || '生成失败')
      }
    })
    await loadSessions(sessionId)
  } catch (error) {
    ElMessage.error(error.message)
    await loadSessions(sessionId).catch(() => {})
  } finally {
    streaming.value = false
    scrollToBottom()
  }
}

function handleInputEnter(event) {
  if (event.shiftKey) return

  event.preventDefault()
  handleSendMessage()
}

async function handleSaveConfig() {
  await createLLMConfig({ ...configForm })
  ElMessage.success('保存成功')
  resetConfigForm()
  await loadConfigs()
}

function handleProviderChange() {
  if (!currentProvider.value) return
  if (!configForm.name) {
    configForm.name = currentProvider.value.name || ''
  }
}

async function handleTestConfig() {
  const data = await testLLMConfig({ ...configForm })
  if (data.success === false) {
    ElMessage.error(data.message || '连接失败')
    return
  }
  ElMessage.success(data.message || '连接成功')
}

async function handleDeleteConfig(configId) {
  await ElMessageBox.confirm('确认删除这个模型配置吗？', '删除配置', { type: 'warning' })
  await deleteLLMConfig(configId)
  await loadConfigs()
}

function resetConfigForm() {
  Object.assign(configForm, {
    name: '',
    provider: 'openai',
    api_key: '',
    model: '',
    base_url: '',
    is_default: false,
    is_active: true,
  })
}

function getSessionTitle(session) {
  return session.messages?.find((message) => message.role === 'user')?.content || '新会话'
}

function getConfigLabel(config) {
  return `${getProviderName(config.provider)} / ${config.model}`
}

function getProviderName(providerId) {
  return providerCatalog.value.find((provider) => provider.id === providerId)?.name || providerId || '-'
}

function formatTime(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', { hour12: false })
}

async function scrollToBottom() {
  await nextTick()
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}
</script>

<style scoped>
.llm-chat-page {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  height: 100%;
  min-height: 520px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  overflow: hidden;
}

.session-panel,
.chat-panel {
  min-height: 0;
}

.session-panel {
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e7ed;
  background: #f8fafc;
}

.panel-header,
.chat-header,
.chat-input-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.panel-header,
.chat-header {
  justify-content: flex-end;
}

.panel-header h1,
.chat-header > div {
  display: none;
}

.config-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 40px;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.session-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid #e4e7ed;
}

.session-list {
  flex: 1;
  min-height: 0;
  padding: 10px;
  overflow-y: auto;
}

.session-pagination {
  flex-shrink: 0;
  padding: 10px 12px;
  margin-top: 0;
  overflow: hidden;
  border-top: 1px solid #e4e7ed;
}

.session-pagination :deep(.el-pagination) {
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px;
}

.session-pagination :deep(.el-pagination__sizes) {
  margin: 0;
}

.session-pagination :deep(.el-select) {
  width: 96px;
}

.session-item {
  position: relative;
  width: 100%;
  padding: 12px 34px 12px 40px;
  color: #606266;
  text-align: left;
  background: transparent;
  border: 0;
  border-radius: 8px;
  cursor: pointer;
}

.session-checkbox {
  position: absolute;
  top: 13px;
  left: 12px;
}

.session-item:hover,
.session-item.is-active {
  background: #ffffff;
  color: #303133;
}

.session-title,
.session-meta {
  display: block;
}

.session-title {
  overflow: hidden;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-meta {
  margin-top: 4px;
  color: #909399;
  font-size: 12px;
}

.session-delete {
  position: absolute;
  top: 10px;
  right: 8px;
  opacity: 0;
}

.session-item:hover .session-delete {
  opacity: 1;
}

.chat-panel {
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.message-list {
  flex: 1;
  min-height: 0;
  max-height: 100%;
  padding: 24px;
  overflow-y: auto;
  background: #ffffff;
}

.message-row {
  display: flex;
  margin-bottom: 16px;
}

.message-row.is-user {
  justify-content: flex-end;
}

.message-row.is-assistant {
  justify-content: flex-start;
}

.message-bubble {
  max-width: min(720px, 78%);
  padding: 12px 14px;
  color: #303133;
  background: #f4f6f8;
  border-radius: 8px;
}

.message-row.is-user .message-bubble {
  color: #ffffff;
  background: #409eff;
}

.message-role {
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 700;
  opacity: 0.78;
}

.message-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.message-time {
  margin-top: 8px;
  font-size: 12px;
  opacity: 0.72;
}

.typing-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 8px;
  background: #409eff;
  border-radius: 50%;
  animation: pulse 1s infinite ease-in-out;
}

.chat-input-bar {
  align-items: flex-end;
  border-top: 1px solid #e4e7ed;
  border-bottom: 0;
}

.chat-input {
  flex: 1;
}

.send-button {
  height: 76px;
}

.config-table {
  margin-top: 16px;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 0.35;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 900px) {
  .llm-chat-page {
    grid-template-columns: 1fr;
  }

  .session-panel {
    max-height: 320px;
    border-right: 0;
    border-bottom: 1px solid #e4e7ed;
  }
}
</style>
