# Browser Executor API Reference

Complete parameter details, enums, and shared conventions for BrowserFlow Browser Executor APIs.

## All Endpoint Parameter Reference

This section is the complete parameter reference for Browser Executor APIs. Use the exact JSON field names below. When a field has allowed values, choose one of the listed values instead of inventing a synonym.

### Shared Variables And Conventions

| Variable / concept | Available values | Meaning |
|---|---|---|
| `identifier` | RefID like `@e1`; CSS selector like `.btn`; XPath like `//button[text()='OK']`; visible text such as `Search` | Identifies a page element. Prefer fresh RefIDs from `snapshot`, `observe`, `clickable-elements`, or `input-elements`. RefIDs may become stale after navigation or DOM changes. |
| `wait_until` | `load`, `dom-stable`, `request-idle`, `page-stable`, `none` | Navigation wait strategy. Use `load` for normal pages, `dom-stable` for DOM-rendered pages, `request-idle` for SPAs/network-heavy pages, `page-stable` for visual/layout stability, and `none` only when the next command waits explicitly. |
| `return_observe` | `true`, `false` | When supported, appends `observe` output to the operation result so the model can verify the page state in the same call. |
| `include_text` | `true`, `false` | Controls whether appended observe includes visible page text. Keep false for speed unless text is needed. |
| `text_limit` / `limit` | Positive integer | Caps returned text or item count. Start small and increase only when necessary. |
| `fields[].type` | `text`, `textarea`, `select`, `checkbox`, `radio`, `number`, `password`, `email`, `search`, `date`, `json`, `boolean` | Optional form-field hint for the model/exported docs. Backend auto-detects actual DOM type, so this is mainly descriptive. |
| `result` response | `data.result.success`, `data.result.error`, `data.result.data` | GoFrame wraps responses as `{code,message,data}`. Browser operation result is usually under `data.result`; inspect success/error/data before reporting completion. |

### `status`

- Route: GET `/status`
- Notes: No parameters.
- Parameters: none.

### `help`

- Route: GET/POST `/help`
- Notes: Use this to inspect command help when unsure.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `command` | `string` | no | Optional command name filter. Empty returns all command help. | Command names from `help`, such as `navigate`, `act`, `tabs`, `storage`. | `act` |

### `export/skill`

- Route: GET `/export/skill`
- Notes: No parameters. Downloads or returns the current generated SKILL.md content.
- Parameters: none.

### `navigate`

- Route: POST `/navigate`
- Notes: Opens an exact URL or stable entry page. Do not encode user search/filter/page requirements into the URL.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `url` | `string` | yes | Target URL. If protocol is missing, backend prefixes `https://`. | Any absolute URL or domain-like string. | `https://example.com` |
| `wait_until` | `string` | no | Navigation wait strategy after opening the URL. Default is `load`. | `load`: wait for page load event; `dom-stable`: wait until DOM stops changing; `request-idle`: wait until network is idle; `page-stable`: wait until page becomes stable; `none`: return immediately. | `load` |
| `timeout` | `number` | no | Maximum wait time in seconds. Default is 60 when omitted or <= 0. | Positive integer seconds. | `30` |
| `return_observe` | `boolean` | no | Append an `observe` result after navigation. Useful when the next step needs current page state. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text in the appended observe result. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. Default is 8000 in appended observe. | Positive integer characters. | `4000` |

### `snapshot`

- Route: GET/POST `/snapshot`
- Notes: No parameters. Returns accessibility text and RefIDs such as `@e1`.
- Parameters: none.

### `clickable-elements`

- Route: GET/POST `/clickable-elements`
- Notes: Returns compact clickable controls from the current snapshot.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `limit` | `number` | no | Maximum number of clickable elements to return. Default is 50 in batch usage. | Positive integer item count. | `50` |

### `input-elements`

- Route: GET/POST `/input-elements`
- Notes: Returns compact input controls from the current snapshot.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `limit` | `number` | no | Maximum number of input elements to return. Default is 30 in batch usage. | Positive integer item count. | `30` |

### `click`

- Route: POST `/click`
- Notes: Clicks one element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target element identifier. | RefID such as `@e1`; CSS selector; XPath; or visible text. Prefer fresh RefIDs. | `@e1` |
| `return_observe` | `boolean` | no | Append current page observation after click. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. | Positive integer characters. | `4000` |

### `type`

- Route: POST `/type`
- Notes: Types text into one element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Input-like target element identifier. | RefID, CSS selector, XPath, or visible text. Prefer `input-elements` or `snapshot` RefIDs. | `@e2` |
| `text` | `string` | no | Text to enter. Empty string is allowed when clearing an input. | Any user text. | `hello` |
| `clear` | `boolean` | no | Clear existing value before typing. Default is true in batch usage. | `true` or `false`. | `true` |
| `return_observe` | `boolean` | no | Append current page observation after typing. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. | Positive integer characters. | `4000` |

### `select`

- Route: POST `/select`
- Notes: Selects an option in a dropdown/select-like element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target select element identifier. | RefID, CSS selector, XPath, or visible text. | `@e3` |
| `value` | `string` | yes | Option visible text, value, or regex-like text accepted by Rod select matching. | Visible label or option value. | `China` |
| `return_observe` | `boolean` | no | Append current page observation after selecting. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. | Positive integer characters. | `4000` |

### `press-key`

- Route: POST `/press-key`
- Notes: Sends one key or shortcut to the active page.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `key` | `string` | yes | Key to press. | `Enter`/`Return`, `Tab`, `Escape`/`Esc`, `Backspace`, `Delete`, `Up`/`ArrowUp`, `Down`/`ArrowDown`, `Left`/`ArrowLeft`, `Right`/`ArrowRight`, `Home`, `End`, `PageUp`, `PageDown`, `Space`, or a single character. | `Enter` |
| `ctrl` | `boolean` | no | Hold Ctrl while pressing key. | `true` or `false`. | `false` |
| `shift` | `boolean` | no | Hold Shift while pressing key. | `true` or `false`. | `false` |
| `alt` | `boolean` | no | Hold Alt while pressing key. | `true` or `false`. | `false` |
| `meta` | `boolean` | no | Hold Meta/Command/Windows key while pressing key. | `true` or `false`. | `false` |

### `wait`

- Route: POST `/wait`
- Notes: Waits for page, element, count, or time state.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | conditional | Target element or selector. Required for element states and `elements-more-than`; omit for page wait states. | RefID/CSS/XPath/text for element states; CSS selector is best for `elements-more-than`. | `.result-item` |
| `state` | `string` | no | Wait state. Default is `load` when omitted. | `load`: page load; `visible`: element visible; `hidden`: element invisible; `enabled`: element enabled; `interactable`: element can receive input; `writable`: input can be written; `stable`: element stable, or page stable if no identifier; `dom-stable`: DOM stable; `request-idle`: network idle; `page-stable`: page stable; `elements-more-than`: selector count above `count`; `time`: sleep for `timeout` seconds. | `dom-stable` |
| `timeout` | `number` | no | Maximum wait seconds. Default is 10 when omitted or <= 0. | Positive integer seconds. | `10` |
| `count` | `number` | conditional | Minimum element count for `elements-more-than`. | Positive integer count. | `5` |

### `reload`

- Route: POST `/reload`
- Notes: No parameters. Reloads the current business page.
- Parameters: none.

### `go-back`

- Route: POST `/go-back`
- Notes: No parameters. Navigates back in browser history.
- Parameters: none.

### `go-forward`

- Route: POST `/go-forward`
- Notes: No parameters. Navigates forward in browser history.
- Parameters: none.

### `hover`

- Route: POST `/hover`
- Notes: Moves the cursor over one element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target element identifier. | RefID, CSS selector, XPath, or visible text. | `@e1` |

### `resize`

- Route: POST `/resize`
- Notes: Changes viewport size for the active page.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `width` | `number` | no | Viewport width in CSS pixels. Default is 1440 when omitted or <= 0. | Positive integer pixels. | `1440` |
| `height` | `number` | no | Viewport height in CSS pixels. Default is 900 when omitted or <= 0. | Positive integer pixels. | `900` |

### `page-info`

- Route: GET/POST `/page-info`
- Notes: No parameters. Returns URL, title, target id, ready state, viewport, scroll, and page counts.
- Parameters: none.

### `observe`

- Route: GET/POST `/observe`
- Notes: Returns compact page state: status, page info, snapshot text, RefIDs, and optional visible text.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `include_text` | `boolean` | no | Include visible page text. Keep false unless the current step needs broad page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum visible text characters when `include_text` is true. Default is 8000 when omitted or <= 0. | Positive integer characters. | `4000` |

### `get-text`

- Route: POST `/get-text`
- Notes: Reads visible text/textContent from one element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target element identifier. | RefID, CSS selector, XPath, or visible text. | `@e1` |

### `get-value`

- Route: POST `/get-value`
- Notes: Reads value, value attribute, or textContent from one element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target element identifier. | RefID, CSS selector, XPath, or visible text. | `@e2` |

### `element-info`

- Route: POST `/element-info`
- Notes: Returns diagnostics for one element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target element identifier. | RefID, CSS selector, XPath, or visible text. | `@e1` |
| `attributes` | `array<string>` | no | Attribute names to read. If empty, defaults to common diagnostic attributes. | Any HTML attribute name. Default set: `id`, `name`, `class`, `type`, `role`, `aria-label`, `placeholder`, `href`, `src`, `title`, `alt`. | `["id","class","href","aria-label"]` |

### `page-text`

- Route: GET/POST `/page-text`
- Notes: Returns visible text from the whole current page.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `limit` | `number` | no | Maximum visible text characters. Default is 8000 when omitted or <= 0. | Positive integer characters. Increase gradually only when needed. | `8000` |

### `page-content`

- Route: GET/POST `/page-content`
- Notes: Returns current page HTML. Use only when narrower APIs cannot answer.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `limit` | `number` | no | Maximum HTML characters. If omitted or <= 0, backend defaults to 8000 through internal limiter. | Positive integer characters. | `12000` |

### `page-structure`

- Route: GET/POST `/page-structure`
- Notes: Returns compact structured page data by category.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `include_links` | `boolean` | no | Include visible links. | `true` or `false`. If all include flags are false, backend includes all categories. | `true` |
| `include_forms` | `boolean` | no | Include forms and fields. | `true` or `false`. | `true` |
| `include_tables` | `boolean` | no | Include table text. | `true` or `false`. | `true` |
| `include_images` | `boolean` | no | Include image src/alt/title data. Usually false unless image info is needed. | `true` or `false`. | `false` |
| `include_buttons` | `boolean` | no | Include buttons and button-like controls. | `true` or `false`. | `true` |
| `limit` | `number` | no | Maximum items per included category. Default is 50 when omitted or <= 0. | Positive integer item count. | `30` |

### `extract`

- Route: POST `/extract`
- Notes: Extracts selected fields from elements matched by a CSS selector.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `selector` | `string` | no | CSS selector to extract. If empty, returns full body visible text instead of structured items. | CSS selector. | `.result-item` |
| `fields` | `array<string>` | no | Fields to keep from each selected element. Default is `text`, `href`, `value`. | `text`: innerText/textContent; `href`: link href; `value`: input value; `html`: outerHTML. | `["text","href"]` |
| `multiple` | `boolean` | no | Whether to return all matched elements. If false, only the first matched item is returned. | `true` or `false`. | `true` |

### `screenshot`

- Route: POST `/screenshot`
- Notes: Captures a page screenshot as base64 image data.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `full_page` | `boolean` | no | Capture the full scrollable page instead of the viewport only. | `true` or `false`. | `true` |
| `format` | `string` | no | Image format. Default is `png` unless `jpeg` is provided. | `png` or `jpeg`. | `png` |
| `quality` | `number` | no | JPEG quality. Ignored for PNG. Used only when > 0. | 1-100 is typical for JPEG. | `80` |

### `element-screenshot`

- Route: POST `/element-screenshot`
- Notes: Captures one element as base64 image data.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | Target element identifier. | RefID, CSS selector, XPath, or visible text. | `@e1` |
| `format` | `string` | no | Image format. Default is `png` unless `jpeg` is provided. | `png` or `jpeg`. | `png` |
| `quality` | `number` | no | JPEG quality. Ignored for PNG. Used only when > 0. | 1-100 is typical for JPEG. | `80` |

### `evaluate`

- Route: POST `/evaluate`
- Notes: Runs JavaScript in the current page. Prefer read-only extraction and semantic APIs first.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `script` | `string` | yes | JavaScript source. Plain bodies are auto-wrapped as `() => { ... }`; function expressions and async functions are accepted as-is. | Normal JavaScript. Return serializable values. | `return document.title` |

### `cookies`

- Route: POST `/cookies`
- Notes: Lists, sets, deletes, or clears browser cookies.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `action` | `string` | no | Cookie operation. Default is `list`. | `list`/`get`: list cookies; `set`: create/update cookie; `delete`/`remove`: delete one cookie; `clear`: clear cookies. | `list` |
| `name` | `string` | conditional | Cookie name. Required for `set`, `delete`, and `remove`. | Cookie key string. | `token` |
| `value` | `string` | conditional | Cookie value. Used by `set`. | Cookie value string. | `abc` |
| `url` | `string` | conditional | Cookie URL scope. If omitted with no `domain`, backend uses current page URL. | URL such as `https://example.com`. | `https://example.com` |
| `domain` | `string` | conditional | Cookie domain scope. Use either URL or domain according to browser cookie rules. | Domain such as `example.com`. | `example.com` |
| `path` | `string` | no | Cookie path scope. | Path string. | `/` |
| `secure` | `boolean` | no | Set cookie Secure flag. | `true` or `false`. | `true` |
| `http_only` | `boolean` | no | Set cookie HttpOnly flag. | `true` or `false`. | `false` |
| `same_site` | `string` | no | Cookie SameSite value. | `Strict`, `Lax`, or `None`. Empty leaves it unset. | `Lax` |
| `expires` | `number` | no | Cookie expiration timestamp. 0 means session cookie. | Unix epoch seconds as number. | `0` |

### `storage`

- Route: POST `/storage`
- Notes: Manages localStorage or sessionStorage on the current page.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `action` | `string` | no | Storage operation. Default is `list`. | `list`: list all keys; `get`: get one key; `set`: set one key; `delete`/`remove`: remove one key; `clear`: clear all keys in selected storage. | `list` |
| `type` | `string` | no | Storage area. Default is `local`. | `local`: localStorage; `session`: sessionStorage. | `local` |
| `key` | `string` | conditional | Storage key. Required for `get`, `set`, `delete`, and `remove`. | Any storage key string. | `token` |
| `value` | `string` | conditional | Storage value. Used by `set`. | String value. JSON should be stringified before sending. | `{"id":1}` |

### `tabs`

- Route: POST `/tabs`
- Notes: Manages browser tabs. Never close or switch to a tab whose URL contains `#/browser-agent`.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `action` | `string` | yes | Tab operation. | `list`: list tabs; `new`: create and activate tab; `switch`: activate tab by index; `close`: close tab by index. | `list` |
| `url` | `string` | conditional | URL for `new`. Defaults to `about:blank` if empty. | Absolute URL or `about:blank`. | `https://example.com` |
| `index` | `number` | conditional | Tab index returned by `tabs` list. Required for `switch` and `close`. | Zero-based tab index from API response. | `1` |

### `scroll`

- Route: POST `/scroll`
- Notes: Scrolls the page or a target element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `direction` | `string` | no | Scroll direction. Default is `down`. | `down`: scroll down by pixels; `up`: scroll up by pixels; `top`: scroll to top; `bottom`: scroll to bottom. | `down` |
| `pixels` | `number` | no | Scroll distance in pixels for `down`/`up`. Default is 700 when omitted or <= 0. | Positive integer pixels. | `700` |
| `identifier` | `string` | no | Optional target scroll container or element. If empty, scrolls window. | RefID, CSS selector, XPath, or visible text. | `@e1` |
| `return_observe` | `boolean` | no | Append current page observation after scrolling. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. | Positive integer characters. | `4000` |

### `mouse`

- Route: POST `/mouse`
- Notes: Runs coordinate-based mouse operations. Prefer semantic element APIs first.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `action` | `string` | yes | Mouse operation. | `move`/`move-to`: move pointer; `click`: click; `double-click`/`dblclick`: double click; `right-click`/`context-menu`: right click; `down`: press mouse button; `up`: release mouse button; `scroll`/`wheel`: mouse wheel scroll. | `click` |
| `x` | `number` | conditional | Viewport X coordinate. Used by move/click/down/up/scroll. | CSS pixel coordinate, often from `element-info.data.box`. | `300` |
| `y` | `number` | conditional | Viewport Y coordinate. Used by move/click/down/up/scroll. | CSS pixel coordinate, often from `element-info.data.box`. | `200` |
| `delta_x` | `number` | conditional | Horizontal wheel delta. Used by `scroll`/`wheel`. | Number of pixels or wheel delta units. | `0` |
| `delta_y` | `number` | conditional | Vertical wheel delta. Used by `scroll`/`wheel`. | Number of pixels or wheel delta units. | `700` |
| `steps` | `number` | no | Movement or wheel steps. Default is 10 when omitted or <= 0. | Positive integer. | `10` |
| `button` | `string` | no | Mouse button. Default is `left`. | `left`, `right`, or `middle`. | `left` |

### `window`

- Route: POST `/window`
- Notes: Reads or changes browser window bounds/state.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `action` | `string` | yes | Window operation. | `info`/`get`: read bounds; `set`/`resize`/`move`: set bounds; `maximize`/`maximized`: maximize; `minimize`/`minimized`: minimize; `fullscreen`: fullscreen; `normal`/`restore`: restore normal state. | `info` |
| `left` | `number` | conditional | Window left position for `set`/`move`. 0 means leave unchanged in current implementation. | Screen pixel coordinate. | `100` |
| `top` | `number` | conditional | Window top position for `set`/`move`. 0 means leave unchanged in current implementation. | Screen pixel coordinate. | `100` |
| `width` | `number` | conditional | Window width for `set`/`resize`. | Positive screen pixel width. | `1280` |
| `height` | `number` | conditional | Window height for `set`/`resize`. | Positive screen pixel height. | `900` |

### `close-page`

- Route: POST `/close-page`
- Notes: No parameters. Closes the active business page only. Backend refuses to close BrowserFlow `#/browser-agent` tabs.
- Parameters: none.

### `fill-form`

- Route: POST `/fill-form`
- Notes: Fills several form fields in one call. It matches fields by name, id, placeholder, aria-label, associated label text, or identifier fallback.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `fields` | `array<object>` | yes | Form fields to fill. Each item has `name`, `value`, and optional `type`. | `name`: field name/id/placeholder/label text; `value`: string/number/boolean; `type`: optional hint for model readability, backend auto-detects actual DOM field type. | `[{"name":"keyword","value":"ai智能体","type":"text"}]` |
| `fields[].name` | `string` | yes | Field locator inside form filling. | Input `name`, `id`, placeholder text, aria-label, visible label text, RefID, CSS selector, XPath, or visible text fallback. | `keyword` |
| `fields[].value` | `any` | no | Value to set. Checkboxes/radios use truthy values. | Strings, numbers, booleans. Truthy checkbox/radio values include `true`, `1`, `yes`, `on`, `checked`. | `ai智能体` |
| `fields[].type` | `string` | no | Optional type hint for the LLM and exported skill. Current backend auto-detects DOM field type and does not require this hint. | Common hints: `text`, `textarea`, `select`, `checkbox`, `radio`, `number`, `password`, `email`, `search`, `date`, `json`, `boolean`. | `text` |
| `submit` | `boolean` | no | Submit the form after filling. Backend tries submit button, common submit labels, then Enter. | `true` or `false`. | `true` |
| `timeout` | `number` | no | Per-field match/fill timeout seconds. Default is 10 when omitted or <= 0. | Positive integer seconds. | `10` |
| `return_observe` | `boolean` | no | Append current page observation after filling. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. | Positive integer characters. | `4000` |

### `drag`

- Route: POST `/drag`
- Notes: Drags one element to another element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `from_identifier` | `string` | yes | Source element identifier. | RefID, CSS selector, XPath, or visible text. | `@e1` |
| `to_identifier` | `string` | yes | Target element identifier. | RefID, CSS selector, XPath, or visible text. | `@e2` |

### `file-upload`

- Route: POST `/file-upload`
- Notes: Sets local file paths on a file input element.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `identifier` | `string` | yes | File input element identifier. | RefID, CSS selector, XPath, or visible text. Prefer an actual `<input type="file">`. | `@e4` |
| `file_paths` | `array<string>` | yes | Local file paths to upload. | Windows absolute paths such as `C:\path\file.png`; paths must exist on the machine running the browser. | `["C:\\path\\file.png"]` |

### `handle-dialog`

- Route: POST `/handle-dialog`
- Notes: Arms a handler for the next JavaScript alert/confirm/prompt/beforeunload dialog. Call this before the action that opens the dialog.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `accept` | `boolean` | no | Whether to accept the dialog. False dismisses/cancels. | `true` or `false`. | `true` |
| `text` | `string` | no | Prompt input text when handling a prompt dialog. | Any prompt text. | `ok` |
| `timeout` | `number` | no | How long the handler waits for the dialog in seconds. Default is 10 when omitted or <= 0. | Positive integer seconds. | `10` |

### `act`

- Route: POST `/act`
- Notes: Runs one high-level action by intent. Prefer this for simple known actions after inspecting the page.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `intent` | `string` | yes | High-level action to run. | `navigate`/`open`/`goto`: open URL from `text` or `value`; `click`/`press`: click `identifier`; `type`/`input`/`fill`: type into `identifier`; `select`/`choose`: choose option; `check`: check checkbox/radio; `uncheck`: uncheck checkbox; `fill-form`/`form`: fill multiple fields; `press-key`/`key`: press key from `text` or `value`; `scroll`: scroll using `text`/`value` as direction; `hover`: hover target. | `click` |
| `identifier` | `string` | conditional | Target element for element actions. | RefID, CSS selector, XPath, or visible text. Required for click/type/select/check/uncheck/hover and optional for scroll. | `@e1` |
| `value` | `any` | conditional | Generic action value. Used as URL for navigate, text for type, option for select, key for press-key, direction for scroll. | String/number/boolean depending on intent. `text` takes priority for text-like intents. | `ai智能体` |
| `text` | `string` | conditional | Text action value. Used for type, navigate URL, select value, key, or scroll direction. | Any string. | `ai智能体` |
| `wait_until` | `string` | no | Navigation wait strategy for navigate/open/goto intents. | `load`, `dom-stable`, `request-idle`, `page-stable`, or `none`. | `load` |
| `fields` | `array<object>` | conditional | Fields for `fill-form`/`form` intent. | Same structure as `fill-form.fields`. | `[{"name":"keyword","value":"ai智能体","type":"text"}]` |
| `submit` | `boolean` | no | Submit after `fill-form`/`form` intent. | `true` or `false`. | `true` |
| `clear` | `boolean` | no | Clear before type/input/fill intent. Default is true in batch usage. | `true` or `false`. | `true` |
| `timeout` | `number` | no | Timeout seconds for navigate/form operations. Default is 10 for act form/lookup operations when omitted or <= 0. | Positive integer seconds. | `10` |
| `return_observe` | `boolean` | no | Append current page observation after the action. | `true` or `false`. | `true` |
| `include_text` | `boolean` | no | When `return_observe` is true, include visible page text. | `true` or `false`. | `false` |
| `text_limit` | `number` | no | Maximum page text characters when `include_text` is true. | Positive integer characters. | `4000` |

### `batch`

- Route: POST `/batch`
- Notes: Executes deterministic operations sequentially. Do not batch across unknown decision points that require fresh observation.

| Field | Type | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|
| `operations` | `array<object>` | yes | Ordered operations to run. Each item has `type`, `params`, and optional `stop_on_error`. | Array of batch operation objects. | `[{"type":"click","params":{"identifier":"@e1"},"stop_on_error":true}]` |
| `operations[].type` | `string` | yes | Command name to run inside batch. | Supported: `navigate`, `click`, `type`, `select`, `press-key`, `wait`, `snapshot`, `clickable-elements`, `input-elements`, `observe`, `page-text`, `page-content`, `get-text`, `get-value`, `element-info`, `extract`, `screenshot`, `evaluate`, `cookies`, `storage`, `page-structure`, `scroll`, `reload`, `go-back`, `go-forward`, `hover`, `resize`, `element-screenshot`, `mouse`, `window`, `close-page`, `fill-form`, `drag`, `file-upload`, `handle-dialog`, `tabs`, `act`. | `navigate` |
| `operations[].params` | `object` | yes | Parameters for the selected operation type. Use the same field names described in this reference. | Object matching the endpoint payload. | `{"url":"https://example.com","wait_until":"load"}` |
| `operations[].stop_on_error` | `boolean` | no | Stop executing later batch operations if this operation returns an error. | `true` or `false`. | `true` |

