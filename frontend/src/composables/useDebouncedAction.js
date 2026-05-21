import { onBeforeUnmount } from 'vue'

// useDebouncedAction runs an action after quiet time 防抖执行动作
export function useDebouncedAction(action, delay = 200) {
  let timer = 0

  function run(...args) {
    cancel()
    timer = window.setTimeout(() => {
      timer = 0
      action(...args)
    }, delay)
  }

  function cancel() {
    if (!timer) return

    window.clearTimeout(timer)
    timer = 0
  }

  onBeforeUnmount(cancel)

  return {
    run,
    cancel,
  }
}
