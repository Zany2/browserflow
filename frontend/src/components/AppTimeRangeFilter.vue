<template>
  <div class="app-time-range-filter">
    <el-date-picker
      v-model="startValue"
      class="time-picker"
      type="datetime"
      clearable
      :placeholder="startPlaceholder"
      :value-format="valueFormat"
    />
    <span class="time-separator">至</span>
    <el-date-picker
      v-model="endValue"
      class="time-picker"
      type="datetime"
      clearable
      :placeholder="endPlaceholder"
      :value-format="valueFormat"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'

const modelValue = defineModel({
  type: Array,
  default: () => [],
})

defineProps({
  startPlaceholder: {
    type: String,
    default: '开始时间',
  },
  endPlaceholder: {
    type: String,
    default: '结束时间',
  },
  valueFormat: {
    type: String,
    default: 'YYYY-MM-DD HH:mm:ss',
  },
})

const startValue = computed({
  get: () => modelValue.value?.[0] || '',
  set: (value) => {
    // Start value update 开始时间可单独更新
    modelValue.value = [value || '', modelValue.value?.[1] || '']
  },
})

const endValue = computed({
  get: () => modelValue.value?.[1] || '',
  set: (value) => {
    // End value update 结束时间可单独更新
    modelValue.value = [modelValue.value?.[0] || '', value || '']
  },
})
</script>

<style scoped>
.app-time-range-filter {
  display: inline-flex;
  align-items: center;
  gap: 0;
  width: max-content;
  min-width: 0;
}

.time-picker {
  box-sizing: border-box;
  flex: 0 0 var(--time-range-picker-width, 230px) !important;
  width: var(--time-range-picker-width, 230px) !important;
  min-width: 0;
}

.time-picker :deep(.el-input__wrapper) {
  min-width: 0;
}

.time-picker :deep(.el-input__inner) {
  min-width: 0;
}

.time-separator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  color: #909399;
  font-size: 13px;
  line-height: 32px;
  text-align: center;
}

@media (max-width: 640px) {
  .app-time-range-filter {
    display: flex;
    width: 100%;
    flex-direction: column;
    gap: 8px;
  }

  .time-picker {
    flex: 1 1 auto !important;
    width: 100%;
  }

  .time-separator {
    display: none;
  }
}
</style>
