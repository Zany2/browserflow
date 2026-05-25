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
	"github.com/Zany2/browserflow/backend/utility/rr"
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
	syncStatusFilter := strings.TrimSpace(req.SyncStatus)
	workflowStatusFilter := strings.TrimSpace(req.WorkflowStatus)
	sourceIP := strings.TrimSpace(req.SourceIP)
	sourceNodeID := normalizeSourceNodeID(sourceIP, req.SourceNodeID)
	sourceIdentity := ""
	if sourceNodeID != "" {
		sourceIdentity = nodeConnectionIdentity(sourceIP, sourceNodeID)
	}
	automaID := strings.TrimSpace(req.AutomaID)
	mode := strings.TrimSpace(req.Mode)
	if !serverMode {
		if req.Refresh || sourceIP != "" || sourceNodeID != "" || mode == "workflow" {
			return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: 0}, nil
		}
		mode = "client"
	}
	if serverMode && sourceIdentity != "" && !workflowcache.IsClientOnline(ctx, sourceIdentity) {
		return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: 0}, nil
	}
	if req.Refresh {
		if sourceIdentity != "" {
			if err = workflowagent.RefreshClientInventory(ctx, sourceIdentity, workflowInventoryRefreshTimeout); err != nil {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), err.Error())
				return nil, nil
			}
		} else if mode == "workflow" || sourceIP != "" {
			clientIPs, listErr := listOnlineCandidateIdentities(ctx, sourceIP, sourceNodeID)
			if listErr != nil {
				return nil, listErr
			}
			if err = workflowagent.RefreshClientInventories(ctx, clientIPs, workflowInventoryRefreshTimeout); err != nil {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), err.Error())
				return nil, nil
			}
		}
	}

	if serverMode && mode == "workflow" {
		if automaID == "" {
			return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: 0}, nil
		}
		serverRecord := serverRecords[automaID]
		if serverRecord == nil {
			return &v1.WorkflowSyncCandidatesRes{List: []v1.WorkflowSyncCandidatesResModel{}, Total: 0}, nil
		}

		serverAutomaName := strings.TrimSpace(serverRecord.AutomaName)
		if serverAutomaName == "" {
			serverAutomaName = serverRecord.Name
		}
		serverAutomaDescription := strings.TrimSpace(serverRecord.AutomaDescription)
		if serverAutomaDescription == "" {
			serverAutomaDescription = serverRecord.Description
		}

		clientIPs, listErr := listOnlineCandidateIdentities(ctx, sourceIP, sourceNodeID)
		if listErr != nil {
			return nil, listErr
		}

		candidates := make([]v1.WorkflowSyncCandidatesResModel, 0, len(clientIPs))
		for _, clientIdentity := range clientIPs {
			clientIdentity = strings.TrimSpace(clientIdentity)
			if clientIdentity == "" || !workflowcache.IsClientOnline(ctx, clientIdentity) {
				continue
			}
			node := getOnlineNodeSnapshot(ctx, clientIdentity)
			if keyword != "" && !strings.Contains(strings.ToLower(strings.Join([]string{clientIdentity, node.ClientIP, node.MachineID, node.NodeID, node.NodeName}, " ")), keyword) {
				continue
			}

			item, ok, loadErr := workflowcache.GetClientWorkflow(ctx, clientIdentity, automaID)
			if loadErr != nil {
				return nil, loadErr
			}

			candidate := v1.WorkflowSyncCandidatesResModel{
				Id:                automaID,
				AutomaId:          automaID,
				WorkflowId:        automaID,
				Source:            "客户端同步",
				SourceIp:          firstNonEmpty(node.ClientIP, clientIdentity),
				MachineId:         node.MachineID,
				NodeId:            firstNonEmpty(node.NodeID, clientIdentity),
				NodeName:          node.NodeName,
				IsProtected:       serverRecord.IsProtected,
				Synced:            false,
				HasUpdate:         false,
				SyncStatus:        "client_missing",
				Online:            true,
				ServerId:          serverRecord.ID,
				ServerName:        serverRecord.Name,
				ServerDesc:        serverRecord.Description,
				ServerAutomaName:  serverAutomaName,
				ServerAutomaDesc:  serverAutomaDescription,
				ServerRevision:    serverRecord.Revision,
				ServerUpdatedAt:   nil,
				LastSyncedAt:      nil,
				AutomaName:        "",
				AutomaDescription: "",
			}
			if !serverRecord.LastSyncedAt.IsZero() {
				candidate.LastSyncedAt = gtime.NewFromTime(serverRecord.LastSyncedAt)
			}
			if !serverRecord.UpdatedAt.IsZero() {
				candidate.ServerUpdatedAt = gtime.NewFromTime(serverRecord.UpdatedAt)
			}

			if ok {
				synced := false
				hasUpdate := true
				status := "has_update"
				serverContentHash := strings.TrimSpace(serverRecord.ContentHash)
				clientContentHash := strings.TrimSpace(item.ContentHash)
				if serverContentHash != "" && clientContentHash != "" && serverContentHash == clientContentHash {
					synced = true
					hasUpdate = false
					status = "synced"
				} else if item.UpdatedAtAutoma > 0 && serverRecord.UpdatedAtAutoma > 0 {
					if item.UpdatedAtAutoma > serverRecord.UpdatedAtAutoma {
						status = "client_newer"
					} else if item.UpdatedAtAutoma < serverRecord.UpdatedAtAutoma {
						hasUpdate = false
						status = "server_newer"
					}
				}
				candidate.Id = item.AutomaId
				candidate.AutomaId = item.AutomaId
				candidate.WorkflowId = item.WorkflowId
				candidate.Name = item.Name
				candidate.Description = item.Description
				candidate.AutomaName = item.Name
				candidate.AutomaDescription = item.Description
				fillCandidateNode(&candidate, item, clientIdentity)
				candidate.AutomaVersion = item.AutomaVersion
				candidate.ExtVersion = item.ExtVersion
				candidate.CreatedAtAutoma = item.CreatedAtAutoma
				candidate.UpdatedAtAutoma = item.UpdatedAtAutoma
				candidate.IsDisabled = item.IsDisabled
				candidate.NodeCount = item.NodeCount
				candidate.EdgeCount = item.EdgeCount
				candidate.ContentHash = item.ContentHash
				candidate.Synced = synced
				candidate.HasUpdate = hasUpdate
				candidate.SyncStatus = status
			}
			if !matchCandidateFilters(syncStatusFilter, workflowStatusFilter, candidate.SyncStatus, candidate.Synced, candidate.HasUpdate, candidate.IsDisabled) {
				continue
			}
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

	var cacheItems []workflowcache.WorkflowItem
	if serverMode {
		if sourceIdentity != "" {
			cacheItems, err = workflowcache.ListClientWorkflows(ctx, sourceIdentity)
		} else if sourceIP != "" {
			clientIPs, listErr := listOnlineCandidateIdentities(ctx, sourceIP, "")
			if listErr != nil {
				return nil, listErr
			}
			for _, identity := range clientIPs {
				var items []workflowcache.WorkflowItem
				items, err = workflowcache.ListClientWorkflows(ctx, identity)
				if err != nil {
					return nil, err
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
			itemIdentity := workflowItemIdentity(item)
			if !workflowcache.IsClientOnline(ctx, itemIdentity) {
				continue
			}
			if automaID != "" && item.AutomaId != automaID {
				continue
			}
			if keyword != "" {
				text := strings.ToLower(strings.Join([]string{item.AutomaId, item.Name, item.Description, item.SourceIp, item.MachineId, item.NodeId, item.NodeName}, " "))
				if !strings.Contains(text, keyword) {
					continue
				}
			}
			itemAutomaID := strings.TrimSpace(item.AutomaId)
			serverRecord := serverRecords[itemAutomaID]
			synced := false
			hasUpdate := true
			status := "not_synced"
			if serverRecord != nil {
				status = "has_update"
				serverContentHash := strings.TrimSpace(serverRecord.ContentHash)
				clientContentHash := strings.TrimSpace(item.ContentHash)
				if serverContentHash != "" && clientContentHash != "" && serverContentHash == clientContentHash {
					synced = true
					hasUpdate = false
					status = "synced"
				} else if item.UpdatedAtAutoma > 0 && serverRecord.UpdatedAtAutoma > 0 {
					if item.UpdatedAtAutoma > serverRecord.UpdatedAtAutoma {
						status = "client_newer"
					} else if item.UpdatedAtAutoma < serverRecord.UpdatedAtAutoma {
						hasUpdate = false
						status = "server_newer"
					}
				}
			}
			if !matchCandidateFilters(syncStatusFilter, workflowStatusFilter, status, synced, hasUpdate, item.IsDisabled) {
				continue
			}
			if serverRecord != nil {
				serverAutomaName := strings.TrimSpace(serverRecord.AutomaName)
				if serverAutomaName == "" {
					serverAutomaName = serverRecord.Name
				}
				serverAutomaDescription := strings.TrimSpace(serverRecord.AutomaDescription)
				if serverAutomaDescription == "" {
					serverAutomaDescription = serverRecord.Description
				}
				candidate := v1.WorkflowSyncCandidatesResModel{Id: item.AutomaId, AutomaId: item.AutomaId, WorkflowId: item.WorkflowId, Name: item.Name, Description: item.Description, AutomaName: item.Name, AutomaDescription: item.Description, Source: "客户端同步", SourceIp: item.SourceIp, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: serverRecord.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, ContentHash: item.ContentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status, Online: workflowcache.IsClientOnline(ctx, itemIdentity), ServerId: serverRecord.ID, ServerName: serverRecord.Name, ServerDesc: serverRecord.Description, ServerAutomaName: serverAutomaName, ServerAutomaDesc: serverAutomaDescription, ServerRevision: serverRecord.Revision}
				if !serverRecord.LastSyncedAt.IsZero() {
					candidate.LastSyncedAt = gtime.NewFromTime(serverRecord.LastSyncedAt)
				}
				if !serverRecord.UpdatedAt.IsZero() {
					candidate.ServerUpdatedAt = gtime.NewFromTime(serverRecord.UpdatedAt)
				}
				fillCandidateNode(&candidate, item, itemIdentity)
				candidates = append(candidates, candidate)
				continue
			}
			candidate := v1.WorkflowSyncCandidatesResModel{Id: item.AutomaId, AutomaId: item.AutomaId, WorkflowId: item.WorkflowId, Name: item.Name, Description: item.Description, AutomaName: item.Name, AutomaDescription: item.Description, Source: "客户端同步", SourceIp: item.SourceIp, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, ContentHash: item.ContentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status, Online: workflowcache.IsClientOnline(ctx, itemIdentity)}
			fillCandidateNode(&candidate, item, itemIdentity)
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
	if serverMode && sourceIP == "" && sourceNodeID == "" && automaID == "" {
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
		synced := false
		hasUpdate := true
		status := "not_synced"
		if serverRecord != nil {
			status = "has_update"
			serverContentHash := strings.TrimSpace(serverRecord.ContentHash)
			clientUpdatedAt := gconv.Int64(item["updatedAt"])
			if serverContentHash != "" && itemContentHash != "" && serverContentHash == itemContentHash {
				synced = true
				hasUpdate = false
				status = "synced"
			} else if clientUpdatedAt > 0 && serverRecord.UpdatedAtAutoma > 0 {
				if clientUpdatedAt > serverRecord.UpdatedAtAutoma {
					status = "client_newer"
				} else if clientUpdatedAt < serverRecord.UpdatedAtAutoma {
					hasUpdate = false
					status = "server_newer"
				}
			}
		}
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

func getOnlineNodeSnapshot(ctx context.Context, identity string) workflowcache.OnlineNode {
	node, ok, err := workflowcache.GetOnlineNode(ctx, identity)
	if err != nil || !ok {
		return workflowcache.OnlineNode{NodeID: strings.TrimSpace(identity)}
	}
	return node
}

func listOnlineCandidateIdentities(ctx context.Context, sourceIP string, sourceNodeID string) ([]string, error) {
	sourceIP = strings.TrimSpace(sourceIP)
	sourceNodeID = strings.TrimSpace(sourceNodeID)
	if sourceNodeID != "" {
		identity := nodeConnectionIdentity(sourceIP, sourceNodeID)
		if identity != "" && workflowcache.IsClientOnline(ctx, identity) {
			return []string{identity}, nil
		}
		return []string{}, nil
	}
	if sourceIP != "" {
		return workflowcache.ListClientNodes(ctx, sourceIP)
	}
	return workflowcache.ListOnlineClients(ctx)
}

func nodeConnectionIdentity(sourceIP string, sourceNodeID string) string {
	sourceIP = strings.TrimSpace(sourceIP)
	sourceNodeID = normalizeSourceNodeID(sourceIP, sourceNodeID)
	if sourceIP == "" {
		return sourceNodeID
	}
	if sourceNodeID == "" || sourceNodeID == sourceIP {
		return sourceIP
	}
	return sourceIP + "|" + sourceNodeID
}

func normalizeSourceNodeID(sourceIP string, sourceNodeID string) string {
	sourceIP = strings.TrimSpace(sourceIP)
	sourceNodeID = strings.TrimSpace(sourceNodeID)
	if sourceIP == "" || sourceNodeID == "" {
		return sourceNodeID
	}

	prefix := sourceIP + "|"
	for strings.HasPrefix(sourceNodeID, prefix) {
		sourceNodeID = strings.TrimSpace(strings.TrimPrefix(sourceNodeID, prefix))
	}
	return sourceNodeID
}

func workflowItemIdentity(item workflowcache.WorkflowItem) string {
	return nodeConnectionIdentity(item.SourceIp, item.NodeId)
}

func fillCandidateNode(candidate *v1.WorkflowSyncCandidatesResModel, item workflowcache.WorkflowItem, identity string) {
	if candidate == nil {
		return
	}
	candidate.SourceIp = item.SourceIp
	candidate.MachineId = item.MachineId
	candidate.NodeId = firstNonEmpty(item.NodeId, identity)
	candidate.NodeName = item.NodeName
	candidate.Online = true
}

func matchCandidateFilters(syncStatusFilter string, workflowStatusFilter string, status string, synced bool, hasUpdate bool, isDisabled bool) bool {
	if syncStatusFilter != "" {
		switch syncStatusFilter {
		case "syncable":
			if status == "client_missing" || status == "server_newer" || !(hasUpdate || !synced) {
				return false
			}
		case "synced":
			if !synced {
				return false
			}
		default:
			if status != syncStatusFilter {
				return false
			}
		}
	}

	if workflowStatusFilter != "" {
		switch workflowStatusFilter {
		case "enabled":
			if isDisabled {
				return false
			}
		case "disabled":
			if !isDisabled {
				return false
			}
		}
	}

	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
