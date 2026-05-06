import request, { API_BASE_URL } from '@/api/request'

export function listBrowserInstances() {
  return request({
    url: '/browser/instances',
    showSuccessMessage: false,
  })
}

export function createBrowserInstance(data) {
  return request({
    url: '/browser/instances',
    method: 'POST',
    data,
    showSuccessMessage: false,
  })
}

export function updateBrowserInstance(id, data) {
  return request({
    url: `/browser/instances/${id}`,
    method: 'PUT',
    data,
    showSuccessMessage: false,
  })
}

export function deleteBrowserInstance(id) {
  return request({
    url: `/browser/instances/${id}`,
    method: 'DELETE',
    showSuccessMessage: false,
  })
}

export function startBrowserInstance(id) {
  return request({
    url: `/browser/instances/${id}/start`,
    method: 'POST',
    showSuccessMessage: false,
  })
}

export function stopBrowserInstance(id) {
  return request({
    url: `/browser/instances/${id}/stop`,
    method: 'POST',
    showSuccessMessage: false,
  })
}

export function switchBrowserInstance(id) {
  return request({
    url: `/browser/instances/${id}/switch`,
    method: 'POST',
    showSuccessMessage: false,
  })
}

export function getBrowserStatus() {
  return request({
    url: '/browser/status',
    showSuccessMessage: false,
  })
}

export function getAgentStatus() {
  return request({
    url: '/agents/status',
    showSuccessMessage: false,
  })
}

export function syncAutomaWorkflows(browserId) {
  const safeBrowserId = String(browserId || '').trim()
  return request({
    url: '/workflows/sync',
    method: 'POST',
    params: { browser_id: safeBrowserId },
    data: {
      // SourceIP keeps backend validation happy; agent sync uses browser_id. source_ip 仅用于通过当前后端校验，执行端同步实际使用 browser_id。
      source_ip: safeBrowserId,
      workflows: [],
    },
    showSuccessMessage: false,
  })
}

export function subscribeBrowserStatus(onStatus, onError) {
  return createStatusSubscriber({
    type: 'browser_subscribe',
    responseType: 'browser_status',
    getPayload: (payload) => payload.browser,
    onMessage: onStatus,
    onError,
    errorMessage: '浏览器状态监听失败',
  })
}

export function subscribeAgentStatus(onStatus, onError) {
  return createStatusSubscriber({
    type: 'agent_status_subscribe',
    responseType: 'agent_status',
    getPayload: (payload) => payload.agents || [],
    onMessage: onStatus,
    onError,
    errorMessage: '执行端状态监听失败',
  })
}

function createStatusSubscriber({ type, responseType, getPayload, onMessage, onError, errorMessage }) {
  let socket = null
  let reconnectTimer = null
  let stopped = false
  let reconnectCount = 0
  let connectionErrorNotified = false

  const connect = () => {
    socket = new WebSocket(getWSURL())

    socket.addEventListener('open', () => {
      reconnectCount = 0
      connectionErrorNotified = false
      // Status subscribe only listens status and does not send agent_register. 状态订阅只监听状态，不发送 agent_register。
      socket.send(JSON.stringify({ type }))
    })

    socket.addEventListener('message', (event) => {
      const payload = JSON.parse(event.data)
      if (payload.type === responseType) {
        onMessage(getPayload(payload))
      }
      if (payload.type === 'error') {
        onError?.(new Error(payload.error || errorMessage))
      }
    })

    socket.addEventListener('close', scheduleReconnect)
    socket.addEventListener('error', () => {
      if (connectionErrorNotified) return
      connectionErrorNotified = true
      onError?.(new Error(`${errorMessage}，正在重连`))
    })
  }

  const scheduleReconnect = () => {
    if (stopped) return
    reconnectCount += 1
    const delay = Math.min(1000 * reconnectCount, 10000)
    reconnectTimer = window.setTimeout(connect, delay)
  }

  connect()

  return () => {
    stopped = true
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    socket?.close()
  }
}

function getWSURL() {
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL
  }

  const baseURL = API_BASE_URL.replace(/\/$/, '')
  if (baseURL.startsWith('http')) {
    return `${baseURL.replace(/^http/, 'ws')}/ws`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}${baseURL}/ws`
}
