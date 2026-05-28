package workflowskill

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// FileName is the exported Skill filename. ??? Skill ????
const FileName = "SKILL.md"

// FilterWorkflows keeps workflows by export scope. ???????????
func FilterWorkflows(workflows []map[string]any, scope string, workflowIDs []string) []map[string]any {
	return filterAgentSkillWorkflows(workflows, scope, workflowIDs)
}

// FilterServerWorkflows keeps Server workflows by server id or Automa id.
func FilterServerWorkflows(workflows []map[string]any, scope string, workflowIDs []string) []map[string]any {
	return filterServerSkillWorkflows(workflows, scope, workflowIDs)
}

// GenerateMarkdown builds SKILL.md content. ?? SKILL.md ???
func GenerateMarkdown(workflows []map[string]any, baseURL string, browserID string) string {
	return generateAgentWorkflowSkillMD(workflows, baseURL, browserID)
}

// GenerateServerMarkdown builds Server-mode SKILL.md content.
func GenerateServerMarkdown(workflows []map[string]any, baseURL string) string {
	return generateServerWorkflowSkillMD(workflows, baseURL)
}

// ContentDisposition builds download header. ????????
func ContentDisposition(fileName string) string {
	return buildAgentSkillContentDisposition(fileName)
}

// BaseURL builds api base url. ?? API ?????
func BaseURL(host string, tls bool) string {
	return agentSkillBaseURL(host, tls)
}

// BaseURLFromServerAddress builds api base url from server.address.
func BaseURLFromServerAddress(address string, tls bool) string {
	return agentSkillBaseURLFromServerAddress(address, tls)
}

// BaseURLFromFrontendURL builds api base url from configured frontend url.
func BaseURLFromFrontendURL(frontendURL string) string {
	return agentSkillBaseURLFromFrontendURL(frontendURL)
}

// filterAgentSkillWorkflows keeps workflows by export scope 按导出范围保留工作流
func filterAgentSkillWorkflows(workflows []map[string]any, scope string, workflowIDs []string) []map[string]any {
	if strings.EqualFold(strings.TrimSpace(scope), "all") {
		return workflows
	}

	// Scoped export uses explicit ids for selected and filtered ranges 选中和筛选范围使用明确 ID
	idSet := make(map[string]bool, len(workflowIDs))
	for _, id := range workflowIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			idSet[id] = true
		}
	}
	if len(idSet) == 0 {
		return []map[string]any{}
	}

	selected := make([]map[string]any, 0, len(workflows))
	for _, workflow := range workflows {
		if idSet[getAgentWorkflowID(workflow)] {
			selected = append(selected, workflow)
		}
	}
	return selected
}

// filterServerSkillWorkflows keeps workflows by Server id or Automa id.
func filterServerSkillWorkflows(workflows []map[string]any, scope string, workflowIDs []string) []map[string]any {
	if strings.EqualFold(strings.TrimSpace(scope), "all") {
		return workflows
	}

	idSet := make(map[string]bool, len(workflowIDs))
	for _, id := range workflowIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			idSet[id] = true
		}
	}
	if len(idSet) == 0 {
		return []map[string]any{}
	}

	selected := make([]map[string]any, 0, len(workflows))
	for _, workflow := range workflows {
		for _, id := range getServerWorkflowFilterIDs(workflow) {
			if idSet[id] {
				selected = append(selected, workflow)
				break
			}
		}
	}
	return selected
}

// generateAgentWorkflowSkillMD builds SKILL.md content 构建 SKILL.md 内容
func generateAgentWorkflowSkillMD(workflows []map[string]any, baseURL string, browserID string) string {
	var sb strings.Builder
	firstWorkflowID := getAgentWorkflowID(workflows[0])
	firstVariables := buildAgentSkillVariableExample(extractAgentWorkflowParameters(workflows[0]))

	sb.WriteString("---\n")
	sb.WriteString("name: browserflow-automa-workflows\n")
	sb.WriteString("description: " + strconv.Quote(markdownLine(buildAgentSkillDescription(workflows))) + "\n")
	sb.WriteString("---\n\n")

	sb.WriteString("# BrowserFlow Automa Workflows\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString("This skill describes Automa workflows that are currently available from a BrowserFlow browser-agent. Use the BrowserFlow HTTP API to open or run these workflows in the browser instance that exported them.\n\n")
	sb.WriteString(fmt.Sprintf("**Total Workflows Available:** %d\n\n", len(workflows)))
	sb.WriteString(fmt.Sprintf("**Recommended Filename:** `%s`\n\n", FileName))
	sb.WriteString(fmt.Sprintf("**API Base URL:** `%s`\n\n", baseURL))
	if strings.TrimSpace(browserID) != "" {
		sb.WriteString(fmt.Sprintf("**Browser Instance ID:** `%s`\n\n", inlineCode(browserID)))
	}

	sb.WriteString("## Quick Start\n\n")
	sb.WriteString("Use this Skill with a compact workflow-call loop:\n\n")
	sb.WriteString("1. Choose the workflow that matches the user's request.\n")
	sb.WriteString("2. Inspect the workflow parameters and collect missing required values.\n")
	sb.WriteString("3. Decide async submit or sync result mode based on whether the user needs returned data.\n")
	sb.WriteString("4. Call the run API with the exported `browser_id` and the exact `variables` object.\n")
	sb.WriteString("5. For sync runs, read returned variables, table data, status, and errors before answering.\n\n")
	sb.WriteString("Fast default for data questions: use sync mode with `wait_result:true`, `timeout:300`, and `return_data.variables:[\"browserflow_output\"]`. Add `return_data.include_table:true` only when the user asks for table rows or tabular output.\n\n")

	sb.WriteString("## Skill Map\n\n")
	sb.WriteString("- `Lightweight Preflight Strategy`: when to verify backend and agent status.\n")
	sb.WriteString("- `Required Step-By-Step Procedure`: required order before running a workflow.\n")
	sb.WriteString("- `Fast Call Decision Table`: how to choose async, sync, variables, and table output.\n")
	sb.WriteString("- `Run Request Parameter Reference`: exact request fields and meanings.\n")
	sb.WriteString("- `Parameter Rules`: how to pass trigger parameters through `variables`.\n")
	sb.WriteString("- `Workflow Listing Response Rules`: how to explain available workflows to users.\n")
	sb.WriteString("- `Execution Mode Rules`: async submit vs sync result mode.\n")
	sb.WriteString("- `All API Parameter Reference`: complete request fields, variables, and allowed values.\n")
	sb.WriteString("- `API Endpoints`: reusable HTTP examples.\n")
	sb.WriteString("- `Available Workflows`: workflow IDs, parameters, and per-workflow examples.\n\n")

	sb.WriteString("## Lightweight Preflight Strategy\n\n")
	sb.WriteString("Use a lightweight preflight strategy instead of repeating every check before every call:\n\n")
	sb.WriteString("- Before the first workflow call in a conversation, verify `/app/runtime` and `/agents/status`.\n")
	sb.WriteString("- A valid preflight requires successful responses from both endpoints, an online agent whose `browser_id` matches this Skill, and `automa_installed:true` for that agent.\n")
	sb.WriteString("- If any preflight check fails, do not call the run or open APIs. Reply to the user with the exact blocker from the response `message`, `error`, or checked status.\n")
	sb.WriteString("- Reuse a successful preflight result for short follow-up calls in the same conversation, especially when they target the same exported `browser_id`.\n")
	sb.WriteString("- For sync result mode, long-running workflows, returned data, or table output, perform preflight unless a recent successful preflight is already available.\n")
	sb.WriteString("- For simple async submit requests, it is acceptable to call directly after a recent successful preflight; if the API fails, inspect status and report the exact reason.\n")
	sb.WriteString("- If any call returns an offline, disconnected, missing Automa, timeout, or unreachable backend error, run preflight again before retrying.\n\n")

	sb.WriteString("## Mandatory Preflight\n\n")
	sb.WriteString("For the first workflow call, or when no recent successful preflight is available, verify that the BrowserFlow backend is reachable.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/app/runtime'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("If the request fails or returns a non-successful `code`, ask the user to start the BrowserFlow backend before continuing. Do not run any workflow until this check passes.\n\n")
	sb.WriteString("Then verify that the browser instance exported with this skill is online.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/agents/status'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("Find an agent whose `browser_id` matches the Browser Instance ID in this skill. It must be online. If the status response fails, the agent is missing, or the agent is offline, stop and ask the user to start that exact browser instance and keep the browser-agent page connected.\n\n")
	sb.WriteString("After confirming the agent is online, verify that its Automa plugin status reports `automa_installed: true`. If Automa is not installed or not available, stop and ask the user to install or enable the Automa extension in that browser instance, then refresh the browser-agent page before continuing.\n\n")
	sb.WriteString("Do not replace the exported `browser_id` with the current browser unless the user explicitly confirms that the workflow exists in the new browser instance.\n")
	sb.WriteString("Do not close the browser-agent tab that exported this Skill. It is the control channel for workflow detection and execution. When closing business tabs, keep every BrowserFlow `browser-agent` tab open unless the user explicitly asks to stop that browser instance.\n\n")

	sb.WriteString("## Required Step-By-Step Procedure\n\n")
	sb.WriteString("Always work in this order. Do not run or open a workflow until the needed checks are complete.\n\n")
	sb.WriteString("1. Decide preflight depth: use a recent successful preflight for short follow-up calls; otherwise call `/app/runtime` and `/agents/status`.\n")
	sb.WriteString("2. Validate preflight responses: both calls must return successful responses. If either call fails, stop and report the readable response reason.\n")
	sb.WriteString("3. Detect client when needed: find the exported `browser_id`, confirm it is online, and confirm `automa_installed: true`.\n")
	sb.WriteString("4. Plan first: break the user request into concrete steps, choose the best matching workflow from this Skill, inspect its `Parameters`, and ask the user for any missing required values.\n")
	sb.WriteString("5. Decide execution mode: use async for action-only requests, sync for requests that need returned data or final completion.\n")
	sb.WriteString("6. Execute only after steps 1-5 pass. If any check fails, stop and report the exact reason instead of calling the run API.\n")
	sb.WriteString("7. For sync runs, inspect the execution result before answering. For async runs, return the `execution_id` and explain that status can be queried later.\n\n")

	appendAgentWorkflowFastCallDecisionTable(&sb)
	appendAgentWorkflowRunParameterReference(&sb)
	appendAgentWorkflowAllAPIParameterReference(&sb)

	sb.WriteString("## Parameter Rules\n\n")
	sb.WriteString("Before running a workflow, inspect its `Parameters` section. If a required parameter has no value, ask the user for it before calling the API. If an optional parameter has a default value, use the default unless the user provides another value. Pass parameters through the `variables` object, and keep parameter names exactly as listed in this skill. BrowserFlow treats this `variables` object as the completed parameter set and instructs Automa not to open its own parameter input page.\n\n")

	sb.WriteString("## Parameter Type Rules\n\n")
	sb.WriteString("- `string`: pass a string value.\n")
	sb.WriteString("- `number`: pass a JSON number, not a quoted string, when the user provides a numeric value.\n")
	sb.WriteString("- `json`: pass a valid JSON object or array. If the user provides plain text, ask them to confirm the JSON structure before executing.\n")
	sb.WriteString("- `checkbox` or boolean parameters: pass a boolean `true` or `false`.\n")
	sb.WriteString("- Example values like `\"\"` are placeholders. Replace required placeholders with real user-provided values before executing.\n\n")

	appendAgentWorkflowListingResponseRules(&sb)

	sb.WriteString("## Execution Mode Rules\n\n")
	sb.WriteString("Before running a workflow, decide whether the user needs the final workflow result.\n\n")
	sb.WriteString("- Use asynchronous execution when the user only asks to start, trigger, submit, open, launch, run, or execute a task and does not ask for returned data or final completion. Set `wait_result` to `false`. Return the `execution.execution_id` to the user so they can query status or results later.\n")
	sb.WriteString("- Use synchronous waiting when the user asks to get, query, search, extract, collect, return, fetch, read, wait for completion, or confirm final success/failure. Set `wait_result` to `true`, set a reasonable `timeout`, and request returned data if needed.\n")
	sb.WriteString("- For data-returning requests, request `return_data.variables: [\"browserflow_output\"]` and read `execution.result.data.variables.browserflow_output` first. If it is missing, report that the workflow completed but did not provide a BrowserFlow output variable.\n")
	sb.WriteString("- If the user asks to check a previous task, use the execution status endpoint with the saved `execution_id` instead of running the workflow again.\n")
	sb.WriteString("- If the intent is ambiguous, prefer async mode for action-only tasks and sync mode for data-returning tasks.\n\n")

	sb.WriteString("## Result Reading Rules\n\n")
	sb.WriteString("- First read the run response. It may contain `result` for the immediate browser-agent result and `execution` for BrowserFlow's execution state.\n")
	sb.WriteString("- For sync runs, check `result.data.variables.browserflow_output`, then `execution.result.data.variables.browserflow_output` before answering data-returning requests.\n")
	sb.WriteString("- For async runs, save `execution.execution_id` and call `/workflows/executions/{execution_id}` when the user asks for status or results.\n")
	sb.WriteString("- Treat `running` and `timeout` as not-final-success states. Query again later for `running`, and report the timeout reason for `timeout`.\n")
	sb.WriteString("- If the response contains an error status or readable error message, report that message instead of only saying the workflow failed.\n\n")

	sb.WriteString("## Failure Handling Rules\n\n")
	sb.WriteString("- Backend unreachable: ask the user to start BrowserFlow and do not execute.\n")
	sb.WriteString("- Preflight response is not successful: report the response `message` or `error` and do not execute.\n")
	sb.WriteString("- Exported `browser_id` is offline or missing: ask the user to open that exact browser instance and keep the browser-agent page connected.\n")
	sb.WriteString("- Automa is not installed or unavailable: ask the user to install or enable the Automa extension, then refresh the browser-agent page.\n")
	sb.WriteString("- Missing required parameters: ask for the missing values before calling the run API.\n")
	sb.WriteString("- Sync execution has no `browserflow_output`: report that the workflow completed but did not provide the expected output variable.\n")
	sb.WriteString("- API returns an execution error: report the readable error message from the response and do not claim success.\n\n")

	sb.WriteString("## API Endpoints\n\n")
	sb.WriteString("### Run Workflow Async\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/workflows/%s/run' \\\n", baseURL, firstWorkflowID))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"browser_id":  browserID,
		"variables":   firstVariables,
		"wait_result": false,
	}))))
	sb.WriteString("```\n\n")

	sb.WriteString("### Run Workflow And Wait For Result\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/workflows/%s/run' \\\n", baseURL, firstWorkflowID))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"browser_id":  browserID,
		"variables":   firstVariables,
		"wait_result": true,
		"timeout":     300,
		"return_data": map[string]any{
			"variables":       []string{"browserflow_output"},
			"include_table":   false,
			"table_limit":     20,
			"include_history": false,
		},
	}))))
	sb.WriteString("```\n\n")

	sb.WriteString("### Query Execution Status\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/workflows/executions/{execution_id}'\n", baseURL))
	sb.WriteString("```\n\n")

	sb.WriteString("### Open Workflow Editor\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/workflows/%s/open' \\\n", baseURL, firstWorkflowID))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"browser_id": browserID,
	}))))
	sb.WriteString("```\n\n")

	sb.WriteString("## Available Workflows\n\n")
	for index, workflow := range workflows {
		appendAgentWorkflowSkillSection(&sb, index+1, workflow, baseURL, browserID)
	}

	sb.WriteString("## Usage Notes\n\n")
	sb.WriteString("- Keep the target browser running and keep the browser-agent page connected before calling the API.\n")
	sb.WriteString("- Never close the browser-agent tab while using this Skill. Close only ordinary business tabs unless the user explicitly asks to stop BrowserFlow control for that browser.\n")
	sb.WriteString("- The target browser must report `automa_installed: true`; otherwise workflow list, open, and run commands may fail.\n")
	sb.WriteString("- Pass trigger parameters through the `variables` object. Parameter names must match the Automa trigger configuration.\n")
	sb.WriteString("- Do not rely on Automa's parameter tab for Skill calls; collect required values before sending the HTTP request.\n")
	sb.WriteString("- If the exported browser instance is no longer available, update `browser_id` only after the user confirms the same workflow exists in another running browser.\n")
	sb.WriteString("- Async run returns after the command is accepted. Sync run waits until Automa reports `success`, `error`, `stopped`, or BrowserFlow reports `timeout`.\n")

	return sb.String()
}

// appendAgentWorkflowSkillSection writes one workflow section 写入单个工作流说明
func appendAgentWorkflowSkillSection(sb *strings.Builder, index int, workflow map[string]any, baseURL string, browserID string) {
	workflowID := getAgentWorkflowID(workflow)
	name := firstAgentSkillString(workflow, "name", "title")
	if name == "" {
		name = workflowID
	}
	if name == "" {
		name = fmt.Sprintf("Workflow %d", index)
	}

	sb.WriteString(fmt.Sprintf("### %d. %s\n\n", index, markdownLine(name)))
	sb.WriteString(fmt.Sprintf("- ID: `%s`\n", inlineCode(workflowID)))
	sb.WriteString(fmt.Sprintf("- Description: %s\n", markdownLine(defaultText(firstAgentSkillString(workflow, "description"), "-"))))
	sb.WriteString(fmt.Sprintf("- Status: %s\n", agentWorkflowStatus(workflow)))
	sb.WriteString(fmt.Sprintf("- Nodes: %d\n", getAgentWorkflowNodeCount(workflow)))
	sb.WriteString(fmt.Sprintf("- Created: %s\n", markdownLine(defaultText(formatAgentSkillTime(firstAgentSkillValue(workflow, "createdAt", "created_at", "created")), "-"))))
	sb.WriteString(fmt.Sprintf("- Updated: %s\n\n", markdownLine(defaultText(formatAgentSkillTime(firstAgentSkillValue(workflow, "updatedAt", "updated_at", "updated")), "-"))))

	params := extractAgentWorkflowParameters(workflow)
	if len(params) == 0 {
		sb.WriteString("Parameters: none detected.\n\n")
		appendAgentWorkflowRunExample(sb, baseURL, workflowID, browserID, nil)
		return
	}

	sb.WriteString("Parameters:\n")
	appendAgentWorkflowParameterDetails(sb, params)
	sb.WriteString("\n")
	appendAgentWorkflowInvocationGuide(sb)
	appendAgentWorkflowRunExample(sb, baseURL, workflowID, browserID, buildAgentSkillVariableExample(params))
}

// appendAgentWorkflowRunExample writes a workflow-specific run example 写入工作流运行示例
func appendAgentWorkflowRunExample(sb *strings.Builder, baseURL string, workflowID string, browserID string, variables map[string]any) {
	if strings.TrimSpace(workflowID) == "" {
		return
	}
	if variables == nil {
		variables = map[string]any{}
	}
	sb.WriteString("Run example:\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/workflows/%s/run' \\\n", baseURL, workflowID))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"browser_id":  browserID,
		"variables":   variables,
		"wait_result": false,
	}))))
	sb.WriteString("```\n\n")
	sb.WriteString("For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.\n\n")
}

func appendAgentWorkflowListingResponseRules(sb *strings.Builder) {
	sb.WriteString("## Workflow Listing Response Rules\n\n")
	sb.WriteString("When the user asks what workflows are available, do not only list names and IDs. For each workflow, also explain how it can be called:\n\n")
	sb.WriteString("- Show the workflow name, ID, status, required parameters, and optional parameters.\n")
	sb.WriteString("- Explain async submit mode: set `wait_result` to `false`; the API only dispatches the workflow and returns `execution.execution_id`; it does not wait for completion and does not return workflow output.\n")
	sb.WriteString("- Explain sync result mode: set `wait_result` to `true`; set `timeout` to the maximum wait time in seconds. The exported examples use `timeout: 300`, so the default recommendation is to wait up to 300 seconds.\n")
	sb.WriteString("- Explain variable output: add `return_data.variables: [\"browserflow_output\"]` and read `execution.result.data.variables.browserflow_output` or `result.data.variables.browserflow_output`.\n")
	sb.WriteString("- Explain table output only when needed: set `return_data.include_table` to `true` and choose a `table_limit`.\n")
	sb.WriteString("- If the user wants to start a workflow without waiting, use async submit mode. If the user wants final data, search results, extracted content, success/failure, or returned variables, use sync result mode.\n")
	sb.WriteString("- Reply in the user's language, but keep API field names exactly as written.\n\n")
	sb.WriteString("Suggested wording when listing workflows:\n\n")
	sb.WriteString("```text\n")
	sb.WriteString("This workflow can be called in two ways:\n")
	sb.WriteString("- Async submit: starts the workflow with wait_result=false and returns execution_id only.\n")
	sb.WriteString("- Sync result: waits up to 300 seconds with wait_result=true. Use return_data.variables=[\"browserflow_output\"] for returned data, and include_table=true when table rows are needed.\n")
	sb.WriteString("Required parameters: ... Optional parameters: ...\n")
	sb.WriteString("```\n\n")
}

func appendAgentWorkflowFastCallDecisionTable(sb *strings.Builder) {
	sb.WriteString("## Fast Call Decision Table\n\n")
	sb.WriteString("Use this table to avoid overthinking the request mode.\n\n")
	sb.WriteString("| User intent | Request mode | Required payload choices | What to tell the user |\n")
	sb.WriteString("|---|---|---|---|\n")
	sb.WriteString("| Start/run/trigger only | Async submit | `wait_result:false` | Return `execution.execution_id`; no final output is available yet. |\n")
	sb.WriteString("| Search/get/query/extract/return data | Sync result | `wait_result:true`, `timeout:300`, `return_data.variables:[\"browserflow_output\"]` | Wait for final status and report returned output or readable error. |\n")
	sb.WriteString("| Need table rows | Sync result with table | Add `return_data.include_table:true` and `table_limit` | Read returned table data or report where the table result is stored. |\n")
	sb.WriteString("| User asks status of previous run | Status query | Call `/workflows/executions/{execution_id}` | Do not rerun the workflow unless the user asks. |\n")
	sb.WriteString("| Missing required parameter | Ask before calling | Do not send placeholder values for required fields | Ask for only the missing values. |\n\n")
}

func appendAgentWorkflowRunParameterReference(sb *strings.Builder) {
	sb.WriteString("## Run Request Parameter Reference\n\n")
	sb.WriteString("The run API is `POST /workflows/{workflow_id}/run`. Build the JSON body with these fields:\n\n")
	sb.WriteString("| Field | Type | Required | Meaning | Example |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	sb.WriteString("| `browser_id` | string | yes | Browser instance that exported this Skill. Keep the exact exported value unless the user confirms another browser owns the same workflow. | `browser_...` |\n")
	sb.WriteString("| `variables` | object | yes | Trigger parameter object. Keys must exactly match each workflow's `Parameters` section. | `{\"key_word\":\"ai智能体\"}` |\n")
	sb.WriteString("| `wait_result` | boolean | yes | `false` submits only; `true` waits for final completion or timeout. | `true` |\n")
	sb.WriteString("| `timeout` | number | no | Maximum wait seconds for sync mode. Use `300` by default unless the user asks otherwise. | `300` |\n")
	sb.WriteString("| `return_data` | object | no | Controls which workflow outputs are returned in sync mode. | `{\"variables\":[\"browserflow_output\"]}` |\n")
	sb.WriteString("| `return_data.variables` | string array | no | Automa variables to return. Use `browserflow_output` for normal data-returning workflows. | `[\"browserflow_output\"]` |\n")
	sb.WriteString("| `return_data.include_table` | boolean | no | Return table output when the user asks for table rows or tabular data. | `true` |\n")
	sb.WriteString("| `return_data.table_limit` | number | no | Maximum table rows to include directly. | `20` |\n")
	sb.WriteString("| `return_data.include_history` | boolean | no | Include execution history details. Usually false for speed. | `false` |\n\n")
	sb.WriteString("Do not invent variables. If a workflow has no parameter named `keyword`, do not send `keyword`; use the exact listed key such as `key_word`.\n\n")
}

func appendAgentWorkflowAllAPIParameterReference(sb *strings.Builder) {
	sb.WriteString("## All API Parameter Reference\n\n")
	sb.WriteString("Use this as the complete parameter reference for the Windows/browser-agent Automa workflow Skill. Keep field names exactly as shown.\n\n")
	appendWorkflowSharedParameterConcepts(sb, false)

	appendWorkflowEndpointReference(sb, "Runtime preflight", "GET `/app/runtime`", "No parameters. Confirms the backend is reachable and reports runtime mode.", nil)
	appendWorkflowEndpointReference(sb, "Agent status", "GET `/agents/status`", "No parameters. Use this to find an online agent whose `browser_id` matches this exported Skill and to confirm Automa is installed.", nil)
	appendWorkflowEndpointReference(sb, "List agent workflows", "GET `/workflows/agent/workflows`", "Lists workflows from one connected browser-agent.", []skillAPIParamDoc{
		{"browser_id", "string", "query", "no", "Browser instance ID to inspect. Use the exported Browser Instance ID when available.", "A BrowserFlow browser id such as `browser_...`. Empty means backend may use the current/default agent context.", "browser_1tgxl..."},
	})
	appendWorkflowEndpointReference(sb, "Export agent workflow Skill", "POST `/workflows/agent/export/skill`", "Exports a Windows/browser-agent workflow Skill. This is normally used by BrowserFlow UI, not by the model during workflow execution.", []skillAPIParamDoc{
		{"browser_id", "string", "body", "no", "Browser instance ID whose detected workflows should be exported.", "A BrowserFlow browser id such as `browser_...`.", "browser_1tgxl..."},
		{"scope", "string", "body", "no", "Export range. Default is `filtered`.", "`selected`: export only selected workflow IDs; `filtered`: export the current filtered list; `all`: export all detected workflows.", "filtered"},
		{"workflow_ids", "array<string>", "body", "conditional", "Workflow IDs used by `selected` and filtered exports.", "Automa workflow IDs exactly as listed in the UI/export data.", "[\"7KKfW4mVvDFBGux2kMgft\"]"},
	})
	appendWorkflowEndpointReference(sb, "Run workflow", "POST `/workflows/{workflow_id}/run`", "Runs a detected workflow through the browser-agent. Use async for submit-only tasks and sync when the user needs final data or status.", []skillAPIParamDoc{
		{"workflow_id", "string", "path", "yes", "Workflow ID from the `Available Workflows` section.", "Automa workflow ID. Keep it exact.", "7KKfW4mVvDFBGux2kMgft"},
		{"browser_id", "string", "body", "yes", "Browser instance that exported this Skill and owns the workflow.", "A BrowserFlow browser id such as `browser_...`. Do not replace it unless the user confirms another browser owns the same workflow.", "browser_1tgxl..."},
		{"variables", "object", "body", "yes", "Trigger parameter object passed to Automa.", "Keys must exactly match the workflow `Parameters` section. Values follow each parameter type: string, number, JSON object/array, or boolean.", "{\"key_word\":\"ai智能体\"}"},
		{"wait_result", "boolean", "body", "yes", "Execution mode.", "`false`: submit only and return `execution.execution_id`; `true`: wait for final status or timeout.", "true"},
		{"timeout", "number", "body", "no", "Maximum wait seconds when `wait_result` is true. Default recommendation is 300.", "Positive integer seconds.", "300"},
		{"return_data", "object", "body", "no", "Returned data configuration for sync runs.", "Use when the user needs workflow output, variables, table rows, or history.", "{\"variables\":[\"browserflow_output\"],\"include_table\":true}"},
		{"return_data.variables", "array<string>", "body", "no", "Automa variable names to return.", "`browserflow_output`: recommended standard variable for workflow output; any custom Automa variable name created by the workflow may also be listed.", "[\"browserflow_output\"]"},
		{"return_data.include_table", "boolean", "body", "no", "Whether to return Automa table data.", "`true` when the user asks for rows/table/list data stored as table output; otherwise `false` for speed.", "false"},
		{"return_data.table_limit", "number", "body", "no", "Maximum table rows to include directly.", "Positive integer row count. Use 20 for concise answers, 100 when the user requests all visible rows.", "20"},
		{"return_data.include_history", "boolean", "body", "no", "Whether to include workflow execution history details.", "`true` only for debugging/auditing; normally `false`.", "false"},
	})
	appendWorkflowEndpointReference(sb, "Query workflow execution", "GET `/workflows/executions/{execution_id}`", "Queries a previous Windows/browser-agent workflow execution.", []skillAPIParamDoc{
		{"execution_id", "string", "path", "yes", "Execution ID returned by an async or sync run response.", "Value from `execution.execution_id`.", "exec_..."},
	})
	appendWorkflowEndpointReference(sb, "Open workflow editor", "POST `/workflows/{workflow_id}/open`", "Opens the workflow editor in the browser-agent. Use only when the user wants to inspect/edit the workflow, not for normal execution.", []skillAPIParamDoc{
		{"workflow_id", "string", "path", "yes", "Workflow ID from the `Available Workflows` section.", "Automa workflow ID. Keep it exact.", "7KKfW4mVvDFBGux2kMgft"},
		{"browser_id", "string", "body", "yes", "Browser instance that should open the workflow editor.", "A BrowserFlow browser id such as `browser_...`.", "browser_1tgxl..."},
	})
}

func appendAgentWorkflowInvocationGuide(sb *strings.Builder) {
	sb.WriteString("Invocation guide:\n")
	sb.WriteString("- Async submit: use `wait_result: false`. This starts the workflow and returns `execution.execution_id`; it does not wait for completion or return workflow data.\n")
	sb.WriteString("- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.\n")
	sb.WriteString("- Variable result: add `return_data.variables: [\"browserflow_output\"]` and read `execution.result.data.variables.browserflow_output` first.\n")
	sb.WriteString("- Table result: set `return_data.include_table: true` and set `table_limit` when the user asks for table data.\n\n")
}

func appendAgentWorkflowParameterDetails(sb *strings.Builder, params []map[string]any) {
	for index, param := range params {
		name := firstAgentSkillString(param, "name", "key")
		if name == "" {
			continue
		}

		defaultValue := formatAgentSkillDefaultValue(firstAgentSkillValue(param, "defaultValue", "default", "value"))
		if defaultValue == "" {
			defaultValue = "none"
		}

		sb.WriteString(fmt.Sprintf("%d. Parameter `%s`\n", index+1, inlineCode(name)))
		sb.WriteString(fmt.Sprintf("   - Key: `%s`\n", inlineCode(name)))
		sb.WriteString(fmt.Sprintf("   - Meaning: %s\n", markdownLine(defaultText(firstAgentSkillString(param, "description", "placeholder"), "-"))))
		sb.WriteString(fmt.Sprintf("   - Type: `%s`\n", inlineCode(defaultText(firstAgentSkillString(param, "type"), "string"))))
		sb.WriteString(fmt.Sprintf("   - Required: %s\n", requiredText(isAgentSkillParamRequired(param))))
		sb.WriteString(fmt.Sprintf("   - Default: `%s`\n", inlineCode(defaultValue)))
	}
}

func appendServerWorkflowListingResponseRules(sb *strings.Builder) {
	sb.WriteString("## Workflow Listing Response Rules\n\n")
	sb.WriteString("When the user asks what workflows are available, do not only list names and IDs. For each workflow, also explain how it can be called through Server-mode tasks:\n\n")
	sb.WriteString("- Show the workflow name, workflow ID, status, dispatch target rules, required parameters, and optional parameters.\n")
	sb.WriteString("- Explain async execution: execute a task without `wait_result: true`; the API dispatches the task and returns task record information, but does not wait for final workflow output.\n")
	sb.WriteString("- Explain sync result mode: set `wait_result` to `true`; set `timeout` to the maximum wait time in seconds. The exported examples use `timeout: 300`, so the default recommendation is to wait up to 300 seconds.\n")
	sb.WriteString("- Explain variable output: add `return_data.variables: [\"browserflow_output\"]` and read `result.data.variables.browserflow_output`, `record.result.data.variables.browserflow_output`, or the task record detail.\n")
	sb.WriteString("- Explain table output only when needed: set `return_data.include_table` to `true`, choose a `table_limit`, then inspect the task record files if the table is stored as a file.\n")
	sb.WriteString("- If the user wants to start a workflow without waiting, use async execution. If the user wants final data, search results, extracted content, success/failure, returned variables, or table data, use sync result mode.\n")
	sb.WriteString("- Reply in the user's language, but keep API field names exactly as written.\n\n")
}

func appendServerWorkflowInvocationGuide(sb *strings.Builder) {
	sb.WriteString("Invocation guide:\n")
	sb.WriteString("- Async execution: execute or create a task without `wait_result: true`. This dispatches the workflow and returns task record information; it does not wait for completion or return final workflow data.\n")
	sb.WriteString("- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.\n")
	sb.WriteString("- Variable result: add `return_data.variables: [\"browserflow_output\"]` and read `result.data.variables.browserflow_output` first, then the task record detail if needed.\n")
	sb.WriteString("- Table result: set `return_data.include_table: true` and set `table_limit`; inspect task record files when the returned table is stored as a file.\n\n")
}

func appendServerWorkflowAllAPIParameterReference(sb *strings.Builder) {
	sb.WriteString("## All API Parameter Reference\n\n")
	sb.WriteString("Use this as the complete parameter reference for the Server-mode Automa workflow Skill. Keep field names exactly as shown.\n\n")
	appendWorkflowSharedParameterConcepts(sb, true)

	appendWorkflowEndpointReference(sb, "Runtime preflight", "GET `/app/runtime`", "No parameters. Confirms the backend is reachable and running in Server mode.", nil)
	appendWorkflowEndpointReference(sb, "List clients", "GET `/clients`", "Lists connected client nodes. Use it to confirm an online node owns the workflow before dispatch.", []skillAPIParamDoc{
		{"page_num", "number", "query", "no", "Page number when pagination is supported.", "Positive integer, starts from 1.", "1"},
		{"page_size", "number", "query", "no", "Page size when pagination is supported.", "Allowed by common pagination: 10, 30, or 60.", "30"},
	})
	appendWorkflowEndpointReference(sb, "Query existing tasks", "GET `/tasks`", "Find reusable tasks before creating a new one.", []skillAPIParamDoc{
		{"page_num", "number", "query", "no", "Page number.", "Positive integer, starts from 1.", "1"},
		{"page_size", "number", "query", "no", "Page size.", "`10`, `30`, or `60`.", "10"},
		{"start_time", "string", "query", "no", "Created/updated time filter start.", "Datetime string accepted by backend validation.", "2026-05-27 00:00:00"},
		{"end_time", "string", "query", "no", "Created/updated time filter end.", "Datetime string accepted by backend validation.", "2026-05-27 23:59:59"},
		{"keyword", "string", "query", "no", "Keyword search over task fields.", "Any search text.", "Skill task"},
		{"workflow_id", "string", "query", "no", "Filter by workflow ID.", "Automa workflow ID from this Skill.", "7KKfW4mVvDFBGux2kMgft"},
		{"workflow_name", "string", "query", "no", "Filter by workflow name.", "Workflow name text.", "普通检索"},
		{"client_id", "string", "query", "no", "Filter by client ID.", "Server client ID.", "client_..."},
		{"machine_id", "string", "query", "no", "Filter by machine ID.", "Client machine ID.", "machine_..."},
		{"node_id", "string", "query", "no", "Filter by node ID.", "Browser/client node ID.", "node_..."},
		{"enabled", "string", "query", "no", "Filter by enabled state.", "Common values are `true` or `false`.", "true"},
	})
	appendWorkflowEndpointReference(sb, "Create task", "POST `/tasks`", "Creates a reusable Server-mode task. Use `run_once_after_create:true` only when the user wants to create and immediately run it.", []skillAPIParamDoc{
		{"name", "string", "body", "yes", "Task name.", "Any readable name.", "Skill task for 普通检索"},
		{"description", "string", "body", "no", "Task description.", "Any readable description.", "Created by BrowserFlow exported Skill"},
		{"workflow_id", "string", "body", "yes", "Workflow ID to run.", "Automa workflow ID from this Skill.", "7KKfW4mVvDFBGux2kMgft"},
		{"workflow_name", "string", "body", "no", "Workflow display name snapshot.", "Workflow name from this Skill.", "普通检索"},
		{"client_id", "string", "body", "no", "Target client ID.", "Use only when targeting a specific client. Prefer leaving empty for automatic dispatch.", "client_..."},
		{"client_name", "string", "body", "no", "Target client name snapshot.", "Readable client name.", "Office PC"},
		{"client_ip", "string", "body", "no", "Target client IP.", "Leave empty for automatic dispatch. If set, also set `node_id` when targeting a specific node.", "192.168.1.10"},
		{"machine_id", "string", "body", "no", "Target machine ID.", "Use only when known and needed.", "machine_..."},
		{"node_id", "string", "body", "no", "Target execution node ID.", "Leave empty for automatic dispatch. If set, also set `client_ip` for specific-node dispatch.", "node_..."},
		{"target_group_id", "number", "body", "no", "Target node group ID.", "Integer group ID when group dispatch is configured.", "1"},
		{"dispatch_mode", "string", "body", "no", "Dispatch mode.", "Use existing project modes when configured; otherwise leave empty for default dispatch.", ""},
		{"queue_policy", "string", "body", "no", "Queue/busy handling policy.", "Use existing project policy values when configured; otherwise leave empty for default behavior.", ""},
		{"max_attempts", "number", "body", "no", "Maximum retry attempts.", "Positive integer. Leave empty/0 for backend default.", "1"},
		{"timeout_seconds", "number", "body", "no", "Task execution timeout seconds.", "Positive integer seconds. Leave empty/0 for backend default.", "300"},
		{"queue_wait_seconds", "number", "body", "no", "Maximum time to wait for an available node.", "Positive integer seconds.", "60"},
		{"queue_retry_interval_seconds", "number", "body", "no", "Retry interval while waiting for a node.", "Positive integer seconds.", "5"},
		{"cron_expression", "string", "body", "no", "Cron schedule. Empty means no schedule.", "Cron expression supported by the backend scheduler.", ""},
		{"run_once_after_create", "boolean", "body", "no", "Whether to immediately execute once after creating.", "`true` or `false`.", "false"},
		{"params", "object", "body", "no", "Workflow trigger parameters.", "Keys must exactly match the workflow `Parameters` section. Values follow parameter type rules.", "{\"key_word\":\"ai智能体\"}"},
		{"enabled", "boolean", "body", "no", "Whether the task is enabled.", "`true` or `false`.", "true"},
	})
	appendWorkflowEndpointReference(sb, "Execute existing task", "POST `/tasks/{task_id}/execute`", "Dispatches an existing task to a client node.", []skillAPIParamDoc{
		{"task_id", "string", "path", "yes", "Task ID to execute.", "Task id from `/tasks` or create response.", "123"},
		{"client_id", "string", "body", "no", "Override target client ID for this execution.", "Use only when selecting a specific client.", "client_..."},
		{"client_ip", "string", "body", "no", "Override target client IP for this execution.", "Leave empty for automatic dispatch. If set with specific node targeting, also set `node_id`.", "192.168.1.10"},
		{"machine_id", "string", "body", "no", "Override target machine ID.", "Use only when known and needed.", "machine_..."},
		{"node_id", "string", "body", "no", "Override target node ID.", "Leave empty for automatic dispatch. If set, also set `client_ip` for a specific node.", "node_..."},
		{"trigger_type", "string", "body", "no", "Execution trigger label.", "Use `skill` for executions started from this Skill. Other common labels may include manual/schedule/system depending on backend usage.", "skill"},
		{"params", "object", "body", "no", "Execution parameter override.", "Keys must exactly match the workflow `Parameters` section. Overrides task saved params for this run.", "{\"key_word\":\"ai智能体\"}"},
		{"wait_result", "boolean", "body", "no", "Execution mode.", "`false`: dispatch only; `true`: wait for final status or timeout.", "true"},
		{"timeout", "number", "body", "no", "Maximum wait seconds when `wait_result` is true. Default recommendation is 300.", "Positive integer seconds.", "300"},
		{"return_data", "object", "body", "no", "Returned data configuration for sync runs.", "Use when the user needs workflow output, variables, table rows, or history.", "{\"variables\":[\"browserflow_output\"],\"include_table\":true}"},
		{"return_data.variables", "array<string>", "body", "no", "Automa variable names to return.", "`browserflow_output`: recommended standard variable for workflow output; any custom Automa variable name created by the workflow may also be listed.", "[\"browserflow_output\"]"},
		{"return_data.include_table", "boolean", "body", "no", "Whether to return Automa table data.", "`true` when the user asks for rows/table/list data stored as table output; otherwise `false` for speed.", "true"},
		{"return_data.table_limit", "number", "body", "no", "Maximum table rows to include directly.", "Positive integer row count.", "100"},
		{"return_data.include_history", "boolean", "body", "no", "Whether to include workflow execution history details.", "`true` only for debugging/auditing; normally `false`.", "false"},
	})
	appendWorkflowEndpointReference(sb, "Query task records", "GET `/task-records`", "Lists task execution records. Use after async execution or when the user asks for historical results.", []skillAPIParamDoc{
		{"page_num", "number", "query", "no", "Page number.", "Positive integer, starts from 1.", "1"},
		{"page_size", "number", "query", "no", "Page size.", "`10`, `30`, or `60`.", "10"},
		{"start_time", "string", "query", "no", "Execution record time filter start.", "Datetime string accepted by backend validation.", "2026-05-27 00:00:00"},
		{"end_time", "string", "query", "no", "Execution record time filter end.", "Datetime string accepted by backend validation.", "2026-05-27 23:59:59"},
		{"task_id", "string", "query", "no", "Filter by task ID.", "Task ID.", "123"},
		{"task_name", "string", "query", "no", "Filter by task name.", "Task name text.", "Skill task"},
		{"workflow_id", "string", "query", "no", "Filter by workflow ID.", "Automa workflow ID from this Skill.", "7KKfW4mVvDFBGux2kMgft"},
		{"workflow_name", "string", "query", "no", "Filter by workflow name.", "Workflow name text.", "普通检索"},
		{"client_id", "string", "query", "no", "Filter by client ID.", "Server client ID.", "client_..."},
		{"client_ip", "string", "query", "no", "Filter by one client IP.", "Client IP.", "192.168.1.10"},
		{"client_ips", "string", "query", "no", "Filter by multiple client IPs.", "Comma-separated client IPs.", "192.168.1.10,192.168.1.11"},
		{"machine_id", "string", "query", "no", "Filter by machine ID.", "Machine ID.", "machine_..."},
		{"node_id", "string", "query", "no", "Filter by one node ID.", "Node ID.", "node_..."},
		{"node_ids", "string", "query", "no", "Filter by multiple node IDs.", "Comma-separated node IDs.", "node_a,node_b"},
		{"status", "string", "query", "no", "Filter by execution status.", "`queued`, `running`, `success`, `error`, `stopped`, `timeout`.", "success"},
		{"keyword", "string", "query", "no", "Keyword search over record fields.", "Any search text.", "ai智能体"},
	})
	appendWorkflowEndpointReference(sb, "Query task record detail", "GET `/task-records/{record_id}`", "Reads one execution record and its result files.", []skillAPIParamDoc{
		{"record_id", "string", "path", "yes", "Task record ID.", "Record id from execute response or `/task-records` list.", "456"},
	})
	appendWorkflowEndpointReference(sb, "Download task record file", "GET `/task-records/files/{file_id}/download`", "Downloads one result file, commonly table output stored as a file.", []skillAPIParamDoc{
		{"file_id", "string", "path", "yes", "Task record file ID.", "File id from task record detail `files` array.", "789"},
	})
}

// generateServerWorkflowSkillMD builds Server-mode SKILL.md content.
func generateServerWorkflowSkillMD(workflows []map[string]any, baseURL string) string {
	var sb strings.Builder
	firstWorkflowID := getServerWorkflowExecutionID(workflows[0])
	firstVariables := buildAgentSkillVariableExample(extractAgentWorkflowParameters(workflows[0]))

	sb.WriteString("---\n")
	sb.WriteString("name: browserflow-server-automa-workflows\n")
	sb.WriteString("description: " + strconv.Quote(markdownLine(buildServerSkillDescription(workflows))) + "\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# BrowserFlow Server Automa Workflows\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString("This skill describes Automa workflows stored in BrowserFlow Server mode. Use the BrowserFlow task APIs to create or run server-side tasks. The server dispatches each task to an online Windows client node that owns the target workflow. A client node is identified by `client_ip` plus `node_id`.\n\n")
	sb.WriteString(fmt.Sprintf("**Total Workflows Available:** %d\n\n", len(workflows)))
	sb.WriteString(fmt.Sprintf("**Recommended Filename:** `%s`\n\n", FileName))
	sb.WriteString(fmt.Sprintf("**API Base URL:** `%s`\n\n", baseURL))
	sb.WriteString("## Mandatory Preflight\n\n")
	sb.WriteString("Before running any workflow, first verify that the BrowserFlow backend is reachable and running in Server mode.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/app/runtime'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("If the request fails, returns a non-successful `code`, or the mode is not `server`, ask the user to start BrowserFlow in Server mode before continuing. Do not create or execute any task until this check passes.\n\n")
	sb.WriteString("Then verify that at least one client node is online and has the target workflow. If a task does not specify a client node, BrowserFlow scans online nodes that own the workflow and chooses an unlocked node.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/clients'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("If the clients response fails, returns a non-successful `code`, has no online nodes, or cannot prove that an online node owns the target workflow, stop and tell the user the exact blocker before creating or executing a task.\n\n")
	sb.WriteString("## Required Step-By-Step Procedure\n\n")
	sb.WriteString("Always work in this order. Do not create or execute a task until the checks are complete.\n\n")
	sb.WriteString("1. Detect runtime: call `/app/runtime` and confirm BrowserFlow is reachable, the response is successful, and the runtime is `server` mode.\n")
	sb.WriteString("2. Detect client nodes: call `/clients`, confirm the response is successful, confirm there is at least one online node, and confirm the target workflow can be dispatched to an online node.\n")
	sb.WriteString("3. Detect workflow and parameters: choose the matching workflow from this Skill, inspect its `Parameters`, and ask the user for any missing required values.\n")
	sb.WriteString("4. Find reusable tasks: query `/tasks?workflow_id={workflow_id}&page_num=1&page_size=10`. Reuse an enabled task when it matches the workflow and parameters, unless the user asks to create a new task.\n")
	sb.WriteString("5. Decide dispatch target: keep both `client_ip` and `node_id` empty for automatic dispatch, or set both fields when the user explicitly requires a specific node.\n")
	sb.WriteString("6. Decide execution mode: use async for action-only requests, sync for requests that need returned data or final completion.\n")
	sb.WriteString("7. Execute only after steps 1-6 pass. If any check fails, stop and report the exact reason instead of creating or executing a task.\n")
	sb.WriteString("8. After execution, inspect the response or task record before answering. For async runs, return the task record or execution identifier available in the response.\n\n")

	sb.WriteString("## Parameter Rules\n\n")
	sb.WriteString("Before creating or executing a task, inspect the workflow's `Parameters` section. If a required parameter has no value, ask the user for it. Pass values through the task `params` object and keep parameter names exactly as listed.\n\n")
	sb.WriteString("## Parameter Type Rules\n\n")
	sb.WriteString("- `string`: pass a string value.\n")
	sb.WriteString("- `number`: pass a JSON number, not a quoted string, when the user provides a numeric value.\n")
	sb.WriteString("- `json`: pass a valid JSON object or array. If the user provides plain text, ask them to confirm the JSON structure before executing.\n")
	sb.WriteString("- `checkbox`: pass a boolean `true` or `false`.\n")
	sb.WriteString("- Example values like `\"\"` are placeholders. Replace required placeholders with real user-provided values before executing.\n\n")

	appendServerWorkflowAllAPIParameterReference(&sb)
	appendServerWorkflowListingResponseRules(&sb)

	sb.WriteString("## Dispatch Rules\n\n")
	sb.WriteString("- Leave both `client_ip` and `node_id` empty when any online node that owns the workflow may execute it.\n")
	sb.WriteString("- Set both `client_ip` and `node_id` when the user explicitly wants a specific execution node.\n")
	sb.WriteString("- Avoid setting only one of `client_ip` or `node_id`; BrowserFlow Server mode identifies execution targets by the pair.\n")
	sb.WriteString("- If the target node is busy, offline, or does not own the workflow, the API creates a failed execution record with a readable reason.\n")
	sb.WriteString("- Per-node Redis locks prevent the same browser node from running multiple Automa workflows concurrently.\n")
	sb.WriteString("- Use `trigger_type: \"skill\"` when executing tasks from this skill so execution records are easy to filter.\n\n")
	sb.WriteString("## Execution Mode Rules\n\n")
	sb.WriteString("- Use asynchronous execution when the user only asks to start, trigger, submit, launch, run, or execute a task and does not ask for returned data or final completion. Set `wait_result` to `false`.\n")
	sb.WriteString("- Use synchronous waiting when the user asks to get, query, search, extract, collect, return, fetch, read, wait for completion, or confirm final success/failure. Set `wait_result` to `true`, set a reasonable `timeout`, and request returned data if needed.\n")
	sb.WriteString("- For variable results, request `return_data.variables: [\"browserflow_output\"]` and read `result.data.variables.browserflow_output` first. If it is missing, report that the workflow completed but did not provide a BrowserFlow output variable.\n")
	sb.WriteString("- For table results, set `return_data.include_table` to `true`. BrowserFlow stores larger table payloads as task record files and returns the execution record for follow-up inspection.\n\n")
	sb.WriteString("## Result Reading Rules\n\n")
	sb.WriteString("- First read the execute response. It may contain `record` for the task record and `result` for the immediate client result.\n")
	sb.WriteString("- If the response contains `record.id`, call `/task-records/{record_id}` before giving the final answer when the user needs status, errors, returned variables, table files, or detailed output.\n")
	sb.WriteString("- For variable output, check `result.data.variables.browserflow_output`, then `record.result.data.variables.browserflow_output`, then the task record detail.\n")
	sb.WriteString("- For table output, inspect the task record detail `files` array. Use the file download endpoint when the table is stored as a file.\n")
	sb.WriteString("- Treat `queued` and `running` as incomplete states. Query the task record again later instead of reporting final success.\n\n")
	sb.WriteString("## Failure Handling Rules\n\n")
	sb.WriteString("- Backend unreachable: ask the user to start BrowserFlow Server and do not execute.\n")
	sb.WriteString("- Preflight response is not successful: report the response `message` or `error` and do not execute.\n")
	sb.WriteString("- Runtime mode is not `server`: ask the user to start or switch to Server mode and do not execute.\n")
	sb.WriteString("- No online nodes: ask the user to open a Windows client browser-agent page and keep it connected.\n")
	sb.WriteString("- No online node owns the workflow: ask the user to install or sync the workflow to a client node, or choose another workflow.\n")
	sb.WriteString("- Target node is busy or locked: report the busy reason and suggest retrying later or choosing another node.\n")
	sb.WriteString("- Missing required parameters: ask for the missing values before creating or executing a task.\n")
	sb.WriteString("- API returns a failed task record: report `record.error_message` or the readable failure reason from the response.\n\n")
	sb.WriteString("## API Endpoints\n\n")
	sb.WriteString("### Query Existing Tasks\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/tasks?workflow_id=%s&page_num=1&page_size=10'\n", baseURL, url.QueryEscape(firstWorkflowID)))
	sb.WriteString("```\n\n")
	sb.WriteString("### Create Task\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/tasks' \\\n", baseURL))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"name":                  "Skill task for " + firstWorkflowID,
		"description":           "Created by BrowserFlow exported Skill",
		"workflow_id":           firstWorkflowID,
		"client_ip":             "",
		"node_id":               "",
		"cron_expression":       "",
		"params":                firstVariables,
		"enabled":               true,
		"run_once_after_create": false,
	}))))
	sb.WriteString("```\n\n")
	sb.WriteString("### Execute Existing Task\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/tasks/{task_id}/execute' \\\n", baseURL))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"trigger_type": "skill",
		"client_ip":    "",
		"node_id":      "",
		"params":       firstVariables,
	}))))
	sb.WriteString("```\n\n")
	sb.WriteString("### Execute Existing Task And Wait For Result\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/tasks/{task_id}/execute' \\\n", baseURL))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"trigger_type": "skill",
		"client_ip":    "",
		"node_id":      "",
		"params":       firstVariables,
		"wait_result":  true,
		"timeout":      300,
		"return_data": map[string]any{
			"variables":       []string{"browserflow_output"},
			"include_table":   true,
			"table_limit":     100,
			"include_history": false,
		},
	}))))
	sb.WriteString("```\n\n")
	sb.WriteString("### Query Task Records\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/task-records?workflow_id=%s&page_num=1&page_size=10'\n", baseURL, url.QueryEscape(firstWorkflowID)))
	sb.WriteString("```\n\n")
	sb.WriteString("### Query Task Record Detail\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/task-records/{record_id}'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("### Download Task Record File\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -O '%s/task-records/files/{file_id}/download'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("## Available Workflows\n\n")
	for index, workflow := range workflows {
		appendServerWorkflowSkillSection(&sb, index+1, workflow, baseURL)
	}
	sb.WriteString("## Usage Notes\n\n")
	sb.WriteString("- These workflows come from the BrowserFlow server database, not from the currently open browser-agent page.\n")
	sb.WriteString("- Prefer creating reusable tasks for repeated use, then execute those task IDs from the skill.\n")
	sb.WriteString("- Leave both `client_ip` and `node_id` empty when any online node that owns the workflow may execute it.\n")
	sb.WriteString("- Set both `client_ip` and `node_id` only when the user explicitly wants a specific execution node.\n")
	sb.WriteString("- A successful execute response means the server accepted and dispatched the task. Use task records to inspect final status and returned data.\n")
	return sb.String()
}

// appendServerWorkflowSkillSection writes one Server workflow section.
func appendServerWorkflowSkillSection(sb *strings.Builder, index int, workflow map[string]any, baseURL string) {
	workflowID := getServerWorkflowExecutionID(workflow)
	name := firstAgentSkillString(workflow, "name", "automa_name", "title")
	if name == "" {
		name = workflowID
	}
	if name == "" {
		name = fmt.Sprintf("Workflow %d", index)
	}

	sb.WriteString(fmt.Sprintf("### %d. %s\n\n", index, markdownLine(name)))
	sb.WriteString(fmt.Sprintf("- Workflow ID: `%s`\n", inlineCode(workflowID)))
	if serverID := firstAgentSkillString(workflow, "server_id"); serverID != "" {
		sb.WriteString(fmt.Sprintf("- Server ID: `%s`\n", inlineCode(serverID)))
	}
	if automaID := firstAgentSkillString(workflow, "automa_id"); automaID != "" && automaID != workflowID {
		sb.WriteString(fmt.Sprintf("- Automa ID: `%s`\n", inlineCode(automaID)))
	}
	if sourceIP := firstAgentSkillString(workflow, "source_ip"); sourceIP != "" {
		sb.WriteString(fmt.Sprintf("- Last Sync Client IP: `%s`\n", inlineCode(sourceIP)))
	}
	if sourceNodeID := firstAgentSkillString(workflow, "source_node_id"); sourceNodeID != "" {
		sb.WriteString(fmt.Sprintf("- Last Sync Node ID: `%s`\n", inlineCode(sourceNodeID)))
	}
	sb.WriteString(fmt.Sprintf("- Description: %s\n", markdownLine(defaultText(firstAgentSkillString(workflow, "description", "automa_description"), "-"))))
	sb.WriteString(fmt.Sprintf("- Status: %s\n", agentWorkflowStatus(workflow)))
	sb.WriteString(fmt.Sprintf("- Nodes: %d\n", getAgentWorkflowNodeCount(workflow)))
	sb.WriteString(fmt.Sprintf("- Created: %s\n", markdownLine(defaultText(formatAgentSkillTime(firstAgentSkillValue(workflow, "createdAt", "created_at", "created")), "-"))))
	sb.WriteString(fmt.Sprintf("- Updated: %s\n\n", markdownLine(defaultText(formatAgentSkillTime(firstAgentSkillValue(workflow, "updatedAt", "updated_at", "updated")), "-"))))

	params := extractAgentWorkflowParameters(workflow)
	if len(params) == 0 {
		sb.WriteString("Parameters: none detected.\n\n")
		appendServerWorkflowTaskExample(sb, baseURL, workflowID, nil)
		return
	}

	sb.WriteString("Parameters:\n")
	appendAgentWorkflowParameterDetails(sb, params)
	sb.WriteString("\n")
	appendServerWorkflowInvocationGuide(sb)
	appendServerWorkflowTaskExample(sb, baseURL, workflowID, buildAgentSkillVariableExample(params))
}

// appendServerWorkflowTaskExample writes a Server task creation example.
func appendServerWorkflowTaskExample(sb *strings.Builder, baseURL string, workflowID string, variables map[string]any) {
	if strings.TrimSpace(workflowID) == "" {
		return
	}
	if variables == nil {
		variables = map[string]any{}
	}
	sb.WriteString("Create task example:\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/tasks' \\\n", baseURL))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString(fmt.Sprintf("  -d %s\n", shellSingleQuote(compactJSON(map[string]any{
		"name":                  "Skill task for " + workflowID,
		"workflow_id":           workflowID,
		"client_ip":             "",
		"node_id":               "",
		"cron_expression":       "",
		"params":                variables,
		"enabled":               true,
		"run_once_after_create": true,
	}))))
	sb.WriteString("```\n\n")
}

func buildAgentSkillDescription(workflows []map[string]any) string {
	names := make([]string, 0, len(workflows))
	for index, workflow := range workflows {
		if index >= 8 {
			break
		}
		name := firstAgentSkillString(workflow, "name", "title")
		if name == "" {
			name = getAgentWorkflowID(workflow)
		}
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "Run BrowserFlow Automa workflows through the local browser-agent HTTP API."
	}
	return "Run BrowserFlow Automa workflows through the local browser-agent HTTP API. Workflows include: " + strings.Join(names, ", ")
}

// extractAgentWorkflowParameters reads trigger parameters 提取触发器参数
// buildServerSkillDescription builds frontmatter description.
func buildServerSkillDescription(workflows []map[string]any) string {
	names := make([]string, 0, len(workflows))
	for index, workflow := range workflows {
		if index >= 8 {
			break
		}
		name := firstAgentSkillString(workflow, "name", "automa_name", "title")
		if name == "" {
			name = getAgentWorkflowID(workflow)
		}
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "Run BrowserFlow Server-mode Automa workflows through the task scheduling API."
	}
	return "Run BrowserFlow Server-mode Automa workflows through the task scheduling API. Workflows include: " + strings.Join(names, ", ")
}

func extractAgentWorkflowParameters(workflow map[string]any) []map[string]any {
	if trigger, ok := agentSkillMapValue(workflow["trigger"]); ok {
		if params := agentSkillMapSliceValue(trigger["parameters"]); len(params) > 0 {
			return params
		}
	}

	drawflow, ok := agentSkillMapValue(workflow["drawflow"])
	if !ok {
		return nil
	}

	if nodes, ok := agentSkillSliceValue(drawflow["nodes"]); ok {
		for _, item := range nodes {
			node, ok := agentSkillMapValue(item)
			if !ok || firstAgentSkillString(node, "label", "name") != "trigger" {
				continue
			}
			if data, ok := agentSkillMapValue(node["data"]); ok {
				return agentSkillMapSliceValue(data["parameters"])
			}
		}
	}

	legacyNodes := nestedAgentSkillMap(drawflow, "drawflow", "Home", "data")
	for _, value := range legacyNodes {
		node, ok := agentSkillMapValue(value)
		if !ok || firstAgentSkillString(node, "name", "label") != "trigger" {
			continue
		}
		if data, ok := agentSkillMapValue(node["data"]); ok {
			return agentSkillMapSliceValue(data["parameters"])
		}
	}

	return nil
}

// buildAgentSkillVariableExample builds variables example 构建变量示例
func buildAgentSkillVariableExample(params []map[string]any) map[string]any {
	variables := make(map[string]any, len(params))
	for _, param := range params {
		name := firstAgentSkillString(param, "name", "key")
		if name == "" {
			continue
		}
		value := firstAgentSkillValue(param, "defaultValue", "default", "value")
		if value == nil {
			value = ""
		}
		variables[name] = value
	}
	return variables
}

// getAgentWorkflowID reads workflow id 读取工作流 ID
func getAgentWorkflowID(workflow map[string]any) string {
	return firstAgentSkillString(workflow, "id", "workflowId", "workflow_id", "automaId", "automa_id")
}

// getServerWorkflowExecutionID reads the Automa id used by Server tasks.
func getServerWorkflowExecutionID(workflow map[string]any) string {
	return firstAgentSkillString(workflow, "automa_id", "workflow_id", "workflowId", "automaId", "id")
}

// getServerWorkflowFilterIDs returns accepted ids for Server export selection.
func getServerWorkflowFilterIDs(workflow map[string]any) []string {
	ids := []string{
		firstAgentSkillString(workflow, "server_id"),
		firstAgentSkillString(workflow, "automa_id"),
		firstAgentSkillString(workflow, "workflow_id", "workflowId", "automaId"),
		firstAgentSkillString(workflow, "id"),
	}
	result := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

// getAgentWorkflowNodeCount counts workflow nodes 统计工作流节点数
func getAgentWorkflowNodeCount(workflow map[string]any) int {
	if count, ok := agentSkillIntValue(firstAgentSkillValue(workflow, "nodeCount", "node_count")); ok {
		return count
	}
	drawflow, ok := agentSkillMapValue(workflow["drawflow"])
	if !ok {
		return 0
	}
	if nodes, ok := agentSkillSliceValue(drawflow["nodes"]); ok {
		return len(nodes)
	}
	return len(nestedAgentSkillMap(drawflow, "drawflow", "Home", "data"))
}

// agentWorkflowStatus formats status 格式化状态
func agentWorkflowStatus(workflow map[string]any) string {
	if agentSkillBoolValue(firstAgentSkillValue(workflow, "isDisabled", "is_disabled", "disabled")) {
		return "disabled"
	}
	return "enabled"
}

// isAgentSkillParamRequired reads required flag 读取必填标记
func isAgentSkillParamRequired(param map[string]any) bool {
	if agentSkillBoolValue(param["required"]) {
		return true
	}
	if data, ok := agentSkillMapValue(param["data"]); ok {
		return agentSkillBoolValue(data["required"])
	}
	return false
}

// firstAgentSkillValue returns first existing value 返回第一个存在的值
func firstAgentSkillValue(data map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			return value
		}
	}
	return nil
}

// firstAgentSkillString returns first non-empty string 返回第一个非空字符串
func firstAgentSkillString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		text := strings.TrimSpace(agentSkillStringValue(data[key]))
		if text != "" {
			return text
		}
	}
	return ""
}

// nestedAgentSkillMap reads nested map 读取嵌套 map
func nestedAgentSkillMap(data map[string]any, keys ...string) map[string]any {
	current := data
	for _, key := range keys {
		next, ok := agentSkillMapValue(current[key])
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

// agentSkillMapValue converts value to map 转换 map
func agentSkillMapValue(value any) (map[string]any, bool) {
	data, ok := value.(map[string]any)
	return data, ok
}

// agentSkillSliceValue converts value to slice 转换切片
func agentSkillSliceValue(value any) ([]any, bool) {
	items, ok := value.([]any)
	return items, ok
}

// agentSkillMapSliceValue converts value to map slice 转换 map 切片
func agentSkillMapSliceValue(value any) []map[string]any {
	switch items := value.(type) {
	case []map[string]any:
		return items
	case []any:
		result := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if data, ok := agentSkillMapValue(item); ok {
				result = append(result, data)
			}
		}
		return result
	default:
		return nil
	}
}

// agentSkillStringValue converts value to string 转换字符串
func agentSkillStringValue(value any) string {
	switch item := value.(type) {
	case string:
		return item
	case json.Number:
		return item.String()
	case float64:
		return strconv.FormatFloat(item, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(item), 'f', -1, 32)
	case int:
		return strconv.Itoa(item)
	case int64:
		return strconv.FormatInt(item, 10)
	case bool:
		return strconv.FormatBool(item)
	default:
		return ""
	}
}

// agentSkillIntValue converts value to int 转换整数
func agentSkillIntValue(value any) (int, bool) {
	switch item := value.(type) {
	case int:
		return item, true
	case int64:
		return int(item), true
	case float64:
		return int(item), true
	case json.Number:
		count, err := item.Int64()
		return int(count), err == nil
	default:
		return 0, false
	}
}

// agentSkillBoolValue converts value to bool 转换布尔值
func agentSkillBoolValue(value any) bool {
	switch item := value.(type) {
	case bool:
		return item
	case string:
		return strings.EqualFold(strings.TrimSpace(item), "true")
	default:
		return false
	}
}

// formatAgentSkillTime formats timestamp 格式化时间戳
func formatAgentSkillTime(value any) string {
	switch item := value.(type) {
	case string:
		return strings.TrimSpace(item)
	case int64:
		return formatAgentSkillUnix(item)
	case int:
		return formatAgentSkillUnix(int64(item))
	case float64:
		return formatAgentSkillUnix(int64(item))
	case json.Number:
		timestamp, err := item.Int64()
		if err != nil {
			return item.String()
		}
		return formatAgentSkillUnix(timestamp)
	default:
		return ""
	}
}

// formatAgentSkillUnix formats unix seconds or milliseconds 格式化秒或毫秒时间戳
func formatAgentSkillUnix(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	if timestamp > 1_000_000_000_000 {
		timestamp = timestamp / 1000
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// formatAgentSkillDefaultValue formats default value 格式化默认值
func formatAgentSkillDefaultValue(value any) string {
	if value == nil {
		return ""
	}
	if text := agentSkillStringValue(value); text != "" {
		return text
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

// compactJSON marshals single-line json 序列化单行 JSON
func compactJSON(value any) string {
	data, err := json.Marshal(normalizeSkillJSONValue(value))
	if err != nil {
		return "{}"
	}
	return string(data)
}

// shellSingleQuote wraps a shell argument with safe single quotes 包装安全单引号参数
func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// buildAgentSkillContentDisposition builds download header 构建下载响应头
func buildAgentSkillContentDisposition(fileName string) string {
	encoded := url.PathEscape(fileName)
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", fileName, encoded)
}

// agentSkillBaseURL builds api base url 构建 API 基础地址
func agentSkillBaseURL(host string, tls bool) string {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1:8001"
	}
	scheme := "http"
	if tls {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/api/v1", scheme, host)
}

func agentSkillBaseURLFromServerAddress(address string, tls bool) string {
	address = strings.TrimSpace(address)
	if address == "" {
		return agentSkillBaseURL("", tls)
	}

	if parsedURL, err := url.Parse(address); err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
		return fmt.Sprintf("%s://%s/api/v1", parsedURL.Scheme, normalizeAgentSkillServerHost(parsedURL.Host))
	}

	return agentSkillBaseURL(normalizeAgentSkillServerHost(address), tls)
}

func normalizeAgentSkillServerHost(host string) string {
	host = strings.TrimRight(strings.TrimSpace(host), "/")
	if strings.HasPrefix(host, ":") {
		return "127.0.0.1" + host
	}

	lowerHost := strings.ToLower(host)
	if strings.HasPrefix(lowerHost, "0.0.0.0:") {
		return "127.0.0.1:" + host[len("0.0.0.0:"):]
	}
	if strings.HasPrefix(lowerHost, "[::]:") {
		return "127.0.0.1:" + host[len("[::]:"):]
	}
	return host
}

func agentSkillBaseURLFromFrontendURL(frontendURL string) string {
	frontendURL = strings.TrimRight(strings.TrimSpace(frontendURL), "/")
	if frontendURL == "" {
		return ""
	}
	if strings.HasSuffix(frontendURL, "/api/v1") {
		return frontendURL
	}
	return frontendURL + "/api/v1"
}

// markdownLine keeps text on one markdown line 保持 Markdown 单行文本
func markdownLine(value string) string {
	value = normalizeSkillText(value)
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(strings.TrimSpace(value))
}

// inlineCode escapes markdown inline code 转义 Markdown 行内代码
func inlineCode(value string) string {
	return strings.ReplaceAll(markdownLine(value), "`", "'")
}

// defaultText returns fallback text 返回兜底文本
func defaultText(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func requiredText(required bool) string {
	if required {
		return "Yes"
	}
	return "No"
}

type skillAPIParamDoc struct {
	Name      string
	Type      string
	Location  string
	Required  string
	Meaning   string
	Available string
	Example   string
}

func appendWorkflowSharedParameterConcepts(sb *strings.Builder, serverMode bool) {
	sb.WriteString("### Shared Variables And Response Rules\n\n")
	sb.WriteString("| Variable / concept | Available values | Meaning |\n")
	sb.WriteString("|---|---|---|\n")
	sb.WriteString("| Workflow trigger parameters | Keys listed under each workflow's `Parameters` section | Pass them through `variables` in Windows/agent mode or `params` in Server mode. Do not invent keys or send placeholders for required values. |\n")
	sb.WriteString("| Parameter types | `string`, `number`, `json`, `checkbox`/boolean | Send JSON values matching the declared type. For `json`, send an object or array; for checkbox/boolean, send `true` or `false`. |\n")
	sb.WriteString("| `browserflow_output` | Automa variable name | Recommended standard output variable. Workflows must set this variable if the model should read a returned value. |\n")
	sb.WriteString("| `wait_result` | `false`, `true` | `false` dispatches only. `true` waits for final success/error/stopped/timeout and can return variables/table data. |\n")
	sb.WriteString("| `timeout` | Positive integer seconds | Maximum sync wait time. Exported examples recommend `300` seconds. |\n")
	sb.WriteString("| `return_data.variables` | `browserflow_output` or custom Automa variable names | Controls which Automa variables should be returned after sync execution. |\n")
	sb.WriteString("| `return_data.include_table` | `true`, `false` | Return table output when the user asks for rows, table data, list data, or tabular results. |\n")
	sb.WriteString("| `return_data.table_limit` | Positive integer rows | Caps directly returned table rows. Use a small value for summaries and a larger value when the user asks for all visible rows. |\n")
	sb.WriteString("| `return_data.include_history` | `true`, `false` | Include detailed execution history only for debugging or auditing. Keep false for normal data requests. |\n")
	sb.WriteString("| Execution statuses | `queued`, `running`, `success`, `error`, `stopped`, `timeout` | Treat `queued`/`running` as incomplete. Treat `success` as completed. Report readable errors for `error`, `stopped`, or `timeout`. |\n")
	if serverMode {
		sb.WriteString("| Dispatch target | Empty `client_ip` and `node_id`, or both set | Leave both empty for automatic dispatch to any online node owning the workflow. Set both only when the user requests a specific node. |\n")
		sb.WriteString("| Server result location | `result`, `record`, task record detail, record files | Read immediate `result` first, then `record.result`, then `/task-records/{record_id}` and files for larger tables. |\n")
	} else {
		sb.WriteString("| Browser target | Exported `browser_id` | Use the browser id embedded in this Skill. Change it only after user confirms the same workflow exists in another browser instance. |\n")
		sb.WriteString("| Windows result location | `result`, `execution.result`, execution detail | Read immediate `result.data.variables`, then `execution.result.data.variables`, then `/workflows/executions/{execution_id}`. |\n")
	}
	sb.WriteString("\n")
}

func appendWorkflowEndpointReference(sb *strings.Builder, title string, route string, note string, docs []skillAPIParamDoc) {
	sb.WriteString("### " + title + "\n\n")
	sb.WriteString("- Route: " + route + "\n")
	if strings.TrimSpace(note) != "" {
		sb.WriteString("- Notes: " + note + "\n")
	}
	if len(docs) == 0 {
		sb.WriteString("- Parameters: none.\n\n")
		return
	}
	sb.WriteString("\n")
	sb.WriteString("| Field | Type | Location | Required | Meaning | Available values / variables | Example |\n")
	sb.WriteString("|---|---|---|---|---|---|---|\n")
	for _, doc := range docs {
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | %s | %s | `%s` |\n",
			inlineCode(doc.Name),
			inlineCode(doc.Type),
			workflowSkillTableCell(doc.Location),
			workflowSkillTableCell(doc.Required),
			workflowSkillTableCell(doc.Meaning),
			workflowSkillTableCell(doc.Available),
			workflowSkillTableCell(doc.Example),
		))
	}
	sb.WriteString("\n")
}

func workflowSkillTableCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func normalizeSkillJSONValue(value any) any {
	switch item := value.(type) {
	case string:
		return normalizeSkillText(item)
	case map[string]any:
		result := make(map[string]any, len(item))
		for key, value := range item {
			result[normalizeSkillText(key)] = normalizeSkillJSONValue(value)
		}
		return result
	case []map[string]any:
		result := make([]map[string]any, 0, len(item))
		for _, value := range item {
			normalized, _ := normalizeSkillJSONValue(value).(map[string]any)
			result = append(result, normalized)
		}
		return result
	case []any:
		result := make([]any, 0, len(item))
		for _, value := range item {
			result = append(result, normalizeSkillJSONValue(value))
		}
		return result
	default:
		return value
	}
}

func normalizeSkillText(value string) string {
	if strings.TrimSpace(value) == "" || skillMojibakeScore(value) == 0 {
		return value
	}

	best := value
	bestScore := skillMojibakeScore(value)
	for _, candidate := range []string{
		repairGB18030Mojibake(value),
		repairLatin1Mojibake(value),
	} {
		if candidate == "" || candidate == value || !utf8.ValidString(candidate) {
			continue
		}
		if score := skillMojibakeScore(candidate); score < bestScore {
			best = candidate
			bestScore = score
		}
	}
	return best
}

func repairGB18030Mojibake(value string) string {
	data, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(value))
	if err != nil || !utf8.Valid(data) {
		return ""
	}
	return string(data)
}

func repairLatin1Mojibake(value string) string {
	data := make([]byte, 0, len(value))
	for _, item := range value {
		if item > 255 {
			return ""
		}
		data = append(data, byte(item))
	}
	if !utf8.Valid(data) {
		return ""
	}
	return string(data)
}

func skillMojibakeScore(value string) int {
	score := 0
	for _, item := range value {
		switch {
		case item == utf8.RuneError:
			score += 4
		case item >= '\uE000' && item <= '\uF8FF':
			score += 4
		case strings.ContainsRune("ÃÂ¤¥€™œš", item):
			score += 3
		}
	}

	for _, marker := range []string{"鏅", "澶", "鍗", "妫", "閿", "绱", "弬", "繚", "畾", "屾", "", "", ""} {
		if strings.Contains(value, marker) {
			score += 2
		}
	}
	return score
}
