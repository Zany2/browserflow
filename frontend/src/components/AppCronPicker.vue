<template>
  <div class="app-cron-picker">
    <div class="cron-row">
      <el-input
        v-model="expressionText"
        clearable
        placeholder="可选，例如 @every 10m 或 # */10 * * * *"
        @input="handleManualInput"
      />
      <el-button :icon="Clock" @click="pickerVisible = true">选择</el-button>
    </div>

    <div v-if="summary" class="cron-summary">{{ summary }}</div>

    <AppDialog
      v-model="pickerVisible"
      title="选择定时规则"
      width="760px"
      confirm-text="应用"
      @confirm="applySelection"
      @closed="resetDraftFromValue"
    >
      <div class="cron-dialog">
        <el-tabs v-model="draft.mode" class="cron-tabs">
          <el-tab-pane label="不调度" name="none">
            <div class="cron-empty">任务仅手动执行，不自动调度。</div>
          </el-tab-pane>

          <el-tab-pane label="预设" name="preset">
            <div class="cron-grid cron-grid--preset">
              <span class="cron-label">预设规则</span>
              <el-select v-model="draft.preset">
                <el-option label="每小时" value="@hourly" />
                <el-option label="每天" value="@daily" />
                <el-option label="每周" value="@weekly" />
                <el-option label="每月" value="@monthly" />
                <el-option label="每年" value="@yearly" />
              </el-select>
            </div>
          </el-tab-pane>

          <el-tab-pane label="固定间隔" name="every">
            <div class="cron-grid">
              <span class="cron-label">每隔</span>
              <el-input-number v-model="draft.everyValue" :min="1" :max="9999" controls-position="right" />
              <el-select v-model="draft.everyUnit">
                <el-option label="秒" value="s" />
                <el-option label="分钟" value="m" />
                <el-option label="小时" value="h" />
              </el-select>
            </div>
          </el-tab-pane>

          <el-tab-pane label="每天" name="daily">
            <div class="cron-grid">
              <span class="cron-label">每天</span>
              <el-time-picker
                v-model="draft.dailyTime"
                format="HH:mm:ss"
                value-format="HH:mm:ss"
                placeholder="选择时间"
              />
              <span>执行</span>
            </div>
          </el-tab-pane>

          <el-tab-pane label="每周" name="weekly">
            <div class="cron-stack">
              <el-checkbox-group v-model="draft.weekDays" class="cron-checkboxes">
                <el-checkbox-button
                  v-for="day in weekOptions"
                  :key="day.value"
                  :label="day.value"
                  class="cron-week-button"
                >
                  {{ day.label }}
                </el-checkbox-button>
              </el-checkbox-group>
              <div class="cron-grid">
                <span class="cron-label">时间</span>
                <el-time-picker
                  v-model="draft.weeklyTime"
                  format="HH:mm:ss"
                  value-format="HH:mm:ss"
                  placeholder="选择时间"
                />
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="每月" name="monthly">
            <div class="cron-grid cron-grid--monthly">
              <span class="cron-label">每月</span>
              <el-input-number v-model="draft.monthDay" :min="1" :max="31" controls-position="right" />
              <span>日</span>
              <el-time-picker
                v-model="draft.monthlyTime"
                format="HH:mm:ss"
                value-format="HH:mm:ss"
                placeholder="选择时间"
              />
            </div>
          </el-tab-pane>

          <el-tab-pane label="每年" name="yearly">
            <div class="cron-grid cron-grid--yearly">
              <span class="cron-label">每年</span>
              <el-select v-model="draft.yearMonth">
                <el-option v-for="month in monthOptions" :key="month.value" :label="month.label" :value="month.value" />
              </el-select>
              <el-input-number v-model="draft.yearDay" :min="1" :max="31" controls-position="right" />
              <span>日</span>
              <el-time-picker
                v-model="draft.yearlyTime"
                format="HH:mm:ss"
                value-format="HH:mm:ss"
                placeholder="选择时间"
              />
            </div>
          </el-tab-pane>

          <el-tab-pane label="时段间隔" name="range">
            <div class="cron-stack">
              <div class="cron-field">
                <span class="cron-label">每隔</span>
                <div class="cron-control-row cron-control-row--compact">
                  <el-input-number v-model="draft.rangeStep" :min="1" :max="59" controls-position="right" />
                  <span class="cron-unit">分钟</span>
                </div>
              </div>
              <div class="cron-field">
                <span class="cron-label">时段</span>
                <div class="cron-control-row">
                  <el-time-picker
                    v-model="draft.rangeStart"
                    format="HH:mm"
                    value-format="HH:mm"
                    placeholder="开始"
                  />
                  <span class="cron-unit">至</span>
                  <el-time-picker
                    v-model="draft.rangeEnd"
                    format="HH:mm"
                    value-format="HH:mm"
                    placeholder="结束"
                  />
                </div>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="高级" name="advanced">
            <div class="cron-stack">
              <el-input v-model="draft.advancedExpression" placeholder="5 段、6 段、预设或 @every 表达式" />
              <span class="cron-help">
                支持 GoFrame gcron：6 段 Cron、5 段 Cron、秒字段 #、@every 10m、@hourly、@daily、@weekly、@monthly、@yearly。
              </span>
            </div>
          </el-tab-pane>
        </el-tabs>

        <div class="cron-preview">
          <span>表达式</span>
          <code>{{ previewExpression || '空' }}</code>
        </div>
      </div>
    </AppDialog>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { Clock } from '@element-plus/icons-vue'
import AppDialog from '@/components/AppDialog.vue'

const value = defineModel({
  type: String,
  default: '',
})

const pickerVisible = ref(false)
const expressionText = ref(value.value || '')

const weekOptions = [
  { label: '周日', value: 0 },
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
]

const monthOptions = [
  { label: '1 月', value: 1 },
  { label: '2 月', value: 2 },
  { label: '3 月', value: 3 },
  { label: '4 月', value: 4 },
  { label: '5 月', value: 5 },
  { label: '6 月', value: 6 },
  { label: '7 月', value: 7 },
  { label: '8 月', value: 8 },
  { label: '9 月', value: 9 },
  { label: '10 月', value: 10 },
  { label: '11 月', value: 11 },
  { label: '12 月', value: 12 },
]

const presetLabels = {
  '@hourly': '每小时',
  '@daily': '每天',
  '@weekly': '每周',
  '@monthly': '每月',
  '@yearly': '每年',
  '@annually': '每年',
  '@midnight': '每天午夜',
}

const draft = reactive({
  mode: 'none',
  preset: '@daily',
  everyValue: 10,
  everyUnit: 'm',
  dailyTime: '09:00:00',
  weekDays: [1],
  weeklyTime: '09:00:00',
  monthDay: 1,
  monthlyTime: '09:00:00',
  yearMonth: 1,
  yearDay: 1,
  yearlyTime: '09:00:00',
  rangeStep: 30,
  rangeStart: '09:00',
  rangeEnd: '18:00',
  advancedExpression: '',
})

const previewExpression = computed(() => buildExpression(draft))
const summary = computed(() => describeExpression(expressionText.value))

watch(value, (nextValue) => {
  expressionText.value = nextValue || ''
})

watch(pickerVisible, (visible) => {
  if (visible) resetDraftFromValue()
})

function handleManualInput(text) {
  value.value = String(text || '').trim()
}

function applySelection() {
  const expression = previewExpression.value
  expressionText.value = expression
  value.value = expression
  pickerVisible.value = false
}

function resetDraftFromValue() {
  const expression = String(value.value || '').trim()
  expressionText.value = expression
  draft.advancedExpression = expression
  Object.assign(draft, parseExpression(expression))
}

function buildExpression(state) {
  switch (state.mode) {
    case 'none':
      return ''
    case 'preset':
      return state.preset
    case 'every':
      return `@every ${Number(state.everyValue || 1)}${state.everyUnit}`
    case 'daily':
      return timeToCron(state.dailyTime, '*', '*', '*')
    case 'weekly':
      return timeToCron(state.weeklyTime, '*', '*', normalizeWeekDays(state.weekDays).join(','))
    case 'monthly':
      return timeToCron(state.monthlyTime, Number(state.monthDay || 1), '*', '*')
    case 'yearly':
      return timeToCron(state.yearlyTime, Number(state.yearDay || 1), Number(state.yearMonth || 1), '*')
    case 'range':
      return buildRangeExpression(state)
    case 'advanced':
      return String(state.advancedExpression || '').trim()
    default:
      return ''
  }
}

function timeToCron(timeText, dayOfMonth, month, dayOfWeek) {
  const [hour, minute, second] = parseTime(timeText)
  return `${second} ${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`
}

function buildRangeExpression(state) {
  const step = Math.min(59, Math.max(1, Number(state.rangeStep || 1)))
  const startHour = parseHour(state.rangeStart, 9)
  const endHour = parseHour(state.rangeEnd, 18)
  const [minHour, maxHour] = startHour <= endHour ? [startHour, endHour] : [endHour, startHour]
  return `# */${step} ${minHour}-${maxHour} * * *`
}

function parseExpression(expression) {
  const text = String(expression || '').trim()
  if (!text) return { mode: 'none', advancedExpression: '' }

  const lowerText = text.toLowerCase()
  if (presetLabels[lowerText]) {
    return { mode: 'preset', preset: lowerText, advancedExpression: text }
  }

  const everyMatch = text.match(/^@every\s+(\d+)(ns|us|µs|ms|s|m|h)$/i)
  if (everyMatch) {
    return {
      mode: 'every',
      everyValue: Number(everyMatch[1]),
      everyUnit: normalizeEveryUnit(everyMatch[2]),
      advancedExpression: text,
    }
  }

  const parts = text.split(/\s+/)
  const cronParts = parts.length === 5 ? ['0', ...parts] : parts
  if (cronParts.length !== 6) return { mode: 'advanced', advancedExpression: text }

  const [second, minute, hour, dayOfMonth, month, dayOfWeek] = cronParts
  if (second === '#' && isStepMinute(minute) && isHourRange(hour) && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
    const [startHour, endHour] = hour.split('-')
    return {
      mode: 'range',
      rangeStep: Number(minute.slice(2)),
      rangeStart: `${padTime(startHour)}:00`,
      rangeEnd: `${padTime(endHour)}:00`,
      advancedExpression: text,
    }
  }
  if (!isSimpleTimeParts(hour, minute, second)) return { mode: 'advanced', advancedExpression: text }
  if (month === '*' && dayOfMonth === '*' && dayOfWeek === '*') {
    return { mode: 'daily', dailyTime: `${padTime(hour)}:${padTime(minute)}:${padTime(second)}`, advancedExpression: text }
  }
  if (month === '*' && dayOfMonth === '*' && isSimpleWeekDays(dayOfWeek)) {
    return {
      mode: 'weekly',
      weeklyTime: `${padTime(hour)}:${padTime(minute)}:${padTime(second)}`,
      weekDays: dayOfWeek.split(',').map((item) => Number(item)).filter((item) => item >= 0 && item <= 6),
      advancedExpression: text,
    }
  }
  if (month === '*' && dayOfWeek === '*' && /^\d+$/.test(dayOfMonth)) {
    return {
      mode: 'monthly',
      monthDay: Number(dayOfMonth),
      monthlyTime: `${padTime(hour)}:${padTime(minute)}:${padTime(second)}`,
      advancedExpression: text,
    }
  }
  if (dayOfWeek === '*' && /^\d+$/.test(dayOfMonth) && /^\d+$/.test(month)) {
    return {
      mode: 'yearly',
      yearMonth: Number(month),
      yearDay: Number(dayOfMonth),
      yearlyTime: `${padTime(hour)}:${padTime(minute)}:${padTime(second)}`,
      advancedExpression: text,
    }
  }
  return { mode: 'advanced', advancedExpression: text }
}

function describeExpression(expression) {
  const text = String(expression || '').trim()
  if (!text) return '不自动调度，仅手动执行。'
  const parsed = parseExpression(text)
  switch (parsed.mode) {
    case 'preset':
      return `${presetLabels[parsed.preset] || '预设规则'}执行。`
    case 'every':
      return `每隔 ${parsed.everyValue} ${formatEveryUnit(parsed.everyUnit)}执行一次。`
    case 'daily':
      return `每天 ${parsed.dailyTime} 执行。`
    case 'weekly':
      return `每周 ${formatWeekDays(parsed.weekDays)} ${parsed.weeklyTime} 执行。`
    case 'monthly':
      return `每月 ${parsed.monthDay} 日 ${parsed.monthlyTime} 执行。`
    case 'yearly':
      return `每年 ${parsed.yearMonth} 月 ${parsed.yearDay} 日 ${parsed.yearlyTime} 执行。`
    case 'range':
      return `每天 ${parsed.rangeStart} 至 ${parsed.rangeEnd} 每隔 ${parsed.rangeStep} 分钟执行。`
    default:
      return '高级 GoFrame gcron 表达式，将由前后端校验后提交。'
  }
}

function parseTime(timeText) {
  const parts = String(timeText || '09:00:00').split(':')
  return [
    clampTime(parts[0], 0, 23),
    clampTime(parts[1], 0, 59),
    clampTime(parts[2], 0, 59),
  ]
}

function parseHour(timeText, fallback) {
  const hour = Number(String(timeText || '').split(':')[0])
  if (!Number.isFinite(hour)) return fallback
  return Math.min(23, Math.max(0, hour))
}

function normalizeWeekDays(days) {
  const result = Array.from(new Set((days || []).map((item) => Number(item)).filter((item) => item >= 0 && item <= 6)))
  return result.length > 0 ? result.sort((a, b) => a - b) : [1]
}

function normalizeEveryUnit(unit) {
  return ['s', 'm', 'h'].includes(unit) ? unit : 'm'
}

function isSimpleTimeParts(hour, minute, second) {
  return [hour, minute, second].every((item) => /^\d+$/.test(String(item || '')))
}

function isSimpleWeekDays(value) {
  return String(value || '')
    .split(',')
    .every((item) => /^\d+$/.test(item) && Number(item) >= 0 && Number(item) <= 6)
}

function isStepMinute(value) {
  return /^\*\/\d+$/.test(value)
}

function isHourRange(value) {
  const parts = String(value || '').split('-')
  if (parts.length !== 2) return false
  const start = Number(parts[0])
  const end = Number(parts[1])
  return Number.isInteger(start) && Number.isInteger(end) && start >= 0 && end <= 23 && start <= end
}

function clampTime(value, min, max) {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue)) return '00'
  return padTime(Math.min(max, Math.max(min, numberValue)))
}

function padTime(value) {
  return String(Number(value) || 0).padStart(2, '0')
}

function formatEveryUnit(unit) {
  if (unit === 's') return '秒'
  if (unit === 'h') return '小时'
  return '分钟'
}

function formatWeekDays(days) {
  const labels = normalizeWeekDays(days).map((day) => weekOptions.find((item) => item.value === day)?.label || day)
  return labels.join('、')
}

resetDraftFromValue()
</script>

<style scoped lang="scss">
.app-cron-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.cron-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
}

.cron-summary,
.cron-help,
.cron-empty {
  color: #909399;
  font-size: 13px;
}

.cron-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cron-tabs {
  min-height: 230px;
}

.cron-tabs :deep(.el-tabs__header) {
  margin-bottom: 18px;
}

.cron-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: #e4e7ed;
}

.cron-tabs :deep(.el-tabs__item) {
  height: 36px;
  padding: 0 18px;
  font-weight: 500;
}

.cron-grid {
  display: grid;
  grid-template-columns: max-content 160px 160px max-content;
  gap: 12px;
  align-items: center;
  justify-content: start;
  max-width: 620px;
}

.cron-grid--preset {
  grid-template-columns: max-content 220px;
}

.cron-grid--monthly {
  grid-template-columns: max-content 120px max-content 160px;
  gap: 10px;
}

.cron-grid--yearly {
  grid-template-columns: max-content 120px 120px max-content 160px;
}

.cron-grid--range {
  grid-template-columns: max-content 160px max-content 160px;
}

.cron-grid :deep(.el-input),
.cron-grid :deep(.el-input-number),
.cron-grid :deep(.el-select),
.cron-grid :deep(.el-date-editor) {
  width: 100%;
}

.cron-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
  align-items: flex-start;
  max-width: 620px;
}

.cron-label {
  color: #606266;
  text-align: left;
  white-space: nowrap;
}

.cron-field {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  width: 100%;
}

.cron-control-row {
  display: inline-grid;
  grid-template-columns: minmax(160px, 220px) auto minmax(160px, 220px);
  gap: 10px;
  align-items: center;
  justify-content: start;
  min-width: 0;
}

.cron-control-row--compact {
  grid-template-columns: minmax(120px, 160px) auto;
}

.cron-control-row :deep(.el-input),
.cron-control-row :deep(.el-input-number),
.cron-control-row :deep(.el-date-editor) {
  width: 100%;
}

.cron-unit {
  color: #606266;
  white-space: nowrap;
}

.cron-checkboxes {
  display: grid;
  grid-template-columns: repeat(7, minmax(60px, 1fr));
  gap: 8px;
  max-width: 560px;
}

.cron-checkboxes :deep(.el-checkbox-button) {
  margin: 0;
}

.cron-checkboxes :deep(.el-checkbox-button__inner) {
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  height: 34px;
  padding: 0 10px;
  color: #606266;
  font-size: 13px;
  line-height: 32px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  box-shadow: none;
}

.cron-checkboxes :deep(.el-checkbox-button:first-child .el-checkbox-button__inner),
.cron-checkboxes :deep(.el-checkbox-button:last-child .el-checkbox-button__inner) {
  border-radius: 4px;
}

.cron-checkboxes :deep(.el-checkbox-button.is-checked .el-checkbox-button__inner) {
  color: #ffffff;
  background: #409eff;
  border-color: #409eff;
  box-shadow: none;
}

.cron-checkboxes :deep(.el-checkbox-button:not(.is-checked) .el-checkbox-button__inner:hover) {
  color: #409eff;
  border-color: #a0cfff;
  background: #ecf5ff;
}

.cron-preview {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 36px;
  padding: 8px 10px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
}

.cron-preview span {
  color: #606266;
}

.cron-preview code {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #303133;
}

@media (max-width: 640px) {
  .cron-row,
  .cron-grid,
  .cron-grid--preset,
  .cron-grid--monthly,
  .cron-grid--yearly,
  .cron-grid--range {
    grid-template-columns: 1fr;
  }

  .cron-field,
  .cron-control-row,
  .cron-control-row--compact {
    grid-template-columns: 1fr;
  }

  .cron-label {
    text-align: left;
  }

  .cron-checkboxes {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
