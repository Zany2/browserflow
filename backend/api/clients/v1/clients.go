package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	"github.com/gogf/gf/v2/frame/g"
)

// ClientListReq 客户端列表请求
type ClientListReq struct {
	g.Meta  `path:"/" method:"get" tags:"客户端" summary:"获取客户端列表"`
	Status  string `json:"status,omitempty" in:"query" dc:"客户端状态"`
	IP      string `json:"ip,omitempty" in:"query" dc:"客户端IP"`
	Keyword string `json:"keyword,omitempty" in:"query" dc:"关键字"`
}

// ClientListResModel client list item 客户端列表项
type ClientListResModel = entity.Clients

// ClientResModel client detail item 客户端详情项
type ClientResModel = entity.Clients

// ClientListRes 客户端列表响应
type ClientListRes struct {
	List  []ClientListResModel `json:"list,omitempty" dc:"客户端列表"`
	Total int                  `json:"total" dc:"客户端总数"`
}

// ClientDetailReq 客户端详情请求
type ClientDetailReq struct {
	g.Meta `path:"/{id}" method:"get" tags:"客户端" summary:"获取客户端详情"`
	ID     string `json:"id" in:"path" dc:"客户端ID"`
}

// ClientDetailRes 客户端详情响应
type ClientDetailRes struct {
	Client *ClientResModel `json:"client,omitempty" dc:"客户端详情"`
}

// ClientUpdateReq 更新客户端请求
type ClientUpdateReq struct {
	g.Meta     `path:"/{id}" method:"put" tags:"客户端" summary:"更新客户端"`
	ID         string `json:"id" in:"path" dc:"客户端ID"`
	ClientName string `json:"client_name" dc:"客户端名称"`
}

// ClientUpdateRes 更新客户端响应
type ClientUpdateRes struct {
	Client  *ClientResModel `json:"client,omitempty" dc:"客户端详情"`
	Message string          `json:"message,omitempty" dc:"更新结果消息"`
}

// ClientCheckReq 客户端连接检查请求
type ClientCheckReq struct {
	g.Meta   `path:"/check" method:"get" tags:"客户端" summary:"检查客户端连接"`
	ClientID string `json:"client_id,omitempty" in:"query" dc:"客户端ID"`
}

// ClientCheckRes 客户端连接检查响应
type ClientCheckRes struct {
	Allowed  bool            `json:"allowed" dc:"是否允许连接"`
	IsBanned bool            `json:"is_banned" dc:"是否已拉黑"`
	Reason   string          `json:"reason,omitempty" dc:"原因"`
	Client   *ClientResModel `json:"client,omitempty" dc:"客户端详情"`
}

// ClientOfflineReq 强制下线请求
type ClientOfflineReq struct {
	g.Meta `path:"/{id}/offline" method:"post" tags:"客户端" summary:"强制客户端下线"`
	ID     string `json:"id" in:"path" dc:"客户端ID"`
	Reason string `json:"reason,omitempty" dc:"下线原因"`
}

// ClientOfflineRes 强制下线响应
type ClientOfflineRes struct {
	Client  *ClientResModel `json:"client,omitempty" dc:"客户端详情"`
	Message string          `json:"message,omitempty" dc:"下线结果消息"`
}

// ClientBatchActionReq 批量客户端操作请求
type ClientBatchActionReq struct {
	IDs    []string `json:"ids" v:"required|min-length:1#客户端列表不能为空|至少选择一个客户端" dc:"客户端ID列表"`
	Reason string   `json:"reason,omitempty" dc:"操作原因"`
}

// ClientBatchOfflineReq 批量强制下线请求
type ClientBatchOfflineReq struct {
	g.Meta `path:"/batch-offline" method:"post" tags:"客户端" summary:"批量强制客户端下线"`
	ClientBatchActionReq
}

// ClientBatchActionRes 批量客户端操作响应
type ClientBatchActionRes struct {
	Total    int `json:"total" dc:"请求数量"`
	Success  int `json:"success" dc:"成功数量"`
	NotFound int `json:"not_found" dc:"未找到数量"`
	Notified int `json:"notified" dc:"通知数量"`
}

// ClientBatchOfflineRes 批量强制下线响应
type ClientBatchOfflineRes struct {
	ClientBatchActionRes
}

// ClientBanReq 拉黑客户端请求
type ClientBanReq struct {
	g.Meta `path:"/{id}/ban" method:"post" tags:"客户端" summary:"拉黑客户端"`
	ID     string `json:"id" in:"path" dc:"客户端ID"`
	Reason string `json:"reason,omitempty" dc:"拉黑原因"`
}

// ClientBanRes 拉黑客户端响应
type ClientBanRes struct {
	Client  *ClientResModel `json:"client,omitempty" dc:"客户端详情"`
	Message string          `json:"message,omitempty" dc:"拉黑结果消息"`
}

// ClientBatchBanReq 批量拉黑客户端请求
type ClientBatchBanReq struct {
	g.Meta `path:"/batch-ban" method:"post" tags:"客户端" summary:"批量拉黑客户端"`
	ClientBatchActionReq
}

// ClientBatchBanRes 批量拉黑客户端响应
type ClientBatchBanRes struct {
	ClientBatchActionRes
}

// ClientUnbanReq 解除拉黑请求
type ClientUnbanReq struct {
	g.Meta `path:"/{id}/unban" method:"post" tags:"客户端" summary:"解除客户端拉黑"`
	ID     string `json:"id" in:"path" dc:"客户端ID"`
}

// ClientUnbanRes 解除拉黑响应
type ClientUnbanRes struct {
	Client  *ClientResModel `json:"client,omitempty" dc:"客户端详情"`
	Message string          `json:"message,omitempty" dc:"解除拉黑结果消息"`
}
