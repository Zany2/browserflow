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

	sb.WriteString("## Quick Start\n\n")
	sb.WriteString("Use this Skill with a compact observe-act-verify loop:\n\n")
	sb.WriteString("1. Check runtime and browser status when needed.\n")
	sb.WriteString("2. Inspect only the controls or page area needed for the next action.\n")
	sb.WriteString("3. Group deterministic small actions that share one user intent.\n")
	sb.WriteString("4. Verify the page state after each action group.\n")
	sb.WriteString("5. Extract only the requested data, then close only task-related business tabs.\n\n")

	sb.WriteString("## Skill Map\n\n")
	sb.WriteString("- `Mandatory Workflow`: required safety and browser-control rules.\n")
	sb.WriteString("- `Targeted Inspection Rules`: how to avoid fetching the whole page by default.\n")
	sb.WriteString("- `Action Grouping Rules`: how to combine small deterministic operations.\n")
	sb.WriteString("- `Search, Filter, And Pagination Rules`: how to operate visible page controls reliably.\n")
	sb.WriteString("- `Core Commands`: HTTP examples for each browser executor API.\n")
	sb.WriteString("- `Troubleshooting`: recovery rules for stale elements, slow pages, and failed operations.\n\n")

	sb.WriteString("## Mandatory Workflow\n\n")
	sb.WriteString("1. Use lightweight preflight: check `status` before the first browser operation in a conversation, before long or sensitive tasks, and after any executor failure. Reuse a recent successful status check for short follow-up operations on the same browser.\n")
	sb.WriteString("2. Before running a task, inspect the current page and create a concrete step plan. Do not start clicking, typing, navigating, or evaluating JavaScript before the plan exists.\n")
	sb.WriteString("3. If only a `#/browser-agent` tab exists, use `tabs` with `action:\"new\"` or call `navigate`; the backend will create a business tab instead of navigating the agent tab.\n")
	sb.WriteString("4. Inspect only the page area or element set needed for the current step. Prefer `input-elements`, `clickable-elements`, `page-structure`, `element-info`, or low-limit `observe` before broad page text or HTML.\n")
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
	sb.WriteString("1. Decide preflight depth: call `/app/runtime` and `/browser-executor/status` for the first operation, long tasks, sensitive tasks, or when no recent successful preflight is available. For short follow-up operations, reuse the recent status check; if an operation fails, run preflight before retrying.\n")
	sb.WriteString("2. Inspect context: call `observe` on the current page. If the current page is `#/browser-agent` or no useful business page exists, open or navigate to the required business page first, then call `observe` again.\n")
	sb.WriteString("3. Build a plan: write a short numbered plan with the intended browser actions, expected page changes, and the data or final state needed for success.\n")
	sb.WriteString("4. Execute one meaningful step at a time. After every navigation, click, submit, form fill, scroll that reveals content, or JavaScript mutation, call `observe`, `snapshot`, `page-structure`, or a targeted getter to verify the result.\n")
	sb.WriteString("5. If verification fails, stop the current action chain, inspect again, revise the plan, and continue from the verified page state. Do not blindly repeat stale RefIDs or selectors.\n")
	sb.WriteString("6. Use `batch` only for deterministic mini-sequences where no observation is needed between operations. Do not batch an entire unknown workflow.\n")
	sb.WriteString("7. Before reporting success, verify the final page state or extracted data. If the task is incomplete, report the exact blocking condition and the last verified state.\n\n")

	sb.WriteString("## Targeted Inspection Rules\n\n")
	sb.WriteString("- Do not fetch or analyze the whole page by default. Decide what the current step needs, then request the smallest useful context.\n")
	sb.WriteString("- For search boxes or form fields, use `input-elements` or `fill-form` matching first. For buttons, tabs, filters, and pagination, use `clickable-elements` or compact `observe` without large page text.\n")
	sb.WriteString("- For lists and result pages, prefer `page-structure` or a focused read-only `evaluate` after the page has been inspected. Extract only the target container or requested number of rows/items.\n")
	sb.WriteString("- Use `page-text` or `page-content` only when narrow APIs cannot answer the question, and keep limits small. Increase limits gradually only if the required target is missing.\n")
	sb.WriteString("- If a page is large, inspect in stages: relevant controls first, target result container next, then focused extraction. Avoid mixing navigation controls, footer links, and unrelated page content into one analysis step.\n\n")

	sb.WriteString("## Action Grouping Rules\n\n")
	sb.WriteString("- A meaningful step can be an action group, not a single low-level browser event. Group small operations that share one user intent when the required controls are already known and no intermediate decision is needed.\n")
	sb.WriteString("- Good action groups include: type a search keyword and submit it; open a filter panel and choose known filter values; choose known filter values and click a known page number; fill several fields in one form; close several verified task-related business tabs.\n")
	sb.WriteString("- Keep verification at the boundary of each action group. After the group finishes, verify the expected page state, active filters, current page, or extracted data once, instead of reporting after every click or keystroke.\n")
	sb.WriteString("- Do not group across an unknown decision point. If the next target depends on newly loaded content, a modal, dynamic layout, login state, or an uncertain RefID, inspect the page first and continue with fresh identifiers.\n")
	sb.WriteString("- Prefer `fill-form`, `act` with `return_observe:true`, or `batch` for deterministic grouped operations. If any grouped operation fails, stop, inspect, and recover from the last verified state.\n\n")

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

	sb.WriteString("## Lightweight Preflight Strategy\n\n")
	sb.WriteString("- First operation in a conversation: call `/app/runtime` and `/browser-executor/status`.\n")
	sb.WriteString("- Short follow-up operations on the same running browser: reuse a recent successful preflight to save time.\n")
	sb.WriteString("- Long, multi-step, data extraction, form submission, upload, account-changing, or destructive tasks: perform preflight unless a recent successful preflight is already available.\n")
	sb.WriteString("- If any command fails because the executor is offline, disconnected, page context is missing, or the backend is unreachable, run preflight again before retrying.\n")
	sb.WriteString("- Direct browser actions still need page verification. Reusing preflight does not replace `observe`, `snapshot`, or another page-state check after page-changing actions.\n\n")

	sb.WriteString("## Stepwise Page Operation Rules\n\n")
	sb.WriteString("- Match the operation depth to the task. For complex tasks, operate step by step: open the site, inspect the page, analyze available controls, perform one action, verify the result, then continue.\n")
	sb.WriteString("- Step by step means verified action groups, not every keystroke or click. Keep the loop efficient while preserving page-state verification.\n")
	sb.WriteString("- Complex tasks include search with filters, sorting, pagination, multi-step forms, login-sensitive pages, data extraction, pages with dynamic content, and tasks where the site behavior is unknown.\n")
	sb.WriteString("- After opening a website or changing pages, inspect the smallest useful target before deciding the next meaningful action: inputs for search/form work, clickables for controls, structure for lists, or element-info for one uncertain target.\n")
	sb.WriteString("- After every meaningful page-changing action group, verify the updated state with a focused inspection before continuing.\n")
	sb.WriteString("- Do not rely on stale RefIDs, guessed selectors, guessed page state, or JavaScript extraction before the current page has been inspected.\n")
	sb.WriteString("- For simple tasks such as opening a user-provided exact URL, reading the current title, or visiting a stable static page without search/filter criteria, a direct URL or one-step action is acceptable, but still verify the result before answering.\n\n")

	sb.WriteString("## Search, Filter, And Pagination Rules\n\n")
	sb.WriteString("- For user-visible search, sorting, filters, tabs, date ranges, categories, and pagination, use real page interactions by default, even when the requested operation looks simple.\n")
	sb.WriteString("- Once controls are identified, combine related operations when safe: for example, enter the query and submit search as one group, or apply known filters and choose the requested page as one group, then verify the final state.\n")
	sb.WriteString("- Do not satisfy search/filter/sort/pagination requirements by manually constructing URL query parameters. Open the site or search page first, inspect available controls, then interact with the visible controls step by step.\n")
	sb.WriteString("- Direct URL navigation is allowed only for opening a user-provided exact URL or a stable entry page. Do not encode requested search keywords, filters, sort orders, date ranges, categories, or page numbers into the URL yourself.\n")
	sb.WriteString("- After applying search, sorting, filters, or pagination, verify the selected labels, active tabs, current page number, result count or visible result changes with `observe`, `snapshot`, or `page-structure`.\n")
	sb.WriteString("- If the page state cannot prove the requested filters are active, use the visible controls to select them manually before extracting data.\n")
	sb.WriteString("- Do not claim a filter, sort order, date range, category, or page number was applied only because a URL parameter exists.\n")
	sb.WriteString("- If URL construction and visible controls disagree, trust the verified visible page state and adjust the page through controls.\n\n")

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
	sb.WriteString("Use these commands for first calls, long/sensitive tasks, or recovery after failures. For short follow-up calls in the same conversation, a recent successful result may be reused.\n\n")
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
	sb.WriteString("- Use the narrowest inspection API that fits the step. `observe` is useful for a compact overview, but avoid large `include_text`/`text_limit` values unless the task needs broad page text.\n")
	sb.WriteString("- Use `input-elements` for text boxes and form fields, `clickable-elements` for buttons/links/tabs/filters/pagination, and `page-structure` for compact structured data such as headings, links, forms, tables, images, and buttons.\n")
	sb.WriteString("- Prefer focused result extraction over whole-page extraction. For example, extract only the visible result list and requested item count, not the header, footer, sidebars, and unrelated links.\n")
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
