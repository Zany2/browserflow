import request from '@/api/request'

let runtimeConfigPromise = null
let runtimeConfigCache = null

export function getRuntimeConfig() {
  if (!runtimeConfigPromise) {
    runtimeConfigPromise = request({
      url: '/app/runtime',
      showSuccessMessage: false,
      showErrorMessage: false,
    })
      .then((config) => {
        runtimeConfigCache = config || {}
        return runtimeConfigCache
      })
      .catch((error) => {
        runtimeConfigPromise = null
        runtimeConfigCache = null
        throw error
      })
  }
  return runtimeConfigPromise
}

export function getCachedRuntimeConfig() {
  return runtimeConfigCache
}

export function resetRuntimeConfigCache() {
  runtimeConfigPromise = null
  runtimeConfigCache = null
}

export function getServerDashboard() {
  return request({
    url: '/app/server-dashboard',
    showSuccessMessage: false,
    showErrorMessage: false,
  })
}
