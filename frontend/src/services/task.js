import request from '@/api/request'
import { API_BASE_URL } from '@/api/request'

export function listTasks(params = {}) {
  return request({
    url: '/tasks',
    params,
    showSuccessMessage: false,
  })
}

export function getTaskDetail(id) {
  return request({
    url: `/tasks/${id}`,
    showSuccessMessage: false,
  })
}

export function createTask(data) {
  return request({
    url: '/tasks',
    method: 'POST',
    data,
    showSuccessMessage: false,
  })
}

export function updateTask(id, data) {
  return request({
    url: `/tasks/${id}`,
    method: 'PUT',
    data,
    showSuccessMessage: false,
  })
}

export function deleteTask(id) {
  return request({
    url: `/tasks/${id}`,
    method: 'DELETE',
    showSuccessMessage: false,
  })
}

export function executeTask(id, data) {
  return request({
    url: `/tasks/${id}/execute`,
    method: 'POST',
    data,
    showSuccessMessage: false,
  })
}

export function listTaskRecords(params = {}) {
  return request({
    url: '/task-records',
    params,
    showSuccessMessage: false,
  })
}

export function getTaskRecordDetail(id) {
  return request({
    url: `/task-records/${id}`,
    showSuccessMessage: false,
  })
}

export function deleteTaskRecords(ids = []) {
  return request({
    url: '/task-records/batch',
    method: 'DELETE',
    data: { ids },
    showSuccessMessage: false,
  })
}

export function getTaskRecordFileDownloadUrl(id) {
  return `${API_BASE_URL}/task-records/files/${encodeURIComponent(id)}/download`
}
