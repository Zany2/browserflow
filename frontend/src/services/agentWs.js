const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
// HEARTBEAT_INTERVAL_MS heartbeat interval 心跳发送间隔
const HEARTBEAT_INTERVAL_MS = 15000
// AUTOMA_STATUS_INTERVAL_MS status check interval Automa 安装状态检测间隔
const AUTOMA_STATUS_INTERVAL_MS = 5000
// AUTOMA_REFRESH_FAIL_LIMIT refresh page after consecutive failed probes 连续失败后刷新页面
const AUTOMA_REFRESH_FAIL_LIMIT = 3
// WORKFLOW_INVENTORY_INTERVAL_MS inventory check interval 工作流清单检查间隔
const WORKFLOW_INVENTORY_INTERVAL_MS = 60000
// AUTOMA_VERSION_PROBE_INTERVAL_MS version fallback probe interval 版本兜底探测间隔
const AUTOMA_VERSION_PROBE_INTERVAL_MS = 30000
const WORKFLOW_INVENTORY_REFRESH_COMMAND = 'automa.workflow.inventory.refresh'

// createAgentSocket creates websocket channel 创建客户端 websocket 通道
export function createAgentSocket({
  browserId,
  token,
  role = 'browser_agent',
  wsUrl,
  getAutomaInstalled,
  getAutomaInfo,
  getWorkflows,
  onCommand,
  onStatus,
  onError,
  onRegistered,
  onMessage,
  beforeConnect,
  onNoReconnect,
  enableHeartbeat = false,
  enableWorkflowInventory = false,
}) {
  let socket = null
  let reconnectTimer = null
  let statusTimer = null
  // heartbeatTimer heartbeat interval timer 心跳定时器
  let heartbeatTimer = null
  // workflowInventoryTimer inventory check timer 工作流清单检查定时器
  let workflowInventoryTimer = null
  let reconnectCount = 0
  let stopped = false
  let lastAutomaStatusHash = ''
  let lastWorkflowInventoryHash = ''
  let lastAutomaVersionProbeAt = 0
  let lastKnownAutomaInstalled = false
  let lastKnownAutomaVersion = ''
  // automaProbeFailCount consecutive failed bridge probes 连续桥接探测失败次数
  let automaProbeFailCount = 0
  // automaRefreshTimer pending page refresh timer 待执行页面刷新定时器
  let automaRefreshTimer = null
  let currentClientIp = ''

  const getCurrentAutomaInstalled = () => Boolean(getAutomaInstalled?.() || lastKnownAutomaInstalled)

  const getCurrentAutomaInfo = async () => {
    const rawInfo = getAutomaInfo ? await getAutomaInfo() : {}
    const installed = Boolean(rawInfo?.installed ?? getCurrentAutomaInstalled())
    let automaVersion = normalizeAutomaVersion(rawInfo)

    if (installed && !automaVersion && getWorkflows && canProbeAutomaVersion()) {
      lastAutomaVersionProbeAt = Date.now()
      automaVersion = await resolveAutomaVersionFromWorkflows()
    }
    if (installed && !automaVersion) {
      automaVersion = lastKnownAutomaVersion
    }
    if (installed && automaVersion) {
      lastKnownAutomaVersion = automaVersion
    }
    if (!installed) {
      lastKnownAutomaVersion = ''
    }
    lastKnownAutomaInstalled = installed

    return {
      installed,
      version: automaVersion,
    }
  }

  const resolveAutomaVersionFromWorkflows = async () => {
    try {
      const workflows = await getWorkflows()
      const workflowList = Array.isArray(workflows) ? workflows : []

      for (const workflow of workflowList) {
        const version = normalizeAutomaVersion(workflow)
        if (version) return version
      }
    } catch {
      return ''
    }

    return ''
  }

  const canProbeAutomaVersion = () => {
    return Date.now() - lastAutomaVersionProbeAt >= AUTOMA_VERSION_PROBE_INTERVAL_MS
  }

  const emitStatus = (value) => {
    onStatus?.(value)
  }

  const emitMessage = (payload) => {
    onMessage?.(payload)
  }

  const sendJSON = (payload) => {
    if (socket?.readyState !== WebSocket.OPEN) return
    socket.send(JSON.stringify(payload))
  }

  const registerAgent = async () => {
    const automaInfo = await getCurrentAutomaInfo()
    const clientInfo = getClientInfo()
    lastAutomaStatusHash = getAutomaStatusHash(automaInfo)
    sendJSON({
      type: 'agent_register',
      role,
      browser_id: browserId,
      client_id: browserId,
      ...clientInfo,
      token,
      automa_installed: automaInfo.installed,
      automa_version: automaInfo.version,
    })
  }

  const sendAutomaStatus = async () => {
    const automaInfo = await getCurrentAutomaInfo()
    const automaStatusHash = getAutomaStatusHash(automaInfo)
    const statusChanged = automaStatusHash !== lastAutomaStatusHash

    lastAutomaStatusHash = automaStatusHash
    sendJSON({
      type: 'agent_status_update',
      browser_id: browserId,
      client_id: browserId,
      automa_installed: automaInfo.installed,
      automa_version: automaInfo.version,
    })
    trackAutomaProbeResult(automaInfo, { refreshOnFailure: true })
    if (statusChanged && automaInfo.installed) {
      resetWorkflowInventoryCache()
      sendWorkflowInventory({ automaInstalled: automaInfo.installed })
    }
  }

  // sendHeartbeat reports liveness only 上报在线心跳
  const sendHeartbeat = () => {
    if (!enableHeartbeat) return
    sendJSON({
      type: 'heartbeat',
      browser_id: browserId,
      client_id: browserId,
      client_ip: currentClientIp,
      client_time: Date.now(),
    })
  }

  const sendResult = (payload) => {
    sendJSON({
      browser_id: browserId,
      client_id: browserId,
      ...payload,
    })
  }

  // sendWorkflowInventory reports client workflow cache 上报客户端工作流清单缓存
  const sendWorkflowInventory = async ({ force = false, automaInstalled = null } = {}) => {
    if (!enableWorkflowInventory) {
      return { workflow_count: 0, skipped: true }
    }
    if (!getWorkflows) {
      return { workflow_count: 0, skipped: true }
    }

    const installed = automaInstalled ?? (await getCurrentAutomaInfo()).installed
    if (!installed) {
      return { workflow_count: 0, skipped: true }
    }

    try {
      const workflows = await getWorkflows()
      const workflowList = Array.isArray(workflows)
        ? [...workflows].sort((prev, next) => getWorkflowUpdatedAt(next) - getWorkflowUpdatedAt(prev))
        : []
      const inventoryHash = JSON.stringify(
        workflowList.map((item) => ({
          id: item?.id || item?.workflow_id || item?.workflowId || '',
          updatedAt: item?.updatedAt || item?.updated_at || item?.updatedAtAutoma || item?.updated_at_automa || '',
        })),
      )
      if (!force && inventoryHash === lastWorkflowInventoryHash) {
        return { workflow_count: workflowList.length, skipped: true }
      }

      lastWorkflowInventoryHash = inventoryHash
      sendJSON({
        type: 'workflow_inventory',
        browser_id: browserId,
        client_id: browserId,
        workflows: workflowList,
        client_time: Date.now(),
      })
      return { workflow_count: workflowList.length, skipped: false }
    } catch (error) {
      onError?.(error)
      if (force) throw error
      return { workflow_count: 0, skipped: true }
    }
  }

  // resetWorkflowInventoryCache forces inventory resend after reconnect. 重连后强制重新上报工作流清单
  const resetWorkflowInventoryCache = () => {
    lastWorkflowInventoryHash = ''
  }

  // trackAutomaProbeResult handles bridge probe state 处理桥接探测结果
  const trackAutomaProbeResult = (automaInfo, { refreshOnFailure = false } = {}) => {
    if (automaInfo?.installed) {
      automaProbeFailCount = 0
      if (automaRefreshTimer) {
        window.clearTimeout(automaRefreshTimer)
        automaRefreshTimer = null
      }
      return
    }

    automaProbeFailCount += 1
    if (!refreshOnFailure || automaProbeFailCount < AUTOMA_REFRESH_FAIL_LIMIT || automaRefreshTimer) return

    // Reload after status report so backend sees unavailable first 状态上报后刷新，确保后端先收到不可用
    automaRefreshTimer = window.setTimeout(() => {
      window.location.reload()
    }, 100)
  }

  const handleMessage = async (event) => {
    let payload = null

    try {
      payload = JSON.parse(event.data)
    } catch {
      onError?.(new Error('WebSocket 消息解析失败'))
      return
    }

    if (isNoReconnectMessage(payload)) {
      emitMessage(payload)
      handleNoReconnectMessage(payload)
      return
    }

    emitMessage(payload)

    updateCurrentClientIp(payload)

    if (payload.type === 'agent_registered') {
      emitStatus('online')
      onRegistered?.(payload)
      sendWorkflowInventory()
      return
    }

    if (payload.type !== 'agent_command') return

    try {
      const data =
        payload.command === WORKFLOW_INVENTORY_REFRESH_COMMAND
          ? await sendWorkflowInventory({ force: true })
          : await onCommand(payload.command, payload.payload || {}, payload)
      sendResult({
        type: 'agent_result',
        command_id: payload.command_id,
        success: true,
        data,
      })
    } catch (error) {
      sendResult({
        type: 'agent_result',
        command_id: payload.command_id,
        success: false,
        error: error.message,
      })
    }
  }

  // updateCurrentClientIp stores backend observed ip 保存后端识别到的客户端 IP
  const updateCurrentClientIp = (payload) => {
    const nextClientIp = payload?.client_ip || payload?.ip || payload?.client?.ip || payload?.remote_ip || ''
    if (nextClientIp) currentClientIp = nextClientIp
  }

  // clearSocketTimers clears connection timers 清理连接相关定时器
  const clearSocketTimers = () => {
    if (statusTimer) window.clearInterval(statusTimer)
    if (heartbeatTimer) window.clearInterval(heartbeatTimer)
    if (workflowInventoryTimer) window.clearInterval(workflowInventoryTimer)
    statusTimer = null
    heartbeatTimer = null
    workflowInventoryTimer = null
  }

  const scheduleReconnect = () => {
    if (stopped) return
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    reconnectCount += 1
    const delay = Math.min(1000 * reconnectCount, 10000)
    emitStatus('reconnecting')
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  const handleNoReconnectMessage = (payload) => {
    // noReconnect closes current socket and keeps precheck retry 拉黑通知关闭当前连接并继续预检
    clearSocketTimers()
    onNoReconnect?.(payload)
    emitStatus('offline')
    socket?.close()
  }

  const isNoReconnectMessage = (payload) => {
    return (
      payload?.type === 'client_banned' ||
      payload?.no_reconnect === true ||
      payload?.data?.no_reconnect === true
    )
  }

  const connect = async () => {
    if (stopped) return

    emitStatus('connecting')

    try {
      // beforeConnect is required to finish before creating websocket. 建立 WebSocket 前必须完成预检
      const allowed = beforeConnect ? await beforeConnect() : true
      if (stopped) return
      if (allowed === false) {
        emitStatus('offline')
        scheduleReconnect()
        return
      }
    } catch (error) {
      if (stopped) return
      onError?.(error)
      emitStatus('offline')
      scheduleReconnect()
      return
    }

    if (stopped) return

    socket = new WebSocket(wsUrl || getWSURL())

    socket.addEventListener('open', () => {
      reconnectCount = 0
      clearSocketTimers()
      resetWorkflowInventoryCache()
      registerAgent()
      statusTimer = window.setInterval(sendAutomaStatus, AUTOMA_STATUS_INTERVAL_MS)
      if (enableHeartbeat) {
        heartbeatTimer = window.setInterval(sendHeartbeat, HEARTBEAT_INTERVAL_MS)
      }
      if (enableWorkflowInventory) {
        workflowInventoryTimer = window.setInterval(sendWorkflowInventory, WORKFLOW_INVENTORY_INTERVAL_MS)
        window.setTimeout(sendWorkflowInventory, 1000)
      }
    })

    socket.addEventListener('message', handleMessage)

    socket.addEventListener('close', () => {
      clearSocketTimers()
      if (stopped) {
        emitStatus('offline')
        return
      }

      scheduleReconnect()
    })

    socket.addEventListener('error', () => {
      onError?.(new Error('客户端 WebSocket 连接失败，正在重试'))
    })
  }

  connect()

  return () => {
    stopped = true
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    if (automaRefreshTimer) window.clearTimeout(automaRefreshTimer)
    clearSocketTimers()
    socket?.close()
  }
}

// getClientInfo collects browser client metadata 采集浏览器客户端信息
function getClientInfo() {
  const userAgent = navigator.userAgent || ''
  const browser = getBrowserInfo(userAgent)
  const os = getOSInfo(userAgent)

  return {
    client_name: `${browser.name} - ${os.name}`,
    user_agent: userAgent,
    browser_name: browser.name,
    browser_version: browser.version,
    os_name: os.name,
    os_version: os.version,
  }
}

// getAutomaStatusHash tracks install and version changes 跟踪安装状态和版本变化
function getAutomaStatusHash(automaInfo) {
  return JSON.stringify({
    installed: Boolean(automaInfo?.installed),
    version: String(automaInfo?.version || ''),
  })
}

// normalizeAutomaVersion supports common metadata field names 兼容常见版本字段名
function normalizeAutomaVersion(info = {}) {
  return String(
    info.automa_version ||
    info.automaVersion ||
    info.version ||
    info.extVersion ||
    info.extensionVersion ||
    '',
  ).trim()
}

function getWorkflowUpdatedAt(workflow = {}) {
  return Number(
    workflow.updatedAt ||
      workflow.updated_at ||
      workflow.updatedAtAutoma ||
      workflow.updated_at_automa ||
      0,
  )
}

// getBrowserInfo parses browser name and version 解析浏览器名称和版本
function getBrowserInfo(userAgent) {
  const matchers = [
    ['Edge', /Edg\/([\d.]+)/],
    ['Chrome', /Chrome\/([\d.]+)/],
    ['Firefox', /Firefox\/([\d.]+)/],
    ['Safari', /Version\/([\d.]+).*Safari/],
  ]

  for (const [name, pattern] of matchers) {
    const matched = userAgent.match(pattern)
    if (matched) {
      return { name, version: matched[1] || '' }
    }
  }

  return { name: 'Unknown', version: '' }
}

// getOSInfo parses operating system information 解析操作系统信息
function getOSInfo(userAgent) {
  if (/Windows NT ([\d.]+)/.test(userAgent)) {
    return { name: 'Windows', version: userAgent.match(/Windows NT ([\d.]+)/)?.[1] || '' }
  }
  if (/Mac OS X ([\d_]+)/.test(userAgent)) {
    return { name: 'macOS', version: userAgent.match(/Mac OS X ([\d_]+)/)?.[1]?.replaceAll('_', '.') || '' }
  }
  if (/Android ([\d.]+)/.test(userAgent)) {
    return { name: 'Android', version: userAgent.match(/Android ([\d.]+)/)?.[1] || '' }
  }
  if (/iPhone OS ([\d_]+)/.test(userAgent)) {
    return { name: 'iOS', version: userAgent.match(/iPhone OS ([\d_]+)/)?.[1]?.replaceAll('_', '.') || '' }
  }
  if (/Linux/.test(userAgent)) {
    return { name: 'Linux', version: '' }
  }

  return { name: 'Unknown', version: '' }
}

export function getWSURL() {
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
