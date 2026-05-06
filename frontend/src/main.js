import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from '@/router'

import 'element-plus/dist/index.css'
import '@/styles/reset.scss'
import '@/styles/common.scss'

const app = createApp(App)

// Element icons 注册 Element Plus 全量图标，便于页面直接使用
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// Bootstrap 应用入口，使用中文语言包注册 Element Plus
app.use(router).use(ElementPlus, { locale: zhCn }).mount('#app')
