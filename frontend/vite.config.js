import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// Vite config Vite 配置，集中管理插件和路径别名
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      // Alias 别名，统一使用 @ 指向 src
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    cors: true,
    headers: {
      // CORS headers 开发服务也完全放开，方便执行浏览器 Agent 访问前端资源
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': '*',
      'Access-Control-Allow-Methods': '*',
      'Access-Control-Allow-Private-Network': 'true',
    },
    proxy: {
      // API proxy 后端接口代理，开发环境转发到 Go 服务
      '/api': {
        target: 'http://localhost:8001',
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
