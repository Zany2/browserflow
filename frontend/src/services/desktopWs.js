import { API_BASE_URL } from '@/api/request'

const reconnectMaxDelay = 10000

let socket = null
let reconnectTimer = null
let reconnectCount = 0
let connecting = false
let connectionErrorNotified = false

const pendingMessages = []
const activeSubscriptions = new Set()
const messageListeners = new Map()
const connectionErrorListeners = new Set()
const connectionStateListeners = new Set()

const connectionState = {
  status: 'idle',
  reconnectCount: 0,
  error: '',
}

export function connectDesktopWs() {
  if (isSocketOpen() || connecting) return

  connecting = true
  updateConnectionState('connecting')
  socket = new WebSocket(getWSURL())

  socket.addEventListener('open', () => {
    connecting = false
    reconnectCount = 0
    connectionErrorNotified = false
    updateConnectionState('open')
    replaySubscriptions()
    flushPendingMessages()
  })

  socket.addEventListener('message', (event) => {
    let payload = null
    try {
      payload = JSON.parse(event.data)
    } catch {
      notifyConnectionError(new Error('WebSocket 消息解析失败'))
      return
    }
    notifyMessage(payload)
  })

  socket.addEventListener('close', scheduleReconnect)
  socket.addEventListener('error', () => {
    const error = new Error('WebSocket 连接失败，正在重连')
    updateConnectionState('error', error.message)
    notifyConnectionError(error)
  })
}

export function subscribeDesktopConnectionState(onChange) {
  connectionStateListeners.add(onChange)
  onChange({ ...connectionState })
  return () => {
    connectionStateListeners.delete(onChange)
  }
}

export function subscribeDesktopStatus({ type, responseType, getPayload, onMessage, onError, errorMessage }) {
  const alreadySubscribed = activeSubscriptions.has(type)
  activeSubscriptions.add(type)
  connectDesktopWs()
  if (!alreadySubscribed && isSocketOpen()) {
    sendDesktopMessage({ type })
  }

  const removeMessageListener = addDesktopMessageListener(responseType, (payload) => {
    onMessage(getPayload(payload))
  })
  const removeErrorListener = addConnectionErrorListener((error) => {
    onError?.(new Error(error.message || errorMessage))
  })

  return () => {
    removeMessageListener()
    removeErrorListener()
  }
}

export function sendDesktopChatMessage(sessionId, message, onChunk) {
  return new Promise((resolve, reject) => {
    const removeMessageListener = addDesktopMessageListener('chat_message', handleChunk)
    const removeDoneListener = addDesktopMessageListener('chat_done', handleChunk)
    const removeErrorListener = addDesktopMessageListener('error', handleChunk)
    const removeConnectionErrorListener = addConnectionErrorListener((error) => {
      cleanup()
      reject(error)
    })

    function cleanup() {
      removeMessageListener()
      removeDoneListener()
      removeErrorListener()
      removeConnectionErrorListener()
    }

    function handleChunk(chunk) {
      if (chunk.session_id && chunk.session_id !== sessionId) return

      try {
        onChunk(chunk)
      } catch (error) {
        cleanup()
        reject(error)
        return
      }

      if (chunk.type === 'chat_done') {
        cleanup()
        resolve()
      }
      if (chunk.type === 'error') {
        cleanup()
        reject(new Error(chunk.error || '生成失败'))
      }
    }

    sendDesktopMessage({
      type: 'chat_send',
      session_id: sessionId,
      message,
    })
  })
}

function sendDesktopMessage(message) {
  connectDesktopWs()
  const payload = JSON.stringify(message)
  if (isSocketOpen()) {
    socket.send(payload)
    return
  }
  pendingMessages.push(payload)
}

function replaySubscriptions() {
  activeSubscriptions.forEach((type) => {
    sendDesktopMessage({ type })
  })
}

function flushPendingMessages() {
  while (pendingMessages.length > 0 && isSocketOpen()) {
    socket.send(pendingMessages.shift())
  }
}

function addDesktopMessageListener(type, handler) {
  if (!messageListeners.has(type)) {
    messageListeners.set(type, new Set())
  }
  messageListeners.get(type).add(handler)
  return () => {
    const listeners = messageListeners.get(type)
    listeners?.delete(handler)
    if (listeners?.size === 0) {
      messageListeners.delete(type)
    }
  }
}

function addConnectionErrorListener(handler) {
  connectionErrorListeners.add(handler)
  return () => {
    connectionErrorListeners.delete(handler)
  }
}

function notifyMessage(payload) {
  const listeners = messageListeners.get(payload?.type)
  listeners?.forEach((handler) => handler(payload))
}

function notifyConnectionError(error) {
  if (connectionErrorNotified) return
  connectionErrorNotified = true
  connectionErrorListeners.forEach((handler) => handler(error))
}

function scheduleReconnect() {
  connecting = false
  socket = null
  if (reconnectTimer) return

  reconnectCount += 1
  updateConnectionState('reconnecting', '', reconnectCount)
  const delay = Math.min(1000 * reconnectCount, reconnectMaxDelay)
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null
    connectDesktopWs()
  }, delay)
}

function isSocketOpen() {
  return socket?.readyState === WebSocket.OPEN
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

function updateConnectionState(status, error = '', nextReconnectCount = reconnectCount) {
  connectionState.status = status
  connectionState.error = error
  connectionState.reconnectCount = nextReconnectCount
  connectionStateListeners.forEach((handler) => handler({ ...connectionState }))
}
