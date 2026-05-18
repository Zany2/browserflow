<template>
  <!-- Hero panel 首页主视觉，先说明价值，再给出当前模式 -->
  <section class="home-page">
    <section class="hero-panel">
      <div class="hero-copy">
        <span class="eyebrow">双模式浏览器自动化平台</span>
        <h1>把 Automa 工作流，接成可调度的自动化能力</h1>
        <p class="summary">
          BrowserFlow 把浏览器执行端、工作流同步、任务下发和结果追踪放进同一条链路里：
          在 Windows 上专注本机调试与执行，部署到服务器后则承担多客户端调度和集中管理。
        </p>
      </div>
      <div class="mode-strip" aria-label="运行模式">
        <span class="mode-pill mode-pill--current">{{ currentModeLabel }}</span>
        <span class="mode-pill mode-pill--desktop">Windows 本地</span>
        <span class="mode-pill mode-pill--server">Server 调度</span>
      </div>
    </section>

    <!-- Quick start 当前模式下的高频入口，减少用户判断成本 -->
    <section class="quick-start-panel">
      <div class="section-heading">
        <div class="section-heading__copy">
          <span>从这里开始</span>
          <p>{{ currentModeDescription }}</p>
        </div>
        <a
          v-if="showAgentEntryLink"
          class="client-agent-link"
          :href="agentEntryUrl"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ agentEntryText }}
        </a>
      </div>
      <div class="quick-start-grid">
        <RouterLink
          v-for="item in currentModeActions"
          :key="item.to"
          class="quick-start-card"
          :to="item.to"
        >
          <strong>{{ item.title }}</strong>
          <span>{{ item.description }}</span>
        </RouterLink>
      </div>
    </section>

    <!-- Runtime modes 双模式说明，压缩成面向决策的描述 -->
    <section class="mode-grid">
      <article class="mode-card mode-card--desktop">
        <h2>Windows 本地使用</h2>
        <p>
          适合单机调试和本地自动化。启动受控浏览器后，可以直接管理工作流、验证模型、测试对话，
          并从当前浏览器里的 Automa 读取或执行流程。
        </p>
        <div class="route-list">
          <span>浏览器</span>
          <span>大模型</span>
          <span>对话</span>
          <span>工作流</span>
        </div>
      </article>

      <article class="mode-card mode-card--server">
        <h2>服务器任务调度</h2>
        <p>
          适合长期在线和多客户端协作。服务端负责同步工作流、绑定客户端、创建任务计划，
          并持续记录每次下发、执行状态和结果。
        </p>
        <div class="route-list">
          <span>工作流管理</span>
          <span>任务配置</span>
          <span>执行记录</span>
          <span>客户端</span>
        </div>
      </article>
    </section>

    <!-- Feature cards 核心能力，围绕一条自动化闭环展开 -->
    <section class="feature-grid">
      <article class="feature-item feature-item--blue">
        <h2>连接执行端</h2>
        <p>让本地浏览器或远程客户端稳定接入，先把“能执行”这件事接牢。</p>
      </article>
      <article class="feature-item feature-item--green">
        <h2>管理工作流</h2>
        <p>读取、同步、导入和保护 Automa 工作流，让流程从浏览器里真正走进系统。</p>
      </article>
      <article class="feature-item feature-item--orange">
        <h2>编排任务</h2>
        <p>把工作流、参数、客户端和 Cron 组合成任务，而不是每次都靠人工点一下。</p>
      </article>
      <article class="feature-item feature-item--purple">
        <h2>追踪结果</h2>
        <p>保留下发、回执、错误和结果，让自动化从“跑过”变成“可追溯”。</p>
      </article>
    </section>

    <!-- Workflow chain 主流程，用最短路径说明产品心智 -->
    <section class="flow-panel">
      <span>连接执行端</span>
      <strong>→</strong>
      <span>同步工作流</span>
      <strong>→</strong>
      <span>编排任务</span>
      <strong>→</strong>
      <span>自动执行</span>
      <strong>→</strong>
      <span>追踪结果</span>
    </section>

  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getRuntimeConfig } from '@/services/app'

const router = useRouter()
const runtimeMode = ref('')

// Mode copy 当前运行模式对应的首页说明文案
const modeCopy = {
  windows: {
    label: '当前：Windows 本地模式',
    description: '先启动浏览器，再读取工作流、验证模型，适合本机调试和单机执行。',
    actions: [
      { to: '/browser', title: '浏览器', description: '启动并查看受控浏览器状态' },
      { to: '/workflows', title: '工作流', description: '读取、打开或执行 Automa 流程' },
      { to: '/llm', title: '大模型', description: '配置供应商、模型与默认参数' },
      { to: '/chat', title: '对话', description: '快速验证模型是否可用' },
    ],
  },
  server: {
    label: '当前：Server 调度模式',
    description: '先接入客户端，再同步工作流和创建任务，适合远程调度与集中管理。',
    actions: [
      { to: '/clients', title: '客户端', description: '查看在线执行端与连接状态' },
      { to: '/automa', title: '工作流管理', description: '同步、导入并维护服务端流程' },
      { to: '/tasks', title: '任务配置', description: '绑定流程、客户端与调度规则' },
      { to: '/task-records', title: '执行记录', description: '追踪下发、回执与执行结果' },
    ],
  },
}

onMounted(async () => {
  try {
    const config = await getRuntimeConfig()
    runtimeMode.value = String(config?.mode || '')
  } catch {
    runtimeMode.value = ''
  }
})

// currentModeMeta 当前模式元信息，失败时回退到通用说明
const currentModeMeta = computed(() => {
  return (
    modeCopy[runtimeMode.value] || {
      label: '当前：运行模式读取中',
      description: '先确认后端运行模式，再进入对应的浏览器控制或任务调度页面。',
      actions: [
        { to: '/browser', title: '浏览器', description: '进入本地执行环境' },
        { to: '/automa', title: '工作流管理', description: '进入服务端流程管理' },
      ],
    }
  )
})

// currentModeLabel 当前运行模式标题
const currentModeLabel = computed(() => currentModeMeta.value.label)

// currentModeDescription 当前运行模式说明
const currentModeDescription = computed(() => currentModeMeta.value.description)

// currentModeActions 当前运行模式推荐入口
const currentModeActions = computed(() => currentModeMeta.value.actions)

// showAgentEntryLink 服务端模式才显示客户端执行页入口
const showAgentEntryLink = computed(() => runtimeMode.value === 'server')

// agentEntryUrl current agent page url 当前模式对应的入口地址
const agentEntryUrl = computed(() => {
  return new URL(router.resolve({ name: 'client-agent' }).href, window.location.href).href
})

// agentEntryText current agent page label 当前模式对应的入口文案
const agentEntryText = computed(() => {
  return '客户端执行页测试入口'
})
</script>

<style scoped lang="scss">
.home-page {
  display: grid;
  gap: 14px;
}

.hero-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 24px;
  padding: 22px 0 8px;
}

.hero-copy {
  min-width: 0;
}

.eyebrow {
  display: inline-flex;
  margin-bottom: 8px;
  color: #1d4ed8;
  font-size: 13px;
  font-weight: 700;
}

h1 {
  margin: 0 0 12px;
  color: #303133;
  font-size: 34px;
  line-height: 1.25;
}

.summary {
  max-width: 900px;
  margin: 0;
  color: #606266;
  line-height: 1.7;
}

.mode-strip {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.mode-pill {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 12px;
  font-size: 13px;
  font-weight: 700;
  border: 1px solid transparent;
  border-radius: 6px;
}

.mode-pill--desktop {
  color: #1d4ed8;
  background: #eff6ff;
  border-color: #bfdbfe;
}

.mode-pill--current {
  color: #047857;
  background: #ecfdf5;
  border-color: #a7f3d0;
}

.mode-pill--server {
  color: #b45309;
  background: #fffbeb;
  border-color: #fde68a;
}

.quick-start-panel {
  display: grid;
  gap: 12px;
  padding: 16px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.section-heading__copy {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 10px;
}

.section-heading__copy span {
  color: #303133;
  font-size: 18px;
  font-weight: 700;
}

.section-heading__copy p {
  margin: 0;
  color: #909399;
}

.quick-start-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.quick-start-card {
  display: grid;
  gap: 4px;
  min-height: 74px;
  padding: 12px 14px;
  background: #f8fbff;
  border: 1px solid #dce8f5;
  border-radius: 8px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.quick-start-card strong {
  color: #303133;
  font-size: 16px;
}

.quick-start-card span {
  color: #606266;
  line-height: 1.5;
}

.quick-start-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 10px 24px rgb(37 99 235 / 10%);
  transform: translateY(-1px);
}

.mode-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.mode-card {
  padding: 18px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-top: 5px solid transparent;
  border-radius: 8px;
}

.mode-card--desktop {
  border-top-color: #2563eb;
}

.mode-card--server {
  border-top-color: #f59e0b;
}

.mode-card h2,
.feature-item h2 {
  margin: 0 0 8px;
  color: #303133;
  font-size: 18px;
}

.mode-card p,
.feature-item p {
  margin: 0;
  color: #606266;
  line-height: 1.6;
}

.route-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.route-list span {
  padding: 5px 9px;
  color: #303133;
  font-size: 13px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.feature-item {
  padding: 14px 16px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-left: 5px solid transparent;
  border-radius: 8px;
}

.feature-item--blue {
  border-left-color: #3b82f6;
  background: #f8fbff;
}

.feature-item--green {
  border-left-color: #10b981;
  background: #f7fffb;
}

.feature-item--orange {
  border-left-color: #f97316;
  background: #fffaf5;
}

.feature-item--purple {
  border-left-color: #8b5cf6;
  background: #fbfaff;
}

.flow-panel {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  padding: 12px 16px;
  color: #606266;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.flow-panel span {
  color: #303133;
  font-weight: 700;
}

.flow-panel strong {
  color: #909399;
}

.client-agent-link {
  flex-shrink: 0;
  color: #909399;
  font-size: 13px;
}

.client-agent-link:hover {
  color: #1677ff;
}

@media (max-width: 768px) {
  .hero-panel,
  .mode-grid,
  .feature-grid,
  .quick-start-grid {
    grid-template-columns: 1fr;
  }

  .mode-strip {
    justify-content: flex-start;
  }

  .section-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .client-agent-link {
    align-self: flex-start;
  }
}
</style>
