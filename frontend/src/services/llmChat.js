import request, { API_BASE_URL } from '@/api/request'
import { sendDesktopChatMessage } from '@/services/desktopWs'

export function listLLMConfigs() {
  return request({
    url: '/llm/configs',
    showSuccessMessage: false,
  })
}

export function listLLMProviders() {
  return request({
    url: '/llm/providers',
    showSuccessMessage: false,
  })
}

export function createLLMConfig(data) {
  return request({
    url: '/llm/configs',
    method: 'POST',
    data,
    showSuccessMessage: false,
  })
}

export function updateLLMConfig(id, data) {
  return request({
    url: `/llm/configs/${id}`,
    method: 'PUT',
    data,
    showSuccessMessage: false,
  })
}

export function deleteLLMConfig(id) {
  return request({
    url: `/llm/configs/${id}`,
    method: 'DELETE',
    showSuccessMessage: false,
  })
}

export function deleteLLMConfigs(ids) {
  return request({
    url: '/llm/configs',
    method: 'DELETE',
    data: { ids },
    showSuccessMessage: false,
  })
}

export function testLLMConfig(data) {
  return request({
    url: '/llm/configs/test',
    method: 'POST',
    data,
    showSuccessMessage: false,
  })
}

export function listChatSessions() {
  return request({
    url: '/chat/sessions',
    showSuccessMessage: false,
  })
}

export function createChatSession(llmConfigId) {
  return request({
    url: '/chat/sessions',
    method: 'POST',
    data: { llm_config_id: llmConfigId },
    showSuccessMessage: false,
  })
}

export function deleteChatSession(id) {
  return request({
    url: `/chat/sessions/${id}`,
    method: 'DELETE',
    showSuccessMessage: false,
  })
}

export function deleteChatSessions(ids) {
  return request({
    url: '/chat/sessions',
    method: 'DELETE',
    data: { ids },
    showSuccessMessage: false,
  })
}

export async function streamChatMessage(sessionId, message, onChunk) {
  const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ message }),
  })

  if (!response.ok || !response.body) {
    throw new Error(await readFetchErrorMessage(response, '发送消息失败'))
  }

  const contentType = response.headers.get('content-type') || ''
  if (contentType.includes('application/json')) {
    throw new Error(await readFetchErrorMessage(response, '发送消息失败'))
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const chunks = buffer.split('\n\n')
    buffer = chunks.pop() || ''

    for (const chunk of chunks) {
      const line = chunk.trim()
      if (!line.startsWith('data:')) continue
      const payload = line.replace(/^data:\s*/, '')
      if (!payload) continue
      // Await chunk handling so the UI can render SSE output character by character. 等待前端逐字渲染完成再处理下一段 SSE。
      await onChunk(JSON.parse(payload))
    }
  }

  if (buffer.trim()) {
    const line = buffer.trim()
    if (line.startsWith('data:')) {
      const payload = line.replace(/^data:\s*/, '')
      if (payload) await onChunk(JSON.parse(payload))
    }
  }
}

export function sendChatMessageWS(sessionId, message, onChunk) {
  return sendDesktopChatMessage(sessionId, message, (chunk) => onChunk(normalizeChatChunk(chunk)))
}

function normalizeChatChunk(chunk) {
  const rawType = String(chunk?.type || '')
  if (!rawType.startsWith('chat_')) return chunk

  return {
    ...chunk,
    // Type normalize for WebSocket and SSE shared message/done/error. 类型归一，WebSocket 与 SSE 共用 message/done/error。
    type: rawType.replace(/^chat_/, ''),
  }
}

async function readFetchErrorMessage(response, fallbackMessage) {
  const data = await response.json().catch(() => ({}))
  return data?.message || data?.error || data?.data?.message || fallbackMessage
}
