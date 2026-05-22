package workflowhash

import (
	"encoding/json"
	"strings"
)

// NormalizeDrawflowForHash normalizes Automa drawflow before hashing. 规范化 Automa 画布数据用于哈希计算。
func NormalizeDrawflowForHash(value any) any {
	// Parse drawflow JSON string first, matching Automa payload variants. 先解析字符串形式的画布 JSON，兼容 Automa 载荷差异。
	parsedValue := parseJSONValue(value)
	clonedValue := cloneHashValue(parsedValue)
	drawflowMap, ok := clonedValue.(map[string]any)
	if !ok {
		return clonedValue
	}

	// Drop Vue Flow runtime node snapshots from edges. 删除连线中的 Vue Flow 运行态节点快照。
	edges, ok := drawflowMap["edges"].([]any)
	if !ok {
		return drawflowMap
	}
	normalizedEdges := make([]any, 0, len(edges))
	for _, edge := range edges {
		normalizedEdges = append(normalizedEdges, normalizeEdgeForHash(edge))
	}
	drawflowMap["edges"] = normalizedEdges
	return drawflowMap
}

// parseJSONValue parses non-empty JSON strings. 解析非空 JSON 字符串。
func parseJSONValue(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return value
	}

	var parsedValue any
	if json.Unmarshal([]byte(text), &parsedValue) != nil {
		return value
	}
	return parsedValue
}

// normalizeEdgeForHash removes editor-only fields from one edge. 移除单条连线中的编辑器临时字段。
func normalizeEdgeForHash(edge any) any {
	edgeMap, ok := edge.(map[string]any)
	if !ok {
		return edge
	}
	delete(edgeMap, "sourceNode")
	delete(edgeMap, "targetNode")
	return edgeMap
}

// cloneHashValue clones JSON-like values before normalization. 归一化前复制 JSON 类数据。
func cloneHashValue(value any) any {
	switch typedValue := value.(type) {
	case []any:
		clonedList := make([]any, 0, len(typedValue))
		for _, item := range typedValue {
			clonedList = append(clonedList, cloneHashValue(item))
		}
		return clonedList
	case map[string]any:
		clonedMap := make(map[string]any, len(typedValue))
		for key, item := range typedValue {
			clonedMap[key] = cloneHashValue(item)
		}
		return clonedMap
	default:
		return value
	}
}
