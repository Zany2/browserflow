package agents

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/agents/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
)

// Status lists online browser agents 返回在线浏览器执行端状态
func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	state.AgentMu.Lock()
	defer state.AgentMu.Unlock()

	statuses := make([]model.AgentStatus, 0, len(state.AgentConnections))
	for browserID, agent := range state.AgentConnections {
		// Keep nil entries defensive 兼容空连接
		if agent == nil {
			statuses = append(statuses, model.AgentStatus{BrowserID: browserID, Online: false})
			continue
		}
		if strings.ToLower(strings.TrimSpace(agent.Role)) == "client_agent" {
			continue
		}
		statuses = append(statuses, model.AgentStatus{
			BrowserID:       agent.BrowserID,
			Online:          true,
			AutomaInstalled: agent.AutomaInstalled,
			ConnectedAt:     agent.ConnectedAt,
			LastSeenAt:      agent.LastSeenAt,
		})
	}

	return &v1.StatusRes{Agents: statuses}, nil
}
