import { nextTick, ref } from 'vue'

// usePagedTableSelection keeps table selections across paged data 保留分页表格的跨页选择
export function usePagedTableSelection({ rows, getRowKey, isSelectable = () => true } = {}) {
  const selectedRowMap = ref(new Map())
  const selectedRows = ref([])
  const selectedKeys = ref([])
  const restoringSelection = ref(false)

  function handleSelectionChange(selection = []) {
    if (restoringSelection.value) return

    // Replace only current page selections 仅替换当前页选择状态
    const currentPageKeys = new Set(getCurrentRows().map(resolveRowKey).filter(Boolean))
    const selectedPageKeys = new Set(selection.map(resolveRowKey).filter(Boolean))

    currentPageKeys.forEach((key) => {
      if (!selectedPageKeys.has(key)) {
        selectedRowMap.value.delete(key)
      }
    })
    selection.forEach((row) => {
      const key = resolveRowKey(row)
      if (key) {
        selectedRowMap.value.set(key, row)
      }
    })
    syncSelectionState()
  }

  async function restoreSelection(tableRef) {
    const table = getTable(tableRef)
    if (!table) return

    restoringSelection.value = true
    await nextTick()
    table.clearSelection?.()
    getCurrentRows().forEach((row) => {
      const key = resolveRowKey(row)
      if (key && selectedRowMap.value.has(key) && isSelectable(row)) {
        table.toggleRowSelection?.(row, true)
      }
    })
    syncSelectionState()
    await nextTick()
    restoringSelection.value = false
  }

  function resetSelection(tableRef) {
    selectedRowMap.value = new Map()
    syncSelectionState()
    getTable(tableRef)?.clearSelection?.()
  }

  function retainSelectionByRows(nextRows = []) {
    const nextKeys = new Set(nextRows.map(resolveRowKey).filter(Boolean))
    selectedRowMap.value.forEach((_, key) => {
      if (!nextKeys.has(key)) {
        selectedRowMap.value.delete(key)
      }
    })
    syncSelectionState()
  }

  function removeSelectionKeys(keys = []) {
    keys.forEach((key) => {
      selectedRowMap.value.delete(String(key || ''))
    })
    syncSelectionState()
  }

  function syncSelectionState() {
    selectedRows.value = Array.from(selectedRowMap.value.values())
    selectedKeys.value = Array.from(selectedRowMap.value.keys())
  }

  function getCurrentRows() {
    return Array.isArray(rows?.value) ? rows.value : []
  }

  function resolveRowKey(row) {
    return String(getRowKey?.(row) || '').trim()
  }

  function getTable(tableRef) {
    return tableRef?.value || tableRef || null
  }

  return {
    selectedRows,
    selectedKeys,
    handleSelectionChange,
    restoreSelection,
    resetSelection,
    retainSelectionByRows,
    removeSelectionKeys,
  }
}
