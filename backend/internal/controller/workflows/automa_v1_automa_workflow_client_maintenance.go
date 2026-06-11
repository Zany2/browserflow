package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/api/workflows/v1"
	"github.com/Zany2/browserflow/backend/internal/consts"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/state"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/Zany2/browserflow/backend/utility/workflowagent"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// WorkflowClientMaintenance installs, updates, or deletes workflows on client nodes. 维护指定客户端节点工作流。
func (c *ControllerV1) WorkflowClientMaintenance(ctx context.Context, req *v1.WorkflowClientMaintenanceReq) (res *v1.WorkflowClientMaintenanceRes, err error) {
	if consts.ResolveRuntimeMode(ctx) != consts.RuntimeModeServer {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "当前接口仅支持 Server 模式")
		return nil, nil
	}

	sourceIP := strings.TrimSpace(req.SourceIP)
	sourceNodeIDs := normalizeSourceNodeIDs(sourceIP, req.SourceNodeIDs)
	if len(sourceNodeIDs) == 0 {
		sourceNodeIDs = normalizeSourceNodeIDs(sourceIP, []string{req.SourceNodeID})
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if sourceIP == "" || len(sourceNodeIDs) == 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "请选择客户端和执行节点")
		return nil, nil
	}

	sourceIdentities := make([]string, 0, len(sourceNodeIDs))
	for _, sourceNodeID := range sourceNodeIDs {
		sourceIdentities = append(sourceIdentities, websockets.NodeConnectionID(sourceIP, sourceNodeID))
	}
	onlineSet, onlineErr := workflowcache.GetOnlineClientSet(ctx, sourceIdentities)
	if onlineErr != nil {
		return nil, onlineErr
	}
	offlineNodes := make([]string, 0)
	for _, sourceNodeID := range sourceNodeIDs {
		sourceIdentity := websockets.NodeConnectionID(sourceIP, sourceNodeID)
		if _, ok := onlineSet[sourceIdentity]; sourceIdentity == "" || !ok {
			offlineNodes = append(offlineNodes, firstNonEmpty(sourceIdentity, sourceNodeID))
		}
	}
	if len(offlineNodes) > 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("以下客户端节点不在线或 WebSocket 未连接：%s", strings.Join(offlineNodes, "、")))
		return nil, nil
	}

	workflowIDs := normalizeWorkflowIDs(req.WorkflowIds)
	if len(workflowIDs) == 0 {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "请选择需要维护的工作流")
		return nil, nil
	}

	payload := map[string]any{
		"action":       action,
		"workflow_ids": workflowIDs,
	}
	if action == "install" || action == "update" {
		workflowPayloads := make([]map[string]any, 0, len(workflowIDs))
		columns := dao.AutomaWorkflows.Columns()
		items := make([]entity.AutomaWorkflows, 0, len(workflowIDs))
		if err = dao.AutomaWorkflows.Ctx(ctx).WhereIn(columns.AutomaId, workflowIDs).Scan(&items); err != nil {
			return nil, err
		}
		workflowMap := make(map[string]entity.AutomaWorkflows, len(items))
		for _, item := range items {
			workflowMap[item.AutomaId] = item
		}
		for _, workflowID := range workflowIDs {
			item, ok := workflowMap[workflowID]
			if !ok || item.Id <= 0 {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("服务端工作流不存在：%s", workflowID))
				return nil, nil
			}
			if item.IsProtected {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("该工作流不允许客户端同步到本地：%s", firstNonEmpty(item.Name, item.AutomaName, workflowID)))
				return nil, nil
			}
			workflowPayload := map[string]any{}
			rawJSON := strings.TrimSpace(item.NormalizedJson)
			if rawJSON == "" {
				rawJSON = strings.TrimSpace(item.RawJson)
			}
			if rawJSON == "" || json.Unmarshal([]byte(rawJSON), &workflowPayload) != nil {
				rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), fmt.Sprintf("服务端工作流 JSON 无效：%s", workflowID))
				return nil, nil
			}
			workflowPayload["id"] = item.AutomaId
			workflowPayload["createdAt"] = firstNonZeroInt64(item.CreatedAtAutoma, workflowPayload["createdAt"], workflowPayload["created_at"])
			workflowPayload["updatedAt"] = firstNonZeroInt64(item.UpdatedAtAutoma, workflowPayload["updatedAt"], workflowPayload["updated_at"])
			workflowPayloads = append(workflowPayloads, workflowPayload)
		}
		payload["workflows"] = workflowPayloads
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	var (
		totalSubmitted int
		totalSucceeded int
		totalFailed    int
		resultIDs      []string
		failedNodes    []string
		failedMessages []string
	)
	for _, sourceNodeID := range sourceNodeIDs {
		sourceIdentity := websockets.NodeConnectionID(sourceIP, sourceNodeID)
		commandID := "cmd_" + guid.S()
		resultCh := make(chan model.AgentCommandResult, 1)
		state.SetPendingCommand(commandID, resultCh)

		node, ok, nodeErr := workflowcache.GetOnlineNode(ctx, sourceIdentity)
		if nodeErr != nil {
			state.RemovePendingCommand(commandID)
			return nil, nodeErr
		}
		if !ok {
			node.ClientIP = sourceIP
			node.NodeID = sourceNodeID
		}

		if sent := websockets.SendConnectionMessage(sourceIdentity, &model.WSResponse{
			Type:      model.WSMessageTypeAgentCommand,
			ClientIP:  sourceIP,
			NodeID:    sourceNodeID,
			NodeName:  node.NodeName,
			MachineID: node.MachineID,
			CommandID: commandID,
			Command:   "automa.workflow.maintain",
			Payload:   payload,
		}); sent <= 0 {
			state.RemovePendingCommand(commandID)
			totalSubmitted += len(workflowIDs)
			totalFailed += len(workflowIDs)
			failedNodes = append(failedNodes, sourceIdentity)
			continue
		}

		timer := time.NewTimer(timeout)
		var result model.AgentCommandResult
		select {
		case result = <-resultCh:
		case <-timer.C:
			state.RemovePendingCommand(commandID)
			totalSubmitted += len(workflowIDs)
			totalFailed += len(workflowIDs)
			failedNodes = append(failedNodes, sourceIdentity)
			failedMessages = append(failedMessages, fmt.Sprintf("%s：等待客户端响应超时", sourceIdentity))
			timer.Stop()
			continue
		case <-ctx.Done():
			state.RemovePendingCommand(commandID)
			timer.Stop()
			return nil, ctx.Err()
		}
		timer.Stop()
		if !result.Success {
			totalSubmitted += len(workflowIDs)
			totalFailed += len(workflowIDs)
			failedNodes = append(failedNodes, sourceIdentity)
			if strings.TrimSpace(result.Error) != "" {
				failedMessages = append(failedMessages, fmt.Sprintf("%s：%s", sourceIdentity, strings.TrimSpace(result.Error)))
			}
			continue
		}

		var resultData struct {
			Submitted   int      `json:"submitted"`
			Succeeded   int      `json:"succeeded"`
			Failed      int      `json:"failed"`
			WorkflowIds []string `json:"workflow_ids"`
			Message     string   `json:"message"`
		}
		if len(result.Data) > 0 {
			_ = json.Unmarshal(result.Data, &resultData)
		}
		if resultData.Submitted <= 0 {
			resultData.Submitted = len(workflowIDs)
		}
		if resultData.Succeeded <= 0 && resultData.Failed <= 0 {
			resultData.Succeeded = resultData.Submitted
		}
		if len(resultData.WorkflowIds) == 0 {
			resultData.WorkflowIds = workflowIDs
		}
		totalSubmitted += resultData.Submitted
		totalSucceeded += resultData.Succeeded
		totalFailed += resultData.Failed
		resultIDs = append(resultIDs, resultData.WorkflowIds...)
	}

	if req.Refresh {
		for _, sourceNodeID := range sourceNodeIDs {
			_ = workflowagent.RefreshClientInventory(ctx, websockets.NodeConnectionID(sourceIP, sourceNodeID), 8*time.Second)
		}
	}

	resultIDs = normalizeWorkflowIDs(resultIDs)
	if len(resultIDs) == 0 {
		resultIDs = workflowIDs
	}
	message := fmt.Sprintf("客户端维护完成：节点 %d 个，成功 %d 个，失败 %d 个", len(sourceNodeIDs), totalSucceeded, totalFailed)
	if len(failedNodes) > 0 {
		message = fmt.Sprintf("%s，失败节点：%s", message, strings.Join(failedNodes, "、"))
	}
	if len(failedMessages) > 0 {
		message = fmt.Sprintf("%s，失败原因：%s", message, strings.Join(failedMessages, "；"))
	}

	return &v1.WorkflowClientMaintenanceRes{
		Submitted:   totalSubmitted,
		Succeeded:   totalSucceeded,
		Failed:      totalFailed,
		WorkflowIds: resultIDs,
		Message:     message,
	}, nil
}

func normalizeSourceNodeIDs(sourceIP string, values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = websockets.NormalizeNodeID(sourceIP, value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeWorkflowIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func firstNonZeroInt64(values ...any) int64 {
	for _, value := range values {
		switch typedValue := value.(type) {
		case int64:
			if typedValue > 0 {
				return typedValue
			}
		case int:
			if typedValue > 0 {
				return int64(typedValue)
			}
		case float64:
			if typedValue > 0 {
				return int64(typedValue)
			}
		case string:
			var parsed int64
			if _, err := fmt.Sscan(strings.TrimSpace(typedValue), &parsed); err == nil && parsed > 0 {
				return parsed
			}
		}
	}
	return 0
}
