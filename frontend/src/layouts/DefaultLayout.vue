<template>
  <div class="default-layout">
    <header class="app-header">
      <!-- Brand entry 品牌入口，点击返回首页 -->
      <RouterLink class="brand" to="/">
        <img class="brand-logo" :src="layoutLogo" alt="" aria-hidden="true" />
        <span>BrowserFlow</span>
      </RouterLink>

      <!-- Navigation main navigation 主导航，按业务模块排列 -->
      <nav class="nav-links" aria-label="Main navigation">
        <RouterLink to="/"><span class="nav-link-text" data-text="首页">首页</span></RouterLink>
        <template v-for="section in navSections" :key="section.key">
          <span class="nav-separator">|</span>
          <template v-for="item in section.items" :key="item.to">
            <RouterLink v-if="!isRouteDisabled(item.to)" :to="item.to">
              <span class="nav-link-text" :data-text="item.label">{{ item.label }}</span>
            </RouterLink>
            <span v-else class="nav-disabled" aria-disabled="true">
              <span class="nav-link-text" :data-text="item.label">{{ item.label }}</span>
            </span>
          </template>
        </template>
      </nav>
    </header>

    <main class="page-container">
      <!-- Page outlet 页面出口，子路由内容渲染在统一布局内 -->
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import layoutLogo from '@/assets/images/layout-logo.png'
import { getRuntimeConfig } from '@/services/app'

const runtimeConfig = ref({
  mode: '',
  disabled_routes: [],
})

// Navigation groups disabled by runtime mode 导航分组，按运行模式禁用部分路由
const navSections = [
  {
    key: 'desktop',
    items: [
      { to: '/browser', label: '浏览器' },
      { to: '/llm', label: '大模型' },
      { to: '/chat', label: '对话' },
      { to: '/workflows', label: '工作流' },
    ],
  },
  {
    key: 'server',
    items: [
      { to: '/automa', label: '工作流管理' },
      { to: '/tasks', label: '任务配置' },
      { to: '/task-records', label: '执行记录' },
      { to: '/clients', label: '客户端' },
    ],
  },
]

const disabledRouteSet = computed(() => new Set(runtimeConfig.value.disabled_routes || []))

onMounted(() => {
  loadRuntimeConfig()
})

async function loadRuntimeConfig() {
  try {
    runtimeConfig.value = await getRuntimeConfig()
  } catch {
    runtimeConfig.value = {
      mode: '',
      disabled_routes: [],
    }
  }
}

function isRouteDisabled(routePath) {
  return disabledRouteSet.value.has(routePath)
}
</script>

<style scoped>
.default-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  height: 100vh;
  overflow: hidden;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 36px;
  min-height: 72px;
  padding: 20px 48px;
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
}

.brand,
.nav-links {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand {
  flex-shrink: 0;
  color: #1f2d3d;
  font-size: 24px;
  font-weight: 800;
}

.brand-logo {
  width: 34px;
  height: 34px;
  object-fit: contain;
}

.nav-links {
  flex: 1;
  justify-content: flex-end;
  color: #606266;
  font-size: 16px;
  font-weight: 500;
}

.nav-links a,
.nav-disabled {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  min-height: 36px;
  padding: 7px 10px;
  color: inherit;
  border-bottom: 2px solid transparent;
  border-radius: 6px 6px 0 0;
  transition:
    color 0.2s ease,
    background-color 0.2s ease,
    border-bottom-color 0.2s ease;
}

.nav-link-text {
  display: inline-grid;
  place-items: center;
  grid-template-columns: 1fr;
  grid-template-rows: 1fr;
}

.nav-link-text::before,
.nav-link-text {
  grid-area: 1 / 1;
}

.nav-link-text::before {
  height: 0;
  overflow: hidden;
  visibility: hidden;
  content: attr(data-text);
  font-size: 17px;
  font-weight: 800;
  white-space: nowrap;
  pointer-events: none;
}

.nav-links a {
  color: inherit;
  font-size: 16px;
  font-weight: 500;
}

.nav-disabled {
  color: #c0c4cc;
  cursor: not-allowed;
  text-decoration: line-through;
  text-decoration-thickness: 2px;
}

.nav-separator {
  color: #c0c4cc;
}

.nav-links a.router-link-exact-active {
  color: #1d4ed8;
  font-size: 17px;
  font-weight: 800;
  background: #eff6ff;
  border-bottom-color: #2563eb;
}

.page-container {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  width: min(1440px, calc(100% - 64px));
  margin: 0 auto;
  padding: 32px 0;
  overflow: auto;
}

@media (max-width: 640px) {
  .app-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 16px;
    padding: 18px 24px;
  }

  .nav-links {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .page-container {
    width: calc(100% - 32px);
  }
}
</style>
