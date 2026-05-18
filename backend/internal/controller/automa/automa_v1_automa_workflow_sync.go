package automa

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/automa/v1"
	localmodel "github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
)

// AutomaWorkflowSync syncs workflows into local file or refreshes agent snapshot 同步工作流到本地文件
func (c *ControllerV1) AutomaWorkflowSync(ctx context.Context, req *v1.AutomaWorkflowSyncReq) (res *v1.AutomaWorkflowSyncRes, err error) {
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

	if len(req.WorkflowJsonDataList) == 0 {
		browserID := strings.TrimSpace(g.RequestFromCtx(ctx).Get("browser_id").String())
		state.AgentMu.Lock()
		var agent *state.AgentConnection
		if browserID != "" {
			agent = state.AgentConnections[browserID]
			if agent == nil {
				state.AgentMu.Unlock()
				return &v1.AutomaWorkflowSyncRes{}, errors.New("browser agent is offline")
			}
		} else {
			for _, item := range state.AgentConnections {
				agent = item
				break
			}
			if agent == nil {
				state.AgentMu.Unlock()
				return &v1.AutomaWorkflowSyncRes{}, errors.New("no browser agent online")
			}
		}
		commandID := "cmd_" + guid.S()
		resultCh := make(chan localmodel.AgentCommandResult, 1)
		state.PendingCommands[commandID] = resultCh
		agent.LastSeenAt = time.Now()
		state.AgentMu.Unlock()
		if sent := websockets.SendConnectionMessage(agent.ConnectionID, &localmodel.WSResponse{Type: "agent_command", BrowserID: agent.BrowserID, CommandID: commandID, Command: "automa.workflow.list", Payload: map[string]any{}}); sent <= 0 {
			state.AgentMu.Lock()
			delete(state.PendingCommands, commandID)
			state.AgentMu.Unlock()
			return &v1.AutomaWorkflowSyncRes{}, errors.New("browser agent is offline")
		}
		select {
		case result := <-resultCh:
			if !result.Success {
				if strings.TrimSpace(result.Error) != "" {
					return &v1.AutomaWorkflowSyncRes{}, errors.New(result.Error)
				}
				return &v1.AutomaWorkflowSyncRes{}, errors.New("browser agent workflow list failed")
			}
			if result.Success {
				data := result.Data
				var wrapped struct {
					Workflows json.RawMessage `json:"workflows"`
				}
				if json.Unmarshal(data, &wrapped) == nil && len(wrapped.Workflows) > 0 {
					data = wrapped.Workflows
				}
				if len(data) == 0 {
					data = json.RawMessage("[]")
				}
				_ = db.SaveAutomaWorkflowSnapshot(&localmodel.AutomaWorkflowSnapshot{ID: "latest", Workflows: data})
			}
			return &v1.AutomaWorkflowSyncRes{}, nil
		case <-ctx.Done():
			state.AgentMu.Lock()
			delete(state.PendingCommands, commandID)
			state.AgentMu.Unlock()
			return &v1.AutomaWorkflowSyncRes{}, ctx.Err()
		}
	}

	selected := map[string]struct{}{}
	for _, workflowID := range req.WorkflowIds {
		workflowID = strings.TrimSpace(workflowID)
		if workflowID != "" {
			selected[workflowID] = struct{}{}
		}
	}
	syncTime := time.Now()
	for index, item := range req.WorkflowJsonDataList {
		bytes, err := json.Marshal(item)
		if err != nil {
			return nil, gerror.Wrapf(err, "workflow %d normalize failed", index+1)
		}
		var payload map[string]any
		if err = json.Unmarshal(bytes, &payload); err != nil {
			return nil, gerror.Wrapf(err, "workflow %d normalize failed", index+1)
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
		if len(selected) > 0 {
			if _, ok := selected[automaID]; !ok {
				continue
			}
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
		parsed := &localmodel.AutomaWorkflowRecord{AutomaID: automaID, Name: strings.TrimSpace(gconv.String(payload["name"])), Description: strings.TrimSpace(gconv.String(payload["description"])), Source: 2, SourceIP: strings.TrimSpace(req.SourceIP), AutomaVersion: strings.TrimSpace(gconv.String(payload["version"])), ExtVersion: strings.TrimSpace(gconv.String(payload["extVersion"])), CreatedAtAutoma: gconv.Int64(payload["createdAt"]), UpdatedAtAutoma: gconv.Int64(payload["updatedAt"]), IsDisabled: gconv.Bool(payload["isDisabled"]), NodeCount: nodeCount, EdgeCount: edgeCount, RawJSON: string(rawBytes), NormalizedJSON: string(rawBytes), ContentHash: contentHash}
		existing, getErr := db.GetAutomaWorkflowRecord(parsed.AutomaID)
		if getErr == nil && existing != nil {
			if existing.IsProtected {
				continue
			}
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
			}
			if parsed.ContentHash == existing.ContentHash {
				continue
			}
		} else {
			parsed.Revision = 1
		}
		if parsed.FirstSyncedAt.IsZero() {
			parsed.FirstSyncedAt = syncTime
		}
		parsed.LastSyncedAt = syncTime
		if err = db.SaveAutomaWorkflowRecord(parsed); err != nil {
			return nil, err
		}
	}
	raw, _ := json.Marshal(req.WorkflowJsonDataList)
	_ = db.SaveAutomaWorkflowSnapshot(&localmodel.AutomaWorkflowSnapshot{ID: "latest", Workflows: raw})
	return &v1.AutomaWorkflowSyncRes{}, nil
}
