package model

// AgentConnectionPayload agent registration payload 客户端注册负载
type AgentConnectionPayload struct {
	ClientID        string
	ClientName      string
	ClientIP        string
	UserAgent       string
	Role            string
	AutomaInstalled bool
	AutomaVersion   string
}
