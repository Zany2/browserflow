---
name: browserflow-automa-workflows
description: "Run BrowserFlow Automa workflows through the local browser-agent HTTP API. Workflows include: bilibili表格返回值测试, Google Keyword Research, bilibili必填参数测试, Generate lorem ipsum, Google search, bilibili返回值测试, Twitter Trends to Google Sheets, Search in ProductHunt"
---

# BrowserFlow Automa Workflows

## Overview

This skill describes Automa workflows that are currently available from a BrowserFlow browser-agent. Use the BrowserFlow HTTP API to open or run these workflows in the browser instance that exported them.

**Total Workflows Available:** 8

**Recommended Filename:** `SKILL_AUTOMA.md`

**API Base URL:** `http://127.0.0.1:8001/api/v1`

**Browser Instance ID:** `browser_q84m3z011jcdidyac75yof0100ezgm2l`

## Mandatory Preflight

Before running any workflow, first verify that the BrowserFlow backend is reachable.

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

## Parameter Rules

Before running a workflow, inspect its `Parameters` section. If a required parameter has no value, ask the user for it before calling the API. If an optional parameter has a default value, use the default unless the user provides another value. Pass parameters through the `variables` object, and keep parameter names exactly as listed in this skill. BrowserFlow treats this `variables` object as the completed parameter set and instructs Automa not to open its own parameter input page.

## Execution Mode Rules

Before running a workflow, decide whether the user needs the final workflow result.

- Use asynchronous execution when the user only asks to start, trigger, submit, open, launch, run, or execute a task and does not ask for returned data or final completion. Set `wait_result` to `false`. Return the `execution.execution_id` to the user so they can query status or results later.
- Use synchronous waiting when the user asks to get, query, search, extract, collect, return, fetch, read, wait for completion, or confirm final success/failure. Set `wait_result` to `true`, set a reasonable `timeout`, and request returned data if needed.
- For data-returning requests, request `return_data.variables: ["browserflow_output"]` and read `execution.result.data.variables.browserflow_output` first. If it is missing, report that the workflow completed but did not provide a BrowserFlow output variable.
- If the user asks to check a previous task, use the execution status endpoint with the saved `execution_id` instead of running the workflow again.
- If the intent is ambiguous, prefer async mode for action-only tasks and sync mode for data-returning tasks.

## API Endpoints

### Run Workflow Async

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/9LJuEPNiZjfz1698i7-Yr/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"key_word":""},"wait_result":false}'
```

### Run Workflow And Wait For Result

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/9LJuEPNiZjfz1698i7-Yr/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","return_data":{"include_history":false,"include_table":false,"table_limit":20,"variables":["browserflow_output"]},"timeout":300,"variables":{"key_word":""},"wait_result":true}'
```

### Query Execution Status

```bash
curl 'http://127.0.0.1:8001/api/v1/workflows/executions/{execution_id}'
```

### Open Workflow Editor

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/9LJuEPNiZjfz1698i7-Yr/open' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l"}'
```

## Available Workflows

### 1. bilibili表格返回值测试

- ID: `9LJuEPNiZjfz1698i7-Yr`
- Description: bilibili表格返回值测试
- Status: enabled
- Nodes: 8
- Created: 2026-05-19 21:00:13
- Updated: 2026-05-19 20:58:23

Parameters:
- `key_word` (string, required): 关键词 Default: `""`

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/9LJuEPNiZjfz1698i7-Yr/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 2. Google Keyword Research

- ID: `Q3hHwu7U34ncZz983Ai3c`
- Description: -
- Status: enabled
- Nodes: 11
- Created: 2026-05-19 20:59:22
- Updated: 2026-05-19 20:59:22

Parameters:
- `keywords` (string): keyword 1,keyword 2 Default: `""`

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/Q3hHwu7U34ncZz983Ai3c/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"keywords":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 3. bilibili必填参数测试

- ID: `Rjm5rLGNY2It8jZ2G1P2e`
- Description: bilibili必填参数测试
- Status: enabled
- Nodes: 5
- Created: 2026-05-19 21:00:13
- Updated: 2026-05-19 20:21:53

Parameters:
- `key_word` (string, required): 关键词 Default: `""`

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/Rjm5rLGNY2It8jZ2G1P2e/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 4. Generate lorem ipsum

- ID: `aBi3vZCB1Drv1y4dOsM1o`
- Description: -
- Status: enabled
- Nodes: 11
- Created: 2026-05-19 20:59:22
- Updated: 2026-05-19 20:59:22

Parameters: none detected.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/aBi3vZCB1Drv1y4dOsM1o/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 5. Google search

- ID: `c3ICcGVv8GNPFrs0E4rf7`
- Description: -
- Status: enabled
- Nodes: 4
- Created: 2026-05-19 20:59:22
- Updated: 2026-05-19 20:59:22

Parameters: none detected.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/c3ICcGVv8GNPFrs0E4rf7/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 6. bilibili返回值测试

- ID: `oMuLzzaXycLXlP-8rzIHg`
- Description: bilibili返回值测试
- Status: enabled
- Nodes: 7
- Created: 2026-05-19 21:00:13
- Updated: 2026-05-19 20:33:03

Parameters:
- `key_word` (string, required): 关键词 Default: `""`

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/oMuLzzaXycLXlP-8rzIHg/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"key_word":""},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 7. Twitter Trends to Google Sheets

- ID: `ueAjEkT4Eqhl8BBCqyRiM`
- Description: Import current twitter trends to Google Sheets
- Status: enabled
- Nodes: 8
- Created: 2026-05-19 20:59:22
- Updated: 2026-05-19 20:59:22

Parameters: none detected.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/ueAjEkT4Eqhl8BBCqyRiM/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

### 8. Search in ProductHunt

- ID: `yNfm8xtoIkNkclXPd3Wnt`
- Description: -
- Status: enabled
- Nodes: 2
- Created: 2026-05-19 20:59:22
- Updated: 2026-05-19 20:59:22

Parameters: none detected.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/yNfm8xtoIkNkclXPd3Wnt/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{},"wait_result":false}'
```

For data-returning requests, use the sync example in the API Endpoints section and keep the same workflow ID and variables.

## Usage Notes

- Keep the target browser running and keep the browser-agent page connected before calling the API.
- The target browser must report `automa_installed: true`; otherwise workflow list, open, and run commands may fail.
- Pass trigger parameters through the `variables` object. Parameter names must match the Automa trigger configuration.
- Do not rely on Automa's parameter tab for Skill calls; collect required values before sending the HTTP request.
- If the exported browser instance is no longer available, choose another running browser and update `browser_id`.
- Async run returns after the command is accepted. Sync run waits until Automa reports `success`, `error`, `stopped`, or BrowserFlow reports `timeout`.
