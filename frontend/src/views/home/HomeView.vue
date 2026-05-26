<template>
  <section class="home-page">
    <section class="hero-panel">
      <div class="hero-copy">
        <p class="eyebrow">Browser automation workspace</p>
        <h1>{{ heroTitle }}</h1>
        <p class="summary">
          {{ heroSummary }}
        </p>
      </div>

      <div class="hero-status">
        <span class="status-label">当前模式</span>
        <strong>{{ runtimeModeText }}</strong>
        <template v-if="backendAvailable && isServerMode">
          <span class="status-note">将客户端地址发给执行电脑打开</span>
          <div class="client-link-row">
            <a class="client-link" :href="clientAgentOpenUrl" target="_blank" rel="noreferrer">
              {{ clientAgentUrl }}
            </a>
            <el-button link type="primary" @click="copyClientAgentUrl">复制</el-button>
          </div>
        </template>
        <span v-else class="status-note" :class="{ 'status-note--error': !backendAvailable }">
          {{ runtimeStatusNote }}
        </span>
      </div>
    </section>

    <section class="quick-grid" aria-label="核心入口">
      <template v-if="backendAvailable">
        <RouterLink
          v-for="item in quickActions"
          :key="item.to"
          class="quick-card"
          :to="item.to"
        >
          <span class="quick-icon">{{ item.icon }}</span>
          <strong>{{ item.title }}</strong>
          <span>{{ item.desc }}</span>
        </RouterLink>
      </template>
      <template v-else>
        <button
          v-for="item in quickActions"
          :key="item.to"
          class="quick-card quick-card--disabled"
          type="button"
          @click="showBackendUnavailable"
        >
          <span class="quick-icon">{{ item.icon }}</span>
          <strong>{{ item.title }}</strong>
          <span>{{ item.desc }}</span>
        </button>
      </template>
    </section>

    <section class="intro-grid">
      <article v-for="card in introCards" :key="card.title" class="intro-card">
        <span class="intro-tag" :class="card.tagClass">{{ card.tag }}</span>
        <h2>{{ card.title }}</h2>
        <p>{{ card.desc }}</p>
      </article>
    </section>

    <section class="flow-panel" aria-label="使用流程">
      <span v-for="(step, index) in flowSteps" :key="step" class="flow-step">
        {{ step }}
        <strong v-if="index < flowSteps.length - 1">→</strong>
      </span>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { APP_MESSAGE_TYPE, appMessage } from '@/components/AppMessage'
import { getRuntimeConfig } from '@/services/app'
import { copyText } from '@/utils/browser'

const runtimeMode = ref('')
const backendAvailable = ref(true)
const frontendUrl = ref('')

// windowsQuickActions lists local workspace entries Windows 本地模式入口
const windowsQuickActions = [
  {
    icon: '01',
    title: '浏览器',
    desc: '维护浏览器实例和运行状态，作为自动化执行现场',
    to: '/browser',
  },
  {
    icon: '02',
    title: '工作流',
    desc: '读取、筛选、执行和导出浏览器里的 Automa 工作流',
    to: '/workflows',
  },
  {
    icon: '03',
    title: '大模型',
    desc: '集中维护模型厂商、模型名称和 API Key',
    to: '/llm',
  },
  {
    icon: '04',
    title: '对话',
    desc: '用流式对话快速验证模型配置和回答效果',
    to: '/chat',
  },
]

// serverQuickActions lists scheduler workspace entries Server 调度模式入口
const serverQuickActions = [
  {
    icon: '01',
    title: '工作流管理',
    desc: '导入、同步和维护服务端可调度的工作流',
    to: '/automa',
  },
  {
    icon: '02',
    title: '任务配置',
    desc: '把工作流编排成可手动或定时下发的任务',
    to: '/tasks',
  },
  {
    icon: '03',
    title: '执行记录',
    desc: '查看任务下发、运行状态和客户端回传结果',
    to: '/task-records',
  },
  {
    icon: '04',
    title: '客户端',
    desc: '管理在线执行电脑、插件状态和连接状态',
    to: '/clients',
  },
]

// windowsIntroCards describes local scenarios and strengths Windows 模式使用场景与优点
const windowsIntroCards = [
  {
    tag: '使用场景',
    tagClass: 'intro-tag--scene',
    title: '网页重复操作自动化',
    desc: '适合登录后页面操作、表单填写、数据读取、流程验证等需要稳定复用的浏览器任务。',
  },
  {
    tag: '使用场景',
    tagClass: 'intro-tag--scene',
    title: 'Automa 工作流调试',
    desc: '从当前浏览器读取工作流，查看参数、打开流程、带参数执行，让调试留在同一个控制台里。',
  },
  {
    tag: '项目优点',
    tagClass: 'intro-tag--advantage',
    title: '执行链路更清晰',
    desc: '浏览器、执行端、工作流和模型配置集中管理，减少在多个页面和工具之间来回切换。',
  },
  {
    tag: '项目优点',
    tagClass: 'intro-tag--advantage',
    title: '本地优先，轻量可控',
    desc: '优先服务本机 Windows 自动化场景，配置简单、反馈直接，也便于把稳定流程继续沉淀成 Skill。',
  },
]

// serverIntroCards describes scheduler scenarios and strengths Server 模式使用场景与优点
const serverIntroCards = [
  {
    tag: '使用场景',
    tagClass: 'intro-tag--scene',
    title: '集中调度多台客户端',
    desc: '适合把多台执行电脑接入同一个控制台，由服务端统一维护客户端、任务和执行状态。',
  },
  {
    tag: '使用场景',
    tagClass: 'intro-tag--scene',
    title: '工作流下发与同步',
    desc: '服务端保存可同步工作流，客户端打开执行页后接收任务并在本地 Automa 中执行。',
  },
  {
    tag: '项目优点',
    tagClass: 'intro-tag--advantage',
    title: '调度和执行解耦',
    desc: '服务端负责配置、派发和记录，客户端负责实际浏览器执行，更适合长期运行的自动化任务。',
  },
  {
    tag: '项目优点',
    tagClass: 'intro-tag--advantage',
    title: '执行过程可追踪',
    desc: '客户端在线状态、任务记录和结果回传集中展示，方便排查任务失败和确认执行进度。',
  },
]

const isServerMode = computed(() => runtimeMode.value === 'server')

// quickActions switches entries by runtime mode 首页入口按运行模式切换
const quickActions = computed(() => (isServerMode.value ? serverQuickActions : windowsQuickActions))

// introCards switches scenario cards by runtime mode 首页说明卡片按运行模式切换
const introCards = computed(() => (isServerMode.value ? serverIntroCards : windowsIntroCards))

// flowSteps switches automation lifecycle by runtime mode 首页流程步骤按运行模式切换
const flowSteps = computed(() =>
  isServerMode.value
    ? ['同步工作流', '生成客户端地址', '客户端连接', '下发任务', '查看记录']
    : ['配置浏览器', '连接执行端', '读取工作流', '填写参数', '执行验证'],
)

// heroTitle switches headline by runtime mode 首页标题按运行模式切换
const heroTitle = computed(() => {
  if (!backendAvailable.value) return 'BrowserFlow 后端服务暂不可用。'
  if (isServerMode.value) return 'BrowserFlow 是一个面向多客户端自动化调度的服务端控制台。'
  return 'BrowserFlow 是一个面向浏览器自动化的本地控制台。'
})

// heroSummary switches project intro by runtime mode 首页简介按运行模式切换
const heroSummary = computed(() => {
  if (!backendAvailable.value) {
    return '当前无法连接后端服务，首页仍可查看项目入口。请先启动后端服务，恢复后刷新页面或重新点击入口。'
  }
  if (isServerMode.value) {
    return '服务器模式用于集中管理工作流、任务、客户端和执行记录，适合把多台执行电脑接入同一个调度中心。'
  }
  return 'Windows 模式把浏览器实例、Automa 工作流、执行端连接、大模型配置和对话验证放在一起，适合把重复网页操作沉淀为可调试、可复用的本地流程。'
})

// clientAgentUrl builds a shareable client page address 生成可分享的客户端执行页地址
const clientAgentUrl = computed(() => {
  const configuredUrl = frontendUrl.value.trim().replace(/\/$/, '')
  if (configuredUrl) return `${configuredUrl}/#/client-agent`
  if (typeof window === 'undefined') return '#/client-agent'

  return `${window.location.origin}${window.location.pathname}#/client-agent`
})

// clientAgentOpenUrl opens a local manual node from the current browser 首页手动打开时使用 node-0
const clientAgentOpenUrl = computed(() => {
  return `${clientAgentUrl.value}?node_id=node-0`
})

onMounted(async () => {
  try {
    const config = await getRuntimeConfig()
    backendAvailable.value = true
    runtimeMode.value = String(config?.mode || '')
    frontendUrl.value = String(config?.frontend_url || '')
  } catch {
    backendAvailable.value = false
    runtimeMode.value = ''
    frontendUrl.value = ''
  }
})

// runtimeModeText displays current runtime mode 当前运行模式文案
const runtimeModeText = computed(() => {
  if (!backendAvailable.value) return '后端不可用'
  if (runtimeMode.value === 'windows') return 'Windows 本地版'
  if (runtimeMode.value === 'server') return '服务器调度版'
  return '自动识别中'
})

// runtimeStatusNote displays runtime health note 显示运行状态提示
const runtimeStatusNote = computed(() => {
  if (!backendAvailable.value) return '后端服务未启动或连接失败，除首页外的路由已暂时禁用。'
  return '公共能力与 Windows 本地能力优先可用'
})

function showBackendUnavailable() {
  appMessage({ type: APP_MESSAGE_TYPE.error, message: '后端服务不可用，请先启动后端' })
}

async function copyClientAgentUrl() {
  await copyText(clientAgentUrl.value)
  appMessage({ type: APP_MESSAGE_TYPE.success, message: '客户端地址已复制' })
}
</script>

<style scoped>
.home-page {
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto;
  gap: 14px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.hero-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 16px;
  padding: 24px;
  background:
    radial-gradient(circle at 92% 12%, rgba(64, 158, 255, 0.18), transparent 34%),
    linear-gradient(135deg, #f8fbff 0%, #ffffff 62%);
  border: 1px solid #dbeafe;
  border-radius: 18px;
}

.hero-copy {
  min-width: 0;
}

.eyebrow {
  margin: 0 0 10px;
  color: #2563eb;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

h1 {
  max-width: 760px;
  margin: 0;
  color: #111827;
  font-size: 32px;
  line-height: 1.22;
}

.summary {
  max-width: 820px;
  margin: 12px 0 0;
  color: #606266;
  font-size: 16px;
  line-height: 1.7;
}

.hero-status {
  align-self: stretch;
  display: grid;
  align-content: center;
  gap: 10px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid #e4e7ed;
  border-radius: 14px;
  box-shadow: 0 18px 45px rgba(37, 99, 235, 0.08);
}

.status-label {
  color: #909399;
  font-size: 13px;
}

.hero-status strong {
  color: #111827;
  font-size: 22px;
}

.status-note {
  color: #606266;
  line-height: 1.6;
}

.status-note--error {
  color: #c2410c;
  font-weight: 700;
}

.client-link-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.client-link {
  min-width: 0;
  overflow: hidden;
  color: #2563eb;
  font-size: 13px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.quick-card,
.intro-card,
.flow-panel {
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 14px;
}

.quick-card {
  display: grid;
  gap: 8px;
  min-height: 118px;
  padding: 16px;
  color: #303133;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.quick-card--disabled {
  width: 100%;
  font: inherit;
  text-align: left;
  cursor: not-allowed;
  opacity: 0.68;
}

.quick-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 14px 32px rgba(37, 99, 235, 0.1);
  transform: translateY(-2px);
}

.quick-card--disabled:hover {
  border-color: #e4e7ed;
  box-shadow: none;
  transform: none;
}

.quick-icon {
  color: #2563eb;
  font-size: 13px;
  font-weight: 900;
}

.quick-card strong {
  color: #111827;
  font-size: 18px;
}

.quick-card span:last-child {
  color: #606266;
  line-height: 1.55;
}

.intro-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  min-height: 0;
}

.intro-card {
  min-width: 0;
  overflow: hidden;
  padding: 16px;
}

.intro-tag {
  display: inline-flex;
  margin-bottom: 12px;
  padding: 5px 10px;
  font-size: 13px;
  font-weight: 800;
  border-radius: 999px;
}

.intro-tag--scene {
  color: #1d4ed8;
  background: #eff6ff;
}

.intro-tag--advantage {
  color: #b45309;
  background: #fffbeb;
}

.intro-card h2 {
  margin: 0 0 8px;
  color: #111827;
  font-size: 18px;
}

.intro-card p {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  color: #606266;
  line-height: 1.7;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 4;
}

.flow-panel {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  padding: 16px 18px;
}

.flow-step {
  color: #303133;
  font-weight: 700;
}

.flow-step strong {
  margin-left: 10px;
  color: #c0c4cc;
}

@media (max-width: 900px) {
  .hero-panel {
    grid-template-columns: 1fr;
  }

  .intro-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .quick-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .home-page {
    height: auto;
    overflow: visible;
  }

  .hero-panel {
    padding: 22px;
  }

  h1 {
    font-size: 28px;
  }

  .quick-grid {
    grid-template-columns: 1fr;
  }

  .intro-grid {
    grid-template-columns: 1fr;
  }
}
</style>
