// Automa event names 扩展事件名，和 Automa content script 保持一致
export const AUTOMA_EVENTS = {
  bridge: '__automa-ext__',
  workflowsResponse: '__automa-ext__get-workflows',
  importWorkflowResponse: '__automa-ext__add-workflow',
  executeWorkflow: 'automa:execute-workflow',
  executeWorkflowResponse: '__browserflow_automa_workflow_result__',
}
