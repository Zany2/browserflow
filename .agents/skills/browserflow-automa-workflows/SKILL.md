---
name: browserflow-server-automa-workflows
description: "Run BrowserFlow Server-mode Automa workflows through the task scheduling API. Workflows include: 单参数检索, 普通检索"
---

# BrowserFlow Server Automa Workflows

## Overview

This skill describes Automa workflows stored in BrowserFlow Server mode. Use the BrowserFlow task APIs to create or run server-side tasks. The server dispatches each task to an online Windows client node that owns the target workflow. A client node is identified by `client_ip` plus `node_id`.

**Total Workflows Available:** 2

**Recommended Filename:** `SKILL.md`

**API Base URL:** `http://127.0.0.1:8001/api/v1`

## Mandatory Preflight

Before running any workflow, first verify that the BrowserFlow backend is reachable and running in Server mode.

```bash
curl 'http://127.0.0.1:8001/api/v1/app/runtime'
```

If the request fails, returns a non-successful `code`, or the mode is not `server`, ask the user to start BrowserFlow in Server mode before continuing. Do not create or execute any task until this check passes.

Then verify that at least one client node is online and has the target workflow. If a task does not specify a client node, BrowserFlow scans online nodes that own the workflow and chooses an unlocked node.

```bash
curl 'http://127.0.0.1:8001/api/v1/clients'
```

If the clients response fails, returns a non-successful `code`, has no online nodes, or cannot prove that an online node owns the target workflow, stop and tell the user the exact blocker before creating or executing a task.

## Required Step-By-Step Procedure

Always work in this order. Do not create or execute a task until the checks are complete.

1. Detect runtime: call `/app/runtime` and confirm BrowserFlow is reachable, the response is successful, and the runtime is `server` mode.
2. Detect client nodes: call `/clients`, confirm the response is successful, confirm there is at least one online node, and confirm the target workflow can be dispatched to an online node.
3. Detect workflow and parameters: choose the matching workflow from this Skill, inspect its `Parameters`, and ask the user for any missing required values.
4. Find reusable tasks: query `/tasks?workflow_id={workflow_id}&page_num=1&page_size=10`. Reuse an enabled task when it matches the workflow and parameters, unless the user asks to create a new task.
5. Decide dispatch target: keep both `client_ip` and `node_id` empty for automatic dispatch, or set both fields when the user explicitly requires a specific node.
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

## All API Parameter Reference

Use this as the complete parameter reference for the Server-mode Automa workflow Skill. Keep field names exactly as shown.

### Shared Variables And Response Rules

| Variable / concept | Available values | Meaning |
|---|---|---|
| Workflow trigger parameters | Keys listed under each workflow's `Parameters` section | Pass them through `variables` in Windows/agent mode or `params` in Server mode. Do not invent keys or send placeholders for required values. |
| Parameter types | `string`, `number`, `json`, `checkbox`/boolean | Send JSON values matching the declared type. For `json`, send an object or array; for checkbox/boolean, send `true` or `false`. |
| `browserflow_output` | Automa variable name | Recommended standard output variable. Workflows must set this variable if the model should read a returned value. |
| `wait_result` | `false`, `true` | `false` dispatches only. `true` waits for final success/error/stopped/timeout and can return variables/table data. |
| `timeout` | Positive integer seconds | Maximum sync wait time. Exported examples recommend `300` seconds. |
| `return_data.variables` | `browserflow_output` or custom Automa variable names | Controls which Automa variables should be returned after sync execution. |
| `return_data.include_table` | `true`, `false` | Return table output when the user asks for rows, table data, list data, or tabular results. |
| `return_data.table_limit` | Positive integer rows | Caps directly returned table rows. Use a small value for summaries and a larger value when the user asks for all visible rows. |
| `return_data.include_history` | `true`, `false` | Include detailed execution history only for debugging or auditing. Keep false for normal data requests. |
| Execution statuses | `queued`, `running`, `success`, `error`, `stopped`, `timeout` | Treat `queued`/`running` as incomplete. Treat `success` as completed. Report readable errors for `error`, `stopped`, or `timeout`. |
| Dispatch target | Empty `client_ip` and `node_id`, or both set | Leave both empty for automatic dispatch to any online node owning the workflow. Set both only when the user requests a specific node. |
| Server result location | `result`, `record`, task record detail, record files | Read immediate `result` first, then `record.result`, then `/task-records/{record_id}` and files for larger tables. |

### Runtime preflight

- Route: GET `/app/runtime`
- Notes: No parameters. Confirms the backend is reachable and running in Server mode.
- Parameters: none.

### List clients

- Route: GET `/clients`
- Notes: Lists connected client nodes. Use it to confirm an online node owns the workflow before dispatch.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `page_num` | `number` | query | no | Page number when pagination is supported. | Positive integer, starts from 1. | `1` |
| `page_size` | `number` | query | no | Page size when pagination is supported. | Allowed by common pagination: 10, 30, or 60. | `30` |

### Query existing tasks

- Route: GET `/tasks`
- Notes: Find reusable tasks before creating a new one.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `page_num` | `number` | query | no | Page number. | Positive integer, starts from 1. | `1` |
| `page_size` | `number` | query | no | Page size. | `10`, `30`, or `60`. | `10` |
| `start_time` | `string` | query | no | Created/updated time filter start. | Datetime string accepted by backend validation. | `2026-05-27 00:00:00` |
| `end_time` | `string` | query | no | Created/updated time filter end. | Datetime string accepted by backend validation. | `2026-05-27 23:59:59` |
| `keyword` | `string` | query | no | Keyword search over task fields. | Any search text. | `Skill task` |
| `workflow_id` | `string` | query | no | Filter by workflow ID. | Automa workflow ID from this Skill. | `7KKfW4mVvDFBGux2kMgft` |
| `workflow_name` | `string` | query | no | Filter by workflow name. | Workflow name text. | `普通检索` |
| `client_id` | `string` | query | no | Filter by client ID. | Server client ID. | `client_...` |
| `machine_id` | `string` | query | no | Filter by machine ID. | Client machine ID. | `machine_...` |
| `node_id` | `string` | query | no | Filter by node ID. | Browser/client node ID. | `node_...` |
| `enabled` | `string` | query | no | Filter by enabled state. | Common values are `true` or `false`. | `true` |

### Create task

- Route: POST `/tasks`
- Notes: Creates a reusable Server-mode task. Use `run_once_after_create:true` only when the user wants to create and immediately run it.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `name` | `string` | body | yes | Task name. | Any readable name. | `Skill task for 普通检索` |
| `description` | `string` | body | no | Task description. | Any readable description. | `Created by BrowserFlow exported Skill` |
| `workflow_id` | `string` | body | yes | Workflow ID to run. | Automa workflow ID from this Skill. | `7KKfW4mVvDFBGux2kMgft` |
| `workflow_name` | `string` | body | no | Workflow display name snapshot. | Workflow name from this Skill. | `普通检索` |
| `client_id` | `string` | body | no | Target client ID. | Use only when targeting a specific client. Prefer leaving empty for automatic dispatch. | `client_...` |
| `client_name` | `string` | body | no | Target client name snapshot. | Readable client name. | `Office PC` |
| `client_ip` | `string` | body | no | Target client IP. | Leave empty for automatic dispatch. If set, also set `node_id` when targeting a specific node. | `192.168.1.10` |
| `machine_id` | `string` | body | no | Target machine ID. | Use only when known and needed. | `machine_...` |
| `node_id` | `string` | body | no | Target execution node ID. | Leave empty for automatic dispatch. If set, also set `client_ip` for specific-node dispatch. | `node_...` |
| `target_group_id` | `number` | body | no | Target node group ID. | Integer group ID when group dispatch is configured. | `1` |
| `dispatch_mode` | `string` | body | no | Dispatch mode. | Use existing project modes when configured; otherwise leave empty for default dispatch. | `` |
| `queue_policy` | `string` | body | no | Queue/busy handling policy. | Use existing project policy values when configured; otherwise leave empty for default behavior. | `` |
| `max_attempts` | `number` | body | no | Maximum retry attempts. | Positive integer. Leave empty/0 for backend default. | `1` |
| `timeout_seconds` | `number` | body | no | Task execution timeout seconds. | Positive integer seconds. Leave empty/0 for backend default. | `300` |
| `queue_wait_seconds` | `number` | body | no | Maximum time to wait for an available node. | Positive integer seconds. | `60` |
| `queue_retry_interval_seconds` | `number` | body | no | Retry interval while waiting for a node. | Positive integer seconds. | `5` |
| `cron_expression` | `string` | body | no | Cron schedule. Empty means no schedule. | Cron expression supported by the backend scheduler. | `` |
| `run_once_after_create` | `boolean` | body | no | Whether to immediately execute once after creating. | `true` or `false`. | `false` |
| `params` | `object` | body | no | Workflow trigger parameters. | Keys must exactly match the workflow `Parameters` section. Values follow parameter type rules. | `{"key_word":"ai智能体"}` |
| `enabled` | `boolean` | body | no | Whether the task is enabled. | `true` or `false`. | `true` |

### Execute existing task

- Route: POST `/tasks/{task_id}/execute`
- Notes: Dispatches an existing task to a client node.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `task_id` | `string` | path | yes | Task ID to execute. | Task id from `/tasks` or create response. | `123` |
| `client_id` | `string` | body | no | Override target client ID for this execution. | Use only when selecting a specific client. | `client_...` |
| `client_ip` | `string` | body | no | Override target client IP for this execution. | Leave empty for automatic dispatch. If set with specific node targeting, also set `node_id`. | `192.168.1.10` |
| `machine_id` | `string` | body | no | Override target machine ID. | Use only when known and needed. | `machine_...` |
| `node_id` | `string` | body | no | Override target node ID. | Leave empty for automatic dispatch. If set, also set `client_ip` for a specific node. | `node_...` |
| `trigger_type` | `string` | body | no | Execution trigger label. | Use `skill` for executions started from this Skill. Other common labels may include manual/schedule/system depending on backend usage. | `skill` |
| `params` | `object` | body | no | Execution parameter override. | Keys must exactly match the workflow `Parameters` section. Overrides task saved params for this run. | `{"key_word":"ai智能体"}` |
| `wait_result` | `boolean` | body | no | Execution mode. | `false`: dispatch only; `true`: wait for final status or timeout. | `true` |
| `timeout` | `number` | body | no | Maximum wait seconds when `wait_result` is true. Default recommendation is 300. | Positive integer seconds. | `300` |
| `return_data` | `object` | body | no | Returned data configuration for sync runs. | Use when the user needs workflow output, variables, table rows, or history. | `{"variables":["browserflow_output"],"include_table":true}` |
| `return_data.variables` | `array<string>` | body | no | Automa variable names to return. | `browserflow_output`: recommended standard variable for workflow output; any custom Automa variable name created by the workflow may also be listed. | `["browserflow_output"]` |
| `return_data.include_table` | `boolean` | body | no | Whether to return Automa table data. | `true` when the user asks for rows/table/list data stored as table output; otherwise `false` for speed. | `true` |
| `return_data.table_limit` | `number` | body | no | Maximum table rows to include directly. | Positive integer row count. | `100` |
| `return_data.include_history` | `boolean` | body | no | Whether to include workflow execution history details. | `true` only for debugging/auditing; normally `false`. | `false` |

### Query task records

- Route: GET `/task-records`
- Notes: Lists task execution records. Use after async execution or when the user asks for historical results.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `page_num` | `number` | query | no | Page number. | Positive integer, starts from 1. | `1` |
| `page_size` | `number` | query | no | Page size. | `10`, `30`, or `60`. | `10` |
| `start_time` | `string` | query | no | Execution record time filter start. | Datetime string accepted by backend validation. | `2026-05-27 00:00:00` |
| `end_time` | `string` | query | no | Execution record time filter end. | Datetime string accepted by backend validation. | `2026-05-27 23:59:59` |
| `task_id` | `string` | query | no | Filter by task ID. | Task ID. | `123` |
| `task_name` | `string` | query | no | Filter by task name. | Task name text. | `Skill task` |
| `workflow_id` | `string` | query | no | Filter by workflow ID. | Automa workflow ID from this Skill. | `7KKfW4mVvDFBGux2kMgft` |
| `workflow_name` | `string` | query | no | Filter by workflow name. | Workflow name text. | `普通检索` |
| `client_id` | `string` | query | no | Filter by client ID. | Server client ID. | `client_...` |
| `client_ip` | `string` | query | no | Filter by one client IP. | Client IP. | `192.168.1.10` |
| `client_ips` | `string` | query | no | Filter by multiple client IPs. | Comma-separated client IPs. | `192.168.1.10,192.168.1.11` |
| `machine_id` | `string` | query | no | Filter by machine ID. | Machine ID. | `machine_...` |
| `node_id` | `string` | query | no | Filter by one node ID. | Node ID. | `node_...` |
| `node_ids` | `string` | query | no | Filter by multiple node IDs. | Comma-separated node IDs. | `node_a,node_b` |
| `status` | `string` | query | no | Filter by execution status. | `queued`, `running`, `success`, `error`, `stopped`, `timeout`. | `success` |
| `keyword` | `string` | query | no | Keyword search over record fields. | Any search text. | `ai智能体` |

### Query task record detail

- Route: GET `/task-records/{record_id}`
- Notes: Reads one execution record and its result files.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `record_id` | `string` | path | yes | Task record ID. | Record id from execute response or `/task-records` list. | `456` |

### Download task record file

- Route: GET `/task-records/files/{file_id}/download`
- Notes: Downloads one result file, commonly table output stored as a file.

| Field | Type | Location | Required | Meaning | Available values / variables | Example |
|---|---|---|---|---|---|---|
| `file_id` | `string` | path | yes | Task record file ID. | File id from task record detail `files` array. | `789` |

## Workflow Listing Response Rules

When the user asks what workflows are available, do not only list names and IDs. For each workflow, also explain how it can be called through Server-mode tasks:

- Show the workflow name, workflow ID, status, dispatch target rules, required parameters, and optional parameters.
- Explain async execution: execute a task without `wait_result: true`; the API dispatches the task and returns task record information, but does not wait for final workflow output.
- Explain sync result mode: set `wait_result` to `true`; set `timeout` to the maximum wait time in seconds. The exported examples use `timeout: 300`, so the default recommendation is to wait up to 300 seconds.
- Explain variable output: add `return_data.variables: ["browserflow_output"]` and read `result.data.variables.browserflow_output`, `record.result.data.variables.browserflow_output`, or the task record detail.
- Explain table output only when needed: set `return_data.include_table` to `true`, choose a `table_limit`, then inspect the task record files if the table is stored as a file.
- If the user wants to start a workflow without waiting, use async execution. If the user wants final data, search results, extracted content, success/failure, returned variables, or table data, use sync result mode.
- Reply in the user's language, but keep API field names exactly as written.

## Dispatch Rules

- Leave both `client_ip` and `node_id` empty when any online node that owns the workflow may execute it.
- Set both `client_ip` and `node_id` when the user explicitly wants a specific execution node.
- Avoid setting only one of `client_ip` or `node_id`; BrowserFlow Server mode identifies execution targets by the pair.
- If the target node is busy, offline, or does not own the workflow, the API creates a failed execution record with a readable reason.
- Per-node Redis locks prevent the same browser node from running multiple Automa workflows concurrently.
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
- Preflight response is not successful: report the response `message` or `error` and do not execute.
- Runtime mode is not `server`: ask the user to start or switch to Server mode and do not execute.
- No online nodes: ask the user to open a Windows client browser-agent page and keep it connected.
- No online node owns the workflow: ask the user to install or sync the workflow to a client node, or choose another workflow.
- Target node is busy or locked: report the busy reason and suggest retrying later or choosing another node.
- Missing required parameters: ask for the missing values before creating or executing a task.
- API returns a failed task record: report `record.error_message` or the readable failure reason from the response.

## API Endpoints

### Query Existing Tasks

```bash
curl 'http://127.0.0.1:8001/api/v1/tasks?workflow_id=XtdNYUUtraUJXUI2SAiRu&page_num=1&page_size=10'
```

### Create Task

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","description":"Created by BrowserFlow exported Skill","enabled":true,"name":"Skill task for XtdNYUUtraUJXUI2SAiRu","node_id":"","params":{"key_word":""},"run_once_after_create":false,"workflow_id":"XtdNYUUtraUJXUI2SAiRu"}'
```

### Execute Existing Task

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","node_id":"","params":{"key_word":""},"trigger_type":"skill"}'
```

### Execute Existing Task And Wait For Result

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/tasks/{task_id}/execute' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","node_id":"","params":{"key_word":""},"return_data":{"include_history":false,"include_table":true,"table_limit":100,"variables":["browserflow_output"]},"timeout":300,"trigger_type":"skill","wait_result":true}'
```

### Query Task Records

```bash
curl 'http://127.0.0.1:8001/api/v1/task-records?workflow_id=XtdNYUUtraUJXUI2SAiRu&page_num=1&page_size=10'
```

### Query Task Record Detail

```bash
curl 'http://127.0.0.1:8001/api/v1/task-records/{record_id}'
```

### Download Task Record File

```bash
curl -O 'http://127.0.0.1:8001/api/v1/task-records/files/{file_id}/download'
```

## Available Workflows

### 1. 单参数检索

- Workflow ID: `XtdNYUUtraUJXUI2SAiRu`
- Server ID: `3`
- Last Sync Client IP: `127.0.0.1`
- Last Sync Node ID: `node-1`
- Description: 单参数检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-19 20:19:40
- Updated: 2026-05-23 23:41:10

Parameters:
1. Parameter `key_word`
   - Key: `key_word`
   - Meaning: 检索关键词
   - Type: `string`
   - Required: Yes
   - Default: `""`

Invocation guide:
- Async execution: execute or create a task without `wait_result: true`. This dispatches the workflow and returns task record information; it does not wait for completion or return final workflow data.
- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.
- Variable result: add `return_data.variables: ["browserflow_output"]` and read `result.data.variables.browserflow_output` first, then the task record detail if needed.
- Table result: set `return_data.include_table: true` and set `table_limit`; inspect task record files when the returned table is stored as a file.

Create task example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for XtdNYUUtraUJXUI2SAiRu","node_id":"","params":{"key_word":""},"run_once_after_create":true,"workflow_id":"XtdNYUUtraUJXUI2SAiRu"}'
```

### 2. 普通检索

- Workflow ID: `Yx03DAsLctZzjj_LDiOCN`
- Server ID: `1`
- Description: 普通检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-23 23:37:56
- Updated: 2026-05-24 00:28:17

Parameters: none detected.

Create task example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/tasks' \
  -H 'Content-Type: application/json' \
  -d '{"client_ip":"","cron_expression":"","enabled":true,"name":"Skill task for Yx03DAsLctZzjj_LDiOCN","node_id":"","params":{},"run_once_after_create":true,"workflow_id":"Yx03DAsLctZzjj_LDiOCN"}'
```

## Usage Notes

- These workflows come from the BrowserFlow server database, not from the currently open browser-agent page.
- Prefer creating reusable tasks for repeated use, then execute those task IDs from the skill.
- Leave both `client_ip` and `node_id` empty when any online node that owns the workflow may execute it.
- Set both `client_ip` and `node_id` only when the user explicitly wants a specific execution node.
- A successful execute response means the server accepted and dispatched the task. Use task records to inspect final status and returned data.
