<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :top="top"
    :destroy-on-close="destroyOnClose"
    :close-on-click-modal="closeOnClickModal"
    class="app-dialog"
    @closed="emit('closed')"
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
const visible = defineModel({
  type: Boolean,
  default: false,
})

defineProps({
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
    default: '72px',
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

function handleCancel() {
  // Cancel close 统一取消按钮关闭行为
  visible.value = false
  emit('cancel')
}
</script>
