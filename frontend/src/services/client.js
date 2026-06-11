import request from '@/api/request'
import { normalizeList } from '@/utils/list'

export function listClients(params = {}) {
  return request({
    url: '/clients',
    params,
    showSuccessMessage: false,
  })
}

export async function listAllClients(params = {}) {
  const rows = []
  let pageNum = 1
  let total = 0

  do {
    const data = await listClients({
      ...params,
      page_num: pageNum,
      page_size: 60,
    })
    const list = normalizeList(data, 'clients')
    rows.push(...list)
    total = Number(data?.total || rows.length)
    pageNum += 1
    if (list.length === 0) break
  } while (rows.length < total)

  return {
    list: rows,
    clients: rows,
    total: rows.length,
  }
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

export function batchUnbanClients(ids) {
  return request({
    url: '/clients/batch-unban',
    method: 'POST',
    data: { ids },
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
