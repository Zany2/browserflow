import { API_BASE_URL } from '@/api/request'

export async function exportBrowserExecutorSkill() {
  const response = await fetch(`${API_BASE_URL}/browser-executor/export/skill`)
  const contentType = response.headers.get('content-type') || ''

  // Error json 解析后端统一 JSON 错误提示
  if (contentType.includes('application/json')) {
    const result = await response.json().catch(() => null)
    throw new Error(result?.message || '导出浏览器控制 Skill 失败')
  }

  if (!response.ok) {
    throw new Error(`导出浏览器控制 Skill 失败，HTTP 状态码：${response.status}`)
  }

  return response.blob()
}
