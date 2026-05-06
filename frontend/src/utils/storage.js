const DEFAULT_COOKIE_PATH = '/'

const stringifyValue = (value) => {
  return typeof value === 'string' ? value : JSON.stringify(value)
}

const parseValue = (value, defaultValue = null) => {
  if (value === null || value === undefined) return defaultValue

  try {
    return JSON.parse(value)
  } catch {
    return value
  }
}

const createStorage = (target) => ({
  // Set storage 设置缓存
  set(key, value) {
    target.setItem(key, stringifyValue(value))
  },

  // Get storage 获取缓存
  get(key, defaultValue = null) {
    return parseValue(target.getItem(key), defaultValue)
  },

  // Remove storage 删除缓存
  remove(key) {
    target.removeItem(key)
  },

  // Clear storage 清空缓存
  clear() {
    target.clear()
  },

  // Has storage 是否存在
  has(key) {
    return target.getItem(key) !== null
  },
})

const getCookieMap = () => {
  return document.cookie.split('; ').reduce((result, item) => {
    const [key, ...value] = item.split('=')
    if (!key) return result

    result[decodeURIComponent(key)] = decodeURIComponent(value.join('='))
    return result
  }, {})
}

export const localCache = createStorage(window.localStorage)

export const sessionCache = createStorage(window.sessionStorage)

export const cookieCache = {
  // Set cookie 设置 Cookie
  set(key, value, options = {}) {
    const {
      expires,
      maxAge,
      path = DEFAULT_COOKIE_PATH,
      domain,
      secure = false,
      sameSite = 'Lax',
    } = options

    const cookieItems = [
      `${encodeURIComponent(key)}=${encodeURIComponent(stringifyValue(value))}`,
    ]

    if (expires instanceof Date) cookieItems.push(`expires=${expires.toUTCString()}`)
    if (typeof maxAge === 'number') cookieItems.push(`max-age=${maxAge}`)
    if (path) cookieItems.push(`path=${path}`)
    if (domain) cookieItems.push(`domain=${domain}`)
    if (secure) cookieItems.push('secure')
    if (sameSite) cookieItems.push(`samesite=${sameSite}`)

    document.cookie = cookieItems.join('; ')
  },

  // Get cookie 获取 Cookie
  get(key, defaultValue = null) {
    const cookieMap = getCookieMap()
    return Object.hasOwn(cookieMap, key)
      ? parseValue(cookieMap[key], defaultValue)
      : defaultValue
  },

  // Remove cookie 删除 Cookie
  remove(key, options = {}) {
    this.set(key, '', {
      ...options,
      expires: new Date(0),
    })
  },

  // Has cookie 是否存在
  has(key) {
    return Object.hasOwn(getCookieMap(), key)
  },
}

export const storage = {
  local: localCache,
  session: sessionCache,
  cookie: cookieCache,
}
