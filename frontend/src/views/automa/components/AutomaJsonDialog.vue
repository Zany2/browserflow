<template>
  <AppDialog
    v-model="visible"
    class="workflow-create-dialog"
    title="新增工作流"
    width="860px"
    confirm-text="保存"
    :loading="loading"
    @confirm="handleSubmit"
  >
    <div class="workflow-create-list">
      <div v-for="(workflow, index) in workflows" :key="workflow.key" class="workflow-create-item">
        <div class="item-header">
          <strong>工作流 {{ index + 1 }}</strong>
          <el-button link type="danger" :disabled="workflows.length === 1" @click="removeWorkflow(index)">
            删除
          </el-button>
        </div>

        <el-form label-width="88px">
          <el-form-item label="名称">
            <el-input v-model="workflow.name" clearable placeholder="未填写时从 JSON 中解析" />
          </el-form-item>

          <el-form-item label="工作流描述">
            <el-input v-model="workflow.description" clearable type="textarea" :rows="2" placeholder="未填写时从 JSON 中解析" />
          </el-form-item>

          <el-form-item label="是否可同步">
            <el-switch v-model="workflow.is_syncable" />
          </el-form-item>

          <el-form-item label="JSON 文件">
            <el-upload v-model:file-list="workflow.fileList" class="workflow-file" :auto-upload="false" :limit="1"
              :accept="accept" :on-exceed="(files) => handleExceed(workflow, files)">
              <el-button>选择文件</el-button>
              <template #tip>
                <span class="upload-tip">只能上传一个 Automa JSON 工作流文件</span>
              </template>
            </el-upload>
          </el-form-item>
        </el-form>
      </div>

      <el-button class="add-button" plain @click="addWorkflow">添加</el-button>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, watch } from 'vue'
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

const accept = '.json,application/json'
const workflows = ref([createWorkflowForm()])

watch(visible, (nextVisible) => {
  if (!nextVisible) workflows.value = [createWorkflowForm()]
})

function createWorkflowForm() {
  return {
    key: `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
    name: '',
    description: '',
    is_syncable: true,
    fileList: [],
  }
}

function addWorkflow() {
  workflows.value.push(createWorkflowForm())
}

function removeWorkflow(index) {
  if (workflows.value.length <= 1) return
  workflows.value.splice(index, 1)
}

function handleExceed(workflow, files) {
  workflow.fileList = files.slice(0, 1).map((file) => ({ name: file.name, raw: file }))
}

function handleSubmit() {
  const payload = workflows.value.map((workflow) => ({
    name: workflow.name.trim(),
    description: workflow.description.trim(),
    is_protected: !workflow.is_syncable,
    file: workflow.fileList[0]?.raw,
  }))

  if (payload.some((workflow) => !workflow.file)) {
    showWarningMessage('请为每个工作流选择 JSON 文件')
    return
  }

  if (payload.some((workflow) => !isJsonFile(workflow.file))) {
    showWarningMessage('只能上传 JSON 文件')
    return
  }

  emit('submit', payload)
}

function isJsonFile(file) {
  return /\.json$/i.test(file?.name || '') || file?.type === 'application/json'
}

function showWarningMessage(message) {
  appMessage({
    type: APP_MESSAGE_TYPE.warning,
    message,
  })
}
</script>

<style scoped lang="scss">
:global(.el-overlay-dialog:has(.workflow-create-dialog)) {
  overflow: hidden;
}

:global(.workflow-create-dialog) {
  display: flex;
  flex-direction: column;
  width: min(860px, calc(100vw - 32px));
  max-height: calc(100vh - 112px);
  overflow: hidden;
}

:global(.workflow-create-dialog .el-dialog__body) {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

:global(.workflow-create-dialog .el-dialog__header),
:global(.workflow-create-dialog .el-dialog__footer) {
  flex: 0 0 auto;
}

.workflow-create-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  flex: 1 1 auto;
  max-height: 100%;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding-right: 4px;
}

.workflow-create-item {
  flex: 0 0 auto;
  padding: 14px;
  border: 1px solid #e4e7ed;
}

.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.workflow-file {
  width: 100%;
}

.upload-tip {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

.add-button {
  align-self: flex-start;
}
</style>
