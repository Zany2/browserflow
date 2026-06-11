<template>
  <AppDialog
    v-model="visible"
    :title="title"
    width="min(760px, calc(100vw - 32px))"
    confirm-text="复制"
    cancel-text="关闭"
    @confirm="emit('copy', value)"
  >
    <pre class="app-long-text-content">{{ value }}</pre>
  </AppDialog>
</template>

<script setup>
import AppDialog from './AppDialog.vue'

const visible = defineModel({
  type: Boolean,
  default: false,
})

defineProps({
  title: {
    type: String,
    default: '',
  },
  value: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['copy'])
</script>

<style scoped>
.app-long-text-content {
  box-sizing: border-box;
  max-height: min(520px, calc(100vh - 240px));
  min-height: 180px;
  margin: 0;
  padding: 12px;
  overflow: auto;
  color: #303133;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  background: #f8fafc;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}
</style>
