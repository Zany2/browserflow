// formatDate formats date-like values 格式化日期时间
export function formatDate(value, { fallback = '', unixSeconds = true } = {}) {
  if (!value) return fallback

  const numericValue = Number(value)
  const dateValue = Number.isFinite(numericValue)
    ? unixSeconds && numericValue > 0 && numericValue < 10000000000
      ? numericValue * 1000
      : numericValue
    : value
  const date = new Date(dateValue)
  if (Number.isNaN(date.getTime())) return fallback

  return [
    date.getFullYear(),
    padDatePart(date.getMonth() + 1),
    padDatePart(date.getDate()),
  ].join('-') + ` ${padDatePart(date.getHours())}:${padDatePart(date.getMinutes())}:${padDatePart(date.getSeconds())}`
}

// formatEmpty returns fallback for empty values 格式化空值
export function formatEmpty(value, fallback = '', { treatDashAsEmpty = false } = {}) {
  if (value === undefined || value === null || value === '' || (treatDashAsEmpty && value === '-')) {
    return fallback
  }
  return String(value)
}

// formatJSON prints data as readable JSON 格式化 JSON 展示
export function formatJSON(value, fallback = {}) {
  return JSON.stringify(value ?? fallback, null, 2)
}

function padDatePart(value) {
  return String(value).padStart(2, '0')
}
