package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	localmodel "github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/llm"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/Zany2/browserflow/backend/utility/storage"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/Zany2/browserflow/backend/utility/workflowhash"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
)

// WorkflowSync syncs workflows into local file or refreshes agent snapshot 同步工作流到本地文件
func (c *ControllerV1) WorkflowSync(ctx context.Context, req *v1.WorkflowSyncReq) (res *v1.WorkflowSyncRes, err error) {
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

	if len(req.WorkflowJsonDataList) == 0 {
		sourceIP := strings.TrimSpace(req.SourceIP)
		sourceNodeID := websockets.NormalizeNodeID(sourceIP, req.SourceNodeID)
		sourceIdentity := websockets.NodeConnectionID(sourceIP, sourceNodeID)
		selectedIDs := make([]string, 0, len(req.WorkflowIds))
		for _, workflowID := range req.WorkflowIds {
			workflowID = strings.TrimSpace(workflowID)
			if workflowID != "" {
				selectedIDs = append(selectedIDs, workflowID)
			}
		}
		if sourceIP != "" && len(selectedIDs) > 0 {
			if !serverMode {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "Windows mode does not read Server client workflow cache")
				return nil, nil
			}
			if !workflowcache.IsClientOnline(ctx, sourceIdentity) {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("客户端 %s 不在线或 WebSocket 未连接", sourceIdentity))
				return nil, nil
			}
			cachedWorkflows := make([]localmodel.JSONMap, 0, len(selectedIDs))
			for _, workflowID := range selectedIDs {
				payload, ok, loadErr := workflowcache.GetClientWorkflowPayload(ctx, sourceIdentity, workflowID)
				if loadErr != nil {
					return nil, loadErr
				}
				if !ok {
					rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("客户端工作流缓存不存在：%s", workflowID))
					return nil, nil
				}
				var workflow localmodel.JSONMap
				payloadBytes, loadErr := json.Marshal(payload)
				if loadErr != nil {
					return nil, loadErr
				}
				if loadErr = json.Unmarshal(payloadBytes, &workflow); loadErr != nil {
					return nil, loadErr
				}
				cachedWorkflows = append(cachedWorkflows, workflow)
			}
			req.WorkflowJsonDataList = cachedWorkflows
			if len(req.WorkflowJsonDataList) == 0 {
				return &v1.WorkflowSyncRes{}, nil
			}
		}
	}

	if len(req.WorkflowJsonDataList) == 0 {
		if serverMode {
			return &v1.WorkflowSyncRes{}, nil
		}

		browserID := strings.TrimSpace(g.RequestFromCtx(ctx).Get("browser_id").String())
		state.AgentMu.Lock()
		var agent *state.AgentConnection
		if browserID != "" {
			agent = state.AgentConnections[browserID]
			if agent == nil {
				state.AgentMu.Unlock()
				return &v1.WorkflowSyncRes{}, errors.New("browser agent is offline")
			}
		} else {
			for _, item := range state.AgentConnections {
				agent = item
				break
			}
			if agent == nil {
				state.AgentMu.Unlock()
				return &v1.WorkflowSyncRes{}, errors.New("no browser agent online")
			}
		}
		commandID := "cmd_" + guid.S()
		resultCh := make(chan localmodel.AgentCommandResult, 1)
		state.SetPendingCommand(commandID, resultCh)
		agent.LastSeenAt = time.Now()
		state.AgentMu.Unlock()
		if sent := websockets.SendConnectionMessage(agent.ConnectionID, &localmodel.WSResponse{Type: "agent_command", BrowserID: agent.BrowserID, CommandID: commandID, Command: "automa.workflow.list", Payload: map[string]any{}}); sent <= 0 {
			state.AgentMu.Lock()
			state.RemovePendingCommand(commandID)
			state.AgentMu.Unlock()
			return &v1.WorkflowSyncRes{}, errors.New("browser agent is offline")
		}
		select {
		case result := <-resultCh:
			if !result.Success {
				if strings.TrimSpace(result.Error) != "" {
					return &v1.WorkflowSyncRes{}, errors.New(result.Error)
				}
				return &v1.WorkflowSyncRes{}, errors.New("browser agent workflow list failed")
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
			return &v1.WorkflowSyncRes{}, nil
		case <-ctx.Done():
			state.AgentMu.Lock()
			state.RemovePendingCommand(commandID)
			state.AgentMu.Unlock()
			return &v1.WorkflowSyncRes{}, ctx.Err()
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
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("第%d个工作流规范化失败", index+1))
			return nil, nil
		}
		var payload map[string]any
		if err = json.Unmarshal(bytes, &payload); err != nil {
			rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("第%d个工作流规范化失败", index+1))
			return nil, nil
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
		automaName := strings.TrimSpace(gconv.String(payload["name"]))
		automaDescription := strings.TrimSpace(gconv.String(payload["description"]))
		sourceIP := strings.TrimSpace(req.SourceIP)
		sourceNodeID := websockets.NormalizeNodeID(sourceIP, req.SourceNodeID)
		parsed := &localmodel.AutomaWorkflowRecord{AutomaID: automaID, Name: automaName, Description: automaDescription, AutomaName: automaName, AutomaDescription: automaDescription, Source: 2, SourceIP: sourceIP, SourceNodeID: sourceNodeID, AutomaVersion: strings.TrimSpace(gconv.String(payload["version"])), ExtVersion: strings.TrimSpace(gconv.String(payload["extVersion"])), CreatedAtAutoma: gconv.Int64(payload["createdAt"]), UpdatedAtAutoma: gconv.Int64(payload["updatedAt"]), IsDisabled: gconv.Bool(payload["isDisabled"]), NodeCount: nodeCount, EdgeCount: edgeCount, RawJSON: string(rawBytes), NormalizedJSON: string(rawBytes), ContentHash: contentHash}
		var existing *localmodel.AutomaWorkflowRecord
		var getErr error
		if serverMode {
			columns := dao.AutomaWorkflows.Columns()
			item := entity.AutomaWorkflows{}
			getErr = dao.AutomaWorkflows.Ctx(ctx).Where(columns.AutomaId, parsed.AutomaID).Scan(&item)
			if getErr == nil && item.Id > 0 {
				existing = &localmodel.AutomaWorkflowRecord{ID: item.Id, AutomaID: item.AutomaId, Name: item.Name, Description: item.Description, AutomaName: item.AutomaName, AutomaDescription: item.AutomaDescription, Source: item.Source, SourceIP: item.SourceIp, SourceNodeID: item.SourceNodeId, SourceUserAgent: item.SourceUserAgent, AutomaVersion: item.AutomaVersion, ExtVersion: item.ExtVersion, CreatedAtAutoma: item.CreatedAtAutoma, UpdatedAtAutoma: item.UpdatedAtAutoma, IsDisabled: item.IsDisabled, IsProtected: item.IsProtected, NodeCount: item.NodeCount, EdgeCount: item.EdgeCount, RawJSON: item.RawJson, NormalizedJSON: item.NormalizedJson, ContentHash: item.ContentHash, Revision: item.Revision}
				if item.FirstSyncedAt != nil && !item.FirstSyncedAt.IsZero() {
					existing.FirstSyncedAt = item.FirstSyncedAt.Time
				}
				if item.LastSyncedAt != nil && !item.LastSyncedAt.IsZero() {
					existing.LastSyncedAt = item.LastSyncedAt.Time
				}
				if item.CreatedAt != nil && !item.CreatedAt.IsZero() {
					existing.CreatedAt = item.CreatedAt.Time
				}
				if item.UpdatedAt != nil && !item.UpdatedAt.IsZero() {
					existing.UpdatedAt = item.UpdatedAt.Time
				}
			}
		} else {
			existing, getErr = db.GetAutomaWorkflowRecord(parsed.AutomaID)
		}
		if getErr == nil && existing != nil {
			parsed.ID = existing.ID
			parsed.CreatedAt = existing.CreatedAt
			parsed.Name = existing.Name
			parsed.Description = existing.Description
			parsed.IsProtected = existing.IsProtected
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
		if serverMode {
			saveData := do.AutomaWorkflows{AutomaId: parsed.AutomaID, Name: parsed.Name, Description: parsed.Description, AutomaName: parsed.AutomaName, AutomaDescription: parsed.AutomaDescription, Source: parsed.Source, SourceIp: parsed.SourceIP, SourceNodeId: parsed.SourceNodeID, SourceUserAgent: parsed.SourceUserAgent, AutomaVersion: parsed.AutomaVersion, ExtVersion: parsed.ExtVersion, CreatedAtAutoma: parsed.CreatedAtAutoma, UpdatedAtAutoma: parsed.UpdatedAtAutoma, IsDisabled: parsed.IsDisabled, IsProtected: parsed.IsProtected, NodeCount: parsed.NodeCount, EdgeCount: parsed.EdgeCount, RawJson: parsed.RawJSON, NormalizedJson: parsed.NormalizedJSON, ContentHash: parsed.ContentHash, Revision: parsed.Revision}
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
	if !serverMode {
		raw, _ := json.Marshal(req.WorkflowJsonDataList)
		_ = db.SaveAutomaWorkflowSnapshot(&localmodel.AutomaWorkflowSnapshot{ID: "latest", Workflows: raw})
	}
	return &v1.WorkflowSyncRes{}, nil
}
