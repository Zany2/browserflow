package workflowcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/utility/tasklock"
	"github.com/Zany2/browserflow/backend/utility/workflowhash"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	onlineTTL    = 45 * time.Second
	inventoryTTL = 6 * time.Hour
	reverseTTL   = 6 * time.Hour
)

// WorkflowItem cached workflow summary.
type WorkflowItem struct {
	Id              string `json:"id"`
	AutomaId        string `json:"automa_id"`
	WorkflowId      string `json:"workflow_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	SourceIp        string `json:"source_ip"`
	MachineId       string `json:"machine_id"`
	NodeId          string `json:"node_id"`
	NodeName        string `json:"node_name"`
	AutomaVersion   string `json:"automa_version"`
	ExtVersion      string `json:"ext_version"`
	CreatedAtAutoma int64  `json:"created_at_automa"`
	UpdatedAtAutoma int64  `json:"updated_at_automa"`
	IsDisabled      bool   `json:"is_disabled"`
	IsProtected     bool   `json:"is_protected"`
	NodeCount       int    `json:"node_count"`
	EdgeCount       int    `json:"edge_count"`
	ContentHash     string `json:"content_hash"`
	ReportedAt      int64  `json:"reported_at"`
}

// OnlineNode describes one online execution node.
type OnlineNode struct {
	ClientIP   string `json:"client_ip"`
	MachineID  string `json:"machine_id"`
	NodeID     string `json:"node_id"`
	NodeName   string `json:"node_name"`
	LastSeenAt int64  `json:"last_seen_at"`
}

// ClearBrowserflowKeys clears BrowserFlow Redis cache keys except active task locks.
func ClearBrowserflowKeys(ctx context.Context) error {
	cursor := "0"
	for {
		result, err := g.Redis().Do(ctx, "SCAN", cursor, "MATCH", "browserflow*", "COUNT", 200)
		if err != nil {
			return err
		}

		values := result.Array()
		if len(values) < 2 {
			return nil
		}

		cursor = gconv.String(values[0])
		keys := gconv.Strings(values[1])
		if len(keys) > 0 {
			args := make([]any, 0, len(keys))
			for _, key := range keys {
				if strings.HasPrefix(key, tasklock.KeyPrefix) {
					continue
				}
				args = append(args, key)
			}
			if len(args) == 0 {
				if cursor == "0" {
					return nil
				}
				continue
			}
			if _, err = g.Redis().Do(ctx, "DEL", args...); err != nil {
				return err
			}
		}

		if cursor == "0" {
			return nil
		}
	}
}

// SaveInventory saves old IP-scoped workflow inventory.
func SaveInventory(ctx context.Context, clientIP string, workflows []g.Map) error {
	return SaveNodeInventory(ctx, "", "", "", clientIP, workflows)
}

// SaveNodeInventory saves execution node workflow inventory.
func SaveNodeInventory(ctx context.Context, nodeID string, nodeName string, machineID string, clientIP string, workflows []g.Map) error {
	clientIP = strings.TrimSpace(clientIP)
	machineID = strings.TrimSpace(machineID)
	rawNodeID := strings.TrimSpace(nodeID)
	nodeID = normalizeNodeIdentity(rawNodeID, clientIP)
	displayNodeID := firstNonEmpty(normalizeNodeID(clientIP, rawNodeID), nodeID)
	nodeName = strings.TrimSpace(nodeName)
	if nodeID == "" {
		return nil
	}

	now := time.Now().UnixMilli()
	summaryKey := clientWorkflowsKey(nodeID)
	payloadKey := clientWorkflowPayloadKey(nodeID)
	workflowIDsKey := clientWorkflowIDsKey(nodeID)
	nextIDs := make(map[string]struct{}, len(workflows))

	for _, workflow := range workflows {
		item, rawJSON, err := buildItem(clientIP, workflow, now)
		if err != nil {
			return err
		}
		if item.AutomaId == "" {
			continue
		}
		item.MachineId = machineID
		item.NodeId = displayNodeID
		item.NodeName = nodeName

		summaryJSON, err := json.Marshal(item)
		if err != nil {
			return err
		}

		nextIDs[item.AutomaId] = struct{}{}
		if _, err = g.Redis().Do(ctx, "HSET", summaryKey, item.AutomaId, string(summaryJSON)); err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "HSET", payloadKey, item.AutomaId, rawJSON); err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "SADD", workflowClientsKey(item.AutomaId), nodeID); err != nil {
			return err
		}
		_, _ = g.Redis().Do(ctx, "EXPIRE", workflowClientsKey(item.AutomaId), int(reverseTTL.Seconds()))
	}

	if err := removeMissingWorkflows(ctx, nodeID, summaryKey, payloadKey, nextIDs); err != nil {
		return err
	}
	if _, err := g.Redis().Do(ctx, "DEL", workflowIDsKey); err != nil {
		return err
	}
	for workflowID := range nextIDs {
		if _, err := g.Redis().Do(ctx, "SADD", workflowIDsKey, workflowID); err != nil {
			return err
		}
	}
	if len(nextIDs) > 0 {
		_, _ = g.Redis().Do(ctx, "EXPIRE", workflowIDsKey, int(inventoryTTL.Seconds()))
	}

	if err := saveOnlineNode(ctx, OnlineNode{
		ClientIP:   clientIP,
		MachineID:  machineID,
		NodeID:     displayNodeID,
		NodeName:   nodeName,
		LastSeenAt: now,
	}); err != nil {
		return err
	}
	if _, err := g.Redis().Do(ctx, "SET", clientWorkflowInventoryUpdatedKey(nodeID), now, "EX", int(inventoryTTL.Seconds())); err != nil {
		return err
	}
	_, _ = g.Redis().Do(ctx, "EXPIRE", summaryKey, int(inventoryTTL.Seconds()))
	_, _ = g.Redis().Do(ctx, "EXPIRE", payloadKey, int(inventoryTTL.Seconds()))
	return nil
}

// TouchClient refreshes old IP-scoped online TTL.
func TouchClient(ctx context.Context, clientIP string) {
	TouchNode(ctx, "", "", "", clientIP)
}

// TouchNode refreshes node online TTL.
func TouchNode(ctx context.Context, nodeID string, nodeName string, machineID string, clientIP string) {
	clientIP = strings.TrimSpace(clientIP)
	machineID = strings.TrimSpace(machineID)
	nodeID = normalizeNodeIdentity(nodeID, clientIP)
	nodeName = strings.TrimSpace(nodeName)
	if nodeID == "" {
		return
	}
	_ = saveOnlineNode(ctx, OnlineNode{
		ClientIP:   clientIP,
		MachineID:  machineID,
		NodeID:     nodeID,
		NodeName:   nodeName,
		LastSeenAt: time.Now().UnixMilli(),
	})
}

// ClearClient removes one old IP-scoped client cache.
func ClearClient(ctx context.Context, clientIP string) error {
	return ClearNode(ctx, clientIP)
}

// ClearNode removes one execution node cache.
func ClearNode(ctx context.Context, nodeID string) error {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil
	}

	if node, ok, nodeErr := GetOnlineNode(ctx, nodeID); nodeErr != nil {
		return nodeErr
	} else if ok && strings.TrimSpace(node.ClientIP) != "" {
		if _, err := g.Redis().Do(ctx, "SREM", clientNodesKey(node.ClientIP), nodeID); err != nil {
			return err
		}
	}

	summaryKey := clientWorkflowsKey(nodeID)
	result, err := g.Redis().Do(ctx, "HKEYS", summaryKey)
	if err != nil {
		return err
	}
	for _, workflowID := range gconv.Strings(result.Val()) {
		workflowID = strings.TrimSpace(workflowID)
		if workflowID == "" {
			continue
		}
		if _, err = g.Redis().Do(ctx, "SREM", workflowClientsKey(workflowID), nodeID); err != nil {
			return err
		}
	}

	if _, err = g.Redis().Do(ctx, "SREM", onlineClientsKey(), nodeID); err != nil {
		return err
	}
	_, err = g.Redis().Do(ctx, "DEL",
		clientOnlineKey(nodeID),
		summaryKey,
		clientWorkflowPayloadKey(nodeID),
		clientWorkflowInventoryUpdatedKey(nodeID),
		clientWorkflowIDsKey(nodeID),
	)
	return err
}

// ListOnlineClients lists online node identities.
func ListOnlineClients(ctx context.Context) ([]string, error) {
	result, err := g.Redis().Do(ctx, "SMEMBERS", onlineClientsKey())
	if err != nil {
		return nil, err
	}

	clients := make([]string, 0)
	for _, nodeID := range gconv.Strings(result.Val()) {
		if IsClientOnline(ctx, nodeID) {
			clients = append(clients, nodeID)
		}
	}
	return clients, nil
}

// GetOnlineClientSet checks several node identities in one Redis round trip.
func GetOnlineClientSet(ctx context.Context, clientIDs []string) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(clientIDs))
	ids := make([]string, 0, len(clientIDs))
	keys := make([]any, 0, len(clientIDs))
	for _, clientID := range clientIDs {
		clientID = strings.TrimSpace(clientID)
		if clientID == "" {
			continue
		}
		if _, ok := seen[clientID]; ok {
			continue
		}
		seen[clientID] = struct{}{}
		ids = append(ids, clientID)
		keys = append(keys, clientOnlineKey(clientID))
	}
	if len(keys) == 0 {
		return map[string]struct{}{}, nil
	}

	result, err := g.Redis().Do(ctx, "MGET", keys...)
	if err != nil {
		return nil, err
	}

	online := make(map[string]struct{}, len(ids))
	for index, value := range result.Array() {
		if index >= len(ids) {
			break
		}
		if strings.TrimSpace(gconv.String(value)) != "" {
			online[ids[index]] = struct{}{}
		}
	}
	return online, nil
}

// ListClientNodes lists online node identities under one client IP.
func ListClientNodes(ctx context.Context, clientIP string) ([]string, error) {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return ListOnlineClients(ctx)
	}

	result, err := g.Redis().Do(ctx, "SMEMBERS", clientNodesKey(clientIP))
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0)
	for _, nodeID := range gconv.Strings(result.Val()) {
		nodeID = strings.TrimSpace(nodeID)
		if nodeID != "" && IsClientOnline(ctx, nodeID) {
			nodes = append(nodes, nodeID)
		}
	}
	if len(nodes) > 0 {
		sort.Strings(nodes)
		return nodes, nil
	}

	onlineNodes, err := ListOnlineClients(ctx)
	if err != nil {
		return nil, err
	}
	for _, nodeID := range onlineNodes {
		node, ok, nodeErr := GetOnlineNode(ctx, nodeID)
		if nodeErr != nil {
			return nil, nodeErr
		}
		if ok && (node.ClientIP == clientIP || nodeID == clientIP || strings.HasPrefix(nodeID, clientIP+"|")) {
			nodes = append(nodes, nodeID)
		}
	}
	sort.Strings(nodes)
	return nodes, nil
}

// IsClientOnline checks node online status.
func IsClientOnline(ctx context.Context, clientIP string) bool {
	result, err := g.Redis().Do(ctx, "EXISTS", clientOnlineKey(strings.TrimSpace(clientIP)))
	return err == nil && result.Int() > 0
}

// GetOnlineNode returns online node snapshot.
func GetOnlineNode(ctx context.Context, nodeID string) (OnlineNode, bool, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return OnlineNode{}, false, nil
	}
	result, err := g.Redis().Do(ctx, "GET", clientOnlineKey(nodeID))
	if err != nil {
		return OnlineNode{}, false, err
	}
	text := strings.TrimSpace(result.String())
	if text == "" {
		return OnlineNode{}, false, nil
	}
	var node OnlineNode
	if err = json.Unmarshal([]byte(text), &node); err != nil {
		return OnlineNode{}, false, err
	}
	if node.NodeID == "" {
		node.NodeID = nodeID
	}
	return node, true, nil
}

// ListClientWorkflows lists workflows reported by one node.
func ListClientWorkflows(ctx context.Context, clientIP string) ([]WorkflowItem, error) {
	result, err := g.Redis().Do(ctx, "HVALS", clientWorkflowsKey(strings.TrimSpace(clientIP)))
	if err != nil {
		return nil, err
	}
	items := parseWorkflowItems(result.Val())
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAtAutoma > items[j].UpdatedAtAutoma
	})
	return items, nil
}

// ListClientWorkflowIDs lists workflow ids reported by one node.
func ListClientWorkflowIDs(ctx context.Context, clientIP string) ([]string, error) {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return []string{}, nil
	}
	result, err := g.Redis().Do(ctx, "SMEMBERS", clientWorkflowIDsKey(clientIP))
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0)
	for _, workflowID := range gconv.Strings(result.Val()) {
		workflowID = strings.TrimSpace(workflowID)
		if workflowID != "" {
			ids = append(ids, workflowID)
		}
	}
	if len(ids) == 0 {
		keyResult, keyErr := g.Redis().Do(ctx, "HKEYS", clientWorkflowsKey(clientIP))
		if keyErr != nil {
			return nil, keyErr
		}
		for _, workflowID := range gconv.Strings(keyResult.Val()) {
			workflowID = strings.TrimSpace(workflowID)
			if workflowID != "" {
				ids = append(ids, workflowID)
			}
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// GetClientWorkflowInventoryUpdatedAt gets inventory update time.
func GetClientWorkflowInventoryUpdatedAt(ctx context.Context, clientIP string) (int64, error) {
	result, err := g.Redis().Do(ctx, "GET", clientWorkflowInventoryUpdatedKey(strings.TrimSpace(clientIP)))
	if err != nil {
		return 0, err
	}
	return gconv.Int64(result.Val()), nil
}

// ListWorkflowClients lists online nodes that have workflow.
func ListWorkflowClients(ctx context.Context, automaID string) ([]WorkflowItem, error) {
	result, err := g.Redis().Do(ctx, "SMEMBERS", workflowClientsKey(strings.TrimSpace(automaID)))
	if err != nil {
		return nil, err
	}

	items := make([]WorkflowItem, 0)
	for _, nodeID := range gconv.Strings(result.Val()) {
		if !IsClientOnline(ctx, nodeID) {
			continue
		}
		item, ok, err := GetClientWorkflow(ctx, nodeID, automaID)
		if err != nil {
			return nil, err
		}
		if ok {
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAtAutoma > items[j].UpdatedAtAutoma
	})
	return items, nil
}

// GetClientWorkflow gets one node workflow summary.
func GetClientWorkflow(ctx context.Context, clientIP string, automaID string) (WorkflowItem, bool, error) {
	result, err := g.Redis().Do(ctx, "HGET", clientWorkflowsKey(strings.TrimSpace(clientIP)), strings.TrimSpace(automaID))
	if err != nil {
		return WorkflowItem{}, false, err
	}

	text := strings.TrimSpace(result.String())
	if text == "" {
		return WorkflowItem{}, false, nil
	}

	var item WorkflowItem
	if err = json.Unmarshal([]byte(text), &item); err != nil {
		return WorkflowItem{}, false, err
	}
	return item, true, nil
}

// GetClientWorkflowPayload gets one node workflow payload.
func GetClientWorkflowPayload(ctx context.Context, clientIP string, automaID string) (g.Map, bool, error) {
	result, err := g.Redis().Do(ctx, "HGET", clientWorkflowPayloadKey(strings.TrimSpace(clientIP)), strings.TrimSpace(automaID))
	if err != nil {
		return nil, false, err
	}

	text := strings.TrimSpace(result.String())
	if text == "" {
		return nil, false, nil
	}

	var workflow g.Map
	if err = json.Unmarshal([]byte(text), &workflow); err != nil {
		return nil, false, err
	}
	return workflow, true, nil
}

func buildItem(clientIP string, workflow g.Map, reportedAt int64) (WorkflowItem, string, error) {
	normalizedJSONBytes, err := json.Marshal(workflow)
	if err != nil {
		return WorkflowItem{}, "", err
	}
	rawJSON := string(normalizedJSONBytes)

	hashDrawflowValue := workflowhash.NormalizeDrawflowForHash(workflow["drawflow"])
	hashTableValue := workflow["table"]
	if hashTableValue == nil {
		hashTableValue = workflow["dataColumns"]
	}
	if hashTableValue == nil {
		hashTableValue = []any{}
	}
	hashSettingsValue := workflow["settings"]
	if hashSettingsValue == nil {
		hashSettingsValue = g.Map{}
	}
	hashGlobalDataValue := workflow["globalData"]
	if hashGlobalDataValue == nil {
		hashGlobalDataValue = ""
	}
	coreWorkflowData := g.Map{
		"name":        strings.TrimSpace(gconv.String(workflow["name"])),
		"icon":        strings.TrimSpace(gconv.String(workflow["icon"])),
		"table":       hashTableValue,
		"drawflow":    hashDrawflowValue,
		"settings":    hashSettingsValue,
		"globalData":  hashGlobalDataValue,
		"description": strings.TrimSpace(gconv.String(workflow["description"])),
	}
	coreJSONBytes, err := json.Marshal(coreWorkflowData)
	if err != nil {
		return WorkflowItem{}, "", err
	}
	sum := sha256.Sum256(coreJSONBytes)
	contentHash := hex.EncodeToString(sum[:])
	automaID := resolveWorkflowID(workflow, contentHash)
	nodeCount, edgeCount := countWorkflowGraph(workflow["drawflow"])

	item := WorkflowItem{
		Id:              automaID,
		AutomaId:        automaID,
		WorkflowId:      automaID,
		Name:            strings.TrimSpace(gconv.String(workflow["name"])),
		Description:     strings.TrimSpace(gconv.String(workflow["description"])),
		SourceIp:        clientIP,
		AutomaVersion:   strings.TrimSpace(gconv.String(workflow["version"])),
		ExtVersion:      strings.TrimSpace(gconv.String(workflow["extVersion"])),
		CreatedAtAutoma: gconv.Int64(workflow["createdAt"]),
		UpdatedAtAutoma: gconv.Int64(workflow["updatedAt"]),
		IsDisabled:      gconv.Bool(workflow["isDisabled"]),
		IsProtected:     gconv.Bool(workflow["isProtected"]),
		NodeCount:       nodeCount,
		EdgeCount:       edgeCount,
		ContentHash:     contentHash,
		ReportedAt:      reportedAt,
	}
	return item, rawJSON, nil
}

func saveOnlineNode(ctx context.Context, node OnlineNode) error {
	node.ClientIP = strings.TrimSpace(node.ClientIP)
	node.MachineID = strings.TrimSpace(node.MachineID)
	node.NodeID = strings.TrimSpace(node.NodeID)
	node.NodeName = strings.TrimSpace(node.NodeName)
	if node.NodeID == "" {
		return nil
	}
	identity := normalizeNodeIdentity(node.NodeID, node.ClientIP)
	node.NodeID = firstNonEmpty(normalizeNodeID(node.ClientIP, node.NodeID), identity)
	body, err := json.Marshal(node)
	if err != nil {
		return err
	}
	if _, err = g.Redis().Do(ctx, "SET", clientOnlineKey(identity), string(body), "EX", int(onlineTTL.Seconds())); err != nil {
		return err
	}
	if _, err = g.Redis().Do(ctx, "SADD", onlineClientsKey(), identity); err != nil {
		return err
	}
	if node.ClientIP != "" {
		if _, err = g.Redis().Do(ctx, "SADD", clientNodesKey(node.ClientIP), identity); err != nil {
			return err
		}
		_, _ = g.Redis().Do(ctx, "EXPIRE", clientNodesKey(node.ClientIP), int(inventoryTTL.Seconds()))
	}
	return err
}

func normalizeNodeIdentity(nodeID string, clientIP string) string {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = normalizeNodeID(clientIP, nodeID)
	if clientIP == "" {
		return nodeID
	}
	if nodeID == "" || nodeID == clientIP {
		return clientIP
	}
	return clientIP + "|" + nodeID
}

func normalizeNodeID(clientIP string, nodeID string) string {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" || nodeID == "" {
		return nodeID
	}

	prefix := clientIP + "|"
	for strings.HasPrefix(nodeID, prefix) {
		nodeID = strings.TrimSpace(strings.TrimPrefix(nodeID, prefix))
	}
	return nodeID
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func resolveWorkflowID(workflow g.Map, contentHash string) string {
	for _, key := range []string{"id", "workflow_id", "workflowId"} {
		value := strings.TrimSpace(gconv.String(workflow[key]))
		if value != "" {
			return value
		}
	}
	if len(contentHash) >= 20 {
		return "generated:" + contentHash[:20]
	}
	return ""
}

func countWorkflowGraph(drawflowValue any) (int, int) {
	if drawflowText, ok := drawflowValue.(string); ok {
		drawflowText = strings.TrimSpace(drawflowText)
		if drawflowText != "" {
			var parsedDrawflow any
			if json.Unmarshal([]byte(drawflowText), &parsedDrawflow) == nil {
				drawflowValue = parsedDrawflow
			}
		}
	}

	nodeCount := 0
	edgeCount := 0
	drawflowMap := gconv.Map(drawflowValue)
	if nodes, ok := drawflowMap["nodes"].([]any); ok {
		nodeCount = len(nodes)
	}
	if edges, ok := drawflowMap["edges"].([]any); ok {
		edgeCount = len(edges)
	}
	return nodeCount, edgeCount
}

func parseWorkflowItems(value any) []WorkflowItem {
	items := make([]WorkflowItem, 0)
	for _, text := range gconv.Strings(value) {
		var item WorkflowItem
		if json.Unmarshal([]byte(text), &item) == nil && item.AutomaId != "" {
			items = append(items, item)
		}
	}
	return items
}

func removeMissingWorkflows(ctx context.Context, clientIP string, summaryKey string, payloadKey string, nextIDs map[string]struct{}) error {
	result, err := g.Redis().Do(ctx, "HKEYS", summaryKey)
	if err != nil {
		return err
	}

	for _, oldID := range gconv.Strings(result.Val()) {
		if _, ok := nextIDs[oldID]; ok {
			continue
		}
		if _, err = g.Redis().Do(ctx, "HDEL", summaryKey, oldID); err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "HDEL", payloadKey, oldID); err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "SREM", workflowClientsKey(oldID), clientIP); err != nil {
			return err
		}
	}
	return nil
}

func onlineClientsKey() string {
	return "browserflow:clients:online"
}

func clientOnlineKey(clientIP string) string {
	return "browserflow:client:online:" + clientIP
}

func clientWorkflowsKey(clientIP string) string {
	return "browserflow:client:workflows:" + clientIP
}

func clientWorkflowPayloadKey(clientIP string) string {
	return "browserflow:client:workflow:payload:" + clientIP
}

func clientWorkflowInventoryUpdatedKey(clientIP string) string {
	return "browserflow:client:workflow:inventory-updated:" + clientIP
}

func clientNodesKey(clientIP string) string {
	return "browserflow:client:nodes:" + clientIP
}

func clientWorkflowIDsKey(clientIP string) string {
	return "browserflow:client:workflow-ids:" + clientIP
}

func workflowClientsKey(automaID string) string {
	return "browserflow:workflow:clients:" + automaID
}
