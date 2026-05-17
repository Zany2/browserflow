package browserexecutor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
)

// GenerateSkill builds SKILL.md content for direct browser control. GenerateSkill builds SKILL.md content.
func GenerateSkill(baseURL string, status model.BrowserExecutorStatus) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: browserflow-browser-executor\n")
	sb.WriteString("description: Control the current BrowserFlow Windows browser directly through HTTP APIs. Use observe, act, accessibility snapshots, RefIDs, page structure, element diagnostics, page text, page HTML, and compact actions to navigate, click, type, fill forms, upload files, drag, handle dialogs, scroll, reload, extract data, screenshot, operate mouse/window, run JavaScript, and manage tabs without Automa workflows.\n")
	sb.WriteString("---\n\n")

	sb.WriteString("# BrowserFlow Browser Executor\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString("Use this skill when the user wants to operate the browser directly with an LLM instead of running a prebuilt Automa workflow.\n\n")
	sb.WriteString("Keep the BrowserFlow `browser-agent` client tab alive. Do not close the whole browser after a task; when cleanup is needed, close only task-related business tabs that were opened or used for that task.\n\n")
	sb.WriteString(fmt.Sprintf("**API Base URL:** `%s/browser-executor`\n\n", baseURL))
	if status.BrowserID != "" {
		sb.WriteString(fmt.Sprintf("**Current Browser Instance ID:** `%s`\n\n", inline(status.BrowserID)))
	}
	if status.URL != "" {
		sb.WriteString(fmt.Sprintf("**Current Page:** `%s`\n\n", inline(status.URL)))
	}

	sb.WriteString("## Mandatory Workflow\n\n")
	sb.WriteString("1. Check `status` before controlling the browser.\n")
	sb.WriteString("2. Before running a task, decide the concrete steps first, then execute them one by one and verify the page state after each meaningful step.\n")
	sb.WriteString("3. If only a `#/browser-agent` tab exists, use `tabs` with `action:\"new\"` or call `navigate`; the backend will create a business tab instead of navigating the agent tab.\n")
	sb.WriteString("4. Prefer `observe` first. It returns status, page info, snapshot text, and optional page text in one call. Use `clickable-elements` or `input-elements` when you only need compact RefID lists.\n")
	sb.WriteString("5. After navigation or a page-changing action, call `observe` or `snapshot` again.\n")
	sb.WriteString("6. Use RefIDs such as `@e1` from `snapshot` for click, type, get-text, and get-value.\n")
	sb.WriteString("7. Prefer `act` for simple intent-driven actions and set `return_observe:true` after page-changing actions.\n")
	sb.WriteString("8. Prefer `fill-form` for multiple fields and `batch` for deterministic sequential actions; do not batch steps that need observation between them.\n")
	sb.WriteString("9. If a RefID fails, call `observe` or `snapshot` again because the page may have changed.\n")
	sb.WriteString("10. Prefer `page-structure` for structured extraction before requesting full `page-content`, and use `element-info` to diagnose one uncertain element.\n")
	sb.WriteString("11. Use `wait` states precisely: `load`, `visible`, `hidden`, `enabled`, `interactable`, `writable`, `stable`, `dom-stable`, `request-idle`, `elements-more-than`, or `time`.\n")
	sb.WriteString("12. Never close the BrowserFlow browser or any `#/browser-agent` client tab. Before using `close-page` or `tabs` with `action:\"close\"`, call `tabs` with `action:\"list\"` and close only task-related business tabs.\n\n")

	sb.WriteString("## Preflight\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/app/runtime'\n", baseURL))
	sb.WriteString(fmt.Sprintf("curl '%s/browser-executor/status'\n", baseURL))
	sb.WriteString("```\n\n")

	sb.WriteString("## Core Commands\n\n")
	appendCurl(&sb, baseURL, "Open URL", "navigate", map[string]any{"url": "https://example.com"})
	appendCurl(&sb, baseURL, "Observe Page", "observe", map[string]any{"include_text": false, "text_limit": 8000})
	appendCurl(&sb, baseURL, "Get Snapshot", "snapshot", nil)
	appendCurl(&sb, baseURL, "Clickable Elements", "clickable-elements", map[string]any{"limit": 50})
	appendCurl(&sb, baseURL, "Input Elements", "input-elements", map[string]any{"limit": 30})
	appendCurl(&sb, baseURL, "Page Structure", "page-structure", map[string]any{"include_links": true, "include_forms": true, "include_tables": true, "include_images": false, "include_buttons": true, "limit": 30})
	appendCurl(&sb, baseURL, "Element Info", "element-info", map[string]any{"identifier": "@e1", "attributes": []string{"id", "class", "href", "aria-label"}})
	appendCurl(&sb, baseURL, "Smart Action", "act", map[string]any{"intent": "click", "identifier": "@e1", "return_observe": true})
	appendCurl(&sb, baseURL, "Click Element", "click", map[string]any{"identifier": "@e1"})
	appendCurl(&sb, baseURL, "Type Text", "type", map[string]any{"identifier": "@e2", "text": "hello", "clear": true})
	appendCurl(&sb, baseURL, "Select Option", "select", map[string]any{"identifier": "@e3", "value": "China"})
	appendCurl(&sb, baseURL, "Fill Form", "fill-form", map[string]any{"fields": []map[string]any{{"name": "email", "value": "user@example.com"}, {"name": "password", "value": "secret"}}, "submit": false, "timeout": 10})
	appendCurl(&sb, baseURL, "Press Key", "press-key", map[string]any{"key": "Enter"})
	appendCurl(&sb, baseURL, "Wait For Element", "wait", map[string]any{"identifier": "@e1", "state": "interactable", "timeout": 10})
	appendCurl(&sb, baseURL, "Wait For DOM Stable", "wait", map[string]any{"state": "dom-stable", "timeout": 10})
	appendCurl(&sb, baseURL, "Hover Element", "hover", map[string]any{"identifier": "@e1"})
	appendCurl(&sb, baseURL, "Drag Element", "drag", map[string]any{"from_identifier": "@e1", "to_identifier": "@e2"})
	appendCurl(&sb, baseURL, "Upload Files", "file-upload", map[string]any{"identifier": "@e4", "file_paths": []string{"C:\\\\path\\\\file.png"}})
	appendCurl(&sb, baseURL, "Arm Dialog Handler", "handle-dialog", map[string]any{"accept": true, "text": "", "timeout": 10})
	appendCurl(&sb, baseURL, "Get Value", "get-value", map[string]any{"identifier": "@e2"})
	appendCurl(&sb, baseURL, "Page Text", "page-text", map[string]any{"limit": 8000})
	appendCurl(&sb, baseURL, "Page Content", "page-content", map[string]any{"limit": 12000})
	appendCurl(&sb, baseURL, "Scroll Page", "scroll", map[string]any{"direction": "down", "pixels": 700})
	appendCurl(&sb, baseURL, "Reload Page", "reload", map[string]any{})
	appendCurl(&sb, baseURL, "Resize Viewport", "resize", map[string]any{"width": 1440, "height": 900})
	appendCurl(&sb, baseURL, "Window Info", "window", map[string]any{"action": "info"})
	appendCurl(&sb, baseURL, "Mouse Click", "mouse", map[string]any{"action": "click", "x": 300, "y": 200, "button": "left"})
	appendCurl(&sb, baseURL, "Extract Text", "extract", map[string]any{"selector": "body", "fields": []string{"text"}, "multiple": false})
	appendCurl(&sb, baseURL, "Screenshot", "screenshot", map[string]any{"full_page": true, "format": "png"})
	appendCurl(&sb, baseURL, "Element Screenshot", "element-screenshot", map[string]any{"identifier": "@e1", "format": "png"})
	appendCurl(&sb, baseURL, "Batch Operations", "batch", map[string]any{"operations": []map[string]any{{"type": "navigate", "params": map[string]any{"url": "https://example.com"}, "stop_on_error": true}, {"type": "observe", "params": map[string]any{"include_text": false}, "stop_on_error": true}}})

	sb.WriteString("## Element Identification\n\n")
	sb.WriteString("1. Prefer RefIDs such as `@e1` from `snapshot`, `observe`, `clickable-elements`, or `input-elements`.\n")
	sb.WriteString("2. Use an exact CSS selector only when the user provides one or the page structure is stable.\n")
	sb.WriteString("3. Use XPath or visible text only for obvious buttons and links.\n")
	sb.WriteString("4. If an identifier fails, refresh with `observe` or `snapshot` because RefIDs may be stale.\n")
	sb.WriteString("5. `fill-form` can match fields by name, id, placeholder, aria-label, or associated label text; use it for multi-field forms.\n\n")

	sb.WriteString("## Efficient Inspection\n\n")
	sb.WriteString("- Use `observe` for the normal control loop: status, page info, snapshot text, RefIDs, and optional page text in one response.\n")
	sb.WriteString("- Use `page-structure` for compact structured data such as headings, links, forms, tables, images, and buttons. Prefer it before `page-content`.\n")
	sb.WriteString("- Use `element-info` when a target element is ambiguous, disabled, hidden, overlapped, or needs attributes/XPath/box coordinates.\n")
	sb.WriteString("- Use `element-screenshot` or `screenshot` only when visual confirmation is needed; base64 can be large.\n")
	sb.WriteString("- Use `mouse` as a coordinate fallback after obtaining coordinates from `element-info`, screenshot inspection, or user instructions. Prefer semantic `click`/`act` first.\n\n")

	sb.WriteString("## Response Format\n\n")
	sb.WriteString("GoFrame wraps responses as `{code,message,data}`. Browser operation data is usually in `data.result`. Check `data.result.success`, `data.result.error`, and `data.result.data` before reporting success.\n\n")

	sb.WriteString("## Troubleshooting\n\n")
	sb.WriteString("- If unsure about a command or parameters, call `help` or `help?command=<name>` before guessing.\n")
	sb.WriteString("- If an element is not found, call `observe`, `snapshot`, `clickable-elements`, or `input-elements` again.\n")
	sb.WriteString("- If a page did not update, call `wait` with `dom-stable`, `request-idle`, or a specific element state, then `observe`.\n")
	sb.WriteString("- If a click fails, use `element-info` to check visibility, disabled state, box coordinates, and XPath; then retry with a better identifier or coordinate `mouse` fallback.\n")
	sb.WriteString("- If extraction is empty, try `page-structure`, `page-text`, `page-content`, or a broader selector with a smaller limit.\n")
	sb.WriteString("- If `status.running` is false, ask the user to reopen the BrowserFlow browser-agent page.\n\n")

	sb.WriteString("## Other Endpoints\n\n")
	for _, command := range Help("") {
		sb.WriteString(fmt.Sprintf("- `%s %s/browser-executor%s` - %s\n", command.Method, baseURL, command.Path, command.Description))
	}
	sb.WriteString("\n## Notes\n\n")
	sb.WriteString("- This skill does not use Automa workflows or Automa trigger parameters.\n")
	sb.WriteString("- Prefer `observe` when you need multiple facts about the page in one round trip.\n")
	sb.WriteString("- Prefer `page-structure` for compact structured extraction before reading raw HTML with `page-content`.\n")
	sb.WriteString("- Prefer `element-info` when one element needs text, attributes, state, coordinates, or XPath for diagnosis.\n")
	sb.WriteString("- Prefer `act` for click/type/select/check/navigate/scroll when the intent is clear.\n")
	sb.WriteString("- Set `return_observe:true` on `act`, `navigate`, `click`, `type`, `select`, `fill-form`, or `scroll` when you need the updated page state.\n")
	sb.WriteString("- Prefer `fill-form` over repeated `type` calls when a page has several fields.\n")
	sb.WriteString("- Prefer `wait` with specific states (`interactable`, `enabled`, `writable`, `dom-stable`, `request-idle`) instead of blind time sleeps.\n")
	sb.WriteString("- Call `handle-dialog` before the action that triggers an alert, confirm, prompt, or beforeunload dialog.\n")
	sb.WriteString("- Prefer `page-text` or `page-content` only when the model needs broad context, and keep limits small.\n")
	sb.WriteString("- Browser control APIs are powerful. Use them only against the local BrowserFlow backend.\n")
	sb.WriteString("- Do not close the browser as a task cleanup step. Keep the BrowserFlow `browser-agent` client tab open so the local executor remains connected.\n")
	sb.WriteString("- If task cleanup is requested, close only pages opened or used for the current task, and never close pages whose URL contains `#/browser-agent`.\n")
	sb.WriteString("- Prefer `snapshot` RefIDs over raw CSS selectors unless the user provides an exact selector.\n")
	sb.WriteString("- `evaluate` accepts normal JavaScript and auto-wraps it as a function when needed, so `return document.title` is valid. Prefer evaluate for read-only extraction; do not mutate page state with evaluate unless normal APIs cannot do it.\n")
	sb.WriteString("- `screenshot` returns base64 image data; summarize it unless the user asks for the raw data.\n")
	sb.WriteString("- `element-screenshot` also returns base64 image data and is cheaper than a full-page screenshot for visual checks.\n")
	return sb.String()
}

func appendCurl(sb *strings.Builder, baseURL string, title string, path string, body map[string]any) {
	sb.WriteString("### " + title + "\n\n")
	if body == nil {
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("curl '%s/browser-executor/%s'\n", baseURL, path))
		sb.WriteString("```\n\n")
		return
	}
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X POST '%s/browser-executor/%s' \\\n", baseURL, path))
	sb.WriteString("  -H 'Content-Type: application/json' \\\n")
	sb.WriteString("  -d '" + compactJSON(body) + "'\n")
	sb.WriteString("```\n\n")
}

func compactJSON(value map[string]any) string {
	parts := make([]string, 0, len(value))
	for key, val := range value {
		parts = append(parts, strconv.Quote(key)+":"+compactJSONValue(val))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func compactJSONValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strconv.Quote(typed)
	case bool:
		return fmt.Sprintf("%t", typed)
	case int, int64, float64:
		return fmt.Sprintf("%v", typed)
	case map[string]any:
		return compactJSON(typed)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, compactJSONValue(item))
		}
		return "[" + strings.Join(items, ",") + "]"
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, strconv.Quote(item))
		}
		return "[" + strings.Join(items, ",") + "]"
	case []map[string]any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, compactJSON(item))
		}
		return "[" + strings.Join(items, ",") + "]"
	default:
		return strconv.Quote(fmt.Sprintf("%v", typed))
	}
}

func inline(value string) string {
	return strings.ReplaceAll(value, "`", "\\`")
}
