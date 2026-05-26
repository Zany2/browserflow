package browserexecutor

import (
	"fmt"
	"strconv"
	"strings"
)

// GenerateSkill builds SKILL.md content for direct browser control. GenerateSkill builds SKILL.md content.
func GenerateSkill(baseURL string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: browserflow-browser-executor-executor\n")
	sb.WriteString("description: Control the current BrowserFlow Windows browser directly through HTTP APIs. Use observe, act, accessibility snapshots, RefIDs, page structure, element diagnostics, page text, page HTML, cookies, storage, and compact actions to navigate, click, type, fill forms, upload files, drag, handle dialogs, scroll, reload, extract data, screenshot, operate mouse/window, run JavaScript, and manage tabs without Automa workflows.\n")
	sb.WriteString("---\n\n")

	sb.WriteString("# BrowserFlow Browser Executor\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString("Use this skill when the user wants to operate the browser directly with an LLM instead of running a prebuilt Automa workflow.\n\n")
	sb.WriteString("Keep the BrowserFlow `browser-agent` client tab alive. Do not close the whole browser after a task; when cleanup is needed, close only task-related business tabs that were opened or used for that task.\n\n")
	sb.WriteString(fmt.Sprintf("**API Base URL:** `%s/browser-executor`\n\n", baseURL))

	sb.WriteString("## Mandatory Workflow\n\n")
	sb.WriteString("1. Check `status` before controlling the browser.\n")
	sb.WriteString("2. Before running a task, inspect the current page and create a concrete step plan. Do not start clicking, typing, navigating, or evaluating JavaScript before the plan exists.\n")
	sb.WriteString("3. If only a `#/browser-agent` tab exists, use `tabs` with `action:\"new\"` or call `navigate`; the backend will create a business tab instead of navigating the agent tab.\n")
	sb.WriteString("4. Prefer `observe` first. It returns status, page info, snapshot text, and optional page text in one call. Use `clickable-elements` or `input-elements` when you only need compact RefID lists.\n")
	sb.WriteString("5. After navigation or a page-changing action, call `observe` or `snapshot` again.\n")
	sb.WriteString("6. Use RefIDs such as `@e1` from `snapshot` for click, type, get-text, and get-value.\n")
	sb.WriteString("7. Prefer `act` for simple intent-driven actions and set `return_observe:true` after page-changing actions.\n")
	sb.WriteString("8. Prefer `fill-form` for multiple fields and `batch` for deterministic sequential actions; do not batch steps that need observation between them.\n")
	sb.WriteString("9. If a RefID fails, call `observe` or `snapshot` again because the page may have changed.\n")
	sb.WriteString("10. Prefer `page-structure` for structured extraction before requesting full `page-content`, and use `element-info` to diagnose one uncertain element.\n")
	sb.WriteString("11. Use `wait` states precisely: `load`, `visible`, `hidden`, `enabled`, `interactable`, `writable`, `stable`, `dom-stable`, `request-idle`, `elements-more-than`, or `time`. For navigation, set `wait_until` to `load`, `dom-stable`, `request-idle`, `page-stable`, or `none`.\n")
	sb.WriteString("12. Use `cookies` and `storage` when the task needs to inspect login/session state or clean up page state. Prefer these semantic APIs over ad hoc JavaScript.\n")
	sb.WriteString("13. Never close the BrowserFlow browser or any `#/browser-agent` client tab. Before using `close-page` or `tabs` with `action:\"close\"`, call `tabs` with `action:\"list\"` and close only task-related business tabs.\n\n")

	sb.WriteString("## Required Step-By-Step Procedure\n\n")
	sb.WriteString("Always work in this order. Do not perform page operations until the planning checks are complete.\n\n")
	sb.WriteString("1. Detect runtime: call `/app/runtime` and `/browser-executor/status`. Confirm the backend is reachable and `status.running` is true.\n")
	sb.WriteString("2. Inspect context: call `observe` on the current page. If the current page is `#/browser-agent` or no useful business page exists, open or navigate to the required business page first, then call `observe` again.\n")
	sb.WriteString("3. Build a plan: write a short numbered plan with the intended browser actions, expected page changes, and the data or final state needed for success.\n")
	sb.WriteString("4. Execute one meaningful step at a time. After every navigation, click, submit, form fill, scroll that reveals content, or JavaScript mutation, call `observe`, `snapshot`, `page-structure`, or a targeted getter to verify the result.\n")
	sb.WriteString("5. If verification fails, stop the current action chain, inspect again, revise the plan, and continue from the verified page state. Do not blindly repeat stale RefIDs or selectors.\n")
	sb.WriteString("6. Use `batch` only for deterministic mini-sequences where no observation is needed between operations. Do not batch an entire unknown workflow.\n")
	sb.WriteString("7. Before reporting success, verify the final page state or extracted data. If the task is incomplete, report the exact blocking condition and the last verified state.\n\n")

	sb.WriteString("## Planning Format\n\n")
	sb.WriteString("Before executing a non-trivial task, produce a compact plan like this:\n\n")
	sb.WriteString("```text\n")
	sb.WriteString("Plan:\n")
	sb.WriteString("1. Verify current browser/page state.\n")
	sb.WriteString("2. Navigate or select the target page.\n")
	sb.WriteString("3. Locate the required controls or data.\n")
	sb.WriteString("4. Perform the action or extraction.\n")
	sb.WriteString("5. Verify the final state and report the result.\n")
	sb.WriteString("```\n\n")
	sb.WriteString("For simple tasks such as reading the current title or URL, the plan can be one sentence, but you still must verify with an API response before answering.\n\n")

	sb.WriteString("## Safety Boundaries\n\n")
	sb.WriteString("Ask the user for explicit confirmation before destructive, irreversible, or externally visible actions, including submitting purchases, orders, payments, account/security changes, deleting data, sending messages or emails, uploading files, clearing cookies/storage, or changing important settings. If the user already gave clear permission for that exact action in the current request, proceed carefully and verify before submitting.\n\n")
	sb.WriteString("Do not use `evaluate` to bypass user confirmation, disable site protections, read unrelated secrets, or mutate sensitive page state when a semantic BrowserFlow API can do the task. Prefer high-level APIs such as `click`, `type`, `fill-form`, `cookies`, and `storage` over custom JavaScript.\n\n")

	sb.WriteString("## Failure Recovery\n\n")
	sb.WriteString("If an action fails or the page state is not what you expected, follow this recovery loop before trying again:\n\n")
	sb.WriteString("1. Stop the current action chain and do not repeat the same stale RefID or selector more than once.\n")
	sb.WriteString("2. Call `observe`, `snapshot`, `page-structure`, or `tabs` to discover the current state.\n")
	sb.WriteString("3. Check for navigation, slow loading, a newly opened tab, modal/dialog, disabled element, validation error, login/session problem, or changed DOM.\n")
	sb.WriteString("4. Revise the plan based on the verified state and continue with fresh RefIDs or a more reliable identifier.\n")
	sb.WriteString("5. If recovery would require a destructive action, credential entry, payment, upload, or account change, ask the user before continuing.\n\n")

	sb.WriteString("## Preflight\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl '%s/app/runtime'\n", baseURL))
	sb.WriteString(fmt.Sprintf("curl '%s/browser-executor/status'\n", baseURL))
	sb.WriteString("```\n\n")

	sb.WriteString("## Core Commands\n\n")
	appendCurl(&sb, baseURL, "Open URL", "navigate", map[string]any{"url": "https://example.com", "wait_until": "load"})
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
	appendCurl(&sb, baseURL, "List Cookies", "cookies", map[string]any{"action": "list"})
	appendCurl(&sb, baseURL, "Set Cookie", "cookies", map[string]any{"action": "set", "name": "token", "value": "abc", "url": "https://example.com", "same_site": "Lax"})
	appendCurl(&sb, baseURL, "List Local Storage", "storage", map[string]any{"action": "list", "type": "local"})
	appendCurl(&sb, baseURL, "Set Local Storage", "storage", map[string]any{"action": "set", "type": "local", "key": "token", "value": "abc"})
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
	sb.WriteString("## Final Response Rules\n\n")
	sb.WriteString("When the task ends, report the outcome with verified evidence. Include what was completed, the final verified page state, and any extracted data the user requested. If the task failed or is incomplete, report the blocker, the last verified page state, and the next suggested action. Do not claim success unless a BrowserFlow API response or observed page state confirms it.\n\n")

	sb.WriteString("## Troubleshooting\n\n")
	sb.WriteString("- If unsure about a command or parameters, call `help` or `help?command=<name>` before guessing.\n")
	sb.WriteString("- If an element is not found, call `observe`, `snapshot`, `clickable-elements`, or `input-elements` again.\n")
	sb.WriteString("- If a page did not update, call `wait` with `dom-stable`, `request-idle`, or a specific element state, then `observe`.\n")
	sb.WriteString("- If a click fails, use `element-info` to check visibility, disabled state, box coordinates, and XPath; then retry with a better identifier or coordinate `mouse` fallback.\n")
	sb.WriteString("- If extraction is empty, try `page-structure`, `page-text`, `page-content`, or a broader selector with a smaller limit.\n")
	sb.WriteString("- If login state or a persisted setting looks wrong, inspect `cookies` and `storage` before retrying page actions.\n")
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
	sb.WriteString("- For `navigate`, choose `wait_until` deliberately. Use `request-idle` for network-heavy SPAs, `dom-stable` for DOM-rendered pages, and `none` only when the next step explicitly waits for something else.\n")
	sb.WriteString("- Call `handle-dialog` before the action that triggers an alert, confirm, prompt, or beforeunload dialog.\n")
	sb.WriteString("- Use `cookies` for authentication/session inspection and `storage` for localStorage/sessionStorage reads or cleanup. Avoid raw `evaluate` for these common state tasks.\n")
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
