<template>
  <main class="agent-page">
    <section class="agent-status">
      <h1>Browser Agent</h1>
      <el-tag :type="status === 'online' ? 'success' : 'info'">
        {{ status === 'online' ? '已连接' : '连接中' }}
      </el-tag>
      <p>此页面用于接收后端下发的命令，并通过页面桥接调用 Automa 扩展。</p>
      <el-descriptions border :column="1">
        <el-descriptions-item label="Browser ID">{{ browserId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最后命令">{{ lastCommand || '-' }}</el-descriptions-item>
        <el-descriptions-item label="执行结果">{{ lastResult || '-' }}</el-descriptions-item>
      </el-descriptions>
    </section>
  </main>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import {
  getAutomaInfo,
  getAutomaWorkflows,
  openAutomaWorkflow,
  runAutomaWorkflow,
} from '@/services/automaBridge'
import { createAgentSocket } from '@/services/agentWs'

const browserId = ref('')
const status = ref('connecting')
const lastCommand = ref('')
const lastResult = ref('')
const closeSocket = ref(null)

onMounted(() => {
  closeSocket.value = createAgentSocket({
    browserId: browserId.value,
    getAutomaInfo,
    getWorkflows: getAutomaWorkflows,
    onCommand: handleCommand,
    onStatus: (nextStatus) => {
      status.value = nextStatus
    },
    onRegistered: (payload) => {
      browserId.value = payload?.browser_id || ''
    },
    onError: (error) => {
      lastResult.value = error.message
      appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
    },
  })
})

onBeforeUnmount(() => {
  closeSocket.value?.()
})

async function handleCommand(command, payload) {
  lastCommand.value = command

  if (command === 'automa.workflow.list') {
    const workflows = await getAutomaWorkflows()
    lastResult.value = `读取到 ${workflows.length} 个工作流`
    return {
      ok: true,
      total: workflows.length,
      workflows,
    }
  }

  if (command === 'automa.workflow.run') {
    runAutomaWorkflow({
      id: payload.id || payload.workflow_id || payload.workflowId,
      publicId: payload.publicId || payload.public_id,
      variables: payload.variables || payload.params || {},
    })
    lastResult.value = '已发送执行命令'
    return { ok: true, status: 'queued' }
  }

  if (command === 'automa.workflow.open') {
    openAutomaWorkflow(payload.id || payload.workflow_id || payload.workflowId)
    lastResult.value = '已打开工作流'
    return { ok: true }
  }

  throw new Error(`不支持的 Agent 命令: ${command}`)
}
</script>

<style scoped>
:global(html),
:global(body),
:global(#app) {
  width: 100%;
  height: 100%;
  min-height: 100%;
}

.agent-page {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  width: 100dvw;
  height: 100dvh;
  overflow: hidden;
  background: #f5f7fa;
}

.agent-status {
  position: fixed;
  top: 50vh;
  left: 50vw;
  top: 50dvh;
  left: 50dvw;
  width: min(640px, calc(100vw - 64px));
  width: min(640px, calc(100dvw - 64px));
  max-width: 640px;
  padding: 24px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  transform: translate3d(-50%, -50%, 0);
}

.agent-status h1 {
  margin: 0 0 12px;
  color: #303133;
  font-size: 24px;
}

.agent-status p {
  margin: 12px 0 18px;
  color: #606266;
}
</style>
