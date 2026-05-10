import { defineComponent, h, ref, render } from 'vue'
import AppConfirmDialog from './AppConfirmDialog.vue'

export const APP_CONFIRM_TYPE = {
  danger: 'danger',
  warning: 'warning',
  info: 'info',
  question: 'question',
}

let appContext = null

export function setAppConfirmContext(context) {
  // App context 复用主应用上下文，保证动态确认框能使用全局组件和语言包
  appContext = context
}

function normalizeConfirmOptions(options = {}) {
  // Confirm options 合并默认配置，页面只传差异项
  return {
    title: options.title || '二次确认',
    message: options.message || '确认继续操作吗？',
    description: options.description || '',
    type: options.type || APP_CONFIRM_TYPE.warning,
    confirmText: options.confirmText || '确认',
    cancelText: options.cancelText || '取消',
    width: options.width || '420px',
    top: options.top || '18vh',
    closeOnClickModal: Boolean(options.closeOnClickModal),
  }
}

export function appConfirm(options = {}) {
  if (typeof document === 'undefined') {
    return Promise.resolve(false)
  }

  const confirmOptions = normalizeConfirmOptions(options)
  const container = document.createElement('div')
  document.body.appendChild(container)

  return new Promise((resolve) => {
    let resolved = false

    const cleanup = () => {
      // Cleanup unmount 动画结束后卸载动态组件
      render(null, container)
      container.remove()
    }

    const finish = (confirmed) => {
      // Finish once 防止取消和关闭事件重复触发
      if (resolved) return
      resolved = true
      resolve(Boolean(confirmed))
    }

    const ConfirmHost = defineComponent({
      setup() {
        const visible = ref(true)

        const closeWithConfirm = () => finish(true)
        const closeWithCancel = () => finish(false)

        return () =>
          h(AppConfirmDialog, {
            ...confirmOptions,
            modelValue: visible.value,
            'onUpdate:modelValue': (value) => {
              visible.value = value
            },
            onConfirm: closeWithConfirm,
            onCancel: closeWithCancel,
            onClosed: cleanup,
          })
      },
    })

    const vnode = h(ConfirmHost)
    if (appContext) {
      vnode.appContext = appContext
    }
    render(vnode, container)
  })
}
