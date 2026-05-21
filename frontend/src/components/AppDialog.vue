<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :top="top"
    :style="dialogStyle"
    :destroy-on-close="destroyOnClose"
    :close-on-click-modal="closeOnClickModal"
    class="app-dialog"
    @closed="handleClosed"
  >
    <slot />

    <template #footer>
      <slot name="footer">
        <el-button @click="handleCancel">{{ cancelText }}</el-button>
        <el-button
          type="primary"
          :disabled="confirmDisabled"
          :loading="loading"
          @click="emit('confirm')"
        >
          {{ confirmText }}
        </el-button>
      </slot>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

const visible = defineModel({
  type: Boolean,
  default: false,
})

const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  width: {
    type: String,
    default: '640px',
  },
  top: {
    type: String,
    default: '48px',
  },
  confirmText: {
    type: String,
    default: '确认',
  },
  cancelText: {
    type: String,
    default: '取消',
  },
  loading: {
    type: Boolean,
    default: false,
  },
  confirmDisabled: {
    type: Boolean,
    default: false,
  },
  destroyOnClose: {
    type: Boolean,
    default: true,
  },
  closeOnClickModal: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['cancel', 'closed', 'confirm'])

const dialogStyle = computed(() => ({
  '--app-dialog-top': props.top,
}))

function handleCancel() {
  // Cancel close 统一取消按钮关闭行为
  visible.value = false
  emit('cancel')
}

function handleClosed() {
  // Focus cleanup 避免关闭弹窗后触发按钮残留焦点高亮
  if (document.activeElement instanceof HTMLElement) {
    document.activeElement.blur()
  }
  emit('closed')
}
</script>

<style scoped lang="scss">
:global(.el-overlay-dialog:has(.app-dialog)) {
  overflow: hidden;
}

:global(.app-dialog) {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - var(--app-dialog-top, 72px) - 24px);
  overflow: hidden;
}

:global(.app-dialog .el-dialog__header),
:global(.app-dialog .el-dialog__footer) {
  flex: 0 0 auto;
}

:global(.app-dialog .el-dialog__body) {
  flex: 1 1 auto;
  min-height: 0;
  padding-right: 28px;
  overflow-y: auto;
  overscroll-behavior: contain;
}
</style>
