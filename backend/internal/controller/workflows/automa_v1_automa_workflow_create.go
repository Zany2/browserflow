package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// WorkflowCreate creates or updates local workflow records 新增或更新本地工作流
func (c *ControllerV1) WorkflowCreate(ctx context.Context, req *v1.WorkflowCreateReq) (res *v1.WorkflowCreateRes, err error) {
	if len(req.WorkflowFiles) == 0 {
		return nil, gerror.New("工作流列表不能为空")
	}
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

	var metas []v1.WorkflowCreateMeta
	if workflowMetas := strings.TrimSpace(req.WorkflowMetas); workflowMetas != "" {
		if err = json.Unmarshal([]byte(workflowMetas), &metas); err != nil {
			return nil, gerror.Wrap(err, "工作流元数据解析失败")
		}
	}

	stats := v1.WorkflowMutationStats{Submitted: len(req.WorkflowFiles)}
	for index, workflowFile := range req.WorkflowFiles {
		if workflowFile == nil {
			return nil, gerror.Newf("第%d个工作流文件不能为空", index+1)
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
			return nil, gerror.Wrapf(err, "第%d个工作流 JSON 格式不正确", index+1)
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
		hashDrawflowValue := payload["drawflow"]
		if drawflowText, ok := hashDrawflowValue.(string); ok {
			drawflowText = strings.TrimSpace(drawflowText)
			if drawflowText != "" {
				var parsedDrawflow any
				if json.Unmarshal([]byte(drawflowText), &parsedDrawflow) == nil {
					hashDrawflowValue = parsedDrawflow
				}
			}
		}
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
		coreJSONBytes, err := json.Marshal(g.Map{"id": strings.TrimSpace(gconv.String(payload["id"])), "name": strings.TrimSpace(gconv.String(payload["name"])), "icon": strings.TrimSpace(gconv.String(payload["icon"])), "table": hashTableValue, "drawflow": hashDrawflowValue, "settings": hashSettingsValue, "globalData": hashGlobalDataValue, "description": strings.TrimSpace(gconv.String(payload["description"]))})
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
		name := strings.TrimSpace(meta.Name)
		if name == "" {
			name = strings.TrimSpace(gconv.String(payload["name"]))
		}
		description := strings.TrimSpace(meta.Description)
		if description == "" {
			description = strings.TrimSpace(gconv.String(payload["description"]))
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
		parsed := &model.AutomaWorkflowRecord{AutomaID: automaID, Name: name, Description: description, Source: meta.Source, IsProtected: meta.IsProtected, AutomaVersion: strings.TrimSpace(gconv.String(payload["version"])), ExtVersion: strings.TrimSpace(gconv.String(payload["extVersion"])), CreatedAtAutoma: gconv.Int64(payload["createdAt"]), UpdatedAtAutoma: gconv.Int64(payload["updatedAt"]), IsDisabled: gconv.Bool(payload["isDisabled"]), NodeCount: nodeCount, EdgeCount: edgeCount, RawJSON: string(rawBytes), NormalizedJSON: string(rawBytes), ContentHash: contentHash}
		existing, getErr := db.GetAutomaWorkflowRecord(parsed.AutomaID)
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
			}
			if parsed.ContentHash == existing.ContentHash {
				stateText = "unchanged"
			}
		} else {
			parsed.Revision = 1
		}
		if stateText != "unchanged" {
			if parsed.FirstSyncedAt.IsZero() {
				parsed.FirstSyncedAt = time.Time{}
			}
			if err = db.SaveAutomaWorkflowRecord(parsed); err != nil {
				return nil, err
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
