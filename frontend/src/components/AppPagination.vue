<template>
  <div class="app-pagination">
    <el-pagination v-model:current-page="currentPageModel" v-model:page-size="pageSizeModel" :page-sizes="pageSizes"
      :total="total" :layout="layout" background />
  </div>
</template>

<script setup>
const currentPageModel = defineModel('currentPage', {
  type: Number,
  default: 1,
})

const pageSizeModel = defineModel('pageSize', {
  type: Number,
  default: 10,
})

defineProps({
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
</style>
