package browserexecutor

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// buildAccessibilitySnapshot reads Chrome Accessibility Tree. buildAccessibilitySnapshot reads Chrome Accessibility Tree.
func buildAccessibilitySnapshot(ctx context.Context, page *rod.Page) (*model.BrowserExecutorAccessibilitySnapshot, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	_ = (proto.AccessibilityDisable{}).Call(page)
	if err := (proto.AccessibilityEnable{}).Call(page); err != nil {
		return nil, fmt.Errorf("failed to enable Accessibility: %w", err)
	}
	defer func() {
		_ = (proto.AccessibilityDisable{}).Call(page)
	}()

	result, err := (proto.AccessibilityGetFullAXTree{}).Call(page)
	if err != nil {
		return nil, fmt.Errorf("failed to read Accessibility Tree: %w", err)
	}
	if result == nil || len(result.Nodes) == 0 {
		return nil, fmt.Errorf("Accessibility Tree is empty")
	}

	snapshot := &model.BrowserExecutorAccessibilitySnapshot{Elements: map[string]*model.BrowserExecutorAccessibilityNode{}}
	for _, axNode := range result.Nodes {
		node := buildNode(axNode)
		if node == nil {
			continue
		}
		snapshot.Elements[node.ID] = node
	}
	return snapshot, nil
}

func buildNode(axNode *proto.AccessibilityAXNode) *model.BrowserExecutorAccessibilityNode {
	if axNode == nil {
		return nil
	}
	role := axValue(axNode.Role)
	name := strings.TrimSpace(axValue(axNode.Name))
	value := strings.TrimSpace(axValue(axNode.Value))
	description := strings.TrimSpace(axValue(axNode.Description))
	if name == "" {
		name = description
	}

	node := &model.BrowserExecutorAccessibilityNode{
		ID:            string(axNode.NodeID),
		BackendNodeID: int(axNode.BackendDOMNodeID),
		Role:          role,
		Name:          name,
		Value:         value,
		Attributes:    map[string]string{},
	}

	for _, prop := range axNode.Properties {
		key := string(prop.Name)
		val := axValue(prop.Value)
		if key == "" || val == "" {
			continue
		}
		node.Attributes[key] = val
		if key == "placeholder" {
			node.Placeholder = val
		}
	}
	return node
}

func axValue(value *proto.AccessibilityAXValue) string {
	if value == nil {
		return ""
	}
	text := value.Value.String()
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}
	return strings.TrimSpace(text)
}

func isInteractiveRole(role string) bool {
	switch strings.ToLower(role) {
	case "button", "link", "menuitem", "checkbox", "radio", "tab", "switch", "option", "combobox":
		return true
	default:
		return false
	}
}

func isUsefulInputRole(role string) bool {
	switch strings.ToLower(role) {
	case "textbox", "searchbox", "combobox", "spinbutton", "slider":
		return true
	default:
		return false
	}
}

func snapshotText(refs []model.BrowserExecutorElementRef) string {
	var builder strings.Builder
	builder.WriteString("=== BrowserFlow Browser Snapshot ===\n")
	builder.WriteString("Use RefIDs like @e1 as identifiers for click/type/get-text.\n\n")

	clickable := make([]model.BrowserExecutorElementRef, 0)
	inputs := make([]model.BrowserExecutorElementRef, 0)
	for _, ref := range refs {
		if isUsefulInputRole(ref.Role) {
			inputs = append(inputs, ref)
			continue
		}
		clickable = append(clickable, ref)
	}

	writeRefs := func(title string, items []model.BrowserExecutorElementRef) {
		builder.WriteString(title)
		builder.WriteString(":\n")
		if len(items) == 0 {
			builder.WriteString("  none\n\n")
			return
		}
		for _, item := range items {
			label := firstNonEmpty(item.Name, item.Placeholder, item.Value, item.Role)
			builder.WriteString(fmt.Sprintf("  %s - %s", item.RefID, label))
			if item.Role != "" {
				builder.WriteString(fmt.Sprintf(" (role: %s)", item.Role))
			}
			if item.Placeholder != "" {
				builder.WriteString(fmt.Sprintf(" [placeholder: %s]", item.Placeholder))
			}
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	writeRefs("Clickable Elements", clickable)
	writeRefs("Input Elements", inputs)
	builder.WriteString("USAGE:\n")
	builder.WriteString("  click: {\"identifier\":\"@e1\"}\n")
	builder.WriteString("  type: {\"identifier\":\"@e2\",\"text\":\"hello\"}\n")
	return builder.String()
}

// FilterElementRefs returns compact refs by interactive category. FilterElementRefs filters refs by type.
func FilterElementRefs(refs []model.BrowserExecutorElementRef, category string, limit int) []model.BrowserExecutorElementRef {
	category = strings.ToLower(strings.TrimSpace(category))
	out := make([]model.BrowserExecutorElementRef, 0)
	for _, ref := range refs {
		isInput := isUsefulInputRole(ref.Role)
		if category == "input" && !isInput {
			continue
		}
		if category == "clickable" && isInput {
			continue
		}
		out = append(out, ref)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func compactAttributes(attrs map[string]string) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	keep := []string{"id", "class", "href", "placeholder", "aria-label", "name", "type", "tag"}
	out := map[string]string{}
	for _, key := range keep {
		if value := strings.TrimSpace(attrs[key]); value != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sortedRefs(refs []model.BrowserExecutorElementRef) []model.BrowserExecutorElementRef {
	out := append([]model.BrowserExecutorElementRef(nil), refs...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].RefID < out[j].RefID
	})
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
