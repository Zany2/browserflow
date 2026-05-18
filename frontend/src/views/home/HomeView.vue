<template>
  <section class="home-page">
    <section class="hero-panel">
      <div class="hero-copy">
        <p class="eyebrow">Browser automation command center</p>
        <h1>把浏览器、模型、工作流和任务调度放在一个控制台里。</h1>
        <p class="summary">
          BrowserFlow 既可以作为 Windows 本地自动化工具，也可以部署为服务器调度中心。
          你可以管理浏览器配置、验证大模型、读取 Automa 工作流，并把可复用流程沉淀为任务。
        </p>
      </div>

      <div class="hero-status">
        <span class="status-label">当前模式</span>
        <strong>{{ runtimeModeText }}</strong>
        <span class="status-note">执行端请在目标浏览器中启动</span>
      </div>
    </section>

    <section class="quick-grid" aria-label="核心入口">
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
    </section>

    <section class="mode-grid">
      <article class="mode-card">
        <span class="mode-tag mode-tag--desktop">Windows 本地</span>
        <h2>适合个人电脑或自动化工作站</h2>
        <p>浏览器执行端由目标浏览器承载；控制台负责管理配置、读取工作流和下发执行请求。</p>
      </article>

      <article class="mode-card">
        <span class="mode-tag mode-tag--server">服务器调度</span>
        <h2>适合长期在线的任务中心</h2>
        <p>统一管理客户端、工作流、定时任务和执行记录，让自动化流程可追踪、可复用。</p>
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
import { getRuntimeConfig } from '@/services/app'

const runtimeMode = ref('')

// quickActions shows primary navigation entries 首页核心入口
const quickActions = [
  {
    icon: '01',
    title: '浏览器',
    desc: '维护本地浏览器实例配置',
    to: '/browser',
  },
  {
    icon: '02',
    title: '大模型',
    desc: '维护厂商、配置名称和模型',
    to: '/llm',
  },
  {
    icon: '03',
    title: '对话',
    desc: '用 SSE 流式验证模型效果',
    to: '/chat',
  },
  {
    icon: '04',
    title: '任务',
    desc: '编排工作流并追踪执行结果',
    to: '/tasks',
  },
]

// flowSteps describes the automation lifecycle 首页流程步骤
const flowSteps = ['配置环境', '验证模型', '同步工作流', '创建任务', '查看结果']

onMounted(async () => {
  try {
    const config = await getRuntimeConfig()
    runtimeMode.value = String(config?.mode || '')
  } catch {
    runtimeMode.value = ''
  }
})

// runtimeModeText displays current runtime mode 当前运行模式文案
const runtimeModeText = computed(() => {
  if (runtimeMode.value === 'windows') return 'Windows 本地版'
  if (runtimeMode.value === 'server') return '服务器调度版'
  return '自动识别中'
})
</script>

<style scoped>
.home-page {
  display: grid;
  gap: 18px;
}

.hero-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 260px;
  gap: 20px;
  padding: 28px;
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
  font-size: 36px;
  line-height: 1.22;
}

.summary {
  max-width: 820px;
  margin: 16px 0 0;
  color: #606266;
  font-size: 16px;
  line-height: 1.8;
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

.quick-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.quick-card,
.mode-card,
.flow-panel {
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 14px;
}

.quick-card {
  display: grid;
  gap: 8px;
  min-height: 138px;
  padding: 18px;
  color: #303133;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.quick-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 14px 32px rgba(37, 99, 235, 0.1);
  transform: translateY(-2px);
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

.mode-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.mode-card {
  padding: 20px;
}

.mode-tag {
  display: inline-flex;
  margin-bottom: 12px;
  padding: 5px 10px;
  font-size: 13px;
  font-weight: 800;
  border-radius: 999px;
}

.mode-tag--desktop {
  color: #1d4ed8;
  background: #eff6ff;
}

.mode-tag--server {
  color: #b45309;
  background: #fffbeb;
}

.mode-card h2 {
  margin: 0 0 8px;
  color: #111827;
  font-size: 18px;
}

.mode-card p {
  margin: 0;
  color: #606266;
  line-height: 1.7;
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
  .hero-panel,
  .mode-grid {
    grid-template-columns: 1fr;
  }

  .quick-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .hero-panel {
    padding: 22px;
  }

  h1 {
    font-size: 28px;
  }

  .quick-grid {
    grid-template-columns: 1fr;
  }
}
</style>
