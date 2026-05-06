<template>
  <!-- Hero panel 首页主视觉，说明项目定位和核心能力 -->
  <section class="home-page">
    <section class="hero-panel">
      <div class="hero-copy">
        <h1>BrowserFlow 控制台</h1>
        <p class="summary">
          BrowserFlow 是一套面向浏览器自动化的双形态项目：在 Windows 上，它可以作为本地工具直接运行，
          管理浏览器实例、大模型配置、对话调试和 Automa 工作流；部署到服务器后，它可以作为任务调度中心，
          统一维护客户端、服务端工作流、定时任务和执行记录。
        </p>
      </div>
      <div class="mode-strip" aria-label="运行模式">
        <span class="mode-pill mode-pill--desktop">Windows 本地版</span>
        <span class="mode-pill mode-pill--server">服务器调度版</span>
      </div>
    </section>

    <section class="mode-grid">
      <article class="mode-card mode-card--desktop">
        <h2>Windows 本地使用</h2>
        <p>
          适合放在个人电脑或自动化工作站上运行。后端可以内嵌前端 dist，打开可执行文件后直接进入控制台，
          用于启动浏览器、连接本机执行端、配置模型、测试对话，并从浏览器扩展侧读取或执行工作流。
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
          适合部署在长期在线的服务器上。服务器负责保存调度数据、接收客户端连接、同步可执行工作流、
          创建任务计划，并记录每一次任务下发、执行状态、错误信息和结果内容。
        </p>
        <div class="route-list">
          <span>工作流管理</span>
          <span>任务配置</span>
          <span>执行记录</span>
          <span>客户端</span>
        </div>
      </article>
    </section>

    <section class="feature-grid">
      <article class="feature-item feature-item--blue">
        <h2>浏览器执行环境</h2>
        <p>维护本地或远程浏览器配置，启动后与执行端保持状态同步，方便工作流调试和自动化运行。</p>
      </article>
      <article class="feature-item feature-item--green">
        <h2>大模型与对话</h2>
        <p>集中管理模型供应商、模型名称、API Key 和默认配置，用对话页快速验证模型是否可用。</p>
      </article>
      <article class="feature-item feature-item--orange">
        <h2>工作流同步</h2>
        <p>本地读取 Automa 工作流，服务器侧可导入、同步、保护和版本化，用于后续任务调度。</p>
      </article>
      <article class="feature-item feature-item--purple">
        <h2>任务编排闭环</h2>
        <p>任务配置、客户端选择、参数下发、执行记录和结果追踪串起来，形成可持续运行的自动化链路。</p>
      </article>
    </section>

    <section class="flow-panel">
      <span>编辑与调试</span>
      <strong>→</strong>
      <span>同步工作流</span>
      <strong>→</strong>
      <span>创建任务</span>
      <strong>→</strong>
      <span>客户端执行</span>
      <strong>→</strong>
      <span>记录结果</span>
    </section>

    <a class="client-agent-link" :href="agentEntryUrl" target="_blank" rel="noopener noreferrer">
      {{ agentEntryText }}
    </a>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getRuntimeConfig } from '@/services/app'

const router = useRouter()
const runtimeMode = ref('')

onMounted(async () => {
  try {
    const config = await getRuntimeConfig()
    runtimeMode.value = String(config?.mode || '')
  } catch {
    runtimeMode.value = ''
  }
})

// agentEntryUrl current agent page url 当前模式对应的入口地址
const agentEntryUrl = computed(() => {
  const routeName = runtimeMode.value === 'windows' ? 'browser-agent' : 'client-agent'
  return new URL(router.resolve({ name: routeName }).href, window.location.href).href
})

// agentEntryText current agent page label 当前模式对应的入口文案
const agentEntryText = computed(() => {
  return runtimeMode.value === 'windows' ? '浏览器执行页测试入口' : '客户端执行页测试入口'
})
</script>

<style scoped>
.home-page {
  display: grid;
  gap: 18px;
}

.hero-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 24px;
  padding: 36px 0 18px;
}

.hero-copy {
  min-width: 0;
}

h1 {
  margin: 0 0 16px;
  color: #303133;
  font-size: 34px;
}

.summary {
  max-width: 900px;
  margin: 0;
  color: #606266;
  line-height: 1.8;
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

.mode-pill--server {
  color: #b45309;
  background: #fffbeb;
  border-color: #fde68a;
}

.mode-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.mode-card {
  padding: 22px;
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
  margin: 0 0 10px;
  color: #303133;
  font-size: 18px;
}

.mode-card p,
.feature-item p {
  margin: 0;
  color: #606266;
  line-height: 1.7;
}

.route-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
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
  padding: 18px;
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
  padding: 16px 18px;
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
  justify-self: end;
  color: #909399;
  font-size: 13px;
}

.client-agent-link:hover {
  color: #1677ff;
}

@media (max-width: 768px) {
  .hero-panel,
  .mode-grid,
  .feature-grid {
    grid-template-columns: 1fr;
  }

  .mode-strip {
    justify-content: flex-start;
  }

  .client-agent-link {
    justify-self: start;
  }
}
</style>
