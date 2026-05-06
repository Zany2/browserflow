import { AUTOMA_EVENTS } from '@/constants/automa'

// isAutomaInstalled checks extension marker 检查扩展注入标记
export function isAutomaInstalled() {
  return Boolean(document.body?.dataset?.atmExtInstalled)
}

// getAutomaVersion reads extension version from injected markers 读取扩展注入的版本标记
export function getAutomaVersion() {
  const dataset = document.body?.dataset || {}
  const versionKeys = [
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
export function getAutomaInfo() {
  return {
    installed: isAutomaInstalled(),
    version: getAutomaVersion(),
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

export function runAutomaWorkflow({ id, publicId, variables = {} }) {
  if (!id && !publicId) {
    throw new Error('执行 Automa 工作流需要 id 或 publicId')
  }

  window.dispatchEvent(
    new CustomEvent(AUTOMA_EVENTS.executeWorkflow, {
      detail: {
        id,
        publicId,
        data: {
          variables,
        },
      },
    }),
  )
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

function normalizeWorkflows(workflows) {
  if (!workflows) return []
  return Array.isArray(workflows) ? workflows : Object.values(workflows)
}
