# Browser Executor Command Examples

Use these examples when composing raw HTTP calls. Prefer compact observations and verify after page-changing actions.

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
  -d '{"url":"https://example.com","wait_until":"load"}'
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
  -d '{"include_links":true,"include_forms":true,"include_tables":true,"include_images":false,"include_buttons":true,"limit":30}'
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
  -d '{"text":"hello","clear":true,"identifier":"@e2"}'
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
  -d '{"timeout":10,"state":"dom-stable"}'
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
  -d '{"timeout":10,"accept":true,"text":""}'
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
  -d '{"operations":[{"type":"navigate","params":{"url":"https://example.com"},"stop_on_error":true},{"params":{"include_text":false},"stop_on_error":true,"type":"observe"}]}'
```

