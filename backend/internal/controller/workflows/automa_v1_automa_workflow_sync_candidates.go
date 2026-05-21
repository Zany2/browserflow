package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/Zany2/browserflow/backend/utility/workflowhash"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	// workflowInventoryRefreshTimeout max wait for client inventory refresh. 客户端清单刷新等待超时
	workflowInventoryRefreshTimeout = 8 * time.Second
)

// WorkflowSyncCandidates lists sync candidates. 获取工作流同步候选
func (c *ControllerV1) WorkflowSyncCandidates(ctx context.Context, req *v1.WorkflowSyncCandidatesReq) (res *v1.WorkflowSyncCandidatesRes, err error) {
	serverMode := consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer
	var (
		db      *storage.BoltDB
		records []*model.AutomaWorkflowRecord
	)
	if serverMode {
		columns := dao.AutomaWorkflows.Columns()
		items := []entity.AutomaWorkflows{}
		if err = dao.AutomaWorkflows.Ctx(ctx).OrderDesc(columns.UpdatedAt).Scan(&items); err != nil {
			return nil, err
		}
		records = make([]*model.AutomaWorkflowRecord, 0, len(items))
		for index := range items {
			item := items[index]
			record := &model.AutomaWorkflowRecord{ID: item.Id, AutomaID: item.AutomaId, Name: item.Name, Description: item.Description, AutomaName: item.AutomaName, AutomaDescription: item.AutomaDescription, Source: item.Source, SourceIP: item.SourceIp, SourceUserAgent: item.SourceUserAgent, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, RawJSON: item.RawJson, NormalizedJSON: item.NormalizedJson, ContentHash: item.ContentHash, Revision: item.Revision}
			if item.FirstSyncedAt != nil && !item.FirstSyncedAt.IsZero() {
				record.FirstSyncedAt = item.FirstSyncedAt.Time
			}
			if item.LastSyncedAt != nil && !item.LastSyncedAt.IsZero() {
				record.LastSyncedAt = item.LastSyncedAt.Time
			}
			if item.CreatedAt != nil && !item.CreatedAt.IsZero() {
				record.CreatedAt = item.CreatedAt.Time
			}
			if item.UpdatedAt != nil && !item.UpdatedAt.IsZero() {
				record.UpdatedAt = item.UpdatedAt.Time
			}
			records = append(records, record)
		}
	} else {
		state.DBMu.Lock()
		if state.DB == nil {
			dbPath := os.Getenv("DB_PATH")
			if dbPath == "" {
				dbPath = g.Cfg().MustGet(ctx, "localStorage.path", "data/browserflow.db").String()
			}
			state.DB, err = storage.NewBoltDB(dbPath)
			if err != nil {
				state.DBMu.Unlock()
				return nil, err
			}
		}
		if state.LLMClient == nil {
			state.LLMClient = llm.NewClient()
		}
		db = state.DB
		state.DBMu.Unlock()

		records, err = db.ListAutomaWorkflowRecords()
		if err != nil {
			return nil, err
		}
	}

	serverRecords := make(map[string]*model.AutomaWorkflowRecord, len(records))
	for _, record := range records {
		recordAutomaID := strings.TrimSpace(record.AutomaID)
		if recordAutomaID != "" {
			serverRecords[recordAutomaID] = record
		}
	}

	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	sourceIP := strings.TrimSpace(req.SourceIP)
	automaID := strings.TrimSpace(req.AutomaID)
	mode := strings.TrimSpace(req.Mode)
	if serverMode && sourceIP != "" && !workflowcache.IsClientOnline(ctx, sourceIP) {
		return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: 0}, nil
	}
	if req.Refresh {
		if sourceIP != "" {
			if err = workflowagent.RefreshClientInventory(ctx, sourceIP, workflowInventoryRefreshTimeout); err != nil {
				return nil, err
			}
		} else if mode == "workflow" {
			clientIPs, listErr := workflowcache.ListOnlineClients(ctx)
			if listErr != nil {
				return nil, listErr
			}
			if err = workflowagent.RefreshClientInventories(ctx, clientIPs, workflowInventoryRefreshTimeout); err != nil {
				return nil, err
			}
		}
	}

	var cacheItems []workflowcache.WorkflowItem
	if sourceIP != "" {
		cacheItems, err = workflowcache.ListClientWorkflows(ctx, sourceIP)
	} else if mode == "workflow" {
		if automaID != "" {
			cacheItems, err = workflowcache.ListWorkflowClients(ctx, automaID)
		} else {
			clientIPs, listErr := workflowcache.ListOnlineClients(ctx)
			if listErr != nil {
				return nil, listErr
			}
			cacheItems = make([]workflowcache.WorkflowItem, 0)
			for _, clientIP := range clientIPs {
				items, listErr := workflowcache.ListClientWorkflows(ctx, clientIP)
				if listErr != nil {
					return nil, listErr
				}
				cacheItems = append(cacheItems, items...)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if cacheItems != nil {
		candidates := make([]v1.WorkflowSyncCandidatesResModel, 0, len(cacheItems))
		for _, item := range cacheItems {
			if !workflowcache.IsClientOnline(ctx, item.SourceIp) {
				continue
			}
			if automaID != "" && item.AutomaId != automaID {
				continue
			}
			if keyword != "" {
				text := strings.ToLower(strings.Join([]string{item.AutomaId, item.Name, item.Description, item.SourceIp}, " "))
				if !strings.Contains(text, keyword) {
					continue
				}
			}
			itemAutomaID := strings.TrimSpace(item.AutomaId)
			serverRecord := serverRecords[itemAutomaID]
			synced, hasUpdate, status := resolveWorkflowSyncState(serverRecord, item.UpdatedAtAutoma, item.ContentHash)
			if serverRecord != nil {
				serverAutomaName := strings.TrimSpace(serverRecord.AutomaName)
				if serverAutomaName == "" {
					serverAutomaName = serverRecord.Name
				}
				serverAutomaDescription := strings.TrimSpace(serverRecord.AutomaDescription)
				if serverAutomaDescription == "" {
					serverAutomaDescription = serverRecord.Description
				}
				candidate := v1.WorkflowSyncCandidatesResModel{Id: item.AutomaId, AutomaId: item.AutomaId, WorkflowId: item.WorkflowId, Name: item.Name, Description: item.Description, AutomaName: item.Name, AutomaDescription: item.Description, Source: "客户端同步", SourceIp: item.SourceIp, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: serverRecord.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, ContentHash: item.ContentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status, Online: workflowcache.IsClientOnline(ctx, item.SourceIp), ServerId: serverRecord.ID, ServerName: serverRecord.Name, ServerDesc: serverRecord.Description, ServerAutomaName: serverAutomaName, ServerAutomaDesc: serverAutomaDescription, ServerRevision: serverRecord.Revision}
				if !serverRecord.LastSyncedAt.IsZero() {
					candidate.LastSyncedAt = gtime.NewFromTime(serverRecord.LastSyncedAt)
				}
				if !serverRecord.UpdatedAt.IsZero() {
					candidate.ServerUpdatedAt = gtime.NewFromTime(serverRecord.UpdatedAt)
				}
				candidates = append(candidates, candidate)
				continue
			}
			candidate := v1.WorkflowSyncCandidatesResModel{Id: item.AutomaId, AutomaId: item.AutomaId, WorkflowId: item.WorkflowId, Name: item.Name, Description: item.Description, AutomaName: item.Name, AutomaDescription: item.Description, Source: "客户端同步", SourceIp: item.SourceIp, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, ContentHash: item.ContentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status, Online: workflowcache.IsClientOnline(ctx, item.SourceIp)}
			candidates = append(candidates, candidate)
		}
		total := len(candidates)
		pageNum := req.PageNum
		if pageNum <= 0 {
			pageNum = 1
		}
		pageSize := req.PageSize
		if pageSize <= 0 {
			pageSize = 30
		}
		start := (pageNum - 1) * pageSize
		if start >= total {
			return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: total}, nil
		}
		end := start + pageSize
		if end > total {
			end = total
		}
		return &v1.WorkflowSyncCandidatesRes{List: candidates[start:end], Total: total}, nil
	}

	items := make([]map[string]any, 0)
	if !serverMode {
		if snapshot, snapshotErr := db.GetAutomaWorkflowSnapshot("latest"); snapshotErr == nil && len(snapshot.Workflows) > 0 {
			_ = json.Unmarshal(snapshot.Workflows, &items)
		}
	}
	if serverMode && sourceIP == "" && automaID == "" {
		return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: 0}, nil
	}
	if len(items) == 0 {
		for _, record := range records {
			var payload map[string]any
			_ = json.Unmarshal([]byte(record.RawJSON), &payload)
			if len(payload) > 0 {
				items = append(items, payload)
			}
		}
	}

	candidates := make([]v1.WorkflowSyncCandidatesResModel, 0, len(items))
	for _, item := range items {
		hashDrawflowValue := workflowhash.NormalizeDrawflowForHash(item["drawflow"])
		hashTableValue := item["table"]
		if hashTableValue == nil {
			hashTableValue = item["dataColumns"]
		}
		if hashTableValue == nil {
			hashTableValue = []any{}
		}
		hashSettingsValue := item["settings"]
		if hashSettingsValue == nil {
			hashSettingsValue = g.Map{}
		}
		hashGlobalDataValue := item["globalData"]
		if hashGlobalDataValue == nil {
			hashGlobalDataValue = ""
		}
		coreJSONBytes, jsonErr := json.Marshal(g.Map{"name": strings.TrimSpace(gconv.String(item["name"])), "icon": strings.TrimSpace(gconv.String(item["icon"])), "table": hashTableValue, "drawflow": hashDrawflowValue, "settings": hashSettingsValue, "globalData": hashGlobalDataValue, "description": strings.TrimSpace(gconv.String(item["description"]))})
		if jsonErr != nil {
			continue
		}
		sum := sha256.Sum256(coreJSONBytes)
		contentHash := hex.EncodeToString(sum[:])
		itemAutomaID := ""
		for _, key := range []string{"id", "workflow_id", "workflowId"} {
			value := strings.TrimSpace(gconv.String(item[key]))
			if value != "" {
				itemAutomaID = value
				break
			}
		}
		if itemAutomaID == "" {
			itemAutomaID = "generated:" + contentHash[:20]
		}
		if automaID != "" && itemAutomaID != automaID {
			continue
		}
		name := strings.TrimSpace(gconv.String(item["name"]))
		description := strings.TrimSpace(gconv.String(item["description"]))
		if keyword != "" {
			text := strings.ToLower(strings.Join([]string{itemAutomaID, name, description, sourceIP}, " "))
			if !strings.Contains(text, keyword) {
				continue
			}
		}
		drawflowValue := item["drawflow"]
		if drawflowText, ok := drawflowValue.(string); ok {
			drawflowText = strings.TrimSpace(drawflowText)
			if drawflowText != "" {
				var parsedDrawflow any
				if json.Unmarshal([]byte(drawflowText), &parsedDrawflow) == nil {
					drawflowValue = parsedDrawflow
				}
			}
		}
		drawflowMap := gconv.Map(drawflowValue)
		nodeCount, edgeCount := 0, 0
		if nodes, ok := drawflowMap["nodes"].([]any); ok {
			nodeCount = len(nodes)
		}
		if edges, ok := drawflowMap["edges"].([]any); ok {
			edgeCount = len(edges)
		}
		itemContentHash := strings.TrimSpace(contentHash)
		serverRecord := serverRecords[itemAutomaID]
		synced, hasUpdate, status := resolveWorkflowSyncState(serverRecord, gconv.Int64(item["updatedAt"]), itemContentHash)
		if serverRecord != nil {
			serverAutomaName := strings.TrimSpace(serverRecord.AutomaName)
			if serverAutomaName == "" {
				serverAutomaName = serverRecord.Name
			}
			serverAutomaDescription := strings.TrimSpace(serverRecord.AutomaDescription)
			if serverAutomaDescription == "" {
				serverAutomaDescription = serverRecord.Description
			}
			candidate := v1.WorkflowSyncCandidatesResModel{Id: itemAutomaID, AutomaId: itemAutomaID, WorkflowId: itemAutomaID, Name: name, Description: description, AutomaName: name, AutomaDescription: description, Source: "客户端同步", SourceIp: sourceIP, AutomaVersion: strings.TrimSpace(gconv.String(item["version"])), ExtVersion: strings.TrimSpace(gconv.String(item["extVersion"])), CreatedAtAutoma: gconv.Int64(item["createdAt"]), UpdatedAtAutoma: gconv.Int64(item["updatedAt"]), IsDisabled: gconv.Bool(item["isDisabled"]), IsProtected: serverRecord.IsProtected, NodeCount: nodeCount, EdgeCount: edgeCount, ContentHash: contentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status, ServerId: serverRecord.ID, ServerName: serverRecord.Name, ServerDesc: serverRecord.Description, ServerAutomaName: serverAutomaName, ServerAutomaDesc: serverAutomaDescription, ServerRevision: serverRecord.Revision}
			if !serverRecord.LastSyncedAt.IsZero() {
				candidate.LastSyncedAt = gtime.NewFromTime(serverRecord.LastSyncedAt)
			}
			if !serverRecord.UpdatedAt.IsZero() {
				candidate.ServerUpdatedAt = gtime.NewFromTime(serverRecord.UpdatedAt)
			}
			candidates = append(candidates, candidate)
			continue
		}
		candidate := v1.WorkflowSyncCandidatesResModel{Id: itemAutomaID, AutomaId: itemAutomaID, WorkflowId: itemAutomaID, Name: name, Description: description, AutomaName: name, AutomaDescription: description, Source: "客户端同步", SourceIp: sourceIP, AutomaVersion: strings.TrimSpace(gconv.String(item["version"])), ExtVersion: strings.TrimSpace(gconv.String(item["extVersion"])), CreatedAtAutoma: gconv.Int64(item["createdAt"]), UpdatedAtAutoma: gconv.Int64(item["updatedAt"]), IsDisabled: gconv.Bool(item["isDisabled"]), NodeCount: nodeCount, EdgeCount: edgeCount, ContentHash: contentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status}
		candidates = append(candidates, candidate)
	}
	total := len(candidates)
	pageNum := req.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}
	start := (pageNum - 1) * pageSize
	if start >= total {
		return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: total}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return &v1.WorkflowSyncCandidatesRes{List: candidates[start:end], Total: total}, nil
}

// resolveWorkflowSyncState compares content hash first, then updatedAt. 优先用内容哈希判断一致，再用更新时间判断新旧。
func resolveWorkflowSyncState(serverRecord *model.AutomaWorkflowRecord, clientUpdatedAt int64, clientContentHash string) (synced bool, hasUpdate bool, status string) {
	if serverRecord == nil {
		return false, true, "not_synced"
	}

	serverContentHash := strings.TrimSpace(serverRecord.ContentHash)
	clientContentHash = strings.TrimSpace(clientContentHash)
	if serverContentHash != "" && clientContentHash != "" && serverContentHash == clientContentHash {
		return true, false, "synced"
	}

	serverUpdatedAt := serverRecord.UpdatedAtAutoma
	if clientUpdatedAt > 0 && serverUpdatedAt > 0 {
		if clientUpdatedAt > serverUpdatedAt {
			return false, true, "client_newer"
		}
		if clientUpdatedAt < serverUpdatedAt {
			return false, false, "server_newer"
		}
	}

	return false, true, "has_update"
}
