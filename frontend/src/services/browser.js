import request from '@/api/request'
import { subscribeDesktopStatus } from '@/services/desktopWs'

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

export function selectBrowserBinPath() {
  return request({
    url: '/browser/paths/browser-bin/select',
    method: 'POST',
    showSuccessMessage: false,
  })
}

export function selectBrowserUserDataDir() {
  return request({
    url: '/browser/paths/user-data-dir/select',
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
  return subscribeDesktopStatus({
    type: 'browser_subscribe',
    responseType: 'browser_status',
    getPayload: (payload) => payload.browser,
    onMessage: onStatus,
    onError,
    errorMessage: '浏览器状态监听失败',
  })
}

export function subscribeAgentStatus(onStatus, onError) {
  return subscribeDesktopStatus({
    type: 'agent_status_subscribe',
    responseType: 'agent_status',
    getPayload: (payload) => payload.agents || [],
    onMessage: onStatus,
    onError,
    errorMessage: '执行端状态监听失败',
  })
}
