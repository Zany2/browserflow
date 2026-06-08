import NProgress from 'nprogress'
import { createRouter, createWebHashHistory } from 'vue-router'
import AutomaView from '@/views/automa/AutomaView.vue'
import ClientView from '@/views/client/ClientView.vue'
import ClientAgentView from '@/views/client-agent/ClientAgentView.vue'
import DefaultLayout from '@/layouts/DefaultLayout.vue'
import BrowserAgentView from '@/views/browser-agent/BrowserAgentView.vue'
import BrowserView from '@/views/browser/BrowserView.vue'
import HomeView from '@/views/home/HomeView.vue'
import LLMChatView from '@/views/llm-chat/LLMChatView.vue'
import LLMConfigView from '@/views/llm-config/LLMConfigView.vue'
import NotFoundView from '@/views/not-found/NotFoundView.vue'
import TaskRecordView from '@/views/task/TaskRecordView.vue'
import TaskView from '@/views/task/TaskView.vue'
import WorkflowListView from '@/views/workflow-list/WorkflowListView.vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import { getRuntimeConfig } from '@/services/app'

import 'nprogress/nprogress.css'
import '@/styles/nprogress.scss'

NProgress.configure({
  showSpinner: false,
})

// Routes page routes 页面路由，统一挂载在默认布局中
export const routes = [
  {
    path: '/browser-agent',
    name: 'browser-agent',
    component: BrowserAgentView,
  },
  {
    path: '/client-agent',
    name: 'client-agent',
    meta: { title: 'BrowserFlow-客户端' },
    component: ClientAgentView,
  },
  {
    path: '/',
    component: DefaultLayout,
    children: [
      {
        path: '',
        name: 'home',
        component: HomeView,
      },
      {
        path: 'browser',
        name: 'browser',
        component: BrowserView,
      },
      {
        path: 'llm',
        name: 'llm',
        component: LLMConfigView,
      },
      {
        path: 'chat',
        name: 'chat',
        component: LLMChatView,
      },
      {
        path: 'workflows',
        name: 'workflows',
        component: WorkflowListView,
      },
      {
        path: 'automa',
        name: 'automa',
        component: AutomaView,
      },
      {
        path: 'tasks',
        name: 'tasks',
        component: TaskView,
      },
      {
        path: 'task-records',
        name: 'task-records',
        component: TaskRecordView,
      },
      {
        path: 'clients',
        name: 'clients',
        component: ClientView,
      },
      {
        path: ':pathMatch(.*)*',
        name: 'not-found',
        component: NotFoundView,
      },
    ],
  },
]

// Router instance uses hash history 路由实例，使用 hash 模式生成带 # 的地址
const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes,
})

let runtimeConfig = {
  mode: '',
  disabled_routes: [],
  backend_available: true,
}

// loadRuntimeConfig loads and caches runtime mode 加载并缓存运行模式
async function loadRuntimeConfig() {
  try {
    runtimeConfig = {
      ...((await getRuntimeConfig()) || runtimeConfig),
      backend_available: true,
    }
  } catch {
    runtimeConfig = {
      mode: '',
      disabled_routes: [],
      backend_available: false,
    }
  }
  return runtimeConfig
}

// Route progress handles route switching progress 路由切换进度条
router.beforeEach(async (to) => {
  NProgress.start()
  const config = await loadRuntimeConfig()
  if (!config.backend_available && to.name !== 'home') {
    appMessage({ type: APP_MESSAGE_TYPE.error, message: '后端服务不可用，请先启动后端' })
    return { name: 'home', replace: true }
  }

  const disabledRoutes = new Set(config.disabled_routes || [])
  if (disabledRoutes.has(to.path)) {
    return { name: 'home', replace: true }
  }
  return true
})

router.afterEach((to) => {
  setDocumentTitle(to)
  NProgress.done()
})

router.onError(() => {
  NProgress.done()
})

function setDocumentTitle(route) {
  // Page title 路由标题，未配置时恢复默认标题
  document.title = route.meta?.title || 'BrowserFlow'
}

export default router
