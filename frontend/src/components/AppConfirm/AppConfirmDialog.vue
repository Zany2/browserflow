<template>
  <AppDialog
    v-model="visible"
    :title="title"
    :width="width"
    :top="top"
    :close-on-click-modal="closeOnClickModal"
    class="app-confirm-dialog"
    @closed="emit('closed')"
  >
    <div class="app-confirm" :class="`app-confirm--${safeType}`">
      <el-icon class="app-confirm__icon">
        <component :is="iconComponent" />
      </el-icon>
      <div class="app-confirm__content">
        <p class="app-confirm__message">{{ message }}</p>
        <p v-if="description" class="app-confirm__description">{{ description }}</p>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleCancel">{{ cancelText }}</el-button>
      <el-button :type="confirmButtonType" @click="handleConfirm">{{ confirmText }}</el-button>
    </template>
  </AppDialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { CircleCloseFilled, InfoFilled, QuestionFilled, WarningFilled } from '@element-plus/icons-vue'
import AppDialog from '@/components/AppDialog.vue'

const CONFIRM_ICON_MAP = {
  danger: CircleCloseFilled,
  warning: WarningFilled,
  info: InfoFilled,
  question: QuestionFilled,
}

const CONFIRM_TYPE_LIST = Object.keys(CONFIRM_ICON_MAP)

const visible = defineModel({
  type: Boolean,
  default: false,
})

const props = defineProps({
  title: {
    type: String,
    default: '二次确认',
  },
  message: {
    type: String,
    default: '确认继续操作吗？',
  },
  description: {
    type: String,
    default: '',
  },
  type: {
    type: String,
    default: 'warning',
  },
  confirmText: {
    type: String,
    default: '确认',
  },
  cancelText: {
    type: String,
    default: '取消',
  },
  width: {
    type: String,
    default: '420px',
  },
  top: {
    type: String,
    default: '18vh',
  },
  closeOnClickModal: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['cancel', 'closed', 'confirm'])

const actionHandled = ref(false)

const safeType = computed(() => {
  return CONFIRM_TYPE_LIST.includes(props.type) ? props.type : 'warning'
})

const iconComponent = computed(() => {
  return CONFIRM_ICON_MAP[safeType.value]
})

const confirmButtonType = computed(() => {
  return safeType.value === 'danger' ? 'danger' : 'primary'
})

watch(visible, (nextVisible, previousVisible) => {
  // Dialog close 右上角或遮罩关闭时统一按取消处理
  if (!nextVisible && previousVisible && !actionHandled.value) {
    actionHandled.value = true
    emit('cancel')
  }

  // Dialog open 每次重新打开时重置操作状态
  if (nextVisible) {
    actionHandled.value = false
  }
})

function handleConfirm() {
  // Confirm close 确认后关闭弹窗并抛出确认事件
  actionHandled.value = true
  visible.value = false
  emit('confirm')
}

function handleCancel() {
  // Cancel close 取消后关闭弹窗并抛出取消事件
  actionHandled.value = true
  visible.value = false
  emit('cancel')
}
</script>

<style scoped lang="scss">
.app-confirm {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  min-height: 48px;
  padding: 4px 0;
}

.app-confirm__icon {
  flex: 0 0 auto;
  margin-top: 2px;
  font-size: 22px;
}

.app-confirm__content {
  min-width: 0;
}

.app-confirm__message {
  margin: 0;
  color: #1f2937;
  font-size: 15px;
  line-height: 24px;
}

.app-confirm__description {
  margin: 6px 0 0;
  color: #6b7280;
  font-size: 13px;
  line-height: 20px;
}

.app-confirm--danger .app-confirm__icon {
  color: #f56c6c;
}

.app-confirm--warning .app-confirm__icon {
  color: #e6a23c;
}

.app-confirm--info .app-confirm__icon {
  color: #409eff;
}

.app-confirm--question .app-confirm__icon {
  color: #909399;
}
</style>
