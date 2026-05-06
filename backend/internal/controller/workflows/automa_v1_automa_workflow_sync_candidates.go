package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// WorkflowSyncCandidates lists local-file sync candidates 获取本地文件同步候选
func (c *ControllerV1) WorkflowSyncCandidates(ctx context.Context, req *v1.WorkflowSyncCandidatesReq) (res *v1.WorkflowSyncCandidatesRes, err error) {
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
	db := state.DB
	state.DBMu.Unlock()

	records, err := db.ListAutomaWorkflowRecords()
	if err != nil {
		return nil, err
	}
	serverRecords := make(map[string]*model.AutomaWorkflowRecord, len(records))
	for _, record := range records {
		serverRecords[record.AutomaID] = record
	}
	items := make([]map[string]any, 0)
	if snapshot, err := db.GetAutomaWorkflowSnapshot("latest"); err == nil && len(snapshot.Workflows) > 0 {
		_ = json.Unmarshal(snapshot.Workflows, &items)
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

	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	candidates := make([]v1.WorkflowSyncCandidatesResModel, 0, len(items))
	for _, item := range items {
		hashDrawflowValue := item["drawflow"]
		if drawflowText, ok := hashDrawflowValue.(string); ok {
			drawflowText = strings.TrimSpace(drawflowText)
			if drawflowText != "" {
				var parsedDrawflow any
				if json.Unmarshal([]byte(drawflowText), &parsedDrawflow) == nil {
					hashDrawflowValue = parsedDrawflow
				}
			}
		}
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
		coreJSONBytes, err := json.Marshal(g.Map{"id": strings.TrimSpace(gconv.String(item["id"])), "name": strings.TrimSpace(gconv.String(item["name"])), "icon": strings.TrimSpace(gconv.String(item["icon"])), "table": hashTableValue, "drawflow": hashDrawflowValue, "settings": hashSettingsValue, "globalData": hashGlobalDataValue, "description": strings.TrimSpace(gconv.String(item["description"]))})
		if err != nil {
			continue
		}
		sum := sha256.Sum256(coreJSONBytes)
		contentHash := hex.EncodeToString(sum[:])
		automaID := ""
		for _, key := range []string{"id", "workflow_id", "workflowId"} {
			value := strings.TrimSpace(gconv.String(item[key]))
			if value != "" {
				automaID = value
				break
			}
		}
		if automaID == "" {
			automaID = "generated:" + contentHash[:20]
		}
		if req.AutomaID != "" && automaID != req.AutomaID {
			continue
		}
		name := strings.TrimSpace(gconv.String(item["name"]))
		description := strings.TrimSpace(gconv.String(item["description"]))
		if keyword != "" {
			text := strings.ToLower(strings.Join([]string{automaID, name, description, req.SourceIP}, " "))
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
		serverRecord := serverRecords[automaID]
		synced, hasUpdate, status := false, true, "not_synced"
		if serverRecord != nil {
			if serverRecord.IsProtected {
				synced, hasUpdate, status = false, false, "protected"
			} else if strings.TrimSpace(serverRecord.ContentHash) != "" && serverRecord.ContentHash == contentHash {
				synced, hasUpdate, status = true, false, "synced"
			} else {
				status = "has_update"
				updatedAtAutoma := gconv.Int64(item["updatedAt"])
				if updatedAtAutoma > 0 && serverRecord.UpdatedAtAutoma > 0 {
					if updatedAtAutoma > serverRecord.UpdatedAtAutoma {
						status = "client_newer"
					}
					if updatedAtAutoma < serverRecord.UpdatedAtAutoma {
						status = "server_newer"
					}
				}
			}
		}
		candidate := v1.WorkflowSyncCandidatesResModel{Id: automaID, AutomaId: automaID, WorkflowId: automaID, Name: name, Description: description, Source: "客户端同步", SourceIp: req.SourceIP, AutomaVersion: strings.TrimSpace(gconv.String(item["version"])), ExtVersion: strings.TrimSpace(gconv.String(item["extVersion"])), CreatedAtAutoma: gconv.Int64(item["createdAt"]), UpdatedAtAutoma: gconv.Int64(item["updatedAt"]), IsDisabled: gconv.Bool(item["isDisabled"]), NodeCount: nodeCount, EdgeCount: edgeCount, ContentHash: contentHash, Synced: synced, HasUpdate: hasUpdate, SyncStatus: status}
		if serverRecord != nil {
			candidate.ServerId = serverRecord.ID
			candidate.ServerName = serverRecord.Name
			candidate.ServerDesc = serverRecord.Description
			candidate.ServerRevision = serverRecord.Revision
			candidate.IsProtected = serverRecord.IsProtected
			if !serverRecord.LastSyncedAt.IsZero() {
				candidate.LastSyncedAt = gtime.NewFromTime(serverRecord.LastSyncedAt)
			}
			if !serverRecord.UpdatedAt.IsZero() {
				candidate.ServerUpdatedAt = gtime.NewFromTime(serverRecord.UpdatedAt)
			}
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
