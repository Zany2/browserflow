package workflowskill

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FileName is the exported Skill filename. ??? Skill ????
const FileName = "SKILL_AUTOMA.md"

// FilterWorkflows keeps workflows by export scope. ???????????
func FilterWorkflows(workflows []map[string]any, scope string, workflowIDs []string) []map[string]any {
	return filterAgentSkillWorkflows(workflows, scope, workflowIDs)
}

// GenerateMarkdown builds SKILL.md content. ?? SKILL.md ???
func GenerateMarkdown(workflows []map[string]any, baseURL string, browserID string) string {
	return generateAgentWorkflowSkillMD(workflows, baseURL, browserID)
}

// ContentDisposition builds download header. ????????
func ContentDisposition(fileName string) string {
	return buildAgentSkillContentDisposition(fileName)
}

// BaseURL builds api base url. ?? API ?????
func BaseURL(host string, tls bool) string {
	return agentSkillBaseURL(host, tls)
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

// generateAgentWorkflowSkillMD builds SKILL.md content 构建 SKILL.md 内容
func generateAgentWorkflowSkillMD(workflows []map[string]any, baseURL string, browserID string) string {
	var sb strings.Builder
	firstWorkflowID := getAgentWorkflowID(workflows[0])
	firstVariables := buildAgentSkillVariableExample(extractAgentWorkflowParameters(workflows[0]))

	sb.WriteString("---\n")
	sb.WriteString("name: browserflow-automa-workflows\n")
	sb.WriteString("description: " + strconv.Quote(buildAgentSkillDescription(workflows)) + "\n")
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

// buildAgentSkillDescription builds frontmatter description 构建 frontmatter 描述
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
	data, err := json.Marshal(value)
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

// markdownLine keeps text on one markdown line 保持 Markdown 单行文本
func markdownLine(value string) string {
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
