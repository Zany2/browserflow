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
        <el-descriptions-item label="Automa">
          <el-tag :type="automaInstalled ? 'success' : 'info'">
            {{ automaInstalled ? `可用${automaVersion ? ` · ${automaVersion}` : ''}` : '检测中' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最后命令">{{ lastCommand || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Command ID">{{ lastCommandId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Execution ID">{{ lastExecutionId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="执行结果">{{ lastResult || '-' }}</el-descriptions-item>
      </el-descriptions>
    </section>
  </main>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import {
  getAutomaInfo,
  getAutomaWorkflows,
  openAutomaWorkflow,
  runAutomaWorkflow,
} from '@/services/automaBridge'
import { createAgentSocket } from '@/services/agentWs'

const route = useRoute()
const browserId = ref(resolveBrowserId())
const status = ref('connecting')
const automaInstalled = ref(false)
const automaVersion = ref('')
const lastCommand = ref('')
const lastCommandId = ref('')
const lastExecutionId = ref('')
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
      updateAutomaInfo(payload)
    },
    onMessage: handleSocketMessage,
    onError: (error) => {
      lastResult.value = error.message
      appMessage({ type: APP_MESSAGE_TYPE.error, message: error.message })
    },
  })
})

onBeforeUnmount(() => {
  closeSocket.value?.()
})

function handleSocketMessage(payload) {
  if (payload?.type === 'agent_registered' || payload?.type === 'agent_status') {
    updateAutomaInfo(payload)
  }
  if (payload?.type === 'agent_command') {
    lastCommandId.value = payload.command_id || ''
    lastExecutionId.value = payload.payload?.execution_id || payload.payload?.executionId || ''
  }
}

async function handleCommand(command, payload, rawMessage) {
  lastCommand.value = command
  lastCommandId.value = rawMessage?.command_id || lastCommandId.value
  lastExecutionId.value = payload.execution_id || payload.executionId || lastExecutionId.value

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
    const waitResult = Boolean(payload.wait_result ?? payload.waitResult ?? false)
    const result = await runAutomaWorkflow({
      id: payload.id || payload.workflow_id || payload.workflowId,
      publicId: payload.publicId || payload.public_id,
      variables: payload.variables || payload.params || {},
      checkParams: payload.check_params ?? payload.checkParams ?? false,
      executionId: payload.execution_id || payload.executionId || '',
      waitResult,
      timeout: payload.timeout || 300,
      returnData: payload.return_data || payload.returnData || null,
    })
    lastResult.value = waitResult ? `执行完成：${result.status || '-'}` : '已提交执行'
    return result
  }


  if (command === 'automa.workflow.open') {
    openAutomaWorkflow(payload.id || payload.workflow_id || payload.workflowId)
    lastResult.value = '已打开工作流'
    return { ok: true }
  }

  throw new Error(`不支持的 Agent 命令: ${command}`)
}

function resolveBrowserId() {
  // Browser id is bound when backend opens this agent page 浏览器 ID 由后端打开执行端页面时写入
  return String(route.query.browser_id || route.query.browserId || '').trim()
}

function updateAutomaInfo(payload) {
  const agent = Array.isArray(payload?.agents) ? payload.agents[0] : payload
  automaInstalled.value = Boolean(agent?.automa_installed ?? agent?.automaInstalled ?? automaInstalled.value)
  automaVersion.value = String(agent?.automa_version || agent?.automaVersion || automaVersion.value || '').trim()
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
