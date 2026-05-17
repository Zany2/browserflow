package browserexecutor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
)

// Batch executes browser operations sequentially. 顺序执行浏览器操作。
func (e *Executor) Batch(ctx context.Context, operations []model.BrowserExecutorBatchAction) (*model.BrowserExecutorOperationResult, error) {
	results := make([]map[string]any, 0, len(operations))
	var err error

	for _, op := range operations {
		var result *model.BrowserExecutorOperationResult
		switch strings.ToLower(strings.TrimSpace(op.Type)) {
		case "navigate":
			url, _ := op.Params["url"].(string)
			timeout := jsonNumberToInt(op.Params["timeout"], 60)
			result, err = e.Navigate(ctx, url, timeout)
		case "click":
			identifier, _ := op.Params["identifier"].(string)
			result, err = e.Click(ctx, identifier)
		case "type":
			identifier, _ := op.Params["identifier"].(string)
			text, _ := op.Params["text"].(string)
			clear := jsonBool(op.Params["clear"], true)
			result, err = e.Type(ctx, identifier, text, clear)
		case "select":
			identifier, _ := op.Params["identifier"].(string)
			value, _ := op.Params["value"].(string)
			result, err = e.Select(ctx, identifier, value)
		case "press-key":
			key, _ := op.Params["key"].(string)
			result, err = e.PressKey(ctx, key, jsonBool(op.Params["ctrl"], false), jsonBool(op.Params["shift"], false), jsonBool(op.Params["alt"], false), jsonBool(op.Params["meta"], false))
		case "wait":
			identifier, _ := op.Params["identifier"].(string)
			state, _ := op.Params["state"].(string)
			timeout := jsonNumberToInt(op.Params["timeout"], 10)
			count := jsonNumberToInt(op.Params["count"], 0)
			result, err = e.Wait(ctx, identifier, state, timeout, count)
		case "snapshot":
			result, err = e.batchSnapshot(ctx)
		case "clickable-elements":
			limit := jsonNumberToInt(op.Params["limit"], 50)
			result, err = e.ElementRefs(ctx, "clickable", limit)
		case "input-elements":
			limit := jsonNumberToInt(op.Params["limit"], 30)
			result, err = e.ElementRefs(ctx, "input", limit)
		case "observe":
			includeText := jsonBool(op.Params["include_text"], false)
			textLimit := jsonNumberToInt(op.Params["text_limit"], 8000)
			observe, observeErr := e.Observe(ctx, includeText, textLimit)
			err = observeErr
			result = &model.BrowserExecutorOperationResult{
				Success:   observe.Success,
				Message:   "Observe completed",
				Error:     observe.Error,
				Timestamp: observe.Timestamp,
				Data: map[string]interface{}{
					"status":        observe.Status,
					"page_info":     observe.PageInfo,
					"snapshot_text": observe.SnapshotText,
					"elements":      observe.Elements,
					"page_text":     observe.PageText,
				},
			}
		case "page-text":
			limit := jsonNumberToInt(op.Params["limit"], 8000)
			result, err = e.PageText(ctx, limit)
		case "page-content":
			limit := jsonNumberToInt(op.Params["limit"], 0)
			result, err = e.PageContent(ctx, limit)
		case "get-value":
			identifier, _ := op.Params["identifier"].(string)
			result, err = e.GetValue(ctx, identifier)
		case "element-info":
			identifier, _ := op.Params["identifier"].(string)
			result, err = e.ElementInfo(ctx, identifier, jsonStringSlice(op.Params["attributes"]))
		case "page-structure":
			result, err = e.PageStructure(ctx, model.BrowserExecutorPageStructureOptions{
				IncludeLinks:   jsonBool(op.Params["include_links"], true),
				IncludeForms:   jsonBool(op.Params["include_forms"], true),
				IncludeTables:  jsonBool(op.Params["include_tables"], true),
				IncludeImages:  jsonBool(op.Params["include_images"], true),
				IncludeButtons: jsonBool(op.Params["include_buttons"], true),
				Limit:          jsonNumberToInt(op.Params["limit"], 50),
			})
		case "scroll":
			direction, _ := op.Params["direction"].(string)
			identifier, _ := op.Params["identifier"].(string)
			pixels := jsonNumberToInt(op.Params["pixels"], 700)
			result, err = e.Scroll(ctx, direction, pixels, identifier)
		case "reload":
			result, err = e.Reload(ctx)
		case "go-back":
			result, err = e.GoBack(ctx)
		case "go-forward":
			result, err = e.GoForward(ctx)
		case "hover":
			identifier, _ := op.Params["identifier"].(string)
			result, err = e.Hover(ctx, identifier)
		case "resize":
			width := jsonNumberToInt(op.Params["width"], 1440)
			height := jsonNumberToInt(op.Params["height"], 900)
			result, err = e.Resize(ctx, width, height)
		case "element-screenshot":
			identifier, _ := op.Params["identifier"].(string)
			format, _ := op.Params["format"].(string)
			quality := jsonNumberToInt(op.Params["quality"], 0)
			result, err = e.ElementScreenshot(ctx, identifier, format, quality)
		case "mouse":
			action, _ := op.Params["action"].(string)
			button, _ := op.Params["button"].(string)
			result, err = e.Mouse(ctx, model.BrowserExecutorMouseOptions{
				Action: action,
				X:      jsonNumberToFloat(op.Params["x"], 0),
				Y:      jsonNumberToFloat(op.Params["y"], 0),
				DeltaX: jsonNumberToFloat(op.Params["delta_x"], 0),
				DeltaY: jsonNumberToFloat(op.Params["delta_y"], 0),
				Steps:  jsonNumberToInt(op.Params["steps"], 10),
				Button: button,
			})
		case "window":
			action, _ := op.Params["action"].(string)
			result, err = e.Window(ctx, model.BrowserExecutorWindowOptions{
				Action: action,
				Left:   jsonNumberToInt(op.Params["left"], 0),
				Top:    jsonNumberToInt(op.Params["top"], 0),
				Width:  jsonNumberToInt(op.Params["width"], 0),
				Height: jsonNumberToInt(op.Params["height"], 0),
			})
		case "close-page":
			result, err = e.ClosePage(ctx)
		case "fill-form":
			fields := jsonFormFields(op.Params["fields"])
			submit := jsonBool(op.Params["submit"], false)
			timeout := jsonNumberToInt(op.Params["timeout"], 10)
			result, err = e.FillForm(ctx, fields, submit, timeout)
		case "drag":
			fromIdentifier, _ := op.Params["from_identifier"].(string)
			toIdentifier, _ := op.Params["to_identifier"].(string)
			result, err = e.Drag(ctx, fromIdentifier, toIdentifier)
		case "file-upload":
			identifier, _ := op.Params["identifier"].(string)
			result, err = e.FileUpload(ctx, identifier, jsonStringSlice(op.Params["file_paths"]))
		case "handle-dialog":
			accept := jsonBool(op.Params["accept"], true)
			text, _ := op.Params["text"].(string)
			timeout := jsonNumberToInt(op.Params["timeout"], 10)
			result, err = e.HandleDialog(ctx, accept, text, timeout)
		case "tabs":
			action, _ := op.Params["action"].(string)
			url, _ := op.Params["url"].(string)
			index := jsonNumberToInt(op.Params["index"], 0)
			result, err = e.Tabs(ctx, action, url, index)
		case "act":
			identifier, _ := op.Params["identifier"].(string)
			intent, _ := op.Params["intent"].(string)
			text, _ := op.Params["text"].(string)
			result, err = e.Act(ctx, model.BrowserExecutorActOptions{
				Intent:     intent,
				Identifier: identifier,
				Value:      op.Params["value"],
				Text:       text,
				Fields:     jsonFormFields(op.Params["fields"]),
				Submit:     jsonBool(op.Params["submit"], false),
				Clear:      jsonBool(op.Params["clear"], true),
				Timeout:    jsonNumberToInt(op.Params["timeout"], 10),
			})
		default:
			result = &model.BrowserExecutorOperationResult{Success: false, Error: "unknown batch operation: " + op.Type, Timestamp: time.Now()}
			err = fmt.Errorf("unknown batch operation: %s", op.Type)
		}

		result, observeErr := e.AppendObserve(ctx, result, jsonBool(op.Params["return_observe"], false), jsonBool(op.Params["include_text"], false), jsonNumberToInt(op.Params["text_limit"], 8000))
		if err == nil {
			err = observeErr
		}

		results = append(results, map[string]any{"type": op.Type, "result": result})
		if err != nil && op.StopOnError {
			break
		}
	}

	return &model.BrowserExecutorOperationResult{
		Success:   true,
		Message:   "Batch completed",
		Data:      map[string]interface{}{"results": results},
		Timestamp: time.Now(),
	}, nil
}

func (e *Executor) batchSnapshot(ctx context.Context) (*model.BrowserExecutorOperationResult, error) {
	snapshot, err := e.Snapshot(ctx)
	return &model.BrowserExecutorOperationResult{
		Success:   snapshot.Success,
		Message:   "Snapshot completed",
		Error:     snapshot.Error,
		Timestamp: snapshot.Timestamp,
		Data: map[string]interface{}{
			"snapshot_text": snapshot.SnapshotText,
			"elements":      snapshot.Elements,
		},
	}, err
}

func jsonNumberToInt(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed)
		}
	}
	return fallback
}

func jsonNumberToFloat(value any, fallback float64) float64 {
	switch typed := value.(type) {
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case float64:
		return typed
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed
		}
	}
	return fallback
}

func jsonBool(value any, fallback bool) bool {
	if value == nil {
		return fallback
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	return fallback
}

func jsonStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return items
	}
	return nil
}

func jsonFormFields(value any) []model.BrowserExecutorFormField {
	rawItems, ok := value.([]any)
	if !ok {
		return nil
	}
	fields := make([]model.BrowserExecutorFormField, 0, len(rawItems))
	for _, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := item["name"].(string)
		fieldType, _ := item["type"].(string)
		fields = append(fields, model.BrowserExecutorFormField{
			Name:  name,
			Value: item["value"],
			Type:  fieldType,
		})
	}
	return fields
}
