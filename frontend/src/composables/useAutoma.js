import { computed, ref } from 'vue'
import {
  getAutomaWorkflows,
  isAutomaInstalled,
  openAutomaWorkflow,
  runAutomaWorkflow,
} from '@/services/automaBridge'

// useAutoma 组合式函数，集中管理页面中的 Automa 扩展交互状态
export function useAutoma() {
  const workflows = ref([])
  const loading = ref(false)
  const error = ref('')

  // Installed 安装状态，依赖 Automa content script 注入的 body 标记
  const installed = computed(() => isAutomaInstalled())

  async function loadWorkflows() {
    loading.value = true
    error.value = ''

    try {
      // Bridge call 扩展桥接，从当前浏览器扩展读取 workflow 列表
      workflows.value = await getAutomaWorkflows()
    } catch (err) {
      workflows.value = []
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  function openWorkflow(workflowId) {
    // Open workflow 打开 Automa 工作流详情页
    openAutomaWorkflow(workflowId)
  }

  function executeWorkflow(workflowId, variables = {}) {
    // Execute workflow 执行工作流，变量会进入 Automa options.data.variables
    runAutomaWorkflow({ id: workflowId, variables })
  }

  return {
    error,
    loading,
    installed,
    workflows,
    loadWorkflows,
    openWorkflow,
    executeWorkflow,
  }
}
