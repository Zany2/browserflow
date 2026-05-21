export const DEFAULT_PAGE_SIZES = [10, 30, 60]

// normalizeList reads common list response shapes 归一化常见列表响应结构
export function normalizeList(data, fallbackKey = 'list') {
  if (Array.isArray(data)) return data

  const list = data?.list || data?.[fallbackKey] || data?.candidates || data?.items || []
  return Array.isArray(list) ? list : []
}

// getSafePage clamps page to available range 限制页码在有效范围内
export function getSafePage({ total, page, size }) {
  const maxPage = Math.max(Math.ceil(Number(total || 0) / Number(size || 1)), 1)
  return Math.min(Number(page || 1), maxPage)
}

// normalizeText returns a trimmed string 归一化文本输入
export function normalizeText(value) {
  return String(value || '').trim()
}
