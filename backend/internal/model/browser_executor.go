package model

import "time"

// BrowserExecutorStatus describes executor availability. 描述执行器可用状态。
type BrowserExecutorStatus struct {
	Running         bool   `json:"running"`
	BrowserID       string `json:"browser_id,omitempty"`
	BrowserName     string `json:"browser_name,omitempty"`
	URL             string `json:"url,omitempty"`
	Title           string `json:"title,omitempty"`
	HasBusinessPage bool   `json:"has_business_page"`
	Message         string `json:"message,omitempty"`
}

// BrowserExecutorOperationResult is the unified operation response. 统一操作响应。
type BrowserExecutorOperationResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// BrowserExecutorSnapshotResult is the accessibility snapshot response. 可访问性快照响应。
type BrowserExecutorSnapshotResult struct {
	Success      bool                        `json:"success"`
	SnapshotText string                      `json:"snapshot_text,omitempty"`
	Elements     []BrowserExecutorElementRef `json:"elements,omitempty"`
	Data         map[string]interface{}      `json:"data,omitempty"`
	Error        string                      `json:"error,omitempty"`
	Timestamp    time.Time                   `json:"timestamp"`
}

// BrowserExecutorElementRef describes an element exposed to the model. 暴露给模型的元素引用。
type BrowserExecutorElementRef struct {
	RefID       string            `json:"ref_id"`
	Role        string            `json:"role"`
	Name        string            `json:"name,omitempty"`
	Value       string            `json:"value,omitempty"`
	Placeholder string            `json:"placeholder,omitempty"`
	BackendID   int               `json:"backend_id,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// BrowserExecutorRefData stores a stable locator for a RefID. 保存 RefID 的稳定定位信息。
type BrowserExecutorRefData struct {
	Role        string
	Name        string
	Nth         int
	BackendID   int
	Tag         string
	Href        string
	Attributes  map[string]string
	Placeholder string
}

// BrowserExecutorHelpCommand documents an executor command. 描述执行器命令。
type BrowserExecutorHelpCommand struct {
	Name        string   `json:"name"`
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters,omitempty"`
	Example     any      `json:"example,omitempty"`
}

// BrowserExecutorBatchAction is an internal batch action. 内部批量动作。
type BrowserExecutorBatchAction struct {
	Type        string
	Params      map[string]any
	StopOnError bool
}

// BrowserExecutorFormField describes one form field to fill. 描述一个待填写表单字段。
type BrowserExecutorFormField struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
	Type  string `json:"type,omitempty"`
}

// BrowserExecutorActOptions describes one smart browser action. 描述一次智能浏览器动作。
type BrowserExecutorActOptions struct {
	Intent     string
	Identifier string
	Value      any
	Text       string
	Fields     []BrowserExecutorFormField
	Submit     bool
	Clear      bool
	Timeout    int
}

// BrowserExecutorPageStructureOptions describes compact page structure extraction. Page structure options.
type BrowserExecutorPageStructureOptions struct {
	IncludeLinks   bool
	IncludeForms   bool
	IncludeTables  bool
	IncludeImages  bool
	IncludeButtons bool
	Limit          int
}

// BrowserExecutorMouseOptions describes a coordinate mouse operation. Mouse options.
type BrowserExecutorMouseOptions struct {
	Action string
	X      float64
	Y      float64
	DeltaX float64
	DeltaY float64
	Steps  int
	Button string
}

// BrowserExecutorWindowOptions describes browser window operation options. Window options.
type BrowserExecutorWindowOptions struct {
	Action string
	Left   int
	Top    int
	Width  int
	Height int
}

// BrowserExecutorObserveResult is compact page context for LLM use. 给大模型使用的紧凑页面上下文。
type BrowserExecutorObserveResult struct {
	Success      bool                        `json:"success"`
	Status       BrowserExecutorStatus       `json:"status"`
	PageInfo     map[string]interface{}      `json:"page_info,omitempty"`
	SnapshotText string                      `json:"snapshot_text,omitempty"`
	Elements     []BrowserExecutorElementRef `json:"elements,omitempty"`
	PageText     string                      `json:"page_text,omitempty"`
	Error        string                      `json:"error,omitempty"`
	Timestamp    time.Time                   `json:"timestamp"`
}

// BrowserExecutorAccessibilitySnapshot contains AX nodes and RefID data. 保存 AX 节点和 RefID 数据。
type BrowserExecutorAccessibilitySnapshot struct {
	Elements map[string]*BrowserExecutorAccessibilityNode
	Refs     []BrowserExecutorElementRef
}

// BrowserExecutorAccessibilityNode is a compact AX node. 精简后的 AX 节点。
type BrowserExecutorAccessibilityNode struct {
	ID            string
	BackendNodeID int
	RefID         string
	Role          string
	Name          string
	Value         string
	Placeholder   string
	Attributes    map[string]string
}
