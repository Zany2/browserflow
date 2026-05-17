<template>
  <AppDialog
    v-model="visible"
    title="文件导入"
    width="680px"
    confirm-text="导入"
    :loading="loading"
    @confirm="handleSubmit"
  >
    <el-upload
      v-model:file-list="fileList"
      class="workflow-upload"
      drag
      :auto-upload="false"
      :limit="1"
      :accept="accept"
      :on-exceed="handleExceed"
    >
      <el-icon class="upload-icon">
        <UploadFilled />
      </el-icon>
      <div class="el-upload__text">拖拽 ZIP 文件到这里，或点击选择文件</div>
      <template #tip>
        <div class="el-upload__tip">只能导入一个 ZIP 压缩包，名称和工作流描述会从工作流 JSON 中解析</div>
      </template>
    </el-upload>
  </AppDialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { UploadFilled } from '@element-plus/icons-vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import AppDialog from '@/components/AppDialog.vue'

const visible = defineModel({
  type: Boolean,
  default: false,
})

defineProps({
  loading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['submit'])

const accept = '.zip,application/zip,application/x-zip-compressed'
const fileList = ref([])

watch(visible, (nextVisible) => {
  if (!nextVisible) fileList.value = []
})

function handleExceed(files) {
  fileList.value = files.slice(0, 1).map((file) => ({ name: file.name, raw: file }))
}

function handleSubmit() {
  const file = fileList.value[0]?.raw
  if (!file) {
    showWarningMessage('请选择要导入的 ZIP 文件')
    return
  }
  if (!isZipFile(file)) {
    showWarningMessage('只能导入 ZIP 文件')
    return
  }

  emit('submit', file)
}

function isZipFile(file) {
  return /\.zip$/i.test(file?.name || '') || /zip/i.test(file?.type || '')
}

function showWarningMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.warning,
    message,
  })
}
</script>

<style scoped lang="scss">
.workflow-upload {
  width: 100%;
}

.upload-icon {
  color: #909399;
  font-size: 48px;
}
</style>
