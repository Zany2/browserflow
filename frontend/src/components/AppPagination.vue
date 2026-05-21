<template>
  <div class="app-pagination" :class="{ 'app-pagination--compact': compact }">
    <template v-if="compact">
      <el-button
        class="compact-button"
        :icon="ArrowLeft"
        :disabled="currentPageModel <= 1"
        circle
        title="上一页"
        aria-label="上一页"
        @click="currentPageModel = Math.max(currentPageModel - 1, 1)"
      />
      <span class="compact-status">{{ currentPageModel }} / {{ totalPages }}</span>
      <el-button
        class="compact-button"
        :icon="ArrowRight"
        :disabled="currentPageModel >= totalPages"
        circle
        title="下一页"
        aria-label="下一页"
        @click="currentPageModel = Math.min(currentPageModel + 1, totalPages)"
      />
    </template>
    <el-pagination
      v-else
      v-model:current-page="currentPageModel"
      v-model:page-size="pageSizeModel"
      :page-sizes="pageSizes"
      :pager-count="pagerCount"
      :total="total"
      :layout="layout"
      background
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'

const currentPageModel = defineModel('currentPage', {
  type: Number,
  default: 1,
})

const pageSizeModel = defineModel('pageSize', {
  type: Number,
  default: 10,
})

const props = defineProps({
  total: {
    type: Number,
    default: 0,
  },
  pageSizes: {
    type: Array,
    default: () => [10, 30, 60],
  },
  layout: {
    type: String,
    default: 'total, sizes, prev, pager, next, jumper',
  },
  pagerCount: {
    type: Number,
    default: 7,
  },
  compact: {
    type: Boolean,
    default: false,
  },
})

const totalPages = computed(() => {
  // Total pages 紧凑分页只展示页码状态，最少保留 1 页避免空列表显示 0/0
  return Math.max(Math.ceil(props.total / pageSizeModel.value), 1)
})
</script>

<style scoped>
.app-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  margin-top: 16px;
}

.app-pagination--compact {
  gap: 8px;
}

.compact-button {
  flex-shrink: 0;
}

.compact-status {
  min-width: 48px;
  color: #606266;
  font-size: 13px;
  line-height: 24px;
  text-align: center;
  white-space: nowrap;
}
</style>
