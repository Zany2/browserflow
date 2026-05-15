import request from '@/api/request'

import { API_BASE_URL } from '@/api/request'

const AUTOMA_SOURCE = {
  import: 1,
  manual: 1,
  client: 2,
  sync: 2,
}

export function listAutomaWorkflows(params = {}) {
  return request({
    url: '/workflows',
    params: normalizeListParams(params),
    showSuccessMessage: false,
  })
}

export function listAgentAutomaWorkflows(params = {}) {
  return request({
    url: '/workflows/agent/workflows',
    params,
    showErrorMessage: false,
    showSuccessMessage: false,
  })
}

export function getAutomaWorkflowDetail(id) {
  return request({
    url: `/workflows/${id}`,
    showSuccessMessage: false,
  })
}

export function createAutomaWorkflow(data) {
  const workflows = Array.isArray(data) ? data : [data]

  return request({
    url: '/workflows',
    method: 'POST',
    data: buildWorkflowFormData(workflows),
    showSuccessMessage: false,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function importAutomaWorkflow(data) {
  return createAutomaWorkflow(data)
}

export function importAutomaWorkflowFiles(file) {
  const formData = new FormData()
  formData.append('file', file)

  return request({
    url: '/workflows/import',
    method: 'POST',
    data: formData,
    showSuccessMessage: false,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function batchImportAutomaWorkflows(workflows) {
  return createAutomaWorkflow(workflows)
}

export function updateAutomaWorkflow(id, data) {
  return request({
    url: `/workflows/${id}`,
    method: 'PUT',
    data: {
      ...data,
      id,
      source: normalizeSource(data?.source),
    },
    showSuccessMessage: false,
  })
}

export function updateAutomaWorkflowProtected(id, data) {
  return request({
    url: `/workflows/${id}/syncable`,
    method: 'PUT',
    data: {
      ...data,
      id,
    },
    showSuccessMessage: false,
  })
}

export function deleteAutomaWorkflow(id) {
  return batchDeleteAutomaWorkflows([id])
}

export function batchDeleteAutomaWorkflows(ids) {
  return request({
    url: '/workflows',
    method: 'DELETE',
    data: { ids },
    showSuccessMessage: false,
  })
}

export function syncAutomaWorkflowCache(sourceIp = '', workflows = []) {
  return request({
    url: '/workflows/sync',
    method: 'POST',
    data: {
      source_ip: sourceIp,
      workflows,
    },
    showSuccessMessage: false,
  })
}

export function listAutomaSyncCandidates(sourceIp, params = {}) {
  return request({
    url: '/workflows/sync-candidates',
    params: {
      ...params,
      mode: 'client',
      source_ip: sourceIp,
    },
    showSuccessMessage: false,
  })
}

export function listAutomaSyncCandidatesByWorkflow(automaId, params = {}) {
  return request({
    url: '/workflows/sync-candidates',
    params: {
      ...params,
      mode: 'workflow',
      automa_id: automaId,
    },
    showSuccessMessage: false,
  })
}

export function syncAutomaWorkflowsByIp(sourceIp, workflowIds, workflows = []) {
  return request({
    url: '/workflows/sync',
    method: 'POST',
    data: {
      source_ip: sourceIp,
      workflow_ids: workflowIds,
      workflows,
    },
    showSuccessMessage: false,
  })
}

export function openAgentAutomaWorkflow(id, browserId = '') {
  return request({
    url: `/workflows/${id}/open`,
    method: 'POST',
    data: {
      browser_id: browserId,
    },
    showErrorMessage: false,
    showSuccessMessage: false,
  })
}

export function runAgentAutomaWorkflow(id, browserId = '', variables = {}) {
  return request({
    url: `/workflows/${id}/run`,
    method: 'POST',
    data: {
      browser_id: browserId,
      variables,
    },
    showErrorMessage: false,
    showSuccessMessage: false,
  })
}

export async function exportAgentAutomaSkill({
  browserId = '',
  scope = 'filtered',
  workflowIds = [],
} = {}) {
  const response = await fetch(`${API_BASE_URL}/workflows/agent/export/skill`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      browser_id: browserId,
      scope,
      workflow_ids: workflowIds,
    }),
  })
  const contentType = response.headers.get('content-type') || ''

  // Error json 后端业务失败时仍返回统一 JSON，需要先解析提示信息
  if (contentType.includes('application/json')) {
    const result = await response.json().catch(() => null)
    throw new Error(result?.message || '导出 Skill 失败')
  }

  if (!response.ok) {
    throw new Error(`导出 Skill 失败，HTTP 状态码：${response.status}`)
  }

  return response.blob()
}

function normalizeListParams(params) {
  const source = normalizeSource(params.source)
  return {
    ...params,
    page_num: params.page_num || 1,
    page_size: params.page_size || 60,
    source,
  }
}

function normalizeSource(source) {
  if (source === '' || source === undefined || source === null) return 0
  if (typeof source === 'number') return source
  return AUTOMA_SOURCE[source] || Number(source) || 0
}

function buildWorkflowFormData(workflows) {
  const formData = new FormData()
  const workflowMetas = []

  workflows.forEach((item, index) => {
    const workflow = item?.data || item?.workflow || item
    const file = item?.file || (workflow instanceof File ? workflow : null)
    const name = normalizeWorkflowMetaText(item?.name || (file ? '' : workflow?.name))
    const description = normalizeWorkflowMetaText(item?.description || (file ? '' : workflow?.description))
    const source = normalizeSource(item?.source || 1)
    const isProtected = Boolean(item?.is_protected ?? item?.isProtected)

    workflowMetas.push({
      name,
      description,
      source: source || 1,
      is_protected: isProtected,
    })
    formData.append('workflow_files', file || workflowToFile(workflow, name, index))
  })

  formData.append('workflow_metas', JSON.stringify(workflowMetas))

  return formData
}

function workflowToFile(workflow, name, index) {
  const content = JSON.stringify(workflow || {}, null, 2)
  const filename = `${sanitizeFilename(name) || `workflow-${index + 1}`}.json`
  return new File([content], filename, { type: 'application/json' })
}

function normalizeWorkflowMetaText(value) {
  return String(value || '').trim()
}

function sanitizeFilename(value) {
  return String(value || '')
    .trim()
    .replace(/[\\/:*?"<>|]+/g, '-')
}
