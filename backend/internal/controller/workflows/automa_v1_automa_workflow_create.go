package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/Zany2/browserflow/backend/utility/workflowhash"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// WorkflowCreate creates or updates local workflow records 新增或更新本地工作流
func (c *ControllerV1) WorkflowCreate(ctx context.Context, req *v1.WorkflowCreateReq) (res *v1.WorkflowCreateRes, err error) {
	if len(req.WorkflowFiles) == 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流列表不能为空")
		return nil, nil
	}
	serverMode := consts.ResolveRuntimeMode(ctx) == consts.RuntimeModeServer
	var db *storage.BoltDB
	if !serverMode {
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
	}

	var metas []v1.WorkflowCreateMeta
	if workflowMetas := strings.TrimSpace(req.WorkflowMetas); workflowMetas != "" {
		if err = json.Unmarshal([]byte(workflowMetas), &metas); err != nil {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "工作流元数据解析失败")
			return nil, nil
		}
	}

	stats := v1.WorkflowMutationStats{Submitted: len(req.WorkflowFiles)}
	for index, workflowFile := range req.WorkflowFiles {
		if workflowFile == nil {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("第%d个工作流文件不能为空", index+1))
			return nil, nil
		}
		file, err := workflowFile.Open()
		if err != nil {
			return nil, err
		}
		rawJSONBytes, readErr := io.ReadAll(file)
		_ = file.Close()
		if readErr != nil {
			return nil, readErr
		}
		var payload map[string]any
		if err = json.Unmarshal(rawJSONBytes, &payload); err != nil {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("第%d个工作流 JSON 格式不正确", index+1))
			return nil, nil
		}

		meta := v1.WorkflowCreateMeta{Source: 1}
		if index < len(metas) {
			meta = metas[index]
		}
		if meta.Source == 0 {
			meta.Source = 1
		}
		rawBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		hashDrawflowValue := workflowhash.NormalizeDrawflowForHash(payload["drawflow"])
		hashTableValue := payload["table"]
		if hashTableValue == nil {
			hashTableValue = payload["dataColumns"]
		}
		if hashTableValue == nil {
			hashTableValue = []any{}
		}
		hashSettingsValue := payload["settings"]
		if hashSettingsValue == nil {
			hashSettingsValue = g.Map{}
		}
		hashGlobalDataValue := payload["globalData"]
		if hashGlobalDataValue == nil {
			hashGlobalDataValue = ""
		}
		coreJSONBytes, err := json.Marshal(g.Map{"name": strings.TrimSpace(gconv.String(payload["name"])), "icon": strings.TrimSpace(gconv.String(payload["icon"])), "table": hashTableValue, "drawflow": hashDrawflowValue, "settings": hashSettingsValue, "globalData": hashGlobalDataValue, "description": strings.TrimSpace(gconv.String(payload["description"]))})
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(coreJSONBytes)
		contentHash := hex.EncodeToString(sum[:])
		automaID := ""
		for _, key := range []string{"id", "workflow_id", "workflowId"} {
			value := strings.TrimSpace(gconv.String(payload[key]))
			if value != "" {
				automaID = value
				break
			}
		}
		if automaID == "" {
			automaID = "generated:" + contentHash[:20]
		}
		automaName := strings.TrimSpace(gconv.String(payload["name"]))
		automaDescription := strings.TrimSpace(gconv.String(payload["description"]))
		name := strings.TrimSpace(meta.Name)
		if name == "" {
			name = automaName
		}
		description := strings.TrimSpace(meta.Description)
		if description == "" {
			description = automaDescription
		}
		drawflowValue := payload["drawflow"]
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
		if meta.Source != 1 && meta.Source != 2 {
			meta.Source = 1
		}
		parsed := &model.AutomaWorkflowRecord{AutomaID: automaID, Name: name, Description: description, AutomaName: automaName, AutomaDescription: automaDescription, Source: meta.Source, IsProtected: meta.IsProtected, AutomaVersion: strings.TrimSpace(gconv.String(payload["version"])), ExtVersion: strings.TrimSpace(gconv.String(payload["extVersion"])), CreatedAtAutoma: gconv.Int64(payload["createdAt"]), UpdatedAtAutoma: gconv.Int64(payload["updatedAt"]), IsDisabled: gconv.Bool(payload["isDisabled"]), NodeCount: nodeCount, EdgeCount: edgeCount, RawJSON: string(rawBytes), NormalizedJSON: string(rawBytes), ContentHash: contentHash}
		var existing *model.AutomaWorkflowRecord
		var getErr error
		if serverMode {
			columns := dao.AutomaWorkflows.Columns()
			item := entity.AutomaWorkflows{}
			getErr = dao.AutomaWorkflows.Ctx(ctx).Where(columns.AutomaId, parsed.AutomaID).Scan(&item)
			if getErr == nil && item.Id > 0 {
				existing = &model.AutomaWorkflowRecord{ID: item.Id, AutomaID: item.AutomaId, Name: item.Name, Description: item.Description, AutomaName: item.AutomaName, AutomaDescription: item.AutomaDescription, Source: item.Source, SourceIP: item.SourceIp, SourceUserAgent: item.SourceUserAgent, IsProtected: item.IsProtected, ContentHash: item.ContentHash, Revision: item.Revision}
				if item.FirstSyncedAt != nil && !item.FirstSyncedAt.IsZero() {
					existing.FirstSyncedAt = item.FirstSyncedAt.Time
				}
				if item.LastSyncedAt != nil && !item.LastSyncedAt.IsZero() {
					existing.LastSyncedAt = item.LastSyncedAt.Time
				}
				if item.CreatedAt != nil && !item.CreatedAt.IsZero() {
					existing.CreatedAt = item.CreatedAt.Time
				}
			}
		} else {
			existing, getErr = db.GetAutomaWorkflowRecord(parsed.AutomaID)
		}
		stateText := "created"
		if getErr == nil && existing != nil {
			parsed.ID = existing.ID
			parsed.CreatedAt = existing.CreatedAt
			parsed.FirstSyncedAt = existing.FirstSyncedAt
			parsed.LastSyncedAt = existing.LastSyncedAt
			parsed.Revision = existing.Revision
			if parsed.Revision <= 0 {
				parsed.Revision = 1
			}
			if strings.TrimSpace(existing.ContentHash) != parsed.ContentHash {
				parsed.Revision++
				stateText = "updated"
				if strings.TrimSpace(meta.Name) == "" {
					parsed.Name = existing.Name
				}
				if strings.TrimSpace(meta.Description) == "" {
					parsed.Description = existing.Description
				}
			}
			if parsed.ContentHash == existing.ContentHash {
				stateText = "unchanged"
				metadataData := do.AutomaWorkflows{}
				metadataChanged := false
				if strings.TrimSpace(meta.Name) != "" && strings.TrimSpace(existing.Name) != parsed.Name {
					metadataData.Name = parsed.Name
					metadataChanged = true
				}
				if strings.TrimSpace(meta.Description) != "" && strings.TrimSpace(existing.Description) != parsed.Description {
					metadataData.Description = parsed.Description
					metadataChanged = true
				}
				if existing.Source != parsed.Source {
					metadataData.Source = parsed.Source
					metadataChanged = true
				}
				if existing.IsProtected != parsed.IsProtected {
					metadataData.IsProtected = parsed.IsProtected
					metadataChanged = true
				}
				if strings.TrimSpace(existing.AutomaName) != parsed.AutomaName {
					metadataData.AutomaName = parsed.AutomaName
					metadataChanged = true
				}
				if strings.TrimSpace(existing.AutomaDescription) != parsed.AutomaDescription {
					metadataData.AutomaDescription = parsed.AutomaDescription
					metadataChanged = true
				}
				if metadataChanged {
					stateText = "updated"
					if serverMode {
						if _, err = dao.AutomaWorkflows.Ctx(ctx).WherePri(parsed.ID).Data(metadataData).Update(); err != nil {
							return nil, err
						}
					} else {
						if strings.TrimSpace(meta.Name) != "" {
							existing.Name = parsed.Name
						}
						if strings.TrimSpace(meta.Description) != "" {
							existing.Description = parsed.Description
						}
						existing.AutomaName = parsed.AutomaName
						existing.AutomaDescription = parsed.AutomaDescription
						existing.Source = parsed.Source
						existing.IsProtected = parsed.IsProtected
						if err = db.SaveAutomaWorkflowRecord(existing); err != nil {
							return nil, err
						}
					}
				}
			}
		} else {
			parsed.Revision = 1
		}
		if stateText != "unchanged" {
			if parsed.FirstSyncedAt.IsZero() {
				parsed.FirstSyncedAt = time.Time{}
			}
			if serverMode {
				saveData := do.AutomaWorkflows{AutomaId: parsed.AutomaID, Name: parsed.Name, Description: parsed.Description, AutomaName: parsed.AutomaName, AutomaDescription: parsed.AutomaDescription, Source: parsed.Source, SourceIp: parsed.SourceIP, SourceUserAgent: parsed.SourceUserAgent, AutomaVersion: parsed.AutomaVersion, ExtVersion: parsed.ExtVersion, CreatedAtAutoma: parsed.CreatedAtAutoma, UpdatedAtAutoma: parsed.UpdatedAtAutoma, IsDisabled: parsed.IsDisabled, IsProtected: parsed.IsProtected, NodeCount: parsed.NodeCount, EdgeCount: parsed.EdgeCount, RawJson: parsed.RawJSON, NormalizedJson: parsed.NormalizedJSON, ContentHash: parsed.ContentHash, Revision: parsed.Revision}
				if !parsed.FirstSyncedAt.IsZero() {
					saveData.FirstSyncedAt = gtime.NewFromTime(parsed.FirstSyncedAt)
				}
				if !parsed.LastSyncedAt.IsZero() {
					saveData.LastSyncedAt = gtime.NewFromTime(parsed.LastSyncedAt)
				}
				if parsed.ID > 0 {
					if _, err = dao.AutomaWorkflows.Ctx(ctx).WherePri(parsed.ID).Data(saveData).Update(); err != nil {
						return nil, err
					}
				} else if _, err = dao.AutomaWorkflows.Ctx(ctx).Data(saveData).InsertAndGetId(); err != nil {
					return nil, err
				}
			} else {
				if err = db.SaveAutomaWorkflowRecord(parsed); err != nil {
					return nil, err
				}
			}
		}
		switch stateText {
		case "created":
			stats.Created++
		case "updated":
			stats.Updated++
		default:
			stats.Unchanged++
		}
	}

	return &v1.WorkflowCreateRes{WorkflowMutationStats: stats}, nil
}
