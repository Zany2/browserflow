---
name: browserflow-server-automa-workflows
description: "Run BrowserFlow Server-mode Automa workflows through the task scheduling API. Workflows include: 普通检索, 多参数检索, 单参数检索返回值变量与表格, 单参数检索返回值变量, 单参数检索"
---

# BrowserFlow Server Automa Workflows

## Overview

This skill describes Automa workflows stored in BrowserFlow Server mode. Use the BrowserFlow task APIs to create or run server-side tasks. The server dispatches each task to an online Windows client that owns the target workflow.

**Total Workflows Available:** 5

**Recommended Filename:** `SKILL.md`

**API Base URL:** `http://192.168.0.103/api/v1`

## Mandatory Preflight

Before running any workflow, first verify that the BrowserFlow backend is reachable and running in Server mode.

```bash
curl 'http://192.168.0.103/api/v1/app/runtime'
```

If the request fails or the mode is not `server`, ask the user to start BrowserFlow in Server mode before continuing.

Then verify that at least one client is online and has the target workflow. If a task does not specify a client, BrowserFlow scans online clients that own the workflow and chooses an unlocked client.

```bash
curl 'http://192.168.0.103/api/v1/clients'
```

## Required Step-By-Step Procedure

Always work in this order. Do not create or execute a task until the checks are complete.

1. Detect runtime: call `/app/runtime` and confirm BrowserFlow is reachable and running in `server` mode.
2. Detect clients: call `/clients`, confirm there is at least one online client, and confirm the target workflow can be dispatched to an online client.
3. Detect workflow and parameters: choose the matching workflow from this Skill, inspect its `Parameters`, and ask the user for any missing required values.
4. Find reusable tasks: query `/tasks?workflow_id={workflow_id}&page_num=1&page_size=10`. Reuse an enabled task when it matches the workflow and parameters, unless the user asks to create a new task.
5. Decide dispatch target: keep `client_ip` empty unless the user explicitly requires a specific client.
6. Decide execution mode: use async for action-only requests, sync for requests that need returned data or final completion.
7. Execute only after steps 1-6 pass. If any check fails, stop and report the exact reason instead of creating or executing a task.
8. After execution, inspect the response or task record before answering. For async runs, return the task record or execution identifier available in the response.

## Parameter Rules

Before creating or executing a task, inspect the workflow's `Parameters` section. If a required parameter has no value, ask the user for it. Pass values through the task `params` object and keep parameter names exactly as listed.

## Parameter Type Rules

- `string`: pass a string value.
- `number`: pass a JSON number, not a quoted string, when the user provides a numeric value.
- `json`: pass a valid JSON object or array. If the user provides plain text, ask them to confirm the JSON structure before executing.
- `checkbox`: pass a boolean `true` or `false`.
- Example values like `""` are placeholders. Replace required placeholders with real user-provided values before executing.

## Dispatch Rules

- If `client_ip` is provided, BrowserFlow dispatches only to that client.
- If `client_ip` is omitted, BrowserFlow scans online clients that own the workflow and dispatches to the first unlocked client.
- If the target client is busy, offline, or does not own the workflow, the API creates a failed execution record with a readable reason.
- Per-client Redis locks prevent the same client from running multiple Automa workflows concurrently.
- Use `trigger_type: "skill"` when executing tasks from this skill so execution records are easy to filter.

## Execution Mode Rules

- Use asynchronous execution when the user only asks to start, trigger, submit, launch, run, or execute a task and does not ask for returned data or final completion. Set `wait_result` to `false`.
- Use synchronous waiting when the user asks to get, query, search, extract, collect, return, fetch, read, wait for completion, or confirm final success/failure. Set `wait_result` to `true`, set a reasonable `timeout`, and request returned data if needed.
- For variable results, request `return_data.variables: ["browserflow_output"]` and read `result.data.variables.browserflow_output` first. If it is missing, report that the workflow completed but did not provide a BrowserFlow output variable.
- For table results, set `return_data.include_table` to `true`. BrowserFlow stores larger table payloads as task record files and returns the execution record for follow-up inspection.

## Result Reading Rules

- First read the execute response. It may contain `record` for the task record and `result` for the immediate client result.
- If the response contains `record.id`, call `/task-records/{record_id}` before giving the final answer when the user needs status, errors, returned variables, table files, or detailed output.
- For variable output, check `result.data.variables.browserflow_output`, then `record.result.data.variables.browserflow_output`, then the task record detail.
- For table output, inspect the task record detail `files` array. Use the file download endpoint when the table is stored as a file.
- Treat `queued` and `running` as incomplete states. Query the task record again later instead of reporting final success.

## Failure Handling Rules

- Backend unreachable: ask the user to start BrowserFlow Server and do not execute.
- Runtime mode is not `server`: ask the user to start or switch to Server mode and do not execute.
- No online clients: ask the user to open a Windows client browser-agent page and keep it connected.
- No online client owns the workflow: ask the user to sync the workflow to a client or choose another workflow.
- Client is busy or locked: report the busy reason and suggest retrying later or choosing another client.
- Missing required parameters: ask for the missing values before creating or executing a task.
- API returns a failed task record: report `record.error_message` or the readable failure reason from the response.

## API Endpoints

### Query Existing Tasks

```bash
curl 'http://192.168.0.103/api/v1/tasks?workflow_id=Yx03DAsLctZzjj_LDiOCN&page_num=1&page_size=10'
```

### Create Task

```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","description":"Created by BrowserFlow exported Skill","enabled":true,"name":"Skill task for Yx03DAsLctZzjj_LDiOCN","params":{},"run_once_after_create":false,"workflow_id":"Yx03DAsLctZzjj_LDiOCN"}'
```

### Execute Existing Task

```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","params":{},"trigger_type":"skill"}'
```

### Execute Existing Task And Wait For Result

```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","params":{},"return_data":{"include_history":false,"include_table":true,"table_limit":100,"variables":["browserflow_output"]},"timeout":300,"trigger_type":"skill","wait_result":true}'
```

### Query Task Records

```bash
curl 'http://192.168.0.103/api/v1/task-records?workflow_id=Yx03DAsLctZzjj_LDiOCN&page_num=1&page_size=10'
```

### Query Task Record Detail

```bash
curl 'http://192.168.0.103/api/v1/task-records/{record_id}'
```

### Download Task Record File

```bash
curl -O 'http://192.168.0.103/api/v1/task-records/files/{file_id}/download'
```

## Available Workflows

### 1. 普通检索

- Workflow ID: `Yx03DAsLctZzjj_LDiOCN`
- Server ID: `19`
- Description: 普通检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-23 23:37:56
- Updated: 2026-05-24 00:28:17

Parameters: none detected.

Create task example:
```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for Yx03DAsLctZzjj_LDiOCN","params":{},"run_once_after_create":true,"workflow_id":"Yx03DAsLctZzjj_LDiOCN"}'
```

### 2. 多参数检索

- Workflow ID: `wbI3CSCL4hRh8xbuTBD5h`
- Server ID: `18`
- Description: 多参数检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-24 00:28:59
- Updated: 2026-05-24 00:31:00

Parameters:
- `key_word` (string, required): 检索关键词 Default: `""`
- `age` (number, required): 年龄 Default: `""`
- `desc` (json): 档案 Default: `""`
- `is_enable` (checkbox): 是否上报 Default: `""`

Create task example:
```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for wbI3CSCL4hRh8xbuTBD5h","params":{"age":"","desc":"","is_enable":"","key_word":""},"run_once_after_create":true,"workflow_id":"wbI3CSCL4hRh8xbuTBD5h"}'
```

### 3. 单参数检索返回值变量与表格

- Workflow ID: `7KKfW4mVvDFBGux2kMgft`
- Server ID: `17`
- Description: 单参数检索返回值变量与表格
- Status: enabled
- Nodes: 9
- Created: 2026-05-23 23:44:00
- Updated: 2026-05-23 23:44:57

Parameters:
- `key_word` (string, required): 检索关键词 Default: `""`

Create task example:
```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for 7KKfW4mVvDFBGux2kMgft","params":{"key_word":""},"run_once_after_create":true,"workflow_id":"7KKfW4mVvDFBGux2kMgft"}'
```

### 4. 单参数检索返回值变量

- Workflow ID: `L-zNN7CUmhQB2l6imY31W`
- Server ID: `16`
- Description: 单参数检索返回值变量
- Status: enabled
- Nodes: 9
- Created: 2026-05-19 20:42:51
- Updated: 2026-05-23 23:43:48

Parameters:
- `key_word` (string, required): 检索关键词 Default: `""`

Create task example:
```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for L-zNN7CUmhQB2l6imY31W","params":{"key_word":""},"run_once_after_create":true,"workflow_id":"L-zNN7CUmhQB2l6imY31W"}'
```

### 5. 单参数检索

- Workflow ID: `XtdNYUUtraUJXUI2SAiRu`
- Server ID: `15`
- Description: 单参数检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-19 20:19:40
- Updated: 2026-05-23 23:41:10

Parameters:
- `key_word` (string, required): 检索关键词 Default: `""`

Create task example:
```bash
curl -X POST 'http://192.168.0.103/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for XtdNYUUtraUJXUI2SAiRu","params":{"key_word":""},"run_once_after_create":true,"workflow_id":"XtdNYUUtraUJXUI2SAiRu"}'
```

## Usage Notes

- These workflows come from the BrowserFlow server database, not from the currently open browser-agent page.
- Prefer creating reusable tasks for repeated use, then execute those task IDs from the skill.
- Leave `client_ip` empty when any online client that owns the workflow may execute it.
- Set `client_ip` only when the user explicitly wants a specific client.
- A successful execute response means the server accepted and dispatched the task. Use task records to inspect final status and returned data.
