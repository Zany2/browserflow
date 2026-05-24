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
const AUTOMA_WORKFLOW_RESULT_EVENT = '__browserflow_automa_workflow_result__'
const WORKFLOW_RUN_COMMANDS = new Set(['automa.workflow.run', 'task.execute', 'task.run'])
const TASK_LOCK_STORAGE_KEY = 'browserflow_client_task_lock'
const TASK_EXECUTION_STORAGE_KEY = 'browserflow_client_task_executions'
const TASK_LOCK_TTL_MS = 10 * 60 * 1000
const TASK_EXECUTION_RETENTION_MS = 24 * 60 * 60 * 1000
const TASK_EXECUTION_MAX_ITEMS = 100

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
  onCommandResult,
  onStatus,
  onError,
  onRegistered,
  onMessage,
  beforeConnect,
  onNoReconnect,
  enableHeartbeat = false,
  enableWorkflowInventory = false,
  enableAutomaStatusPolling = true,
  reportAutomaStatusOnWindowLoad = false,
}) {
  let socket = null
  let reconnectTimer = null
  let statusTimer = null
  let windowLoadStatusTimer = 0
  let windowLoadStatusHandler = null
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
  // activeBrowserId keeps backend-assigned id after first registration 保留后端确认后的浏览器 ID
  let activeBrowserId = String(browserId || '').trim()
  // automaProbeFailCount consecutive failed bridge probes 连续桥接探测失败次数
  let automaProbeFailCount = 0
  // automaRefreshTimer pending page refresh timer 待执行页面刷新定时器
  let automaRefreshTimer = null
  let currentClientIp = ''
  // workflowResultCommandIds keeps async workflow command mapping 保存异步工作流命令映射
  let workflowResultCommandIds = new Map()
  // workflowCommandPayloads keeps command context for final result logs 保存命令上下文用于最终结果展示
  let workflowCommandPayloads = new Map()
  let activeExecutionId = ''

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
      browser_id: activeBrowserId,
      client_id: activeBrowserId,
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
      browser_id: activeBrowserId,
      client_id: activeBrowserId,
      automa_installed: automaInfo.installed,
      automa_version: automaInfo.version,
    })
    trackAutomaProbeResult(automaInfo, { refreshOnFailure: true })
    if (statusChanged && automaInfo.installed) {
      resetWorkflowInventoryCache()
      sendWorkflowInventory({ automaInstalled: automaInfo.installed })
    }
  }

  // scheduleWindowLoadAutomaStatus reports one status after page load 页面完全加载后补充上报一次插件状态
  const scheduleWindowLoadAutomaStatus = () => {
    if (!reportAutomaStatusOnWindowLoad) return

    clearWindowLoadAutomaStatus()
    const report = () => {
      if (stopped) return
      sendAutomaStatus()
    }
    if (document.readyState === 'complete') {
      windowLoadStatusTimer = window.setTimeout(report, 0)
      return
    }

    windowLoadStatusHandler = report
    window.addEventListener('load', windowLoadStatusHandler, { once: true })
  }

  // clearWindowLoadAutomaStatus clears one-shot load report 清理页面加载后的一次性上报
  const clearWindowLoadAutomaStatus = () => {
    if (windowLoadStatusTimer) window.clearTimeout(windowLoadStatusTimer)
    if (windowLoadStatusHandler) window.removeEventListener('load', windowLoadStatusHandler)
    windowLoadStatusTimer = 0
    windowLoadStatusHandler = null
  }

  // sendHeartbeat reports liveness only 上报在线心跳
  const sendHeartbeat = () => {
    if (!enableHeartbeat) return
    renewLocalTaskLock(activeExecutionId)
    sendJSON({
      type: 'heartbeat',
      browser_id: activeBrowserId,
      client_id: activeBrowserId,
      client_ip: currentClientIp,
      execution_id: activeExecutionId || undefined,
      client_time: Date.now(),
    })
  }

  const sendResult = (payload) => {
    sendJSON({
      browser_id: activeBrowserId,
      client_id: activeBrowserId,
      ...payload,
    })
  }

  const sendTaskStatusResult = (payload, success, data, error = '') => {
    sendResult({
      type: 'agent_result',
      command_id: payload.command_id,
      success,
      data,
      error: success ? undefined : error,
    })
  }

  // trackWorkflowCommand remembers async workflow command id 记录异步工作流命令标识
  const trackWorkflowCommand = (payload) => {
    if (!WORKFLOW_RUN_COMMANDS.has(payload?.command)) return

    const commandId = String(payload.command_id || '').trim()
    const data = payload.payload || {}
    const waitResult = Boolean(data.wait_result ?? data.waitResult ?? false)
    const executionId = resolveWorkflowExecutionId(data, commandId)

    if (!commandId || !executionId || waitResult) return
    workflowResultCommandIds.set(executionId, {
      command: payload.command,
      commandId,
    })
    workflowCommandPayloads.set(executionId, data)
  }

  // untrackWorkflowCommand removes failed async workflow mapping 清理失败的异步工作流映射
  const untrackWorkflowCommand = (payload) => {
    const data = payload?.payload || {}
    const executionId = resolveWorkflowExecutionId(data, String(payload?.command_id || '').trim())
    if (executionId) {
      workflowResultCommandIds.delete(executionId)
      workflowCommandPayloads.delete(executionId)
    }
  }

  // ensureLocalTaskLock prevents overlapping workflow runs in one client page 本地租约防止客户端并发执行
  const ensureLocalTaskLock = (payload) => {
    if (!WORKFLOW_RUN_COMMANDS.has(payload?.command)) return

    cleanupExpiredLocalTaskLock()
    const data = payload.payload || {}
    const executionId = resolveWorkflowExecutionId(data, String(payload.command_id || '').trim())
    if (!executionId) return

    const currentLock = readLocalTaskLock()
    if (currentLock && currentLock.execution_id !== executionId && Number(currentLock.expires_at || 0) > Date.now()) {
      throw new Error(`客户端正在执行其他任务，请稍后重试（执行标识：${currentLock.execution_id}）`)
    }

    writeLocalTaskLock({
      execution_id: executionId,
      command_id: String(payload.command_id || '').trim(),
      command: payload.command,
      task_record_id: data.task_record_id || data.taskRecordId || '',
      task_id: data.task_id || data.taskId || '',
      task_name: data.task_name || data.taskName || '',
      workflow_id: data.workflow_id || data.workflowId || data.id || '',
      started_at: currentLock?.started_at || Date.now(),
      last_heartbeat_at: Date.now(),
      expires_at: Date.now() + TASK_LOCK_TTL_MS,
    })
    activeExecutionId = executionId
    saveTaskExecution({
      execution_id: executionId,
      command_id: String(payload.command_id || '').trim(),
      task_record_id: data.task_record_id || data.taskRecordId || '',
      task_id: data.task_id || data.taskId || '',
      workflow_id: data.workflow_id || data.workflowId || data.id || '',
      status: 'running',
      result: null,
      error: '',
      updated_at: Date.now(),
    })
  }

  const clearFinishedLocalTaskLock = (payload, data) => {
    const executionId = resolveWorkflowExecutionId(payload?.payload || {}, String(payload?.command_id || '').trim())
    if (!executionId) return
    const status = String(data?.status || '').toLowerCase()
    if (['queued', 'submitted', 'pending', 'running'].includes(status)) return
    clearLocalTaskLock(executionId)
    if (activeExecutionId === executionId) activeExecutionId = ''
  }

  // handleWorkflowResultEvent forwards async final result 回传异步工作流最终结果
  const handleWorkflowResultEvent = (event) => {
    const detail = event.detail || {}
    const executionId = String(
      detail.execution_id || detail.executionId || detail.request_id || detail.requestId || '',
    ).trim()
    const commandContext = workflowResultCommandIds.get(executionId)
    if (!commandContext) return

    workflowResultCommandIds.delete(executionId)
    const commandPayload = workflowCommandPayloads.get(executionId) || {}
    workflowCommandPayloads.delete(executionId)
    clearLocalTaskLock(executionId)
    if (activeExecutionId === executionId) activeExecutionId = ''
    const success = detail.ok !== false && detail.status !== 'error'
    const finalStatus = success ? normalizeExecutionStatus(detail.status, 'success') : 'failed'
    saveTaskExecution({
      execution_id: executionId,
      command_id: commandContext.commandId,
      task_record_id: commandPayload.task_record_id || commandPayload.taskRecordId || '',
      task_id: commandPayload.task_id || commandPayload.taskId || '',
      workflow_id: commandPayload.workflow_id || commandPayload.workflowId || commandPayload.id || '',
      status: finalStatus,
      result: detail,
      error: success ? '' : String(detail.message || detail.error || '').trim(),
      automa_execution_id: resolveAutomaExecutionId(detail),
      updated_at: Date.now(),
    })
    onCommandResult?.({
      command: commandContext.command,
      payload: commandPayload,
      result: detail,
      success,
      error: success ? '' : String(detail.message || detail.error || '').trim(),
      async: true,
    })
    sendResult({
      type: 'agent_result',
      command_id: commandContext.commandId,
      success,
      data: detail,
      error: success ? undefined : String(detail.message || detail.error || '').trim(),
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
        browser_id: activeBrowserId,
        client_id: activeBrowserId,
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
      activeBrowserId = String(payload.browser_id || payload.client_id || activeBrowserId).trim()
      emitStatus('online')
      onRegistered?.(payload)
      sendWorkflowInventory()
      return
    }

    if (payload.type === 'agent_result_ack') {
      clearAckedTaskExecution(payload.command_id)
      return
    }

    if (payload.type !== 'agent_command') return

    if (payload.command === 'task.status.query') {
      handleTaskStatusQuery(payload)
      return
    }

    try {
      ensureLocalTaskLock(payload)
    } catch (error) {
      onCommandResult?.({
        command: payload.command,
        payload: payload.payload || {},
        result: null,
        success: false,
        error: error.message,
        async: false,
      })
      sendResult({
        type: 'agent_result',
        command_id: payload.command_id,
        success: false,
        error: error.message,
      })
      return
    }

    trackWorkflowCommand(payload)

    try {
      const data =
        payload.command === WORKFLOW_INVENTORY_REFRESH_COMMAND
          ? await sendWorkflowInventory({ force: true })
          : await onCommand(payload.command, payload.payload || {}, payload)
      onCommandResult?.({
        command: payload.command,
        payload: payload.payload || {},
        result: data,
        success: true,
        error: '',
        async: false,
      })
      sendResult({
        type: 'agent_result',
        command_id: payload.command_id,
        success: true,
        data,
      })
      if (WORKFLOW_RUN_COMMANDS.has(payload.command)) {
        const executionId = resolveWorkflowExecutionId(payload.payload || {}, String(payload.command_id || '').trim())
        const normalizedStatus = normalizeExecutionStatus(data?.status, 'success')
        saveTaskExecution({
          execution_id: executionId,
          command_id: String(payload.command_id || '').trim(),
          task_record_id: payload.payload?.task_record_id || payload.payload?.taskRecordId || '',
          task_id: payload.payload?.task_id || payload.payload?.taskId || '',
          workflow_id: payload.payload?.workflow_id || payload.payload?.workflowId || payload.payload?.id || '',
          status: normalizedStatus,
          result: data,
          error: '',
          automa_execution_id: resolveAutomaExecutionId(data),
          updated_at: Date.now(),
        })
      }
      clearFinishedLocalTaskLock(payload, data)
    } catch (error) {
      untrackWorkflowCommand(payload)
      const executionId = resolveWorkflowExecutionId(payload?.payload || {}, String(payload?.command_id || '').trim())
      clearLocalTaskLock(executionId)
      if (WORKFLOW_RUN_COMMANDS.has(payload.command)) {
        saveTaskExecution({
          execution_id: executionId,
          command_id: String(payload.command_id || '').trim(),
          task_record_id: payload.payload?.task_record_id || payload.payload?.taskRecordId || '',
          task_id: payload.payload?.task_id || payload.payload?.taskId || '',
          workflow_id: payload.payload?.workflow_id || payload.payload?.workflowId || payload.payload?.id || '',
          status: 'failed',
          result: null,
          error: error.message,
          updated_at: Date.now(),
        })
      }
      onCommandResult?.({
        command: payload.command,
        payload: payload.payload || {},
        result: null,
        success: false,
        error: error.message,
        async: false,
      })
      sendResult({
        type: 'agent_result',
        command_id: payload.command_id,
        success: false,
        error: error.message,
      })
    }
  }

  const handleTaskStatusQuery = (payload) => {
    const data = payload.payload || {}
    const commandId = String(data.command_id || payload.command_id || '').trim()
    const executionId = resolveWorkflowExecutionId(data, commandId)
    const cached = readTaskExecution(executionId, commandId)
    if (cached) {
      const status = normalizeExecutionStatus(cached.status, 'unknown')
      sendTaskStatusResult(
        payload,
        !['failed', 'unknown', 'not_found', 'missing', 'lost'].includes(status),
        buildTaskStatusQueryData(cached, status),
        cached.error || '',
      )
      return
    }

    const currentLock = readLocalTaskLock()
    if (currentLock && (currentLock.execution_id === executionId || currentLock.command_id === commandId)) {
      renewLocalTaskLock(currentLock.execution_id)
      sendTaskStatusResult(payload, true, {
        status: 'running',
        execution_id: currentLock.execution_id,
        command_id: currentLock.command_id,
        task_record_id: currentLock.task_record_id || data.task_record_id || '',
        task_id: currentLock.task_id || data.task_id || '',
        workflow_id: currentLock.workflow_id || data.workflow_id || '',
        updated_at: Date.now(),
      })
      return
    }

    const message = 'client has no local execution state for this task'
    sendTaskStatusResult(
      payload,
      false,
      {
        status: 'unknown',
        execution_id: executionId,
        command_id: commandId,
        task_record_id: data.task_record_id || '',
        task_id: data.task_id || '',
        workflow_id: data.workflow_id || '',
        message,
        updated_at: Date.now(),
      },
      message,
    )
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
    clearWindowLoadAutomaStatus()
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

    socket.addEventListener('open', async () => {
      reconnectCount = 0
      clearSocketTimers()
      resetWorkflowInventoryCache()
      try {
        await registerAgent()
      } catch (error) {
        onError?.(error)
      }
      scheduleWindowLoadAutomaStatus()
      if (enableAutomaStatusPolling) {
        statusTimer = window.setInterval(sendAutomaStatus, AUTOMA_STATUS_INTERVAL_MS)
      }
      if (enableHeartbeat) {
        heartbeatTimer = window.setInterval(sendHeartbeat, HEARTBEAT_INTERVAL_MS)
      }
      if (enableWorkflowInventory) {
        workflowInventoryTimer = window.setInterval(sendWorkflowInventory, WORKFLOW_INVENTORY_INTERVAL_MS)
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

  cleanupExpiredLocalTaskLock()
  window.addEventListener(AUTOMA_WORKFLOW_RESULT_EVENT, handleWorkflowResultEvent)

  connect()

  return () => {
    stopped = true
    window.removeEventListener(AUTOMA_WORKFLOW_RESULT_EVENT, handleWorkflowResultEvent)
    workflowResultCommandIds.clear()
    workflowCommandPayloads.clear()
    activeExecutionId = ''
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    if (automaRefreshTimer) window.clearTimeout(automaRefreshTimer)
    clearSocketTimers()
    socket?.close()
  }
}

// resolveWorkflowExecutionId returns the bridge execution id 解析工作流桥接执行标识
function resolveWorkflowExecutionId(data = {}, fallbackId = '') {
  return String(data.execution_id || data.executionId || data.task_id || data.taskId || fallbackId || '').trim()
}

function readLocalTaskLock() {
  try {
    const rawValue = window.localStorage.getItem(TASK_LOCK_STORAGE_KEY)
    return rawValue ? JSON.parse(rawValue) : null
  } catch {
    return null
  }
}

function writeLocalTaskLock(lock) {
  try {
    window.localStorage.setItem(TASK_LOCK_STORAGE_KEY, JSON.stringify(lock))
  } catch {
    // localStorage may be unavailable in restricted browser contexts.
  }
}

function cleanupExpiredLocalTaskLock() {
  const currentLock = readLocalTaskLock()
  if (currentLock && Number(currentLock.expires_at || 0) <= Date.now()) {
    clearLocalTaskLock(currentLock.execution_id)
  }
}

function renewLocalTaskLock(executionId) {
  if (!executionId) return
  const currentLock = readLocalTaskLock()
  if (!currentLock || currentLock.execution_id !== executionId) return
  writeLocalTaskLock({
    ...currentLock,
    last_heartbeat_at: Date.now(),
    expires_at: Date.now() + TASK_LOCK_TTL_MS,
  })
}

function clearLocalTaskLock(executionId = '') {
  try {
    if (executionId) {
      const currentLock = readLocalTaskLock()
      if (currentLock && currentLock.execution_id && currentLock.execution_id !== executionId) return
    }
    window.localStorage.removeItem(TASK_LOCK_STORAGE_KEY)
  } catch {
    // Ignore storage cleanup failures.
  }
}

// getClientInfo collects browser client metadata 采集浏览器客户端信息
function readTaskExecution(executionId = '', commandId = '') {
  const executions = readTaskExecutions()
  const normalizedExecutionId = String(executionId || '').trim()
  const normalizedCommandId = String(commandId || '').trim()
  return executions.find((item) => {
    return (
      (normalizedExecutionId && item.execution_id === normalizedExecutionId) ||
      (normalizedCommandId && item.command_id === normalizedCommandId)
    )
  })
}

function saveTaskExecution(execution) {
  const nextExecution = {
    execution_id: String(execution.execution_id || '').trim(),
    command_id: String(execution.command_id || '').trim(),
    task_record_id: execution.task_record_id || '',
    task_id: execution.task_id || '',
    workflow_id: execution.workflow_id || '',
    status: normalizeExecutionStatus(execution.status, 'unknown'),
    result: execution.result ?? null,
    error: execution.error || '',
    automa_execution_id: execution.automa_execution_id || '',
    updated_at: Number(execution.updated_at || Date.now()),
  }
  if (!nextExecution.execution_id && !nextExecution.command_id) return

  const executions = readTaskExecutions().filter((item) => {
    return item.execution_id !== nextExecution.execution_id && item.command_id !== nextExecution.command_id
  })
  executions.unshift(nextExecution)
  writeTaskExecutions(executions)
}

function clearAckedTaskExecution(commandId = '') {
  commandId = String(commandId || '').trim()
  if (!commandId) return

  const executions = readTaskExecutions()
  const nextExecutions = executions.filter((item) => {
    if (item.command_id !== commandId) return true
    return ['queued', 'submitted', 'pending', 'running'].includes(normalizeExecutionStatus(item.status, 'unknown'))
  })
  if (nextExecutions.length !== executions.length) {
    writeTaskExecutions(nextExecutions)
  }
}

function readTaskExecutions() {
  try {
    const rawValue = window.localStorage.getItem(TASK_EXECUTION_STORAGE_KEY)
    const parsed = rawValue ? JSON.parse(rawValue) : []
    if (!Array.isArray(parsed)) return []
    return parsed.filter((item) => {
      return Date.now() - Number(item?.updated_at || 0) <= TASK_EXECUTION_RETENTION_MS
    })
  } catch {
    return []
  }
}

function writeTaskExecutions(executions) {
  try {
    const trimmed = executions
      .filter((item) => Date.now() - Number(item?.updated_at || 0) <= TASK_EXECUTION_RETENTION_MS)
      .slice(0, TASK_EXECUTION_MAX_ITEMS)
    window.localStorage.setItem(TASK_EXECUTION_STORAGE_KEY, JSON.stringify(trimmed))
  } catch {
    // localStorage may be unavailable in restricted browser contexts.
  }
}

function buildTaskStatusQueryData(cached, status) {
  return {
    status,
    execution_id: cached.execution_id,
    command_id: cached.command_id,
    task_record_id: cached.task_record_id || '',
    task_id: cached.task_id || '',
    workflow_id: cached.workflow_id || '',
    result: cached.result ?? null,
    error: cached.error || '',
    automa_execution_id: cached.automa_execution_id || '',
    updated_at: cached.updated_at || Date.now(),
  }
}

function normalizeExecutionStatus(status, fallback = 'unknown') {
  const value = String(status || '').trim().toLowerCase()
  if (['queued', 'submitted', 'pending', 'running'].includes(value)) return 'running'
  if (['success', 'finished', 'done', 'completed'].includes(value)) return 'success'
  if (['error', 'failed', 'fail', 'timeout', 'stopped', 'cancelled', 'canceled'].includes(value)) return 'failed'
  if (['unknown', 'not_found', 'missing', 'lost'].includes(value)) return value
  return fallback
}

function resolveAutomaExecutionId(data = {}) {
  return String(
    data.automa_execution_id ||
      data.automaExecutionId ||
      data.automa_run_id ||
      data.automaRunId ||
      data.history_id ||
      data.historyId ||
      '',
  ).trim()
}

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
