import request from '@/api/request'

export function listClients(params = {}) {
  return request({
    url: '/clients',
    params,
    showSuccessMessage: false,
  })
}

export function getClientDetail(id) {
  return request({
    url: `/clients/${id}`,
    showSuccessMessage: false,
  })
}

export function updateClient(id, data = {}) {
  return request({
    url: `/clients/${id}`,
    method: 'PUT',
    data,
    showSuccessMessage: false,
  })
}

export function checkClient(params = {}) {
  return request({
    url: '/clients/check',
    params,
    showSuccessMessage: false,
    showErrorMessage: false,
  })
}

export function offlineClient(id, reason = '') {
  return request({
    url: `/clients/${id}/offline`,
    method: 'POST',
    data: { reason },
    showSuccessMessage: false,
  })
}

export function batchOfflineClients(ids, reason = '') {
  return request({
    url: '/clients/batch-offline',
    method: 'POST',
    data: { ids, reason },
    showSuccessMessage: false,
  })
}

export function banClient(id, reason = '') {
  return request({
    url: `/clients/${id}/ban`,
    method: 'POST',
    data: { reason },
    showSuccessMessage: false,
  })
}

export function batchBanClients(ids, reason = '') {
  return request({
    url: '/clients/batch-ban',
    method: 'POST',
    data: { ids, reason },
    showSuccessMessage: false,
  })
}

export function unbanClient(id) {
  return request({
    url: `/clients/${id}/unban`,
    method: 'POST',
    showSuccessMessage: false,
  })
}
