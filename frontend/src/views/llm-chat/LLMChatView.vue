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
        <AppSelectionSummary :count="selectedSessionIds.length" unit="会话" />
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
        v-if="sessions.length > sessionPageSize"
        v-model:current-page="sessionCurrentPage"
        v-model:page-size="sessionPageSize"
        class="session-pagination"
        compact
        :total="sessions.length"
      />
    </aside>

    <main class="chat-panel">
      <div ref="messageListRef" class="message-list" @scroll="handleMessageListScroll">
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

      <footer class="chat-input-bar" :style="chatInputBarStyle">
        <div
          class="chat-input-resizer"
          role="separator"
          aria-orientation="horizontal"
          title="拖动调整输入框高度"
          @pointerdown="handleInputResizeStart"
        ></div>
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
  </section>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Delete, Plus, Promotion } from '@element-plus/icons-vue'
import { APP_CONFIRM_TYPE, appConfirm } from '@/components/AppConfirm'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppPagination from '@/components/AppPagination.vue'
import AppSelectionSummary from '@/components/AppSelectionSummary.vue'
import { getSafePage } from '@/utils/list'
import {
  createChatSession,
  deleteChatSession,
  deleteChatSessions,
  listChatSessions,
  listLLMConfigs,
  listLLMProviders,
  streamChatMessage,
} from '@/services/llmChat'

const configs = ref([])
const providerCatalog = ref([])
const sessions = ref([])
const currentSession = ref(null)
const selectedConfigId = ref('')
const selectedSessionIds = ref([])
const sessionCurrentPage = ref(1)
const sessionPageSize = ref(10)
// STREAM_CHAR_DELAY controls typewriter speed 流式逐字显示间隔
const STREAM_CHAR_DELAY = 18
// SCROLL_BOTTOM_THRESHOLD keeps auto-scroll only when viewer stays near bottom 靠近底部时才自动跟随
const SCROLL_BOTTOM_THRESHOLD = 80
const CHAT_INPUT_MIN_HEIGHT = 76
const CHAT_INPUT_MAX_HEIGHT = 260
const inputMessage = ref('')
const streaming = ref(false)
const messageListRef = ref(null)
const chatInputHeight = ref(CHAT_INPUT_MIN_HEIGHT)
const shouldStickToBottom = ref(true)

let inputResizeStartY = 0
let inputResizeStartHeight = CHAT_INPUT_MIN_HEIGHT

const activeConfigs = computed(() => configs.value.filter((config) => config.is_active))
const canSend = computed(() => Boolean(currentSession.value && inputMessage.value.trim() && !streaming.value))
const chatInputBarStyle = computed(() => ({
  '--chat-input-height': `${chatInputHeight.value}px`,
}))
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

watch(
  () => currentSession.value?.messages?.length,
  () => scrollToBottomIfNeeded(),
)

watch(
  () => currentSession.value?.id,
  () => {
    shouldStickToBottom.value = true
    scrollToBottom()
  },
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

onBeforeUnmount(() => {
  stopInputResize()
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
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请先在大模型配置页面配置并启用模型' })
    return
  }

  const data = await createChatSession(selectedConfigId.value)
  sessions.value = sortSessionsByUpdatedDesc([data.session, ...sessions.value])
  sessionCurrentPage.value = 1
  currentSession.value = data.session
}

async function handleDeleteSession(sessionId) {
  const confirmed = await appConfirm({
    title: '删除会话',
    message: '确认删除这个会话吗？',
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

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

  const confirmed = await appConfirm({
    title: '批量删除会话',
    message: `确认删除选中的 ${ids.length} 个会话吗？`,
    type: APP_CONFIRM_TYPE.danger,
    confirmText: '删除',
  })
  if (!confirmed) return

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
    appMessage({ type: APP_MESSAGE_TYPE.warning, message: '请输入对话内容' })
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
    timestamp: '',
  }

  currentSession.value.messages.push(userMessage, assistantMessage)
  const activeSession = currentSession.value
  const assistantMessageIndex = activeSession.messages.length - 1
  shouldStickToBottom.value = true
  scrollToBottom()

  const sessionId = activeSession.id
  try {
    await streamChatMessage(sessionId, messageText, async (chunk) => {
      if (chunk.type === 'message') {
        const messageItem = activeSession.messages?.[assistantMessageIndex]
        if (!messageItem) return

        // Reactive message 通过响应式数组项更新，保证逐字追加能触发界面刷新。
        messageItem.id = chunk.message_id || messageItem.id
        await appendAssistantContent(messageItem, chunk.content)
      }
      if (chunk.type === 'done') {
        const messageItem = activeSession.messages?.[assistantMessageIndex]
        if (messageItem) {
          messageItem.id = chunk.message_id || messageItem.id
          messageItem.timestamp = chunk.timestamp || new Date().toISOString()
        }
      }
      if (chunk.type === 'error') {
        throw new Error(chunk.error || '生成失败')
      }
    })
    await loadSessions(sessionId)
  } catch (error) {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
    await loadSessions(sessionId).catch(() => {})
  } finally {
    streaming.value = false
    scrollToBottomIfNeeded()
  }
}

function handleInputEnter(event) {
  if (event.shiftKey) return

  event.preventDefault()
  handleSendMessage()
}

function handleInputResizeStart(event) {
  // Resize start 顶部拖拽条控制输入区高度，向上拖动变高。
  inputResizeStartY = event.clientY
  inputResizeStartHeight = chatInputHeight.value
  window.addEventListener('pointermove', handleInputResizeMove)
  window.addEventListener('pointerup', stopInputResize)
  event.preventDefault()
}

function handleInputResizeMove(event) {
  // Resize move 输入框位于底部，鼠标上移时高度增加。
  const nextHeight = inputResizeStartHeight + inputResizeStartY - event.clientY
  chatInputHeight.value = clampInputHeight(nextHeight)
}

function stopInputResize() {
  window.removeEventListener('pointermove', handleInputResizeMove)
  window.removeEventListener('pointerup', stopInputResize)
}

function clampInputHeight(value) {
  return Math.min(Math.max(value, CHAT_INPUT_MIN_HEIGHT), CHAT_INPUT_MAX_HEIGHT)
}

function getSessionTitle(session) {
  return session.messages?.find((message) => message.role === 'user')?.content || '新会话'
}

function getConfigLabel(config) {
  return [getProviderName(config.provider), config.name, config.model].filter(Boolean).join(' / ')
}

function getProviderName(providerId) {
  return providerCatalog.value.find((provider) => provider.id === providerId)?.name || providerId || ''
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
    shouldStickToBottom.value = true
  }
}

async function scrollToBottomIfNeeded() {
  if (!shouldStickToBottom.value) return
  await scrollToBottom()
}

function handleMessageListScroll() {
  shouldStickToBottom.value = isMessageListNearBottom()
}

function isMessageListNearBottom() {
  const element = messageListRef.value
  if (!element) return true
  return element.scrollHeight - element.scrollTop - element.clientHeight <= SCROLL_BOTTOM_THRESHOLD
}

async function appendAssistantContent(assistantMessage, content) {
  // Typewriter output renders each SSE chunk one character at a time 逐字追加 SSE 内容
  for (const char of Array.from(String(content || ''))) {
    assistantMessage.content += char
    await scrollToBottomIfNeeded()
    await sleep(STREAM_CHAR_DELAY)
  }
}

function sleep(ms) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms)
  })
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
.chat-input-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.panel-header,
.panel-header {
  justify-content: flex-end;
}

.panel-header h1 {
  display: none;
}

.config-bar {
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
  flex-wrap: nowrap;
  gap: 4px;
}

.session-pagination.app-pagination--compact {
  justify-content: space-between;
}

.session-pagination :deep(.el-pagination button),
.session-pagination :deep(.el-pager li) {
  min-width: 28px;
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
  position: relative;
  padding-top: 22px;
  border-top: 1px solid #e4e7ed;
  border-bottom: 0;
}

.chat-input-resizer {
  position: absolute;
  top: 7px;
  left: 16px;
  right: 16px;
  height: 8px;
  cursor: ns-resize;
  touch-action: none;
}

.chat-input-resizer::before {
  display: block;
  width: 56px;
  height: 3px;
  margin: 2px auto 0;
  background: #cbd5e1;
  border-radius: 999px;
  content: '';
}

.chat-input-resizer:hover::before {
  background: #409eff;
}

.chat-input {
  flex: 1;
}

.chat-input :deep(.el-textarea__inner) {
  height: var(--chat-input-height);
  min-height: var(--chat-input-height);
  resize: none;
}

.send-button {
  height: var(--chat-input-height);
  min-height: 76px;
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
