package tasks

import "strings"

func normalizeDispatchMode(value string, machineID string, nodeID string, targetGroupID int64, clientIP string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "auto", "group", "machine", "node", "ip":
		return value
	}
	if strings.TrimSpace(nodeID) != "" {
		return "node"
	}
	if targetGroupID > 0 {
		return "group"
	}
	if strings.TrimSpace(machineID) != "" {
		return "machine"
	}
	if strings.TrimSpace(clientIP) != "" {
		return "ip"
	}
	return "auto"
}

func normalizeQueuePolicy(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "queue", "fail", "skip":
		return value
	default:
		return "queue"
	}
}
