# Browser Executor Troubleshooting

Read this file when an operation fails, page state is uncertain, or the task is sensitive.

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

- `GET /browser-executor/status` - Check whether the current browser is controllable
- `GET/POST /browser-executor/help` - Show command help
- `GET /browser-executor/export/skill` - Export Browser Executor Skill
- `POST /browser-executor/navigate` - Open URL and wait for load/dom-stable/request-idle/page-stable/none
- `GET/POST /browser-executor/snapshot` - Get page snapshot and RefIDs
- `GET/POST /browser-executor/clickable-elements` - Get compact clickable element RefIDs
- `GET/POST /browser-executor/input-elements` - Get compact input element RefIDs
- `GET/POST /browser-executor/observe` - Get compact page context for LLMs
- `POST /browser-executor/act` - Run a smart action by intent
- `POST /browser-executor/click` - Click element
- `POST /browser-executor/type` - Type text
- `POST /browser-executor/select` - Select dropdown option
- `POST /browser-executor/press-key` - Send key or shortcut
- `POST /browser-executor/wait` - Wait for page load, element states, DOM stability, network idle, or fixed time
- `POST /browser-executor/reload` - Reload current page
- `POST /browser-executor/go-back` - Go back in browser history
- `POST /browser-executor/go-forward` - Go forward in browser history
- `POST /browser-executor/hover` - Hover over element
- `POST /browser-executor/resize` - Resize viewport
- `GET/POST /browser-executor/page-info` - Get current page URL and title
- `GET/POST /browser-executor/page-text` - Get compact visible page text
- `GET/POST /browser-executor/page-content` - Get compact page HTML
- `GET/POST /browser-executor/page-structure` - Get compact structured headings, links, forms, tables, images, and buttons
- `POST /browser-executor/get-text` - Get element text
- `POST /browser-executor/get-value` - Get element value
- `POST /browser-executor/element-info` - Get element diagnostics including text, attributes, state, box, and XPath
- `POST /browser-executor/extract` - Extract page text or selector data
- `POST /browser-executor/screenshot` - Capture screenshot as base64
- `POST /browser-executor/element-screenshot` - Capture a single element screenshot as base64
- `POST /browser-executor/evaluate` - Execute JavaScript with automatic function wrapping
- `POST /browser-executor/cookies` - List, set, delete, or clear browser cookies
- `POST /browser-executor/storage` - List, get, set, delete, or clear localStorage/sessionStorage on the current page
- `POST /browser-executor/tabs` - Manage tabs list/new/switch/close
- `POST /browser-executor/scroll` - Scroll page or element
- `POST /browser-executor/mouse` - Run coordinate mouse operations move/click/double-click/right-click/down/up/scroll
- `POST /browser-executor/window` - Read or change browser window bounds and state
- `POST /browser-executor/close-page` - Close current page
- `POST /browser-executor/fill-form` - Fill multiple form fields in one call
- `POST /browser-executor/drag` - Drag one element to another
- `POST /browser-executor/file-upload` - Upload local files to a file input
- `POST /browser-executor/handle-dialog` - Arm a handler for the next JavaScript dialog
- `POST /browser-executor/batch` - Execute operations in sequence

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
