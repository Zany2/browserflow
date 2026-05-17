import { AUTOMA_EVENTS } from '@/constants/automa'

const AUTOMA_INSTALL_PROBE_TIMEOUT_MS = 1500
const AUTOMA_INSTALL_PROBE_RETRY_INTERVAL_MS = 200
const AUTOMA_INSTALL_PROBE_CACHE_MS = 4000

let lastInstallProbeAt = 0
let lastInstallProbeResult = false
let installProbePromise = null

// isAutomaInstalled checks extension marker 检查扩展注入标记
export function isAutomaInstalled() {
  return Boolean(getInjectedAutomaMarker())
}

// getAutomaVersion reads extension version from injected markers 读取扩展注入的版本标记
export function getAutomaVersion() {
  const dataset = document.body?.dataset || {}
  const versionKeys = [
    'atmExtInstalled',
    'atmExtVersion',
    'atmVersion',
    'automaVersion',
    'automaExtVersion',
    'extensionVersion',
    'extVersion',
  ]

  for (const key of versionKeys) {
    const version = String(dataset[key] || '').trim()
    if (version) return version
  }

  return ''
}

// getAutomaInfo returns current extension metadata 返回当前扩展元信息
export async function getAutomaInfo() {
  const version = getAutomaVersion()
  const bridgeReady = await probeAutomaBridgeInstalled()

  return {
    installed: bridgeReady,
    version: bridgeReady ? version : '',
  }
}

export function getAutomaWorkflows({
  timeout = 6000,
  retryInterval = 300,
} = {}) {
  return new Promise((resolve, reject) => {
    let settled = false
    let timeoutTimer = 0
    let retryTimer = 0

    const cleanup = () => {
      window.clearTimeout(timeoutTimer)
      window.clearInterval(retryTimer)
      window.removeEventListener(AUTOMA_EVENTS.workflowsResponse, onResponse)
    }

    function onResponse(event) {
      if (settled) return

      settled = true
      cleanup()
      resolve(normalizeWorkflows(event.detail))
    }

    const dispatchRequest = () => {
      dispatchBridgeEvent({
        type: 'send-message',
        data: {
          type: 'get-workflows',
        },
      })
    }

    window.addEventListener(AUTOMA_EVENTS.workflowsResponse, onResponse)
    dispatchRequest()
    retryTimer = window.setInterval(dispatchRequest, retryInterval)

    timeoutTimer = window.setTimeout(() => {
      if (settled) return

      settled = true
      cleanup()
      reject(new Error('未检测到 Automa 响应，请确认扩展已安装并启用'))
    }, timeout)
  })
}

export function openAutomaWorkflow(workflowId) {
  if (!workflowId) return

  dispatchBridgeEvent({
    type: 'open-workflow',
    data: {
      workflowId,
    },
  })
}

export function importAutomaWorkflow(workflow) {
  if (!workflow) {
    throw new Error('导入工作流数据不能为空')
  }

  const requestId = createBridgeRequestId()

  return new Promise((resolve, reject) => {
    let timeoutTimer = 0

    // Import ack 导入回执，用于确认扩展侧已经写入本地存储
    const cleanup = () => {
      window.clearTimeout(timeoutTimer)
      window.removeEventListener(AUTOMA_EVENTS.importWorkflowResponse, onResponse)
    }

    function onResponse(event) {
      const detail = event.detail || {}
      if (detail.requestId !== requestId) return

      cleanup()
      if (detail.ok === false) {
        reject(new Error(detail.error || '同步到本地 Automa 失败'))
        return
      }

      resolve(detail.workflow || detail)
    }

    window.addEventListener(AUTOMA_EVENTS.importWorkflowResponse, onResponse)
    dispatchBridgeEvent({
      type: 'add-workflow',
      data: {
        requestId,
        workflow,
      },
    })

    timeoutTimer = window.setTimeout(() => {
      cleanup()
      reject(new Error('未收到 Automa 导入响应，请确认扩展已安装并启用'))
    }, 6000)
  })
}

export function runAutomaWorkflow({
  id,
  publicId,
  variables = {},
  checkParams = false,
  executionId = '',
  waitResult = false,
  timeout = 300,
  returnData = null,
} = {}) {
  if (!id && !publicId) {
    throw new Error('Automa workflow execution requires id or publicId')
  }

  const requestId = executionId || createBridgeRequestId()
  const detail = {
    id,
    publicId,
    options: {
      // Check params is handled by BrowserFlow before calling Automa.
      checkParams,
      browserFlowRequestId: requestId,
      browserFlowWaitResult: Boolean(waitResult),
      browserFlowReturnData: returnData || null,
    },
    data: {
      variables,
    },
  }

  if (waitResult) {
    return waitAutomaWorkflowResult({
      requestId,
      timeoutMs: Math.max(Number(timeout) || 300, 5) * 1000,
      dispatch: () => dispatchAutomaWorkflow(detail),
    })
  }

  dispatchAutomaWorkflow(detail)
  return Promise.resolve({
    ok: true,
    status: 'queued',
    request_id: requestId,
    execution_id: requestId,
  })
}

function dispatchAutomaWorkflow(detail) {
  window.dispatchEvent(
    new CustomEvent(AUTOMA_EVENTS.executeWorkflow, {
      detail,
    }),
  )
}

function waitAutomaWorkflowResult({ requestId, timeoutMs, dispatch }) {
  return new Promise((resolve, reject) => {
    let timeoutTimer = 0

    const cleanup = () => {
      window.clearTimeout(timeoutTimer)
      window.removeEventListener(AUTOMA_EVENTS.executeWorkflowResponse, onResponse)
    }

    function onResponse(event) {
      const detail = event.detail || {}
      if (detail.request_id !== requestId && detail.requestId !== requestId) return

      cleanup()
      resolve({
        ok: detail.ok !== false && detail.status !== 'error',
        ...detail,
        request_id: requestId,
        execution_id: detail.execution_id || requestId,
      })
    }

    window.addEventListener(AUTOMA_EVENTS.executeWorkflowResponse, onResponse)
    dispatch()

    timeoutTimer = window.setTimeout(() => {
      cleanup()
      reject(new Error(`Timed out waiting for Automa workflow result after ${Math.round(timeoutMs / 1000)} seconds`))
    }, timeoutMs)
  })
}

function dispatchBridgeEvent(detail) {
  window.dispatchEvent(
    new CustomEvent(AUTOMA_EVENTS.bridge, {
      detail,
    }),
  )
}

function createBridgeRequestId() {
  return `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

function getInjectedAutomaMarker() {
  return String(document.body?.dataset?.atmExtInstalled || '').trim()
}

async function probeAutomaBridgeInstalled() {
  if (installProbePromise) {
    return installProbePromise
  }

  const now = Date.now()
  if (now - lastInstallProbeAt < AUTOMA_INSTALL_PROBE_CACHE_MS) {
    return lastInstallProbeResult
  }

  // Bridge probe 使用工作流桥接响应判断扩展是否仍然可用
  installProbePromise = getAutomaWorkflows({
    timeout: AUTOMA_INSTALL_PROBE_TIMEOUT_MS,
    retryInterval: AUTOMA_INSTALL_PROBE_RETRY_INTERVAL_MS,
  })
    .then(() => true)
    .catch(() => false)
    .then((installed) => {
      lastInstallProbeAt = Date.now()
      lastInstallProbeResult = installed
      return installed
    })
    .finally(() => {
      installProbePromise = null
    })

  return installProbePromise
}

function normalizeWorkflows(workflows) {
  if (!workflows) return []
  return Array.isArray(workflows) ? workflows : Object.values(workflows)
}
