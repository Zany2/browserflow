---
name: browserflow-browser-executor
description: Control the current BrowserFlow Windows browser directly through HTTP APIs. Use observe, act, accessibility snapshots, RefIDs, page structure, element diagnostics, page text, page HTML, cookies, storage, and compact actions to navigate, click, type, fill forms, upload files, drag, handle dialogs, scroll, reload, extract data, screenshot, operate mouse/window, run JavaScript, and manage tabs without Automa workflows.
---

# BrowserFlow Browser Executor

## Overview

Use this skill when the user wants to operate the browser directly with an LLM instead of running a prebuilt Automa workflow.

Keep the BrowserFlow `browser-agent` client tab alive. Do not close the whole browser after a task; when cleanup is needed, close only task-related business tabs that were opened or used for that task.

**API Base URL:** `http://127.0.0.1:8001/api/v1/browser-executor`

## Reference Files

- `references/api-reference.md`: complete endpoint parameters, enums, and shared conventions. Read when command parameters are uncertain.
- `references/command-examples.md`: curl examples for preflight and common browser operations. Read when composing raw HTTP calls.
- `references/troubleshooting.md`: safety boundaries, recovery rules, element identification, response format, and notes. Read when an operation fails or the task is sensitive.

## Quick Start

Use this Skill with a compact observe-act-verify loop:

1. Check runtime and browser status when needed.
2. Inspect only the controls or page area needed for the next action.
3. Group deterministic small actions that share one user intent.
4. Verify the page state after each action group.
5. Extract only the requested data, then close only task-related business tabs.

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

## Fast Path Decision Table

Use this table before choosing APIs. It is designed for models that need an explicit shortest safe route.

| User goal | Preferred command sequence | Notes |
|---|---|---|
| Open a site or exact URL | `status` if needed -> `navigate` -> focused `observe` | Direct URL is allowed only for an exact URL or stable entry page. |
| Search with keyword | `navigate` -> `input-elements` -> `fill-form` or `act` -> focused `observe` | Enter query and submit as one action group when controls are clear. |
| Apply filters/sort/page | `clickable-elements` or compact `observe` -> `act` or `batch` -> focused `observe` | Use visible controls, not handcrafted query params. |
| Fill a form | `input-elements` -> `fill-form` -> `wait` -> focused `observe` | Prefer one form call over many `type` calls. |
| Extract list/table data | focused `page-structure` or read-only `evaluate` -> optional targeted `extract` | Extract only the requested container/count. |
| Click a known button/link | `clickable-elements` or `snapshot` -> `act` with `return_observe:true` | Use RefIDs like `@e1` when available. |
| Diagnose one element | `element-info` -> retry with fresh identifier or coordinate fallback | Use this when hidden, disabled, overlapped, or stale. |
| Close task tabs | `tabs` list -> `tabs` close only business tabs | Never close URLs containing `#/browser-agent`. |

Avoid slow paths: do not start with full `page-content`, large `page-text`, full-page screenshots, repeated single keystrokes, or repeated preflight checks unless the task truly needs them.

## High-Frequency Parameter Reference

Use these field meanings when building JSON payloads. Keep JSON field names exactly as shown.

### `navigate` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `url` | `string` | yes | Target page URL. Use exact user-provided URL or stable entry page only. | `https://example.com` |
| `wait_until` | `string` | no | `load`, `dom-stable`, `request-idle`, `page-stable`, or `none`. Prefer `load` or `dom-stable`. | `load` |
| `timeout` | `number` | no | Maximum wait seconds. | `15` |
| `return_observe` | `boolean` | no | Return page observation after navigation. | `true` |

### `observe` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `include_text` | `boolean` | no | Whether to include visible page text. Keep false for control discovery. | `false` |
| `text_limit` | `number` | no | Maximum text characters when text is included. | `4000` |

### `input-elements / clickable-elements` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `limit` | `number` | no | Maximum controls returned. Keep modest for speed. | `30` |

### `page-structure` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `include_links` | `boolean` | no | Include links. | `true` |
| `include_forms` | `boolean` | no | Include forms and fields. | `true` |
| `include_tables` | `boolean` | no | Include tables. | `true` |
| `include_buttons` | `boolean` | no | Include buttons. | `true` |
| `include_images` | `boolean` | no | Include images. Usually false unless image data is needed. | `false` |
| `limit` | `number` | no | Maximum structured items returned. | `30` |

### `act` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `intent` | `string` | yes | High-level action such as `click`, `type`, `select`, `check`, `navigate`, or `scroll`. | `click` |
| `identifier` | `string` | conditional | RefID, selector, XPath, or visible text target. Required for element actions. | `@e1` |
| `text` | `string` | conditional | Text for typing or prompt-like actions. | `ai智能体` |
| `value` | `string` | conditional | Option value for select-like actions. | `latest` |
| `return_observe` | `boolean` | no | Return updated page state after the action. | `true` |

### `click / type / fill-form` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `identifier` | `string` | conditional | RefID, selector, XPath, or visible text. Required for single element click/type. | `@e2` |
| `text` | `string` | conditional | Text to type. | `hello` |
| `clear` | `boolean` | no | Clear existing input before typing. | `true` |
| `fields` | `array` | conditional | For `fill-form`: list of fields with `name` and `value`. | `[{"name":"keyword","value":"ai智能体"}]` |
| `submit` | `boolean` | no | Submit the form after filling. | `true` |
| `timeout` | `number` | no | Maximum wait seconds. | `10` |

### `wait` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `state` | `string` | yes | `load`, `visible`, `hidden`, `enabled`, `interactable`, `writable`, `stable`, `dom-stable`, `request-idle`, `elements-more-than`, or `time`. | `dom-stable` |
| `identifier` | `string` | conditional | Target element for element states. | `@e1` |
| `timeout` | `number` | no | Maximum wait seconds. | `10` |
| `count` | `number` | conditional | Expected minimum count for `elements-more-than`. | `5` |

### `evaluate` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `script` | `string` | yes | JavaScript to run. Prefer read-only extraction. | `return document.title` |

### `tabs` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `action` | `string` | yes | `list`, `new`, `switch`, `close`, or supported tab action. | `list` |
| `url` | `string` | conditional | URL for creating a new tab. | `https://example.com` |
| `index` | `number` | conditional | Tab index returned by `tabs` list for switch/close. | `1` |

### `batch` parameters

| Field | Type | Required | Meaning | Example |
|---|---|---|---|---|
| `operations` | `array` | yes | Sequential operations. Each item has `type`, `params`, and optional `stop_on_error`. | `[{"type":"click","params":{"identifier":"@e1"},"stop_on_error":true}]` |
| `type` | `string` | yes | Operation command name inside one batch item. | `click` |
| `params` | `object` | yes | Payload for that command. | `{"identifier":"@e1"}` |
| `stop_on_error` | `boolean` | no | Stop batch when this operation fails. | `true` |

## Safety Boundaries

Ask the user for explicit confirmation before destructive, irreversible, or externally visible actions, including submitting purchases, orders, payments, account/security changes, deleting data, sending messages or emails, uploading files, clearing cookies/storage, or changing important settings. If the user already gave clear permission for that exact action in the current request, proceed carefully and verify before submitting.

Do not use `evaluate` to bypass user confirmation, disable site protections, read unrelated secrets, or mutate sensitive page state when a semantic BrowserFlow API can do the task. Prefer high-level APIs such as `click`, `type`, `fill-form`, `cookies`, and `storage` over custom JavaScript.

