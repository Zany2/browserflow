package tasks

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/utility/workflowcache"
	"github.com/gogf/gf/v2/util/gconv"
)

type dispatchTarget struct {
	ClientIP string
	NodeID   string
}

func normalizeDispatchMode(value string, nodeID string, targetGroupID int64, clientIP string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "auto", "group", "node", "ip":
		return value
	}
	if strings.TrimSpace(nodeID) != "" {
		return "node"
	}
	if targetGroupID > 0 {
		return "group"
	}
	if strings.TrimSpace(clientIP) != "" {
		return "ip"
	}
	return "auto"
}

func normalizeQueuePolicy(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "queue", "fail", "skip":
		return value
	default:
		return "queue"
	}
}

func normalizeMaxAttempts(value int) int {
	if value <= 0 {
		return 3
	}
	if value > 20 {
		return 20
	}
	return value
}

func normalizeTimeoutSeconds(value int) int {
	if value <= 0 {
		return 300
	}
	if value < 30 {
		return 30
	}
	if value > 86400 {
		return 86400
	}
	return value
}

func normalizeQueueWaitSeconds(value int) int {
	if value <= 0 {
		return 60
	}
	if value < 10 {
		return 10
	}
	if value > 86400 {
		return 86400
	}
	return value
}

func normalizeQueueRetryIntervalSeconds(value int) int {
	if value <= 0 {
		return 5
	}
	if value < 1 {
		return 1
	}
	if value > 3600 {
		return 3600
	}
	return value
}

func buildDispatchTargets(ctx context.Context, workflowID string, dispatchMode string, clientIP string, nodeID string, targetGroupID int64) ([]dispatchTarget, string, error) {
	workflowID = strings.TrimSpace(workflowID)
	dispatchMode = normalizeDispatchMode(dispatchMode, nodeID, targetGroupID, clientIP)
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if workflowID == "" {
		return nil, "工作流不能为空", nil
	}

	switch dispatchMode {
	case "node":
		if clientIP == "" || nodeID == "" {
			return nil, "指定节点调度必须同时选择客户端和执行节点", nil
		}
		targetIdentity := nodeConnectionIdentity(clientIP, nodeID)
		if !workflowcache.IsClientOnline(ctx, targetIdentity) {
			return nil, "执行节点不在线或 WebSocket 未连接", nil
		}
		if _, ok, err := workflowcache.GetClientWorkflow(ctx, targetIdentity, workflowID); err != nil {
			return nil, "", err
		} else if !ok {
			return nil, "执行节点没有该工作流", nil
		}
		return []dispatchTarget{{ClientIP: clientIP, NodeID: nodeID}}, "", nil
	case "ip":
		if clientIP == "" {
			return nil, "指定客户端调度必须选择客户端", nil
		}
		targets, err := listWorkflowTargets(ctx, workflowID, func(item workflowcache.WorkflowItem) bool {
			return strings.TrimSpace(item.SourceIp) == clientIP
		})
		if err != nil {
			return nil, "", err
		}
		if len(targets) == 0 {
			return nil, "该客户端下没有在线且拥有该工作流的执行节点", nil
		}
		return targets, "", nil
	case "group":
		if targetGroupID <= 0 {
			return nil, "指定分组调度必须选择节点分组", nil
		}
		members, err := listNodeGroupTargets(ctx, targetGroupID)
		if err != nil {
			return nil, "", err
		}
		if len(members) == 0 {
			return nil, "节点分组下没有可用执行节点", nil
		}
		memberSet := make(map[string]struct{}, len(members))
		for _, member := range members {
			memberSet[nodeConnectionIdentity(member.ClientIP, member.NodeID)] = struct{}{}
		}
		targets, err := listWorkflowTargets(ctx, workflowID, func(item workflowcache.WorkflowItem) bool {
			_, ok := memberSet[nodeConnectionIdentity(item.SourceIp, item.NodeId)]
			return ok
		})
		if err != nil {
			return nil, "", err
		}
		if len(targets) == 0 {
			return nil, "节点分组下没有在线且拥有该工作流的执行节点", nil
		}
		return targets, "", nil
	default:
		targets, err := listWorkflowTargets(ctx, workflowID, nil)
		if err != nil {
			return nil, "", err
		}
		if len(targets) == 0 {
			return nil, "没有在线客户端拥有该工作流", nil
		}
		return targets, "", nil
	}
}

func listWorkflowTargets(ctx context.Context, workflowID string, keep func(workflowcache.WorkflowItem) bool) ([]dispatchTarget, error) {
	items, err := workflowcache.ListWorkflowClients(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	targets := make([]dispatchTarget, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if keep != nil && !keep(item) {
			continue
		}
		target := dispatchTarget{
			ClientIP: strings.TrimSpace(item.SourceIp),
			NodeID:   strings.TrimSpace(item.NodeId),
		}
		identity := nodeConnectionIdentity(target.ClientIP, target.NodeID)
		if identity == "" {
			continue
		}
		if _, ok := seen[identity]; ok {
			continue
		}
		seen[identity] = struct{}{}
		targets = append(targets, target)
	}
	return targets, nil
}

func listNodeGroupTargets(ctx context.Context, groupID int64) ([]dispatchTarget, error) {
	columns := dao.NodeGroupMembers.Columns()
	records, err := dao.NodeGroupMembers.Ctx(ctx).
		Fields(columns.ClientIp, columns.NodeId).
		Where(columns.GroupId, groupID).
		OrderAsc(columns.Id).
		All()
	if err != nil {
		return nil, err
	}
	targets := make([]dispatchTarget, 0, len(records))
	for _, record := range records {
		targets = append(targets, dispatchTarget{
			ClientIP: strings.TrimSpace(gconv.String(record[columns.ClientIp])),
			NodeID:   strings.TrimSpace(gconv.String(record[columns.NodeId])),
		})
	}
	return targets, nil
}

func nodeConnectionIdentity(clientIP string, nodeID string) string {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" {
		return nodeID
	}
	if nodeID == "" || nodeID == clientIP {
		return clientIP
	}
	return clientIP + "|" + nodeID
}

func busyPolicyMessage(policy string) (string, string) {
	switch normalizeQueuePolicy(policy) {
	case "skip":
		return "cancelled", "所有匹配的执行节点都在执行其他任务，本次任务已按繁忙策略跳过"
	case "fail":
		return "failed", "所有匹配的执行节点都在执行其他任务，任务执行失败"
	default:
		return "failed", "所有匹配的执行节点都在执行其他任务，请稍后重试"
	}
}
