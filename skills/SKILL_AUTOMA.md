---
name: browserflow-automa-workflows
description: "Run BrowserFlow Automa workflows through the local browser-agent HTTP API. Workflows include: bilibili测试, baidu测试"
---

# BrowserFlow Automa Workflows

## Overview

This skill describes Automa workflows that are currently available from a BrowserFlow browser-agent. Use the BrowserFlow HTTP API to open or run these workflows in the browser instance that exported them.

**Total Workflows Available:** 2

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

## API Endpoints

### Run Workflow

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/Wpo4FUxYuiTzOdX5F1blF/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"id_card":"111","name":""}}'
```

### Open Workflow Editor

```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/Wpo4FUxYuiTzOdX5F1blF/open' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l"}'
```

## Available Workflows

### 1. bilibili测试

- ID: `Wpo4FUxYuiTzOdX5F1blF`
- Description: bilibili测试
- Status: enabled
- Nodes: 6
- Created: 2026-05-17 00:47:21
- Updated: 2026-05-17 00:52:49

Parameters:
- `id_card` (string): 身份证号 Default: `111`
- `name` (string): 姓名 Default: `""`

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/Wpo4FUxYuiTzOdX5F1blF/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{"id_card":"111","name":""}}'
```

### 2. baidu测试

- ID: `yA6L9FqA7zAOzRp-YaFUp`
- Description: baidu测试
- Status: enabled
- Nodes: 4
- Created: 2026-05-17 00:47:21
- Updated: 2026-05-17 00:50:36

Parameters: none detected.

Run example:
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/workflows/yA6L9FqA7zAOzRp-YaFUp/run' \
  -H 'Content-Type: application/json' \
  -d '{"browser_id":"browser_q84m3z011jcdidyac75yof0100ezgm2l","variables":{}}'
```

## Usage Notes

- Keep the target browser running and keep the browser-agent page connected before calling the API.
- The target browser must report `automa_installed: true`; otherwise workflow list, open, and run commands may fail.
- Pass trigger parameters through the `variables` object. Parameter names must match the Automa trigger configuration.
- Do not rely on Automa's parameter tab for Skill calls; collect required values before sending the HTTP request.
- If the exported browser instance is no longer available, choose another running browser and update `browser_id`.
- The run API confirms that the command was sent to the browser-agent. It does not currently prove that the Automa workflow completed successfully.
