---
name: browserflow-browser-executor
description: Control the current BrowserFlow Windows browser directly through HTTP APIs. Use observe, act, accessibility snapshots, RefIDs, page structure, element diagnostics, page text, page HTML, and compact actions to navigate, click, type, fill forms, upload files, drag, handle dialogs, scroll, reload, extract data, screenshot, operate mouse/window, run JavaScript, and manage tabs without Automa workflows.
---

# BrowserFlow Browser Executor

## Overview

Use this skill when the user wants to operate the browser directly with an LLM instead of running a prebuilt Automa workflow.

Keep the BrowserFlow `browser-agent` client tab alive. Do not close the whole browser after a task; when cleanup is needed, close only task-related business tabs that were opened or used for that task.

**API Base URL:** `http://127.0.0.1:8001/api/v1/browser-executor`

**Current Browser Instance ID:** `browser_q84m3z011jcdidyac75yof0100ezgm2l`

**Current Page:** `http://localhost:5173/#/browser-agent?browser_id=browser_q84m3z011jcdidyac75yof0100ezgm2l`

## Mandatory Workflow

1. Check `status` before controlling the browser.
2. Before running a task, decide the concrete steps first, then execute them one by one and verify the page state after each meaningful step.
3. If only a `#/browser-agent` tab exists, use `tabs` with `action:"new"` or call `navigate`; the backend will create a business tab instead of navigating the agent tab.
4. Prefer `observe` first. It returns status, page info, snapshot text, and optional page text in one call. Use `clickable-elements` or `input-elements` when you only need compact RefID lists.
5. After navigation or a page-changing action, call `observe` or `snapshot` again.
6. Use RefIDs such as `@e1` from `snapshot` for click, type, get-text, and get-value.
7. Prefer `act` for simple intent-driven actions and set `return_observe:true` after page-changing actions.
8. Prefer `fill-form` for multiple fields and `batch` for deterministic sequential actions; do not batch steps that need observation between them.
9. If a RefID fails, call `observe` or `snapshot` again because the page may have changed.
10. Prefer `page-structure` for structured extraction before requesting full `page-content`, and use `element-info` to diagnose one uncertain element.
11. Use `wait` states precisely: `load`, `visible`, `hidden`, `enabled`, `interactable`, `writable`, `stable`, `dom-stable`, `request-idle`, `elements-more-than`, or `time`.
12. Never close the BrowserFlow browser or any `#/browser-agent` client tab. Before using `close-page` or `tabs` with `action:"close"`, call `tabs` with `action:"list"` and close only task-related business tabs.

## Preflight

```bash
curl 'http://127.0.0.1:8001/api/v1/app/runtime'
curl 'http://127.0.0.1:8001/api/v1/browser-executor/status'
```

## Core Commands

### Open URL

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/browser-executor/navigate' \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com"}'
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
  -d '{"include_images":false,"include_buttons":true,"limit":30,"include_links":true,"include_forms":true,"include_tables":true}'
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
  -d '{"return_observe":true,"intent":"click","identifier":"@e1"}'
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
  -d '{"accept":true,"text":"","timeout":10}'
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
  -d '{"x":300,"y":200,"button":"left","action":"click"}'
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

- Use `observe` for the normal control loop: status, page info, snapshot text, RefIDs, and optional page text in one response.
- Use `page-structure` for compact structured data such as headings, links, forms, tables, images, and buttons. Prefer it before `page-content`.
- Use `element-info` when a target element is ambiguous, disabled, hidden, overlapped, or needs attributes/XPath/box coordinates.
- Use `element-screenshot` or `screenshot` only when visual confirmation is needed; base64 can be large.
- Use `mouse` as a coordinate fallback after obtaining coordinates from `element-info`, screenshot inspection, or user instructions. Prefer semantic `click`/`act` first.

## Response Format

GoFrame wraps responses as `{code,message,data}`. Browser operation data is usually in `data.result`. Check `data.result.success`, `data.result.error`, and `data.result.data` before reporting success.

## Troubleshooting

- If unsure about a command or parameters, call `help` or `help?command=<name>` before guessing.
- If an element is not found, call `observe`, `snapshot`, `clickable-elements`, or `input-elements` again.
- If a page did not update, call `wait` with `dom-stable`, `request-idle`, or a specific element state, then `observe`.
- If a click fails, use `element-info` to check visibility, disabled state, box coordinates, and XPath; then retry with a better identifier or coordinate `mouse` fallback.
- If extraction is empty, try `page-structure`, `page-text`, `page-content`, or a broader selector with a smaller limit.
- If `status.running` is false, ask the user to reopen the BrowserFlow browser-agent page.

## Other Endpoints

- `GET http://127.0.0.1:8001/api/v1/browser-executor/status` - Check whether the current browser is controllable
- `GET/POST http://127.0.0.1:8001/api/v1/browser-executor/help` - Show command help
- `GET http://127.0.0.1:8001/api/v1/browser-executor/export/skill` - Export Browser Executor Skill
- `POST http://127.0.0.1:8001/api/v1/browser-executor/navigate` - Open URL
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
- Call `handle-dialog` before the action that triggers an alert, confirm, prompt, or beforeunload dialog.
- Prefer `page-text` or `page-content` only when the model needs broad context, and keep limits small.
- Browser control APIs are powerful. Use them only against the local BrowserFlow backend.
- Do not close the browser as a task cleanup step. Keep the BrowserFlow `browser-agent` client tab open so the local executor remains connected.
- If task cleanup is requested, close only pages opened or used for the current task, and never close pages whose URL contains `#/browser-agent`.
- Prefer `snapshot` RefIDs over raw CSS selectors unless the user provides an exact selector.
- `evaluate` accepts normal JavaScript and auto-wraps it as a function when needed, so `return document.title` is valid. Prefer evaluate for read-only extraction; do not mutate page state with evaluate unless normal APIs cannot do it.
- `screenshot` returns base64 image data; summarize it unless the user asks for the raw data.
- `element-screenshot` also returns base64 image data and is cheaper than a full-page screenshot for visual checks.
