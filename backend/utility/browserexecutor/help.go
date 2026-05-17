package browserexecutor

import (
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
)

const BasePath = "/api/v1/browser-executor"

// Help returns available executor commands. Help returns available executor commands.
func Help(command string) []model.BrowserExecutorHelpCommand {
	commands := []model.BrowserExecutorHelpCommand{
		{Name: "status", Method: "GET", Path: "/status", Description: "Check whether the current browser is controllable"},
		{Name: "help", Method: "GET/POST", Path: "/help", Description: "Show command help"},
		{Name: "export_skill", Method: "GET", Path: "/export/skill", Description: "Export Browser Executor Skill"},
		{Name: "navigate", Method: "POST", Path: "/navigate", Description: "Open URL", Parameters: []string{"url", "timeout"}, Example: map[string]any{"url": "https://example.com"}},
		{Name: "snapshot", Method: "GET/POST", Path: "/snapshot", Description: "Get page snapshot and RefIDs"},
		{Name: "clickable_elements", Method: "GET/POST", Path: "/clickable-elements", Description: "Get compact clickable element RefIDs", Parameters: []string{"limit"}},
		{Name: "input_elements", Method: "GET/POST", Path: "/input-elements", Description: "Get compact input element RefIDs", Parameters: []string{"limit"}},
		{Name: "observe", Method: "GET/POST", Path: "/observe", Description: "Get compact page context for LLMs", Parameters: []string{"include_text", "text_limit"}},
		{Name: "act", Method: "POST", Path: "/act", Description: "Run a smart action by intent", Parameters: []string{"intent", "identifier", "value", "text", "fields", "return_observe"}},
		{Name: "click", Method: "POST", Path: "/click", Description: "Click element", Parameters: []string{"identifier"}, Example: map[string]any{"identifier": "@e1"}},
		{Name: "type", Method: "POST", Path: "/type", Description: "Type text", Parameters: []string{"identifier", "text", "clear"}, Example: map[string]any{"identifier": "@e2", "text": "hello", "clear": true}},
		{Name: "select", Method: "POST", Path: "/select", Description: "Select dropdown option", Parameters: []string{"identifier", "value"}, Example: map[string]any{"identifier": "@e3", "value": "China"}},
		{Name: "press_key", Method: "POST", Path: "/press-key", Description: "Send key or shortcut", Parameters: []string{"key", "ctrl", "shift", "alt", "meta"}, Example: map[string]any{"key": "Enter"}},
		{Name: "wait", Method: "POST", Path: "/wait", Description: "Wait for page load, element states, DOM stability, network idle, or fixed time", Parameters: []string{"identifier", "state", "timeout", "count"}},
		{Name: "reload", Method: "POST", Path: "/reload", Description: "Reload current page"},
		{Name: "go_back", Method: "POST", Path: "/go-back", Description: "Go back in browser history"},
		{Name: "go_forward", Method: "POST", Path: "/go-forward", Description: "Go forward in browser history"},
		{Name: "hover", Method: "POST", Path: "/hover", Description: "Hover over element", Parameters: []string{"identifier"}},
		{Name: "resize", Method: "POST", Path: "/resize", Description: "Resize viewport", Parameters: []string{"width", "height"}},
		{Name: "page_info", Method: "GET/POST", Path: "/page-info", Description: "Get current page URL and title"},
		{Name: "page_text", Method: "GET/POST", Path: "/page-text", Description: "Get compact visible page text", Parameters: []string{"limit"}},
		{Name: "page_content", Method: "GET/POST", Path: "/page-content", Description: "Get compact page HTML", Parameters: []string{"limit"}},
		{Name: "page_structure", Method: "GET/POST", Path: "/page-structure", Description: "Get compact structured headings, links, forms, tables, images, and buttons", Parameters: []string{"include_links", "include_forms", "include_tables", "include_images", "include_buttons", "limit"}},
		{Name: "get_text", Method: "POST", Path: "/get-text", Description: "Get element text", Parameters: []string{"identifier"}},
		{Name: "get_value", Method: "POST", Path: "/get-value", Description: "Get element value", Parameters: []string{"identifier"}},
		{Name: "element_info", Method: "POST", Path: "/element-info", Description: "Get element diagnostics including text, attributes, state, box, and XPath", Parameters: []string{"identifier", "attributes"}},
		{Name: "extract", Method: "POST", Path: "/extract", Description: "Extract page text or selector data", Parameters: []string{"selector", "fields", "multiple"}},
		{Name: "screenshot", Method: "POST", Path: "/screenshot", Description: "Capture screenshot as base64", Parameters: []string{"full_page", "format", "quality"}},
		{Name: "element_screenshot", Method: "POST", Path: "/element-screenshot", Description: "Capture a single element screenshot as base64", Parameters: []string{"identifier", "format", "quality"}},
		{Name: "evaluate", Method: "POST", Path: "/evaluate", Description: "Execute JavaScript with automatic function wrapping", Parameters: []string{"script"}},
		{Name: "tabs", Method: "POST", Path: "/tabs", Description: "Manage tabs list/new/switch/close", Parameters: []string{"action", "url", "index"}},
		{Name: "scroll", Method: "POST", Path: "/scroll", Description: "Scroll page or element", Parameters: []string{"direction", "pixels", "identifier"}},
		{Name: "mouse", Method: "POST", Path: "/mouse", Description: "Run coordinate mouse operations move/click/double-click/right-click/down/up/scroll", Parameters: []string{"action", "x", "y", "delta_x", "delta_y", "steps", "button"}},
		{Name: "window", Method: "POST", Path: "/window", Description: "Read or change browser window bounds and state", Parameters: []string{"action", "left", "top", "width", "height"}},
		{Name: "close_page", Method: "POST", Path: "/close-page", Description: "Close current page"},
		{Name: "fill_form", Method: "POST", Path: "/fill-form", Description: "Fill multiple form fields in one call", Parameters: []string{"fields", "submit", "timeout"}},
		{Name: "drag", Method: "POST", Path: "/drag", Description: "Drag one element to another", Parameters: []string{"from_identifier", "to_identifier"}},
		{Name: "file_upload", Method: "POST", Path: "/file-upload", Description: "Upload local files to a file input", Parameters: []string{"identifier", "file_paths"}},
		{Name: "handle_dialog", Method: "POST", Path: "/handle-dialog", Description: "Arm a handler for the next JavaScript dialog", Parameters: []string{"accept", "text", "timeout"}},
		{Name: "batch", Method: "POST", Path: "/batch", Description: "Execute operations in sequence", Parameters: []string{"operations"}},
	}
	if strings.TrimSpace(command) == "" {
		return commands
	}
	filtered := make([]model.BrowserExecutorHelpCommand, 0, 1)
	for _, item := range commands {
		if strings.EqualFold(item.Name, command) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
