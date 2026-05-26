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

	sb.WriteString("## Mandatory Preflight\n\n")
	sb.WriteString("Before running any workflow, first verify that the BrowserFlow backend is reachable.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/app/runtime'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("If the request fails, ask the user to start the BrowserFlow backend before continuing.\n\n")
	sb.WriteString("Then verify that the browser instance exported with this skill is online.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/agents/status'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("Find an agent whose `browser_id` matches the Browser Instance ID in this skill. It must be online. If it is missing or offline, ask the user to start that exact browser instance and keep the browser-agent page connected.\n\n")
	sb.WriteString("After confirming the agent is online, verify that its Automa plugin status reports `automa_installed: true`. If Automa is not installed or not available, ask the user to install or enable the Automa extension in that browser instance, then refresh the browser-agent page before continuing.\n\n")
	sb.WriteString("Do not replace the exported `browser_id` with the current browser unless the user explicitly confirms that the workflow exists in the new browser instance.\n\n")

	sb.WriteString("## Required Step-By-Step Procedure\n\n")
	sb.WriteString("Always work in this order. Do not run or open a workflow until the checks are complete.\n\n")
	sb.WriteString("1. Detect runtime: call `/app/runtime` and confirm the BrowserFlow backend is reachable.\n")
	sb.WriteString("2. Detect client: call `/agents/status`, find the exported `browser_id`, confirm it is online, and confirm `automa_installed: true`.\n")
	sb.WriteString("3. Detect workflow and parameters: choose the matching workflow from this Skill, inspect its `Parameters`, and ask the user for any missing required values.\n")
	sb.WriteString("4. Decide execution mode: use async for action-only requests, sync for requests that need returned data or final completion.\n")
	sb.WriteString("5. Execute only after steps 1-4 pass. If any check fails, stop and report the exact reason instead of calling the run API.\n")
	sb.WriteString("6. For sync runs, inspect the execution result before answering. For async runs, return the `execution_id` and explain that status can be queried later.\n\n")

	sb.WriteString("## Parameter Rules\n\n")
	sb.WriteString("Before running a workflow, inspect its `Parameters` section. If a required parameter has no value, ask the user for it before calling the API. If an optional parameter has a default value, use the default unless the user provides another value. Pass parameters through the `variables` object, and keep parameter names exactly as listed in this skill. BrowserFlow treats this `variables` object as the completed parameter set and instructs Automa not to open its own parameter input page.\n\n")

	sb.WriteString("## Execution Mode Rules\n\n")
	sb.WriteString("Before running a workflow, decide whether the user needs the final workflow result.\n\n")
	sb.WriteString("- Use asynchronous execution when the user only asks to start, trigger, submit, open, launch, run, or execute a task and does not ask for returned data or final completion. Set `wait_result` to `false`. Return the `execution.execution_id` to the user so they can query status or results later.\n")
	sb.WriteString("- Use synchronous waiting when the user asks to get, query, search, extract, collect, return, fetch, read, wait for completion, or confirm final success/failure. Set `wait_result` to `true`, set a reasonable `timeout`, and request returned data if needed.\n")
	sb.WriteString("- For data-returning requests, request `return_data.variables: [\"browserflow_output\"]` and read `execution.result.data.variables.browserflow_output` first. If it is missing, report that the workflow completed but did not provide a BrowserFlow output variable.\n")
	sb.WriteString("- If the user asks to check a previous task, use the execution status endpoint with the saved `execution_id` instead of running the workflow again.\n")
	sb.WriteString("- If the intent is ambiguous, prefer async mode for action-only tasks and sync mode for data-returning tasks.\n\n")

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
	sb.WriteString("- The target browser must report `automa_installed: true`; otherwise workflow list, open, and run commands may fail.\n")
	sb.WriteString("- Pass trigger parameters through the `variables` object. Parameter names must match the Automa trigger configuration.\n")
	sb.WriteString("- Do not rely on Automa's parameter tab for Skill calls; collect required values before sending the HTTP request.\n")
	sb.WriteString("- If the exported browser instance is no longer available, choose another running browser and update `browser_id`.\n")
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
	for _, param := range params {
		name := firstAgentSkillString(param, "name", "key")
		if name == "" {
			continue
		}
		paramType := defaultText(firstAgentSkillString(param, "type"), "string")
		description := defaultText(firstAgentSkillString(param, "description", "placeholder"), "-")
		required := ""
		if isAgentSkillParamRequired(param) {
			required = ", required"
		}
		defaultValue := formatAgentSkillDefaultValue(firstAgentSkillValue(param, "defaultValue", "default", "value"))
		if defaultValue != "" {
			sb.WriteString(fmt.Sprintf("- `%s` (%s%s): %s Default: `%s`\n", inlineCode(name), markdownLine(paramType), required, markdownLine(description), inlineCode(defaultValue)))
		} else {
			sb.WriteString(fmt.Sprintf("- `%s` (%s%s): %s\n", inlineCode(name), markdownLine(paramType), required, markdownLine(description)))
		}
	}
	sb.WriteString("\n")
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
	sb.WriteString("If the request fails or the mode is not `server`, ask the user to start BrowserFlow in Server mode before continuing.\n\n")
	sb.WriteString("Then verify that at least one client node is online and has the target workflow. If a task does not specify a client node, BrowserFlow scans online nodes that own the workflow and chooses an unlocked node.\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/clients'\n", baseURL))
	sb.WriteString("```\n\n")
	sb.WriteString("## Required Step-By-Step Procedure\n\n")
	sb.WriteString("Always work in this order. Do not create or execute a task until the checks are complete.\n\n")
	sb.WriteString("1. Detect runtime: call `/app/runtime` and confirm BrowserFlow is reachable and running in `server` mode.\n")
	sb.WriteString("2. Detect client nodes: call `/clients`, confirm there is at least one online node, and confirm the target workflow can be dispatched to an online node.\n")
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
	for _, param := range params {
		name := firstAgentSkillString(param, "name", "key")
		if name == "" {
			continue
		}
		paramType := defaultText(firstAgentSkillString(param, "type"), "string")
		description := defaultText(firstAgentSkillString(param, "description", "placeholder"), "-")
		required := ""
		if isAgentSkillParamRequired(param) {
			required = ", required"
		}
		defaultValue := formatAgentSkillDefaultValue(firstAgentSkillValue(param, "defaultValue", "default", "value"))
		if defaultValue != "" {
			sb.WriteString(fmt.Sprintf("- `%s` (%s%s): %s Default: `%s`\n", inlineCode(name), markdownLine(paramType), required, markdownLine(description), inlineCode(defaultValue)))
		} else {
			sb.WriteString(fmt.Sprintf("- `%s` (%s%s): %s\n", inlineCode(name), markdownLine(paramType), required, markdownLine(description)))
		}
	}
	sb.WriteString("\n")
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
