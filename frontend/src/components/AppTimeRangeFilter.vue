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
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  width: 100%;
}

.time-picker {
  width: 100%;
}

.time-separator {
  flex-shrink: 0;
  color: #909399;
  font-size: 13px;
}

@media (max-width: 640px) {
  .app-time-range-filter {
    grid-template-columns: 1fr;
  }

  .time-separator {
    display: none;
  }
}
</style>
