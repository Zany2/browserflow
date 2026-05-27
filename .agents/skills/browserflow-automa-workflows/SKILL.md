---
name: browserflow-automa-workflows
description: "Run BrowserFlow Automa workflows through the local browser-agent HTTP API. Workflows include: 单参数检索返回值变量与表格, 单参数检索返回值变量, 单参数检索, 普通检索, 多参数检索"
---

# BrowserFlow Automa Workflows

## Overview

This skill describes Automa workflows that are currently available from a BrowserFlow browser-agent. Use the BrowserFlow HTTP API to open or run these workflows in the browser instance that exported them.

**Total Workflows Available:** 5

**Recommended Filename:** `SKILL.md`

**API Base URL:** `http://127.0.0.1:8001/api/v1`

**Browser Instance ID:** `browser_1tgxl525rc0disnlu24vqzo500bfjhb2`

## Quick Start

Use this Skill with a compact workflow-call loop:

1. Choose the workflow that matches the user's request.
2. Inspect the workflow parameters and collect missing required values.
3. Decide async submit or sync result mode based on whether the user needs returned data.
4. Call the run API with the exported `browser_id` and the exact `variables` object.
5. For sync runs, read returned variables, table data, status, and errors before answering.

## Skill Map

- `Lightweight Preflight Strategy`: when to verify backend and agent status.
- `Required Step-By-Step Procedure`: required order before running a workflow.
- `Parameter Rules`: how to pass trigger parameters through `variables`.
- `Workflow Listing Response Rules`: how to explain available workflows to users.
- `Execution Mode Rules`: async submit vs sync result mode.
- `API Endpoints`: reusable HTTP examples.
- `Available Workflows`: workflow IDs, parameters, and per-workflow examples.

## Lightweight Preflight Strategy

Use a lightweight preflight strategy instead of repeating every check before every call:

- Before the first workflow call in a conversation, verify `/app/runtime` and `/agents/status`.
- Reuse a successful preflight result for short follow-up calls in the same conversation, especially when they target the same exported `browser_id`.
- For sync result mode, long-running workflows, returned data, or table output, perform preflight unless a recent successful preflight is already available.
- For simple async submit requests, it is acceptable to call directly after a recent successful preflight; if the API fails, inspect status and report the exact reason.
- If any call returns an offline, disconnected, missing Automa, timeout, or unreachable backend error, run preflight again before retrying.

## Mandatory Preflight

For the first workflow call, or when no recent successful preflight is available, verify that the BrowserFlow backend is reachable.

```bash
curl 'http://127.0.0.1:8001/api/v1/app/runtime'
```

If the request fails, ask the user to start the BrowserFlow backend before continuing.

Then verify that the browser instance exported with this skill is online.

```bash
curl 'http://127.0.0.1:8001/api/v1/agents/status'
```

Find an agent whose `browser_id` matches the Browser Instance ID in this skill. It must be online. If it is missing or offline, ask the user to start that exact browser instance and keep the browser-agent page connected.

After confirming the agent is online, verify that its Automa plugin status reports `automa_installed: true`. If Automa is not installed or not available, ask the user to install or enable the Automa extension in that browser instance, then refresh the browser-agent page before continuing.

Do not replace the exported `browser_id` with the current browser unless the user explicitly confirms that the workflow exists in the new browser instance.
Do not close the browser-agent tab that exported this Skill. It is the control channel for workflow detection and execution. When closing business tabs, keep every BrowserFlow `browser-agent` tab open unless the user explicitly asks to stop that browser instance.

## Required Step-By-Step Procedure

Always work in this order. Do not run or open a workflow until the needed checks are complete.

1. Decide preflight depth: use a recent successful preflight for short follow-up calls; otherwise call `/app/runtime` and `/agents/status`.
2. Detect client when needed: find the exported `browser_id`, confirm it is online, and confirm `automa_installed: true`.
3. Plan first: break the user request into concrete steps, choose the best matching workflow from this Skill, inspect its `Parameters`, and ask the user for any missing required values.
4. Decide execution mode: use async for action-only requests, sync for requests that need returned data or final completion.
5. Execute only after steps 1-4 pass. If any check fails, stop and report the exact reason instead of calling the run API.
6. For sync runs, inspect the execution result before answering. For async runs, return the `execution_id` and explain that status can be queried later.

## Parameter Rules

Before running a workflow, inspect its `Parameters` section. If a required parameter has no value, ask the user for it before calling the API. If an optional parameter has a default value, use the default unless the user provides another value. Pass parameters through the `variables` object, and keep parameter names exactly as listed in this skill. BrowserFlow treats this `variables` object as the completed parameter set and instructs Automa not to open its own parameter input page.

## Parameter Type Rules

- `string`: pass a string value.
- `number`: pass a JSON number, not a quoted string, when the user provides a numeric value.
- `json`: pass a valid JSON object or array. If the user provides plain text, ask them to confirm the JSON structure before executing.
- `checkbox` or boolean parameters: pass a boolean `true` or `false`.
- Example values like `""` are placeholders. Replace required placeholders with real user-provided values before executing.

## Workflow Listing Response Rules

When the user asks what workflows are available, do not only list names and IDs. For each workflow, also explain how it can be called:

- Show the workflow name, ID, status, required parameters, and optional parameters.
- Explain async submit mode: set `wait_result` to `false`; the API only dispatches the workflow and returns `execution.execution_id`; it does not wait for completion and does not return workflow output.
- Explain sync result mode: set `wait_result` to `true`; set `timeout` to the maximum wait time in seconds. The exported examples use `timeout: 300`, so the default recommendation is to wait up to 300 seconds.
- Explain variable output: add `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` or `result.data.variables.browserflow_output`.
- Explain table output only when needed: set `return_data.include_table` to `true` and choose a `table_limit`.
- If the user wants to start a workflow without waiting, use async submit mode. If the user wants final data, search results, extracted content, success/failure, or returned variables, use sync result mode.
- Reply in the user's language, but keep API field names exactly as written.

## Execution Mode Rules

Before running a workflow, decide whether the user needs the final workflow result.

- Use asynchronous execution when the user only asks to start, trigger, submit, open, launch, run, or execute a task and does not ask for returned data or final completion. Set `wait_result` to `false`. Return the `execution.execution_id` to the user so they can query status or results later.
- Use synchronous waiting when the user asks to get, query, search, extract, collect, return, fetch, read, wait for completion, or confirm final success/failure. Set `wait_result` to `true`, set a reasonable `timeout`, and request returned data if needed.
- For data-returning requests, request `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` first. If it is missing, report that the workflow completed but did not provide a BrowserFlow output variable.
- If the user asks to check a previous task, use the execution status endpoint with the saved `execution_id` instead of running the workflow again.
- If the intent is ambiguous, prefer async mode for action-only tasks and sync mode for data-returning tasks.

## Result Reading Rules

- First read the run response. It may contain `result` for the immediate browser-agent result and `execution` for BrowserFlow's execution state.
- For sync runs, check `result.data.variables.browserflow_output`, then `execution.result.data.variables.browserflow_output` before answering data-returning requests.
- For async runs, save `execution.execution_id` and call `/workflows/executions/{execution_id}` when the user asks for status or results.
- Treat `running` and `timeout` as not-final-success states. Query again later for `running`, and report the timeout reason for `timeout`.
- If the response contains an error status or readable error message, report that message instead of only saying the workflow failed.

## Failure Handling Rules

- Backend unreachable: ask the user to start BrowserFlow and do not execute.
- Exported `browser_id` is offline or missing: ask the user to open that exact browser instance and keep the browser-agent page connected.
- Automa is not installed or unavailable: ask the user to install or enable the Automa extension, then refresh the browser-agent page.
- Missing required parameters: ask for the missing values before calling the run API.
- Sync execution has no `browserflow_output`: report that the workflow completed but did not provide the expected output variable.
- API returns an execution error: report the readable error message from the response and do not claim success.

## API Endpoints

### Run Workflow Async

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/7KKfW4mVvDFBGux2kMgft/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","variables":{"key_word":""},"wait_result":false}'
```

### Run Workflow And Wait For Result

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/7KKfW4mVvDFBGux2kMgft/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","return_data":{"include_history":false,"include_table":false,"table_limit":20,"variables":["browserflow_output"]},"timeout":300,"variables":{"key_word":""},"wait_result":true}'
```

### Query Execution Status

```bash
curl 'http://127.0.0.1:8001/api/v1/workflows/executions/{execution_id}'
```

### Open Workflow Editor

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/7KKfW4mVvDFBGux2kMgft/open' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2"}'
```

## Available Workflows

### 1. 单参数检索返回值变量与表格

- ID: `7KKfW4mVvDFBGux2kMgft`
- Description: 单参数检索返回值变量与表格
- Status: enabled
- Nodes: 9
- Created: 2026-05-23 23:44:00
- Updated: 2026-05-23 23:44:57

Parameters:
1. Parameter `key_word`
   - Key: `key_word`
   - Meaning: 检索关键词
   - Type: `string`
   - Required: Yes
   - Default: `""`

Invocation guide:
- Async submit: use `wait_result: false`. This starts the workflow and returns `execution.execution_id`; it does not wait for completion or return workflow data.
- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.
- Variable result: add `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` first.
- Table result: set `return_data.include_table: true` and set `table_limit` when the user asks for table data.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/7KKfW4mVvDFBGux2kMgft/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","variables":{"key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 2. 单参数检索返回值变量

- ID: `L-zNN7CUmhQB2l6imY31W`
- Description: 单参数检索返回值变量
- Status: enabled
- Nodes: 9
- Created: 2026-05-19 20:42:51
- Updated: 2026-05-27 10:16:36

Parameters:
1. Parameter `key_word`
   - Key: `key_word`
   - Meaning: 检索关键词
   - Type: `string`
   - Required: Yes
   - Default: `""`

Invocation guide:
- Async submit: use `wait_result: false`. This starts the workflow and returns `execution.execution_id`; it does not wait for completion or return workflow data.
- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.
- Variable result: add `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` first.
- Table result: set `return_data.include_table: true` and set `table_limit` when the user asks for table data.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/L-zNN7CUmhQB2l6imY31W/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","variables":{"key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 3. 单参数检索

- ID: `XtdNYUUtraUJXUI2SAiRu`
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
- Async submit: use `wait_result: false`. This starts the workflow and returns `execution.execution_id`; it does not wait for completion or return workflow data.
- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.
- Variable result: add `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` first.
- Table result: set `return_data.include_table: true` and set `table_limit` when the user asks for table data.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/XtdNYUUtraUJXUI2SAiRu/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","variables":{"key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 4. 普通检索

- ID: `Yx03DAsLctZzjj_LDiOCN`
- Description: 普通检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-23 23:37:56
- Updated: 2026-05-26 23:27:24

Parameters: none detected.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/Yx03DAsLctZzjj_LDiOCN/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","variables":{},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 5. 多参数检索

- ID: `wbI3CSCL4hRh8xbuTBD5h`
- Description: 多参数检索
- Status: enabled
- Nodes: 7
- Created: 2026-05-24 00:28:59
- Updated: 2026-05-24 00:31:00

Parameters:
1. Parameter `key_word`
   - Key: `key_word`
   - Meaning: 检索关键词
   - Type: `string`
   - Required: Yes
   - Default: `""`
2. Parameter `age`
   - Key: `age`
   - Meaning: 年龄
   - Type: `number`
   - Required: Yes
   - Default: `""`
3. Parameter `desc`
   - Key: `desc`
   - Meaning: 档案
   - Type: `json`
   - Required: No
   - Default: `""`
4. Parameter `is_enable`
   - Key: `is_enable`
   - Meaning: 是否上报
   - Type: `checkbox`
   - Required: No
   - Default: `""`

Invocation guide:
- Async submit: use `wait_result: false`. This starts the workflow and returns `execution.execution_id`; it does not wait for completion or return workflow data.
- Sync result: use `wait_result: true` with `timeout: 300` unless the user asks for a different maximum wait time. This waits up to 300 seconds for a final status.
- Variable result: add `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` first.
- Table result: set `return_data.include_table: true` and set `table_limit` when the user asks for table data.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/wbI3CSCL4hRh8xbuTBD5h/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_1tgxl525rc0disnlu24vqzo500bfjhb2","variables":{"age":"","desc":"","is_enable":"","key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

## Usage Notes

- Keep the target browser running and keep the browser-agent page connected before calling the API.
- Never close the browser-agent tab while using this Skill. Close only ordinary business tabs unless the user explicitly asks to stop BrowserFlow control for that browser.
- The target browser must report `automa_installed: true`; otherwise workflow list, open, and run commands may fail.
- Pass trigger parameters through the `variables` object. Parameter names must match the Automa trigger configuration.
- Do not rely on Automa's parameter tab for Skill calls; collect required values before sending the HTTP request.
- If the exported browser instance is no longer available, update `browser_id` only after the user confirms the same workflow exists in another running browser.
- Async run returns after the command is accepted. Sync run waits until Automa reports `success`, `error`, `stopped`, or BrowserFlow reports `timeout`.
