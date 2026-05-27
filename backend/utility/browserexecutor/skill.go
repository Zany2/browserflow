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
	sb.WriteString("- `Fast Path Decision Table`: the shortest safe command path for common tasks.\n")
	sb.WriteString("- `High-Frequency Parameter Reference`: field meanings for the APIs most models should use first.\n")
	sb.WriteString("- `All Endpoint Parameter Reference`: complete parameter details, enums, and variable meanings for every browser executor API.\n")
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

	appendFastPathDecisionTable(&sb)
	appendHighFrequencyParameterReference(&sb)
	appendAllEndpointParameterReference(&sb)

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

func appendFastPathDecisionTable(sb *strings.Builder) {
	sb.WriteString("## Fast Path Decision Table\n\n")
	sb.WriteString("Use this table before choosing APIs. It is designed for models that need an explicit shortest safe route.\n\n")
	sb.WriteString("| User goal | Preferred command sequence | Notes |\n")
	sb.WriteString("|---|---|---|\n")
	sb.WriteString("| Open a site or exact URL | `status` if needed -> `navigate` -> focused `observe` | Direct URL is allowed only for an exact URL or stable entry page. |\n")
	sb.WriteString("| Search with keyword | `navigate` -> `input-elements` -> `fill-form` or `act` -> focused `observe` | Enter query and submit as one action group when controls are clear. |\n")
	sb.WriteString("| Apply filters/sort/page | `clickable-elements` or compact `observe` -> `act` or `batch` -> focused `observe` | Use visible controls, not handcrafted query params. |\n")
	sb.WriteString("| Fill a form | `input-elements` -> `fill-form` -> `wait` -> focused `observe` | Prefer one form call over many `type` calls. |\n")
	sb.WriteString("| Extract list/table data | focused `page-structure` or read-only `evaluate` -> optional targeted `extract` | Extract only the requested container/count. |\n")
	sb.WriteString("| Click a known button/link | `clickable-elements` or `snapshot` -> `act` with `return_observe:true` | Use RefIDs like `@e1` when available. |\n")
	sb.WriteString("| Diagnose one element | `element-info` -> retry with fresh identifier or coordinate fallback | Use this when hidden, disabled, overlapped, or stale. |\n")
	sb.WriteString("| Close task tabs | `tabs` list -> `tabs` close only business tabs | Never close URLs containing `#/browser-agent`. |\n\n")
	sb.WriteString("Avoid slow paths: do not start with full `page-content`, large `page-text`, full-page screenshots, repeated single keystrokes, or repeated preflight checks unless the task truly needs them.\n\n")
}

func appendHighFrequencyParameterReference(sb *strings.Builder) {
	sb.WriteString("## High-Frequency Parameter Reference\n\n")
	sb.WriteString("Use these field meanings when building JSON payloads. Keep JSON field names exactly as shown.\n\n")
	appendParameterReference(sb, "navigate", []shortParameterDoc{
		{"url", "string", "yes", "Target page URL. Use exact user-provided URL or stable entry page only.", "https://example.com"},
		{"wait_until", "string", "no", "`load`, `dom-stable`, `request-idle`, `page-stable`, or `none`. Prefer `load` or `dom-stable`.", "load"},
		{"timeout", "number", "no", "Maximum wait seconds.", "15"},
		{"return_observe", "boolean", "no", "Return page observation after navigation.", "true"},
	})
	appendParameterReference(sb, "observe", []shortParameterDoc{
		{"include_text", "boolean", "no", "Whether to include visible page text. Keep false for control discovery.", "false"},
		{"text_limit", "number", "no", "Maximum text characters when text is included.", "4000"},
	})
	appendParameterReference(sb, "input-elements / clickable-elements", []shortParameterDoc{
		{"limit", "number", "no", "Maximum controls returned. Keep modest for speed.", "30"},
	})
	appendParameterReference(sb, "page-structure", []shortParameterDoc{
		{"include_links", "boolean", "no", "Include links.", "true"},
		{"include_forms", "boolean", "no", "Include forms and fields.", "true"},
		{"include_tables", "boolean", "no", "Include tables.", "true"},
		{"include_buttons", "boolean", "no", "Include buttons.", "true"},
		{"include_images", "boolean", "no", "Include images. Usually false unless image data is needed.", "false"},
		{"limit", "number", "no", "Maximum structured items returned.", "30"},
	})
	appendParameterReference(sb, "act", []shortParameterDoc{
		{"intent", "string", "yes", "High-level action such as `click`, `type`, `select`, `check`, `navigate`, or `scroll`.", "click"},
		{"identifier", "string", "conditional", "RefID, selector, XPath, or visible text target. Required for element actions.", "@e1"},
		{"text", "string", "conditional", "Text for typing or prompt-like actions.", "ai智能体"},
		{"value", "string", "conditional", "Option value for select-like actions.", "latest"},
		{"return_observe", "boolean", "no", "Return updated page state after the action.", "true"},
	})
	appendParameterReference(sb, "click / type / fill-form", []shortParameterDoc{
		{"identifier", "string", "conditional", "RefID, selector, XPath, or visible text. Required for single element click/type.", "@e2"},
		{"text", "string", "conditional", "Text to type.", "hello"},
		{"clear", "boolean", "no", "Clear existing input before typing.", "true"},
		{"fields", "array", "conditional", "For `fill-form`: list of fields with `name` and `value`.", "[{\"name\":\"keyword\",\"value\":\"ai智能体\"}]"},
		{"submit", "boolean", "no", "Submit the form after filling.", "true"},
		{"timeout", "number", "no", "Maximum wait seconds.", "10"},
	})
	appendParameterReference(sb, "wait", []shortParameterDoc{
		{"state", "string", "yes", "`load`, `visible`, `hidden`, `enabled`, `interactable`, `writable`, `stable`, `dom-stable`, `request-idle`, `elements-more-than`, or `time`.", "dom-stable"},
		{"identifier", "string", "conditional", "Target element for element states.", "@e1"},
		{"timeout", "number", "no", "Maximum wait seconds.", "10"},
		{"count", "number", "conditional", "Expected minimum count for `elements-more-than`.", "5"},
	})
	appendParameterReference(sb, "evaluate", []shortParameterDoc{
		{"script", "string", "yes", "JavaScript to run. Prefer read-only extraction.", "return document.title"},
	})
	appendParameterReference(sb, "tabs", []shortParameterDoc{
		{"action", "string", "yes", "`list`, `new`, `switch`, `close`, or supported tab action.", "list"},
		{"url", "string", "conditional", "URL for creating a new tab.", "https://example.com"},
		{"index", "number", "conditional", "Tab index returned by `tabs` list for switch/close.", "1"},
	})
	appendParameterReference(sb, "batch", []shortParameterDoc{
		{"operations", "array", "yes", "Sequential operations. Each item has `type`, `params`, and optional `stop_on_error`.", "[{\"type\":\"click\",\"params\":{\"identifier\":\"@e1\"},\"stop_on_error\":true}]"},
		{"type", "string", "yes", "Operation command name inside one batch item.", "click"},
		{"params", "object", "yes", "Payload for that command.", "{\"identifier\":\"@e1\"}"},
		{"stop_on_error", "boolean", "no", "Stop batch when this operation fails.", "true"},
	})
}

func appendAllEndpointParameterReference(sb *strings.Builder) {
	sb.WriteString("## All Endpoint Parameter Reference\n\n")
	sb.WriteString("This section is the complete parameter reference for Browser Executor APIs. Use the exact JSON field names below. When a field has allowed values, choose one of the listed values instead of inventing a synonym.\n\n")
	appendSharedParameterConcepts(sb)

	appendEndpointReference(sb, "status", "GET `/status`", "No parameters.", nil)
	appendEndpointReference(sb, "help", "GET/POST `/help`", "Use this to inspect command help when unsure.", []parameterDoc{
		{"command", "string", "no", "Optional command name filter. Empty returns all command help.", "Command names from `help`, such as `navigate`, `act`, `tabs`, `storage`.", "act"},
	})
	appendEndpointReference(sb, "export/skill", "GET `/export/skill`", "No parameters. Downloads or returns the current generated SKILL.md content.", nil)
	appendEndpointReference(sb, "navigate", "POST `/navigate`", "Opens an exact URL or stable entry page. Do not encode user search/filter/page requirements into the URL.", []parameterDoc{
		{"url", "string", "yes", "Target URL. If protocol is missing, backend prefixes `https://`.", "Any absolute URL or domain-like string.", "https://example.com"},
		{"wait_until", "string", "no", "Navigation wait strategy after opening the URL. Default is `load`.", "`load`: wait for page load event; `dom-stable`: wait until DOM stops changing; `request-idle`: wait until network is idle; `page-stable`: wait until page becomes stable; `none`: return immediately.", "load"},
		{"timeout", "number", "no", "Maximum wait time in seconds. Default is 60 when omitted or <= 0.", "Positive integer seconds.", "30"},
		{"return_observe", "boolean", "no", "Append an `observe` result after navigation. Useful when the next step needs current page state.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text in the appended observe result.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true. Default is 8000 in appended observe.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "snapshot", "GET/POST `/snapshot`", "No parameters. Returns accessibility text and RefIDs such as `@e1`.", nil)
	appendEndpointReference(sb, "clickable-elements", "GET/POST `/clickable-elements`", "Returns compact clickable controls from the current snapshot.", []parameterDoc{
		{"limit", "number", "no", "Maximum number of clickable elements to return. Default is 50 in batch usage.", "Positive integer item count.", "50"},
	})
	appendEndpointReference(sb, "input-elements", "GET/POST `/input-elements`", "Returns compact input controls from the current snapshot.", []parameterDoc{
		{"limit", "number", "no", "Maximum number of input elements to return. Default is 30 in batch usage.", "Positive integer item count.", "30"},
	})
	appendEndpointReference(sb, "click", "POST `/click`", "Clicks one element.", []parameterDoc{
		{"identifier", "string", "yes", "Target element identifier.", "RefID such as `@e1`; CSS selector; XPath; or visible text. Prefer fresh RefIDs.", "@e1"},
		{"return_observe", "boolean", "no", "Append current page observation after click.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "type", "POST `/type`", "Types text into one element.", []parameterDoc{
		{"identifier", "string", "yes", "Input-like target element identifier.", "RefID, CSS selector, XPath, or visible text. Prefer `input-elements` or `snapshot` RefIDs.", "@e2"},
		{"text", "string", "no", "Text to enter. Empty string is allowed when clearing an input.", "Any user text.", "hello"},
		{"clear", "boolean", "no", "Clear existing value before typing. Default is true in batch usage.", "`true` or `false`.", "true"},
		{"return_observe", "boolean", "no", "Append current page observation after typing.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "select", "POST `/select`", "Selects an option in a dropdown/select-like element.", []parameterDoc{
		{"identifier", "string", "yes", "Target select element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e3"},
		{"value", "string", "yes", "Option visible text, value, or regex-like text accepted by Rod select matching.", "Visible label or option value.", "China"},
		{"return_observe", "boolean", "no", "Append current page observation after selecting.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "press-key", "POST `/press-key`", "Sends one key or shortcut to the active page.", []parameterDoc{
		{"key", "string", "yes", "Key to press.", "`Enter`/`Return`, `Tab`, `Escape`/`Esc`, `Backspace`, `Delete`, `Up`/`ArrowUp`, `Down`/`ArrowDown`, `Left`/`ArrowLeft`, `Right`/`ArrowRight`, `Home`, `End`, `PageUp`, `PageDown`, `Space`, or a single character.", "Enter"},
		{"ctrl", "boolean", "no", "Hold Ctrl while pressing key.", "`true` or `false`.", "false"},
		{"shift", "boolean", "no", "Hold Shift while pressing key.", "`true` or `false`.", "false"},
		{"alt", "boolean", "no", "Hold Alt while pressing key.", "`true` or `false`.", "false"},
		{"meta", "boolean", "no", "Hold Meta/Command/Windows key while pressing key.", "`true` or `false`.", "false"},
	})
	appendEndpointReference(sb, "wait", "POST `/wait`", "Waits for page, element, count, or time state.", []parameterDoc{
		{"identifier", "string", "conditional", "Target element or selector. Required for element states and `elements-more-than`; omit for page wait states.", "RefID/CSS/XPath/text for element states; CSS selector is best for `elements-more-than`.", ".result-item"},
		{"state", "string", "no", "Wait state. Default is `load` when omitted.", "`load`: page load; `visible`: element visible; `hidden`: element invisible; `enabled`: element enabled; `interactable`: element can receive input; `writable`: input can be written; `stable`: element stable, or page stable if no identifier; `dom-stable`: DOM stable; `request-idle`: network idle; `page-stable`: page stable; `elements-more-than`: selector count above `count`; `time`: sleep for `timeout` seconds.", "dom-stable"},
		{"timeout", "number", "no", "Maximum wait seconds. Default is 10 when omitted or <= 0.", "Positive integer seconds.", "10"},
		{"count", "number", "conditional", "Minimum element count for `elements-more-than`.", "Positive integer count.", "5"},
	})
	appendEndpointReference(sb, "reload", "POST `/reload`", "No parameters. Reloads the current business page.", nil)
	appendEndpointReference(sb, "go-back", "POST `/go-back`", "No parameters. Navigates back in browser history.", nil)
	appendEndpointReference(sb, "go-forward", "POST `/go-forward`", "No parameters. Navigates forward in browser history.", nil)
	appendEndpointReference(sb, "hover", "POST `/hover`", "Moves the cursor over one element.", []parameterDoc{
		{"identifier", "string", "yes", "Target element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e1"},
	})
	appendEndpointReference(sb, "resize", "POST `/resize`", "Changes viewport size for the active page.", []parameterDoc{
		{"width", "number", "no", "Viewport width in CSS pixels. Default is 1440 when omitted or <= 0.", "Positive integer pixels.", "1440"},
		{"height", "number", "no", "Viewport height in CSS pixels. Default is 900 when omitted or <= 0.", "Positive integer pixels.", "900"},
	})
	appendEndpointReference(sb, "page-info", "GET/POST `/page-info`", "No parameters. Returns URL, title, target id, ready state, viewport, scroll, and page counts.", nil)
	appendEndpointReference(sb, "observe", "GET/POST `/observe`", "Returns compact page state: status, page info, snapshot text, RefIDs, and optional visible text.", []parameterDoc{
		{"include_text", "boolean", "no", "Include visible page text. Keep false unless the current step needs broad page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum visible text characters when `include_text` is true. Default is 8000 when omitted or <= 0.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "get-text", "POST `/get-text`", "Reads visible text/textContent from one element.", []parameterDoc{
		{"identifier", "string", "yes", "Target element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e1"},
	})
	appendEndpointReference(sb, "get-value", "POST `/get-value`", "Reads value, value attribute, or textContent from one element.", []parameterDoc{
		{"identifier", "string", "yes", "Target element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e2"},
	})
	appendEndpointReference(sb, "element-info", "POST `/element-info`", "Returns diagnostics for one element.", []parameterDoc{
		{"identifier", "string", "yes", "Target element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e1"},
		{"attributes", "array<string>", "no", "Attribute names to read. If empty, defaults to common diagnostic attributes.", "Any HTML attribute name. Default set: `id`, `name`, `class`, `type`, `role`, `aria-label`, `placeholder`, `href`, `src`, `title`, `alt`.", "[\"id\",\"class\",\"href\",\"aria-label\"]"},
	})
	appendEndpointReference(sb, "page-text", "GET/POST `/page-text`", "Returns visible text from the whole current page.", []parameterDoc{
		{"limit", "number", "no", "Maximum visible text characters. Default is 8000 when omitted or <= 0.", "Positive integer characters. Increase gradually only when needed.", "8000"},
	})
	appendEndpointReference(sb, "page-content", "GET/POST `/page-content`", "Returns current page HTML. Use only when narrower APIs cannot answer.", []parameterDoc{
		{"limit", "number", "no", "Maximum HTML characters. If omitted or <= 0, backend defaults to 8000 through internal limiter.", "Positive integer characters.", "12000"},
	})
	appendEndpointReference(sb, "page-structure", "GET/POST `/page-structure`", "Returns compact structured page data by category.", []parameterDoc{
		{"include_links", "boolean", "no", "Include visible links.", "`true` or `false`. If all include flags are false, backend includes all categories.", "true"},
		{"include_forms", "boolean", "no", "Include forms and fields.", "`true` or `false`.", "true"},
		{"include_tables", "boolean", "no", "Include table text.", "`true` or `false`.", "true"},
		{"include_images", "boolean", "no", "Include image src/alt/title data. Usually false unless image info is needed.", "`true` or `false`.", "false"},
		{"include_buttons", "boolean", "no", "Include buttons and button-like controls.", "`true` or `false`.", "true"},
		{"limit", "number", "no", "Maximum items per included category. Default is 50 when omitted or <= 0.", "Positive integer item count.", "30"},
	})
	appendEndpointReference(sb, "extract", "POST `/extract`", "Extracts selected fields from elements matched by a CSS selector.", []parameterDoc{
		{"selector", "string", "no", "CSS selector to extract. If empty, returns full body visible text instead of structured items.", "CSS selector.", ".result-item"},
		{"fields", "array<string>", "no", "Fields to keep from each selected element. Default is `text`, `href`, `value`.", "`text`: innerText/textContent; `href`: link href; `value`: input value; `html`: outerHTML.", "[\"text\",\"href\"]"},
		{"multiple", "boolean", "no", "Whether to return all matched elements. If false, only the first matched item is returned.", "`true` or `false`.", "true"},
	})
	appendEndpointReference(sb, "screenshot", "POST `/screenshot`", "Captures a page screenshot as base64 image data.", []parameterDoc{
		{"full_page", "boolean", "no", "Capture the full scrollable page instead of the viewport only.", "`true` or `false`.", "true"},
		{"format", "string", "no", "Image format. Default is `png` unless `jpeg` is provided.", "`png` or `jpeg`.", "png"},
		{"quality", "number", "no", "JPEG quality. Ignored for PNG. Used only when > 0.", "1-100 is typical for JPEG.", "80"},
	})
	appendEndpointReference(sb, "element-screenshot", "POST `/element-screenshot`", "Captures one element as base64 image data.", []parameterDoc{
		{"identifier", "string", "yes", "Target element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e1"},
		{"format", "string", "no", "Image format. Default is `png` unless `jpeg` is provided.", "`png` or `jpeg`.", "png"},
		{"quality", "number", "no", "JPEG quality. Ignored for PNG. Used only when > 0.", "1-100 is typical for JPEG.", "80"},
	})
	appendEndpointReference(sb, "evaluate", "POST `/evaluate`", "Runs JavaScript in the current page. Prefer read-only extraction and semantic APIs first.", []parameterDoc{
		{"script", "string", "yes", "JavaScript source. Plain bodies are auto-wrapped as `() => { ... }`; function expressions and async functions are accepted as-is.", "Normal JavaScript. Return serializable values.", "return document.title"},
	})
	appendEndpointReference(sb, "cookies", "POST `/cookies`", "Lists, sets, deletes, or clears browser cookies.", []parameterDoc{
		{"action", "string", "no", "Cookie operation. Default is `list`.", "`list`/`get`: list cookies; `set`: create/update cookie; `delete`/`remove`: delete one cookie; `clear`: clear cookies.", "list"},
		{"name", "string", "conditional", "Cookie name. Required for `set`, `delete`, and `remove`.", "Cookie key string.", "token"},
		{"value", "string", "conditional", "Cookie value. Used by `set`.", "Cookie value string.", "abc"},
		{"url", "string", "conditional", "Cookie URL scope. If omitted with no `domain`, backend uses current page URL.", "URL such as `https://example.com`.", "https://example.com"},
		{"domain", "string", "conditional", "Cookie domain scope. Use either URL or domain according to browser cookie rules.", "Domain such as `example.com`.", "example.com"},
		{"path", "string", "no", "Cookie path scope.", "Path string.", "/"},
		{"secure", "boolean", "no", "Set cookie Secure flag.", "`true` or `false`.", "true"},
		{"http_only", "boolean", "no", "Set cookie HttpOnly flag.", "`true` or `false`.", "false"},
		{"same_site", "string", "no", "Cookie SameSite value.", "`Strict`, `Lax`, or `None`. Empty leaves it unset.", "Lax"},
		{"expires", "number", "no", "Cookie expiration timestamp. 0 means session cookie.", "Unix epoch seconds as number.", "0"},
	})
	appendEndpointReference(sb, "storage", "POST `/storage`", "Manages localStorage or sessionStorage on the current page.", []parameterDoc{
		{"action", "string", "no", "Storage operation. Default is `list`.", "`list`: list all keys; `get`: get one key; `set`: set one key; `delete`/`remove`: remove one key; `clear`: clear all keys in selected storage.", "list"},
		{"type", "string", "no", "Storage area. Default is `local`.", "`local`: localStorage; `session`: sessionStorage.", "local"},
		{"key", "string", "conditional", "Storage key. Required for `get`, `set`, `delete`, and `remove`.", "Any storage key string.", "token"},
		{"value", "string", "conditional", "Storage value. Used by `set`.", "String value. JSON should be stringified before sending.", "{\"id\":1}"},
	})
	appendEndpointReference(sb, "tabs", "POST `/tabs`", "Manages browser tabs. Never close or switch to a tab whose URL contains `#/browser-agent`.", []parameterDoc{
		{"action", "string", "yes", "Tab operation.", "`list`: list tabs; `new`: create and activate tab; `switch`: activate tab by index; `close`: close tab by index.", "list"},
		{"url", "string", "conditional", "URL for `new`. Defaults to `about:blank` if empty.", "Absolute URL or `about:blank`.", "https://example.com"},
		{"index", "number", "conditional", "Tab index returned by `tabs` list. Required for `switch` and `close`.", "Zero-based tab index from API response.", "1"},
	})
	appendEndpointReference(sb, "scroll", "POST `/scroll`", "Scrolls the page or a target element.", []parameterDoc{
		{"direction", "string", "no", "Scroll direction. Default is `down`.", "`down`: scroll down by pixels; `up`: scroll up by pixels; `top`: scroll to top; `bottom`: scroll to bottom.", "down"},
		{"pixels", "number", "no", "Scroll distance in pixels for `down`/`up`. Default is 700 when omitted or <= 0.", "Positive integer pixels.", "700"},
		{"identifier", "string", "no", "Optional target scroll container or element. If empty, scrolls window.", "RefID, CSS selector, XPath, or visible text.", "@e1"},
		{"return_observe", "boolean", "no", "Append current page observation after scrolling.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "mouse", "POST `/mouse`", "Runs coordinate-based mouse operations. Prefer semantic element APIs first.", []parameterDoc{
		{"action", "string", "yes", "Mouse operation.", "`move`/`move-to`: move pointer; `click`: click; `double-click`/`dblclick`: double click; `right-click`/`context-menu`: right click; `down`: press mouse button; `up`: release mouse button; `scroll`/`wheel`: mouse wheel scroll.", "click"},
		{"x", "number", "conditional", "Viewport X coordinate. Used by move/click/down/up/scroll.", "CSS pixel coordinate, often from `element-info.data.box`.", "300"},
		{"y", "number", "conditional", "Viewport Y coordinate. Used by move/click/down/up/scroll.", "CSS pixel coordinate, often from `element-info.data.box`.", "200"},
		{"delta_x", "number", "conditional", "Horizontal wheel delta. Used by `scroll`/`wheel`.", "Number of pixels or wheel delta units.", "0"},
		{"delta_y", "number", "conditional", "Vertical wheel delta. Used by `scroll`/`wheel`.", "Number of pixels or wheel delta units.", "700"},
		{"steps", "number", "no", "Movement or wheel steps. Default is 10 when omitted or <= 0.", "Positive integer.", "10"},
		{"button", "string", "no", "Mouse button. Default is `left`.", "`left`, `right`, or `middle`.", "left"},
	})
	appendEndpointReference(sb, "window", "POST `/window`", "Reads or changes browser window bounds/state.", []parameterDoc{
		{"action", "string", "yes", "Window operation.", "`info`/`get`: read bounds; `set`/`resize`/`move`: set bounds; `maximize`/`maximized`: maximize; `minimize`/`minimized`: minimize; `fullscreen`: fullscreen; `normal`/`restore`: restore normal state.", "info"},
		{"left", "number", "conditional", "Window left position for `set`/`move`. 0 means leave unchanged in current implementation.", "Screen pixel coordinate.", "100"},
		{"top", "number", "conditional", "Window top position for `set`/`move`. 0 means leave unchanged in current implementation.", "Screen pixel coordinate.", "100"},
		{"width", "number", "conditional", "Window width for `set`/`resize`.", "Positive screen pixel width.", "1280"},
		{"height", "number", "conditional", "Window height for `set`/`resize`.", "Positive screen pixel height.", "900"},
	})
	appendEndpointReference(sb, "close-page", "POST `/close-page`", "No parameters. Closes the active business page only. Backend refuses to close BrowserFlow `#/browser-agent` tabs.", nil)
	appendEndpointReference(sb, "fill-form", "POST `/fill-form`", "Fills several form fields in one call. It matches fields by name, id, placeholder, aria-label, associated label text, or identifier fallback.", []parameterDoc{
		{"fields", "array<object>", "yes", "Form fields to fill. Each item has `name`, `value`, and optional `type`.", "`name`: field name/id/placeholder/label text; `value`: string/number/boolean; `type`: optional hint for model readability, backend auto-detects actual DOM field type.", "[{\"name\":\"keyword\",\"value\":\"ai智能体\",\"type\":\"text\"}]"},
		{"fields[].name", "string", "yes", "Field locator inside form filling.", "Input `name`, `id`, placeholder text, aria-label, visible label text, RefID, CSS selector, XPath, or visible text fallback.", "keyword"},
		{"fields[].value", "any", "no", "Value to set. Checkboxes/radios use truthy values.", "Strings, numbers, booleans. Truthy checkbox/radio values include `true`, `1`, `yes`, `on`, `checked`.", "ai智能体"},
		{"fields[].type", "string", "no", "Optional type hint for the LLM and exported skill. Current backend auto-detects DOM field type and does not require this hint.", "Common hints: `text`, `textarea`, `select`, `checkbox`, `radio`, `number`, `password`, `email`, `search`, `date`, `json`, `boolean`.", "text"},
		{"submit", "boolean", "no", "Submit the form after filling. Backend tries submit button, common submit labels, then Enter.", "`true` or `false`.", "true"},
		{"timeout", "number", "no", "Per-field match/fill timeout seconds. Default is 10 when omitted or <= 0.", "Positive integer seconds.", "10"},
		{"return_observe", "boolean", "no", "Append current page observation after filling.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "drag", "POST `/drag`", "Drags one element to another element.", []parameterDoc{
		{"from_identifier", "string", "yes", "Source element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e1"},
		{"to_identifier", "string", "yes", "Target element identifier.", "RefID, CSS selector, XPath, or visible text.", "@e2"},
	})
	appendEndpointReference(sb, "file-upload", "POST `/file-upload`", "Sets local file paths on a file input element.", []parameterDoc{
		{"identifier", "string", "yes", "File input element identifier.", "RefID, CSS selector, XPath, or visible text. Prefer an actual `<input type=\"file\">`.", "@e4"},
		{"file_paths", "array<string>", "yes", "Local file paths to upload.", "Windows absolute paths such as `C:\\path\\file.png`; paths must exist on the machine running the browser.", "[\"C:\\\\path\\\\file.png\"]"},
	})
	appendEndpointReference(sb, "handle-dialog", "POST `/handle-dialog`", "Arms a handler for the next JavaScript alert/confirm/prompt/beforeunload dialog. Call this before the action that opens the dialog.", []parameterDoc{
		{"accept", "boolean", "no", "Whether to accept the dialog. False dismisses/cancels.", "`true` or `false`.", "true"},
		{"text", "string", "no", "Prompt input text when handling a prompt dialog.", "Any prompt text.", "ok"},
		{"timeout", "number", "no", "How long the handler waits for the dialog in seconds. Default is 10 when omitted or <= 0.", "Positive integer seconds.", "10"},
	})
	appendEndpointReference(sb, "act", "POST `/act`", "Runs one high-level action by intent. Prefer this for simple known actions after inspecting the page.", []parameterDoc{
		{"intent", "string", "yes", "High-level action to run.", "`navigate`/`open`/`goto`: open URL from `text` or `value`; `click`/`press`: click `identifier`; `type`/`input`/`fill`: type into `identifier`; `select`/`choose`: choose option; `check`: check checkbox/radio; `uncheck`: uncheck checkbox; `fill-form`/`form`: fill multiple fields; `press-key`/`key`: press key from `text` or `value`; `scroll`: scroll using `text`/`value` as direction; `hover`: hover target.", "click"},
		{"identifier", "string", "conditional", "Target element for element actions.", "RefID, CSS selector, XPath, or visible text. Required for click/type/select/check/uncheck/hover and optional for scroll.", "@e1"},
		{"value", "any", "conditional", "Generic action value. Used as URL for navigate, text for type, option for select, key for press-key, direction for scroll.", "String/number/boolean depending on intent. `text` takes priority for text-like intents.", "ai智能体"},
		{"text", "string", "conditional", "Text action value. Used for type, navigate URL, select value, key, or scroll direction.", "Any string.", "ai智能体"},
		{"wait_until", "string", "no", "Navigation wait strategy for navigate/open/goto intents.", "`load`, `dom-stable`, `request-idle`, `page-stable`, or `none`.", "load"},
		{"fields", "array<object>", "conditional", "Fields for `fill-form`/`form` intent.", "Same structure as `fill-form.fields`.", "[{\"name\":\"keyword\",\"value\":\"ai智能体\",\"type\":\"text\"}]"},
		{"submit", "boolean", "no", "Submit after `fill-form`/`form` intent.", "`true` or `false`.", "true"},
		{"clear", "boolean", "no", "Clear before type/input/fill intent. Default is true in batch usage.", "`true` or `false`.", "true"},
		{"timeout", "number", "no", "Timeout seconds for navigate/form operations. Default is 10 for act form/lookup operations when omitted or <= 0.", "Positive integer seconds.", "10"},
		{"return_observe", "boolean", "no", "Append current page observation after the action.", "`true` or `false`.", "true"},
		{"include_text", "boolean", "no", "When `return_observe` is true, include visible page text.", "`true` or `false`.", "false"},
		{"text_limit", "number", "no", "Maximum page text characters when `include_text` is true.", "Positive integer characters.", "4000"},
	})
	appendEndpointReference(sb, "batch", "POST `/batch`", "Executes deterministic operations sequentially. Do not batch across unknown decision points that require fresh observation.", []parameterDoc{
		{"operations", "array<object>", "yes", "Ordered operations to run. Each item has `type`, `params`, and optional `stop_on_error`.", "Array of batch operation objects.", "[{\"type\":\"click\",\"params\":{\"identifier\":\"@e1\"},\"stop_on_error\":true}]"},
		{"operations[].type", "string", "yes", "Command name to run inside batch.", "Supported: `navigate`, `click`, `type`, `select`, `press-key`, `wait`, `snapshot`, `clickable-elements`, `input-elements`, `observe`, `page-text`, `page-content`, `get-text`, `get-value`, `element-info`, `extract`, `screenshot`, `evaluate`, `cookies`, `storage`, `page-structure`, `scroll`, `reload`, `go-back`, `go-forward`, `hover`, `resize`, `element-screenshot`, `mouse`, `window`, `close-page`, `fill-form`, `drag`, `file-upload`, `handle-dialog`, `tabs`, `act`.", "navigate"},
		{"operations[].params", "object", "yes", "Parameters for the selected operation type. Use the same field names described in this reference.", "Object matching the endpoint payload.", "{\"url\":\"https://example.com\",\"wait_until\":\"load\"}"},
		{"operations[].stop_on_error", "boolean", "no", "Stop executing later batch operations if this operation returns an error.", "`true` or `false`.", "true"},
	})
}

func appendSharedParameterConcepts(sb *strings.Builder) {
	sb.WriteString("### Shared Variables And Conventions\n\n")
	sb.WriteString("| Variable / concept | Available values | Meaning |\n")
	sb.WriteString("|---|---|---|\n")
	sb.WriteString("| `identifier` | RefID like `@e1`; CSS selector like `.btn`; XPath like `//button[text()='OK']`; visible text such as `Search` | Identifies a page element. Prefer fresh RefIDs from `snapshot`, `observe`, `clickable-elements`, or `input-elements`. RefIDs may become stale after navigation or DOM changes. |\n")
	sb.WriteString("| `wait_until` | `load`, `dom-stable`, `request-idle`, `page-stable`, `none` | Navigation wait strategy. Use `load` for normal pages, `dom-stable` for DOM-rendered pages, `request-idle` for SPAs/network-heavy pages, `page-stable` for visual/layout stability, and `none` only when the next command waits explicitly. |\n")
	sb.WriteString("| `return_observe` | `true`, `false` | When supported, appends `observe` output to the operation result so the model can verify the page state in the same call. |\n")
	sb.WriteString("| `include_text` | `true`, `false` | Controls whether appended observe includes visible page text. Keep false for speed unless text is needed. |\n")
	sb.WriteString("| `text_limit` / `limit` | Positive integer | Caps returned text or item count. Start small and increase only when necessary. |\n")
	sb.WriteString("| `fields[].type` | `text`, `textarea`, `select`, `checkbox`, `radio`, `number`, `password`, `email`, `search`, `date`, `json`, `boolean` | Optional form-field hint for the model/exported docs. Backend auto-detects actual DOM type, so this is mainly descriptive. |\n")
	sb.WriteString("| `result` response | `data.result.success`, `data.result.error`, `data.result.data` | GoFrame wraps responses as `{code,message,data}`. Browser operation result is usually under `data.result`; inspect success/error/data before reporting completion. |\n\n")
}

func appendEndpointReference(sb *strings.Builder, command string, route string, note string, docs []parameterDoc) {
	sb.WriteString("### `" + command + "`\n\n")
	sb.WriteString("- Route: " + route + "\n")
	if note != "" {
		sb.WriteString("- Notes: " + note + "\n")
	}
	if len(docs) == 0 {
		sb.WriteString("- Parameters: none.\n\n")
		return
	}
	sb.WriteString("\n")
	appendDetailedParameterTable(sb, docs)
}

type parameterDoc struct {
	Name      string
	Type      string
	Required  string
	Meaning   string
	Available string
	Example   string
}

type shortParameterDoc struct {
	Name     string
	Type     string
	Required string
	Meaning  string
	Example  string
}

func appendParameterReference(sb *strings.Builder, command string, docs []shortParameterDoc) {
	sb.WriteString("### `" + command + "` parameters\n\n")
	sb.WriteString("| Field | Type | Required | Meaning | Example |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, doc := range docs {
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | `%s` |\n",
			doc.Name,
			doc.Type,
			doc.Required,
			doc.Meaning,
			strings.ReplaceAll(doc.Example, "|", "\\|"),
		))
	}
	sb.WriteString("\n")
}

func appendDetailedParameterTable(sb *strings.Builder, docs []parameterDoc) {
	sb.WriteString("| Field | Type | Required | Meaning | Available values / variables | Example |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")
	for _, doc := range docs {
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | %s | `%s` |\n",
			doc.Name,
			doc.Type,
			doc.Required,
			escapeTableCell(doc.Meaning),
			escapeTableCell(doc.Available),
			escapeTableCell(doc.Example),
		))
	}
	sb.WriteString("\n")
}

func escapeTableCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
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
