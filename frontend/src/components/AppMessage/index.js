import { ElMessage } from 'element-plus'

export const APP_MESSAGE_TYPE = {
  success: 'success',
  warning: 'warning',
  info: 'info',
  error: 'error',
}

const MESSAGE_TYPES = Object.values(APP_MESSAGE_TYPE)

const getMessageType = (type) => {
  return MESSAGE_TYPES.includes(type) ? type : APP_MESSAGE_TYPE.info
}

// App message wrapper.
export const appMessage = ({
  type = APP_MESSAGE_TYPE.info,
  message = 'Message',
  html = '',
  offset = 80,
  duration = 1000,
} = {}) => {
  return ElMessage({
    type: getMessageType(type),
    message: html || message,
    dangerouslyUseHTMLString: Boolean(html),
    offset,
    duration,
  })
}
