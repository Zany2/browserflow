package workflowcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	onlineTTL    = 90 * time.Second   // onlineTTL client online heartbeat ttl 客户端在线心跳过期时间
	inventoryTTL = 24 * time.Hour     // inventoryTTL workflow inventory ttl 工作流清单过期时间
	reverseTTL   = 7 * 24 * time.Hour // reverseTTL workflow reverse index ttl 工作流反向索引过期时间
)

// WorkflowItem cached workflow summary 缓存工作流摘要
type WorkflowItem struct {
	Id              string `json:"id"`                // Id workflow id 工作流 ID
	AutomaId        string `json:"automa_id"`         // AutomaId original Automa id Automa 原始 ID
	WorkflowId      string `json:"workflow_id"`       // WorkflowId normalized workflow id 规范化工作流 ID
	Name            string `json:"name"`              // Name workflow name 工作流名称
	Description     string `json:"description"`       // Description workflow description 工作流描述
	SourceIp        string `json:"source_ip"`         // SourceIp reporting client ip 上报客户端 IP
	AutomaVersion   string `json:"automa_version"`    // AutomaVersion Automa version Automa 版本
	ExtVersion      string `json:"ext_version"`       // ExtVersion extension version 扩展版本
	CreatedAtAutoma int64  `json:"created_at_automa"` // CreatedAtAutoma Automa create time Automa 创建时间
	UpdatedAtAutoma int64  `json:"updated_at_automa"` // UpdatedAtAutoma Automa update time Automa 更新时间
	IsDisabled      bool   `json:"is_disabled"`       // IsDisabled disabled status 是否禁用
	IsProtected     bool   `json:"is_protected"`      // IsProtected protected status 是否受保护
	NodeCount       int    `json:"node_count"`        // NodeCount workflow node count 节点数量
	EdgeCount       int    `json:"edge_count"`        // EdgeCount workflow edge count 连线数量
	ContentHash     string `json:"content_hash"`      // ContentHash normalized content hash 内容哈希
	ReportedAt      int64  `json:"reported_at"`       // ReportedAt client report time 客户端上报时间
}

// ClearBrowserflowKeys clears browserflow cache keys 清理 browserflow 缓存键
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
				args = append(args, key)
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

// SaveInventory saves client workflow inventory 保存客户端工作流清单
func SaveInventory(ctx context.Context, clientIP string, workflows []g.Map) error {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return nil
	}

	now := time.Now().UnixMilli()
	summaryKey := clientWorkflowsKey(clientIP)
	payloadKey := clientWorkflowPayloadKey(clientIP)
	nextIDs := make(map[string]struct{}, len(workflows))

	for _, workflow := range workflows {
		// Build summary and raw payload 构建摘要和原始载荷
		item, rawJSON, err := buildItem(clientIP, workflow, now)
		if err != nil {
			return err
		}
		if item.AutomaId == "" {
			continue
		}

		summaryJSON, err := json.Marshal(item)
		if err != nil {
			return err
		}

		// Save summary, payload, and reverse index 保存摘要、载荷和反向索引
		nextIDs[item.AutomaId] = struct{}{}
		if _, err = g.Redis().Do(ctx, "HSET", summaryKey, item.AutomaId, string(summaryJSON)); err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "HSET", payloadKey, item.AutomaId, rawJSON); err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "SADD", workflowClientsKey(item.AutomaId), clientIP); err != nil {
			return err
		}
		_, _ = g.Redis().Do(ctx, "EXPIRE", workflowClientsKey(item.AutomaId), int(reverseTTL.Seconds()))
	}

	if err := removeMissingWorkflows(ctx, clientIP, summaryKey, payloadKey, nextIDs); err != nil {
		return err
	}

	// Refresh client online and inventory ttl 刷新客户端在线状态和清单过期时间
	onlineBody, err := json.Marshal(g.Map{
		"client_ip":    clientIP,
		"last_seen_at": now,
	})
	if err != nil {
		return err
	}
	if _, err = g.Redis().Do(ctx, "SET", clientOnlineKey(clientIP), string(onlineBody), "EX", int(onlineTTL.Seconds())); err != nil {
		return err
	}
	if _, err = g.Redis().Do(ctx, "SADD", onlineClientsKey(), clientIP); err != nil {
		return err
	}
	if _, err = g.Redis().Do(ctx, "SET", clientWorkflowInventoryUpdatedKey(clientIP), now, "EX", int(inventoryTTL.Seconds())); err != nil {
		return err
	}
	_, _ = g.Redis().Do(ctx, "EXPIRE", summaryKey, int(inventoryTTL.Seconds()))
	_, _ = g.Redis().Do(ctx, "EXPIRE", payloadKey, int(inventoryTTL.Seconds()))
	return nil
}

// TouchClient refreshes client online ttl 刷新客户端在线状态
func TouchClient(ctx context.Context, clientIP string) {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return
	}
	now := time.Now().UnixMilli()
	onlineBody, _ := json.Marshal(g.Map{
		"client_ip":    clientIP,
		"last_seen_at": now,
	})
	_, _ = g.Redis().Do(ctx, "SET", clientOnlineKey(clientIP), string(onlineBody), "EX", int(onlineTTL.Seconds()))
	_, _ = g.Redis().Do(ctx, "SADD", onlineClientsKey(), clientIP)
}

// ListOnlineClients lists currently online clients 列出当前在线客户端
func ListOnlineClients(ctx context.Context) ([]string, error) {
	result, err := g.Redis().Do(ctx, "SMEMBERS", onlineClientsKey())
	if err != nil {
		return nil, err
	}

	clients := make([]string, 0)
	for _, clientIP := range gconv.Strings(result.Val()) {
		if IsClientOnline(ctx, clientIP) {
			clients = append(clients, clientIP)
		}
	}
	return clients, nil
}

// IsClientOnline checks client online status 检查客户端是否在线
func IsClientOnline(ctx context.Context, clientIP string) bool {
	result, err := g.Redis().Do(ctx, "EXISTS", clientOnlineKey(strings.TrimSpace(clientIP)))
	return err == nil && result.Int() > 0
}

// ListClientWorkflows lists workflows reported by client 列出客户端上报的工作流
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

// GetClientWorkflowInventoryUpdatedAt gets inventory update time 获取客户端清单更新时间
func GetClientWorkflowInventoryUpdatedAt(ctx context.Context, clientIP string) (int64, error) {
	result, err := g.Redis().Do(ctx, "GET", clientWorkflowInventoryUpdatedKey(strings.TrimSpace(clientIP)))
	if err != nil {
		return 0, err
	}
	return gconv.Int64(result.Val()), nil
}

// ListWorkflowClients lists online clients that have workflow 列出拥有指定工作流的在线客户端
func ListWorkflowClients(ctx context.Context, automaID string) ([]WorkflowItem, error) {
	result, err := g.Redis().Do(ctx, "SMEMBERS", workflowClientsKey(strings.TrimSpace(automaID)))
	if err != nil {
		return nil, err
	}

	items := make([]WorkflowItem, 0)
	for _, clientIP := range gconv.Strings(result.Val()) {
		if !IsClientOnline(ctx, clientIP) {
			continue
		}
		item, ok, err := GetClientWorkflow(ctx, clientIP, automaID)
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

// GetClientWorkflow gets one client workflow summary 获取客户端工作流摘要
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

// GetClientWorkflowPayload gets one client workflow payload 获取客户端工作流原始载荷
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

// buildItem builds workflow summary and raw json 构建工作流摘要和原始 JSON
func buildItem(clientIP string, workflow g.Map, reportedAt int64) (WorkflowItem, string, error) {
	normalizedJSONBytes, err := json.Marshal(workflow)
	if err != nil {
		return WorkflowItem{}, "", err
	}
	rawJSON := string(normalizedJSONBytes)

	// Normalize hash inputs to avoid unstable content hash 规范化哈希输入，避免内容哈希抖动
	hashDrawflowValue := workflow["drawflow"]
	if drawflowText, ok := hashDrawflowValue.(string); ok {
		drawflowText = strings.TrimSpace(drawflowText)
		if drawflowText != "" {
			var parsedDrawflow any
			if json.Unmarshal([]byte(drawflowText), &parsedDrawflow) == nil {
				hashDrawflowValue = parsedDrawflow
			}
		}
	}
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
		"id":          strings.TrimSpace(gconv.String(workflow["id"])),
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

	// Fill cache item from normalized workflow data 填充规范化后的缓存条目
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

// resolveWorkflowID resolves stable workflow id 解析稳定的工作流 ID
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

// countWorkflowGraph counts drawflow nodes and edges 统计工作流节点和连线数量
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

// parseWorkflowItems parses cached workflow summaries 解析缓存工作流摘要
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

// removeMissingWorkflows removes stale workflow cache 删除已不存在的工作流缓存
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

// onlineClientsKey returns online clients set key 返回在线客户端集合键
func onlineClientsKey() string {
	return "browserflow:clients:online"
}

// clientOnlineKey returns client online key 返回客户端在线状态键
func clientOnlineKey(clientIP string) string {
	return "browserflow:client:online:" + clientIP
}

// clientWorkflowsKey returns client workflow summary hash key 返回客户端工作流摘要哈希键
func clientWorkflowsKey(clientIP string) string {
	return "browserflow:client:workflows:" + clientIP
}

// clientWorkflowPayloadKey returns client workflow payload hash key 返回客户端工作流载荷哈希键
func clientWorkflowPayloadKey(clientIP string) string {
	return "browserflow:client:workflow:payload:" + clientIP
}

// clientWorkflowInventoryUpdatedKey returns inventory update key 返回工作流清单更新时间键
func clientWorkflowInventoryUpdatedKey(clientIP string) string {
	return "browserflow:client:workflow:inventory-updated:" + clientIP
}

// workflowClientsKey returns workflow reverse index key 返回工作流反向索引键
func workflowClientsKey(automaID string) string {
	return "browserflow:workflow:clients:" + automaID
}
