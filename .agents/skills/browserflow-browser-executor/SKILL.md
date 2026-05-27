---
name: browserflow-browser-executor-executor
description: Control the current BrowserFlow Windows browser directly through HTTP APIs. Use observe, act, accessibility snapshots, RefIDs, page structure, element diagnostics, page text, page HTML, cookies, storage, and compact actions to navigate, click, type, fill forms, upload files, drag, handle dialogs, scroll, reload, extract data, screenshot, operate mouse/window, run JavaScript, and manage tabs without Automa workflows.
---

# BrowserFlow Browser Executor

## Overview

Use this skill when the user wants to operate the browser directly with an LLM instead of running a prebuilt Automa workflow.

Keep the BrowserFlow `browser-agent` client tab alive. Do not close the whole browser after a task; when cleanup is needed, close only task-related business tabs that were opened or used for that task.

**API Base URL:** `http://127.0.0.1:8001/api/v1/browser-executor`

## Quick Start

Use this Skill with a compact observe-act-verify loop:

1. Check runtime and browser status when needed.
2. Inspect only the controls or page area needed for the next action.
3. Group deterministic small actions that share one user intent.
4. Verify the page state after each action group.
5. Extract only the requested data, then close only task-related business tabs.

## Skill Map

- `Mandatory Workflow`: required safety and browser-control rules.
- `Targeted Inspection Rules`: how to avoid fetching the whole page by default.
- `Action Grouping Rules`: how to combine small deterministic operations.
- `Search, Filter, And Pagination Rules`: how to operate visible page controls reliably.
- `Core Commands`: HTTP examples for each browser executor API.
- `Troubleshooting`: recovery rules for stale elements, slow pages, and failed operations.

## Mandatory Workflow

1. Use lightweight preflight: check `status` before the first browser operation in a conversation, before long or sensitive tasks, and after any executor failure. Reuse a recent successful status check for short follow-up operations on the same browser.
2. Before running a task, inspect the current page and create a concrete step plan. Do not start clicking, typing, navigating, or evaluating JavaScript before the plan exists.
3. If only a `#/browser-agent` tab exists, use `tabs` with `action:"new"` or call `navigate`; the backend will create a business tab instead of navigating the agent tab.
4. Inspect only the page area or element set needed for the current step. Prefer `input-elements`, `clickable-elements`, `page-structure`, `element-info`, or low-limit `observe` before broad page text or HTML.
5. After navigation or a page-changing action, call `observe` or `snapshot` again.
6. Use RefIDs such as `@e1` from `snapshot` for click, type, get-text, and get-value.
7. Prefer `act` for simple intent-driven actions and set `return_observe:true` after page-changing actions.
8. Prefer `fill-form` for multiple fields and `batch` for deterministic sequential actions; do not batch steps that need observation between them.
9. If a RefID fails, call `observe` or `snapshot` again because the page may have changed.
10. Prefer `page-structure` for structured extraction before requesting full `page-content`, and use `element-info` to diagnose one uncertain element.
11. Use `wait` states precisely: `load`, `visible`, `hidden`, `enabled`, `interactable`, `writable`, `stable`, `dom-stable`, `request-idle`, `elements-more-than`, or `time`. For navigation, set `wait_until` to `load`, `dom-stable`, `request-idle`, `page-stable`, or `none`.
12. Use `cookies` and `storage` when the task needs to inspect login/session state or clean up page state. Prefer these semantic APIs over ad hoc JavaScript.
13. Never close the BrowserFlow browser or any `#/browser-agent` client tab. Before using `close-page` or `tabs` with `action:"close"`, call `tabs` with `action:"list"` and close only task-related business tabs.

## Required Step-By-Step Procedure

Always work in this order. Do not perform page operations until the planning checks are complete.

1. Decide preflight depth: call `/app/runtime` and `/browser-executor/status` for the first operation, long tasks, sensitive tasks, or when no recent successful preflight is available. For short follow-up operations, reuse the recent status check; if an operation fails, run preflight before retrying.
2. Inspect context: call `observe` on the current page. If the current page is `#/browser-agent` or no useful business page exists, open or navigate to the required business page first, then call `observe` again.
3. Build a plan: write a short numbered plan with the intended browser actions, expected page changes, and the data or final state needed for success.
4. Execute one meaningful step at a time. After every navigation, click, submit, form fill, scroll that reveals content, or JavaScript mutation, call `observe`, `snapshot`, `page-structure`, or a targeted getter to verify the result.
5. If verification fails, stop the current action chain, inspect again, revise the plan, and continue from the verified page state. Do not blindly repeat stale RefIDs or selectors.
6. Use `batch` only for deterministic mini-sequences where no observation is needed between operations. Do not batch an entire unknown workflow.
7. Before reporting success, verify the final page state or extracted data. If the task is incomplete, report the exact blocking condition and the last verified state.

## Targeted Inspection Rules

- Do not fetch or analyze the whole page by default. Decide what the current step needs, then request the smallest useful context.
- For search boxes or form fields, use `input-elements` or `fill-form` matching first. For buttons, tabs, filters, and pagination, use `clickable-elements` or compact `observe` without large page text.
- For lists and result pages, prefer `page-structure` or a focused read-only `evaluate` after the page has been inspected. Extract only the target container or requested number of rows/items.
- Use `page-text` or `page-content` only when narrow APIs cannot answer the question, and keep limits small. Increase limits gradually only if the required target is missing.
- If a page is large, inspect in stages: relevant controls first, target result container next, then focused extraction. Avoid mixing navigation controls, footer links, and unrelated page content into one analysis step.

## Action Grouping Rules

- A meaningful step can be an action group, not a single low-level browser event. Group small operations that share one user intent when the required controls are already known and no intermediate decision is needed.
- Good action groups include: type a search keyword and submit it; open a filter panel and choose known filter values; choose known filter values and click a known page number; fill several fields in one form; close several verified task-related business tabs.
- Keep verification at the boundary of each action group. After the group finishes, verify the expected page state, active filters, current page, or extracted data once, instead of reporting after every click or keystroke.
- Do not group across an unknown decision point. If the next target depends on newly loaded content, a modal, dynamic layout, login state, or an uncertain RefID, inspect the page first and continue with fresh identifiers.
- Prefer `fill-form`, `act` with `return_observe:true`, or `batch` for deterministic grouped operations. If any grouped operation fails, stop, inspect, and recover from the last verified state.

## Planning Format

Before executing a non-trivial task, produce a compact plan like this:

```text
Plan:
1. Verify current browser/page state.
2. Navigate or select the target page.
3. Locate the required controls or data.
4. Perform the action or extraction.
5. Verify the final state and report the result.
```

For simple tasks such as reading the current title or URL, the plan can be one sentence, but you still must verify with an API response before answering.

## Lightweight Preflight Strategy

- First operation in a conversation: call `/app/runtime` and `/browser-executor/status`.
- Short follow-up operations on the same running browser: reuse a recent successful preflight to save time.
- Long, multi-step, data extraction, form submission, upload, account-changing, or destructive tasks: perform preflight unless a recent successful preflight is already available.
- If any command fails because the executor is offline, disconnected, page context is missing, or the backend is unreachable, run preflight again before retrying.
- Direct browser actions still need page verification. Reusing preflight does not replace `observe`, `snapshot`, or another page-state check after page-changing actions.

## Stepwise Page Operation Rules

- Match the operation depth to the task. For complex tasks, operate step by step: open the site, inspect the page, analyze available controls, perform one action, verify the result, then continue.
- Step by step means verified action groups, not every keystroke or click. Keep the loop efficient while preserving page-state verification.
- Complex tasks include search with filters, sorting, pagination, multi-step forms, login-sensitive pages, data extraction, pages with dynamic content, and tasks where the site behavior is unknown.
- After opening a website or changing pages, inspect the smallest useful target before deciding the next meaningful action: inputs for search/form work, clickables for controls, structure for lists, or element-info for one uncertain target.
- After every meaningful page-changing action group, verify the updated state with a focused inspection before continuing.
- Do not rely on stale RefIDs, guessed selectors, guessed page state, or JavaScript extraction before the current page has been inspected.
- For simple tasks such as opening a user-provided exact URL, reading the current title, or visiting a stable static page without search/filter criteria, a direct URL or one-step action is acceptable, but still verify the result before answering.

## Search, Filter, And Pagination Rules

- For user-visible search, sorting, filters, tabs, date ranges, categories, and pagination, use real page interactions by default, even when the requested operation looks simple.
- Once controls are identified, combine related operations when safe: for example, enter the query and submit search as one group, or apply known filters and choose the requested page as one group, then verify the final state.
- Do not satisfy search/filter/sort/pagination requirements by manually constructing URL query parameters. Open the site or search page first, inspect available controls, then interact with the visible controls step by step.
- Direct URL navigation is allowed only for opening a user-provided exact URL or a stable entry page. Do not encode requested search keywords, filters, sort orders, date ranges, categories, or page numbers into the URL yourself.
- After applying search, sorting, filters, or pagination, verify the selected labels, active tabs, current page number, result count or visible result changes with `observe`, `snapshot`, or `page-structure`.
- If the page state cannot prove the requested filters are active, use the visible controls to select them manually before extracting data.
- Do not claim a filter, sort order, date range, category, or page number was applied only because a URL parameter exists.
- If URL construction and visible controls disagree, trust the verified visible page state and adjust the page through controls.

## Safety Boundaries

Ask the user for explicit confirmation before destructive, irreversible, or externally visible actions, including submitting purchases, orders, payments, account/security changes, deleting data, sending messages or emails, uploading files, clearing cookies/storage, or changing important settings. If the user already gave clear permission for that exact action in the current request, proceed carefully and verify before submitting.

Do not use `evaluate` to bypass user confirmation, disable site protections, read unrelated secrets, or mutate sensitive page state when a semantic BrowserFlow API can do the task. Prefer high-level APIs such as `click`, `type`, `fill-form`, `cookies`, and `storage` over custom JavaScript.

## Failure Recovery

If an action fails or the page state is not what you expected, follow this recovery loop before trying again:

1. Stop the current action chain and do not repeat the same stale RefID or selector more than once.
2. Call `observe`, `snapshot`, `page-structure`, or `tabs` to discover the current state.
3. Check for navigation, slow loading, a newly opened tab, modal/dialog, disabled element, validation error, login/session problem, or changed DOM.
4. Revise the plan based on the verified state and continue with fresh RefIDs or a more reliable identifier.
5. If recovery would require a destructive action, credential entry, payment, upload, or account change, ask the user before continuing.

## Preflight

Use these commands for first calls, long/sensitive tasks, or recovery after failures. For short follow-up calls in the same conversation, a recent successful result may be reused.

```bash
curl 'http://127.0.0.1:8001/api/v1/app/runtime'
curl 'http://127.0.0.1:8001/api/v1/browser-executor/status'
```

## Core Commands

### Open URL

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/navigate' \
  -H 'Content-Type: application/json' \
  -d '{"wait_until":"load","url":"https://example.com"}'
```

### Observe Page

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/observe' \
  -H 'Content-Type: application/json' \
  -d '{"include_text":false,"text_limit":8000}'
```

### Get Snapshot

```bash
curl 'http://127.0.0.1:8001/api/v1/browser-executor/snapshot'
```

### Clickable Elements

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/clickable-elements' \
  -H 'Content-Type: application/json' \
  -d '{"limit":50}'
```

### Input Elements

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/input-elements' \
  -H 'Content-Type: application/json' \
  -d '{"limit":30}'
```

### Page Structure

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/page-structure' \
  -H 'Content-Type: application/json' \
  -d '{"include_buttons":true,"limit":30,"include_links":true,"include_forms":true,"include_tables":true,"include_images":false}'
```

### Element Info

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/element-info' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e1","attributes":["id","class","href","aria-label"]}'
```

### Smart Action

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/act' \
  -H 'Content-Type: application/json' \
  -d '{"intent":"click","identifier":"@e1","return_observe":true}'
```

### Click Element

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/click' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e1"}'
```

### Type Text

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/type' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e2","text":"hello","clear":true}'
```

### Select Option

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/select' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e3","value":"China"}'
```

### Fill Form

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/fill-form' \
  -H 'Content-Type: application/json' \
  -d '{"fields":[{"name":"email","value":"user@example.com"},{"name":"password","value":"secret"}],"submit":false,"timeout":10}'
```

### Press Key

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/press-key' \
  -H 'Content-Type: application/json' \
  -d '{"key":"Enter"}'
```

### Wait For Element

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/wait' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e1","state":"interactable","timeout":10}'
```

### Wait For DOM Stable

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/wait' \
  -H 'Content-Type: application/json' \
  -d '{"state":"dom-stable","timeout":10}'
```

### Hover Element

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/hover' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e1"}'
```

### Drag Element

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/drag' \
  -H 'Content-Type: application/json' \
  -d '{"from_identifier":"@e1","to_identifier":"@e2"}'
```

### Upload Files

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/file-upload' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e4","file_paths":["C:\\\\path\\\\file.png"]}'
```

### Arm Dialog Handler

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/handle-dialog' \
  -H 'Content-Type: application/json' \
  -d '{"text":"","timeout":10,"accept":true}'
```

### Get Value

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/get-value' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e2"}'
```

### Page Text

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/page-text' \
  -H 'Content-Type: application/json' \
  -d '{"limit":8000}'
```

### Page Content

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/page-content' \
  -H 'Content-Type: application/json' \
  -d '{"limit":12000}'
```

### List Cookies

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/cookies' \
  -H 'Content-Type: application/json' \
  -d '{"action":"list"}'
```

### Set Cookie

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/cookies' \
  -H 'Content-Type: application/json' \
  -d '{"action":"set","name":"token","value":"abc","url":"https://example.com","same_site":"Lax"}'
```

### List Local Storage

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/storage' \
  -H 'Content-Type: application/json' \
  -d '{"action":"list","type":"local"}'
```

### Set Local Storage

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/storage' \
  -H 'Content-Type: application/json' \
  -d '{"action":"set","type":"local","key":"token","value":"abc"}'
```

### Scroll Page

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/scroll' \
  -H 'Content-Type: application/json' \
  -d '{"direction":"down","pixels":700}'
```

### Reload Page

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/reload' \
  -H 'Content-Type: application/json' \
  -d '{}'
```

### Resize Viewport

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/resize' \
  -H 'Content-Type: application/json' \
  -d '{"width":1440,"height":900}'
```

### Window Info

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/window' \
  -H 'Content-Type: application/json' \
  -d '{"action":"info"}'
```

### Mouse Click

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/mouse' \
  -H 'Content-Type: application/json' \
  -d '{"action":"click","x":300,"y":200,"button":"left"}'
```

### Extract Text

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/extract' \
  -H 'Content-Type: application/json' \
  -d '{"selector":"body","fields":["text"],"multiple":false}'
```

### Screenshot

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/screenshot' \
  -H 'Content-Type: application/json' \
  -d '{"full_page":true,"format":"png"}'
```

### Element Screenshot

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/element-screenshot' \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"@e1","format":"png"}'
```

### Batch Operations

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/batch' \
  -H 'Content-Type: application/json' \
  -d '{"operations":[{"type":"navigate","params":{"url":"https://example.com"},"stop_on_error":true},{"type":"observe","params":{"include_text":false},"stop_on_error":true}]}'
```

## Element Identification

1. Prefer RefIDs such as `@e1` from `snapshot`, `observe`, `clickable-elements`, or `input-elements`.
2. Use an exact CSS selector only when the user provides one or the page structure is stable.
3. Use XPath or visible text only for obvious buttons and links.
4. If an identifier fails, refresh with `observe` or `snapshot` because RefIDs may be stale.
5. `fill-form` can match fields by name, id, placeholder, aria-label, or associated label text; use it for multi-field forms.

## Efficient Inspection

- Use the narrowest inspection API that fits the step. `observe` is useful for a compact overview, but avoid large `include_text`/`text_limit` values unless the task needs broad page text.
- Use `input-elements` for text boxes and form fields, `clickable-elements` for buttons/links/tabs/filters/pagination, and `page-structure` for compact structured data such as headings, links, forms, tables, images, and buttons.
- Prefer focused result extraction over whole-page extraction. For example, extract only the visible result list and requested item count, not the header, footer, sidebars, and unrelated links.
- Use `element-info` when a target element is ambiguous, disabled, hidden, overlapped, or needs attributes/XPath/box coordinates.
- Use `element-screenshot` or `screenshot` only when visual confirmation is needed; base64 can be large.
- Use `mouse` as a coordinate fallback after obtaining coordinates from `element-info`, screenshot inspection, or user instructions. Prefer semantic `click`/`act` first.

## Response Format

GoFrame wraps responses as `{code,message,data}`. Browser operation data is usually in `data.result`. Check `data.result.success`, `data.result.error`, and `data.result.data` before reporting success.

## Final Response Rules

When the task ends, report the outcome with verified evidence. Include what was completed, the final verified page state, and any extracted data the user requested. If the task failed or is incomplete, report the blocker, the last verified page state, and the next suggested action. Do not claim success unless a BrowserFlow API response or observed page state confirms it.

## Troubleshooting

- If unsure about a command or parameters, call `help` or `help?command=<name>` before guessing.
- If an element is not found, call `observe`, `snapshot`, `clickable-elements`, or `input-elements` again.
- If a page did not update, call `wait` with `dom-stable`, `request-idle`, or a specific element state, then `observe`.
- If a click fails, use `element-info` to check visibility, disabled state, box coordinates, and XPath; then retry with a better identifier or coordinate `mouse` fallback.
- If extraction is empty, try `page-structure`, `page-text`, `page-content`, or a broader selector with a smaller limit.
- If login state or a persisted setting looks wrong, inspect `cookies` and `storage` before retrying page actions.
- If `status.running` is false, ask the user to reopen the BrowserFlow browser-agent page.

## Other Endpoints

- `GET http://127.0.0.1:8001/api/v1/browser-executor/status` - Check whether the current browser is controllable
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/help` - Show command help
- `GET http://127.0.0.1:8001/api/v1/browser-executor/export/skill` - Export Browser Executor Skill
- `POST http://127.0.0.1:8001/api/v1/browser-executor/navigate` - Open URL and wait for load/dom-stable/request-idle/page-stable/none
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/snapshot` - Get page snapshot and RefIDs
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/clickable-elements` - Get compact clickable element RefIDs
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/input-elements` - Get compact input element RefIDs
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/observe` - Get compact page context for LLMs
- `POST http://127.0.0.1:8001/api/v1/browser-executor/act` - Run a smart action by intent
- `POST http://127.0.0.1:8001/api/v1/browser-executor/click` - Click element
- `POST http://127.0.0.1:8001/api/v1/browser-executor/type` - Type text
- `POST http://127.0.0.1:8001/api/v1/browser-executor/select` - Select dropdown option
- `POST http://127.0.0.1:8001/api/v1/browser-executor/press-key` - Send key or shortcut
- `POST http://127.0.0.1:8001/api/v1/browser-executor/wait` - Wait for page load, element states, DOM stability, network idle, or fixed time
- `POST http://127.0.0.1:8001/api/v1/browser-executor/reload` - Reload current page
- `POST http://127.0.0.1:8001/api/v1/browser-executor/go-back` - Go back in browser history
- `POST http://127.0.0.1:8001/api/v1/browser-executor/go-forward` - Go forward in browser history
- `POST http://127.0.0.1:8001/api/v1/browser-executor/hover` - Hover over element
- `POST http://127.0.0.1:8001/api/v1/browser-executor/resize` - Resize viewport
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/page-info` - Get current page URL and title
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/page-text` - Get compact visible page text
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/page-content` - Get compact page HTML
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/page-structure` - Get compact structured headings, links, forms, tables, images, and buttons
- `POST http://127.0.0.1:8001/api/v1/browser-executor/get-text` - Get element text
- `POST http://127.0.0.1:8001/api/v1/browser-executor/get-value` - Get element value
- `POST http://127.0.0.1:8001/api/v1/browser-executor/element-info` - Get element diagnostics including text, attributes, state, box, and XPath
- `POST http://127.0.0.1:8001/api/v1/browser-executor/extract` - Extract page text or selector data
- `POST http://127.0.0.1:8001/api/v1/browser-executor/screenshot` - Capture screenshot as base64
- `POST http://127.0.0.1:8001/api/v1/browser-executor/element-screenshot` - Capture a single element screenshot as base64
- `POST http://127.0.0.1:8001/api/v1/browser-executor/evaluate` - Execute JavaScript with automatic function wrapping
- `POST http://127.0.0.1:8001/api/v1/browser-executor/cookies` - List, set, delete, or clear browser cookies
- `POST http://127.0.0.1:8001/api/v1/browser-executor/storage` - List, get, set, delete, or clear localStorage/sessionStorage on the current page
- `POST http://127.0.0.1:8001/api/v1/browser-executor/tabs` - Manage tabs list/new/switch/close
- `POST http://127.0.0.1:8001/api/v1/browser-executor/scroll` - Scroll page or element
- `POST http://127.0.0.1:8001/api/v1/browser-executor/mouse` - Run coordinate mouse operations move/click/double-click/right-click/down/up/scroll
- `POST http://127.0.0.1:8001/api/v1/browser-executor/window` - Read or change browser window bounds and state
- `POST http://127.0.0.1:8001/api/v1/browser-executor/close-page` - Close current page
- `POST http://127.0.0.1:8001/api/v1/browser-executor/fill-form` - Fill multiple form fields in one call
- `POST http://127.0.0.1:8001/api/v1/browser-executor/drag` - Drag one element to another
- `POST http://127.0.0.1:8001/api/v1/browser-executor/file-upload` - Upload local files to a file input
- `POST http://127.0.0.1:8001/api/v1/browser-executor/handle-dialog` - Arm a handler for the next JavaScript dialog
- `POST http://127.0.0.1:8001/api/v1/browser-executor/batch` - Execute operations in sequence

## Notes

- This skill does not use Automa workflows or Automa trigger parameters.
- Prefer `observe` when you need multiple facts about the page in one round trip.
- Prefer `page-structure` for compact structured extraction before reading raw HTML with `page-content`.
- Prefer `element-info` when one element needs text, attributes, state, coordinates, or XPath for diagnosis.
- Prefer `act` for click/type/select/check/navigate/scroll when the intent is clear.
- Set `return_observe:true` on `act`, `navigate`, `click`, `type`, `select`, `fill-form`, or `scroll` when you need the updated page state.
- Prefer `fill-form` over repeated `type` calls when a page has several fields.
- Prefer `wait` with specific states (`interactable`, `enabled`, `writable`, `dom-stable`, `request-idle`) instead of blind time sleeps.
- For `navigate`, choose `wait_until` deliberately. Use `request-idle` for network-heavy SPAs, `dom-stable` for DOM-rendered pages, and `none` only when the next step explicitly waits for something else.
- Call `handle-dialog` before the action that triggers an alert, confirm, prompt, or beforeunload dialog.
- Use `cookies` for authentication/session inspection and `storage` for localStorage/sessionStorage reads or cleanup. Avoid raw `evaluate` for these common state tasks.
- Prefer `page-text` or `page-content` only when the model needs broad context, and keep limits small.
- Browser control APIs are powerful. Use them only against the local BrowserFlow backend.
- Do not close the browser as a task cleanup step. Keep the BrowserFlow `browser-agent` client tab open so the local executor remains connected.
- If task cleanup is requested, close only pages opened or used for the current task, and never close pages whose URL contains `#/browser-agent`.
- Prefer `snapshot` RefIDs over raw CSS selectors unless the user provides an exact selector.
- `evaluate` accepts normal JavaScript and auto-wraps it as a function when needed, so `return document.title` is valid. Prefer evaluate for read-only extraction; do not mutate page state with evaluate unless normal APIs cannot do it.
- `screenshot` returns base64 image data; summarize it unless the user asks for the raw data.
- `element-screenshot` also returns base64 image data and is cheaper than a full-page screenshot for visual checks.
