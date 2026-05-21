<template>
  <span class="app-selection-summary">
    已选择 {{ count }} 个{{ unit }}
  </span>
</template>

<script setup>
defineProps({
  count: {
    type: Number,
    default: 0,
  },
  unit: {
    type: String,
    default: '',
  },
})
</script>

<style scoped lang="scss">
.app-selection-summary {
  flex-shrink: 0;
  color: #606266;
  font-size: 13px;
  line-height: 24px;
  white-space: nowrap;
}
</style>
