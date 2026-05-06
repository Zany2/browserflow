import request from '@/api/request'

export function getRuntimeConfig() {
  return request({
    url: '/app/runtime',
    showSuccessMessage: false,
    showErrorMessage: false,
  })
}
