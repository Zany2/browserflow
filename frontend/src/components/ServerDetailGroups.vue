<template>
  <div class="server-detail-groups">
    <section v-for="group in groups" :key="group.key" class="server-detail-group">
      <div class="server-detail-group-title">{{ group.title }}</div>
      <div class="server-detail-form" :style="detailFormStyle">
        <div
          v-for="field in group.fields"
          :key="field.key"
          class="server-detail-field"
          :class="{ 'server-detail-field--wide': field.type === 'textarea' }"
        >
          <span class="server-detail-label">
            <span>{{ field.label }}</span>
            <el-tooltip v-if="field.help" :content="field.help" placement="top">
              <el-icon class="server-detail-help-icon">
                <InfoFilled />
              </el-icon>
            </el-tooltip>
          </span>
          <div class="server-detail-control" :class="{ 'server-detail-control--switch': isSwitchField(field) }">
            <el-input
              v-if="field.type === 'textarea' && field.editable"
              v-model="model[field.key]"
              clearable
              type="textarea"
              :rows="field.rows || 3"
              :maxlength="field.maxlength"
            />
            <el-switch
              v-else-if="field.type === 'syncable-switch'"
              :model-value="!model.is_protected"
              :before-change="() => onToggleSyncable(model)"
              @click.stop
            />
            <el-switch v-else-if="field.type === 'switch'" v-model="model[field.key]" />
            <el-input
              v-else-if="field.editable"
              v-model="model[field.key]"
              clearable
              :maxlength="field.maxlength"
            />
            <div v-else class="server-detail-value" :class="{ 'server-detail-value--textarea': field.type === 'textarea' }">
              <span class="server-detail-text" :class="{ 'server-detail-text--empty': !field.value }">
                {{ field.value || '' }}
              </span>
              <el-tooltip v-if="field.copyable && field.value" content="复制" placement="top">
                <el-button
                  class="server-detail-copy"
                  text
                  circle
                  :icon="CopyDocument"
                  :aria-label="`复制${field.label}`"
                  @click="emit('copy', field.value)"
                />
              </el-tooltip>
              <el-tooltip v-if="field.expandable && field.value" content="查看完整内容" placement="top">
                <el-button
                  class="server-detail-copy"
                  text
                  circle
                  :icon="View"
                  :aria-label="`查看${field.label}`"
                  @click="emit('expand', field)"
                />
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { CopyDocument, InfoFilled, View } from '@element-plus/icons-vue'

const props = defineProps({
  groups: {
    type: Array,
    default: () => [],
  },
  model: {
    type: Object,
    required: true,
  },
  labelWidth: {
    type: String,
    default: '132px',
  },
  toggleSyncable: {
    type: Function,
    default: null,
  },
})

const emit = defineEmits(['copy', 'expand'])

const detailFormStyle = computed(() => ({
  '--server-detail-label-width': props.labelWidth,
}))

function isSwitchField(field) {
  return field.type === 'switch' || field.type === 'syncable-switch'
}

function onToggleSyncable(row) {
  if (!props.toggleSyncable) return false
  return props.toggleSyncable(row)
}
</script>

<style scoped lang="scss">
.server-detail-groups {
  display: grid;
  gap: 16px;
}

.server-detail-group {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.server-detail-group-title {
  display: flex;
  align-items: center;
  min-height: 24px;
  color: #303133;
  font-size: 14px;
  font-weight: 600;
}

.server-detail-group + .server-detail-group {
  padding-top: 2px;
  border-top: 1px solid #ebeef5;
}

.server-detail-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 20px;
  row-gap: 12px;
  padding-right: 4px;
}

.server-detail-field {
  display: grid;
  grid-template-columns: var(--server-detail-label-width) minmax(0, 1fr);
  align-items: start;
  gap: 6px;
  min-width: 0;
}

.server-detail-field--wide {
  grid-column: 1 / -1;
}

.server-detail-label {
  display: inline-flex;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 4px;
  padding-top: 7px;
  color: #606266;
  font-size: 13px;
  text-align: right;
}

.server-detail-help-icon {
  flex: 0 0 auto;
  margin-top: 1px;
  color: #909399;
  font-size: 14px;
  cursor: help;
}

.server-detail-control {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 6px;
  min-width: 0;
}

.server-detail-control--switch {
  display: inline-flex;
  width: max-content;
  min-width: 0;
}

.server-detail-control :deep(.el-input__wrapper) {
  min-height: 32px;
}

.server-detail-value {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  height: 32px;
  min-width: 0;
  padding: 4px 8px 4px 10px;
  color: #303133;
  line-height: 20px;
  background: #f8fafc;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

.server-detail-text {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-detail-text--empty {
  color: #a8abb2;
}

.server-detail-value--textarea {
  align-items: flex-start;
  height: auto;
  min-height: 78px;
}

.server-detail-value--textarea .server-detail-text {
  display: -webkit-box;
  overflow: hidden;
  white-space: normal;
  word-break: break-all;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.server-detail-copy {
  flex: 0 0 auto;
  width: 24px;
  height: 24px;
  margin-left: 6px;
  padding: 0;
  color: #909399;
}

.server-detail-copy :deep(.el-icon) {
  font-size: 14px;
}

.server-detail-copy:hover {
  color: #409eff;
}

@media (max-width: 900px) {
  .server-detail-form {
    grid-template-columns: 1fr;
  }

  .server-detail-field--wide {
    grid-column: auto;
  }
}

@media (max-width: 640px) {
  .server-detail-field {
    grid-template-columns: 1fr;
  }

  .server-detail-label {
    justify-content: flex-start;
    text-align: left;
  }
}
</style>
