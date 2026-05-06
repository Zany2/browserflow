import axios from 'axios'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const SUCCESS_CODE = 20000

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
})

const getHttpErrorMessage = (error) => {
  if (error.code === 'ECONNABORTED') return '请求超时'
  if (!error.response) return '请求失败，后端服务异常'
  return error.response.data?.message || error.response.data?.error || `请求失败，HTTP 状态码：${error.response.status}`
}

const showResponseMessage = (result, config = {}) => {
  if (config.showErrorMessage === false) return
  if (!result?.message) return

  const isSuccess = result.code === SUCCESS_CODE
  if (isSuccess && config.showSuccessMessage === false) return

  appMessage({
    type: isSuccess ? APP_MESSAGE_TYPE.success : APP_MESSAGE_TYPE.error,
    message: result.message,
  })
}

// Response unwrap 响应解包，统一按后端 code 判断业务成功/失败
request.interceptors.response.use(
  (response) => {
    const result = response.data ?? {}

    if (typeof result.code !== 'number') {
      return Promise.reject(new Error('响应格式错误，缺少业务状态码'))
    }

    showResponseMessage(result, response.config)

    if (result.code !== SUCCESS_CODE) {
      return Promise.reject(new Error(result.message || '请求失败'))
    }

    return result.data
  },
  (error) => {
    const message = getHttpErrorMessage(error)
    if (error.config?.showErrorMessage !== false) {
      appMessage({
        type: APP_MESSAGE_TYPE.error,
        message,
      })
    }
    return Promise.reject(new Error(message))
  },
)

export default request
