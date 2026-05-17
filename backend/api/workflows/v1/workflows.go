package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
)

// WorkflowListReq 获取工作流列表
type WorkflowListReq struct {
	g.Meta `path:"/" method:"get" tags:"工作流" summary:"获取工作流列表"`
	rr.CommonPageReq
	rr.CommonTimeReq
	Keyword  string `json:"keyword" in:"query" dc:"关键字"`
	Source   int    `json:"source" in:"query" d:"0" v:"in:0,1,2#来源只能是0、1、2" dc:"工作流来源"`
	SourceIP string `json:"source_ip" in:"query" dc:"客户端来源地址"`
	Syncable int    `json:"syncable" in:"query" d:"0" v:"in:0,1,2#同步筛选只能是0、1、2" dc:"同步筛选"`
}

// WorkflowListResModel 工作流列表项
type WorkflowListResModel struct {
	Id              int64       `json:"id" dc:"服务端主键"`
	AutomaId        string      `json:"automa_id" dc:"Automa 原始工作流标识"`
	Name            string      `json:"name" dc:"工作流名称"`
	Description     string      `json:"description" dc:"工作流描述"`
	Source          string      `json:"source" dc:"工作流来源"`
	SourceIp        string      `json:"source_ip" dc:"来源客户端地址"`
	CreatedAtAutoma int64       `json:"created_at_automa" dc:"Automa 原始创建时间"`
	UpdatedAtAutoma int64       `json:"updated_at_automa" dc:"Automa 原始更新时间"`
	IsDisabled      bool        `json:"is_disabled" dc:"是否禁用"`
	IsProtected     bool        `json:"is_protected" dc:"是否受保护"`
	NodeCount       int         `json:"node_count" dc:"节点数量"`
	EdgeCount       int         `json:"edge_count" dc:"连线数量"`
	ContentHash     string      `json:"content_hash" dc:"内容哈希"`
	Revision        int         `json:"revision" dc:"版本号"`
	CreatedAt       *gtime.Time `json:"created_at" dc:"服务端创建时间"`
	UpdatedAt       *gtime.Time `json:"updated_at" dc:"服务端更新时间"`
}

// WorkflowListRes 工作流列表响应
type WorkflowListRes struct {
	List  []WorkflowListResModel `json:"list" dc:"工作流列表"`
	Total int                    `json:"total" dc:"工作流总数"`
}

// WorkflowDetailReq 获取工作流详情
type WorkflowDetailReq struct {
	g.Meta `path:"/{id}" method:"get" tags:"工作流" summary:"获取工作流详情"`
	ID     string `json:"id" in:"path" v:"required#ID不能为空" dc:"工作流标识"`
}

// WorkflowDetailRes 工作流详情响应
type WorkflowDetailRes struct {
	Id              int64       `json:"id" dc:"服务端主键"`
	AutomaId        string      `json:"automa_id" dc:"Automa 原始工作流标识"`
	Name            string      `json:"name" dc:"工作流名称"`
	Description     string      `json:"description" dc:"工作流描述"`
	Source          string      `json:"source" dc:"工作流来源"`
	SourceIp        string      `json:"source_ip" dc:"来源客户端地址"`
	SourceUserAgent string      `json:"source_user_agent" dc:"来源用户代理"`
	AutomaVersion   string      `json:"automa_version" dc:"Automa 版本"`
	ExtVersion      string      `json:"ext_version" dc:"扩展版本"`
	CreatedAtAutoma int64       `json:"created_at_automa" dc:"Automa 原始创建时间"`
	UpdatedAtAutoma int64       `json:"updated_at_automa" dc:"Automa 原始更新时间"`
	IsDisabled      bool        `json:"is_disabled" dc:"是否禁用"`
	IsProtected     bool        `json:"is_protected" dc:"是否受保护"`
	NodeCount       int         `json:"node_count" dc:"节点数量"`
	EdgeCount       int         `json:"edge_count" dc:"连线数量"`
	RawJson         string      `json:"raw_json" dc:"原始 JSON"`
	NormalizedJson  string      `json:"normalized_json" dc:"规范化 JSON"`
	ContentHash     string      `json:"content_hash" dc:"内容哈希"`
	Revision        int         `json:"revision" dc:"版本号"`
	FirstSyncedAt   *gtime.Time `json:"first_synced_at" dc:"首次同步时间"`
	LastSyncedAt    *gtime.Time `json:"last_synced_at" dc:"最近同步时间"`
	CreatedAt       *gtime.Time `json:"created_at" dc:"服务端创建时间"`
	UpdatedAt       *gtime.Time `json:"updated_at" dc:"服务端更新时间"`
}

// WorkflowCreateMeta 创建工作流元数据
type WorkflowCreateMeta struct {
	Name        string `json:"name" dc:"工作流名称"`
	Description string `json:"description" dc:"工作流描述"`
	Source      int    `json:"source" d:"1" v:"in:1,2#来源必须是1或2" dc:"工作流来源"`
	IsProtected bool   `json:"is_protected" d:"false" dc:"是否受保护"`
}

// WorkflowCreateReq 创建工作流
type WorkflowCreateReq struct {
	g.Meta        `path:"/" method:"post" tags:"工作流" summary:"创建工作流" mime:"multipart/form-data"`
	WorkflowFiles ghttp.UploadFiles `json:"workflow_files" type:"file" dc:"工作流 JSON 文件列表"`
	WorkflowMetas string            `json:"workflow_metas" dc:"工作流元数据 JSON"`
}

// WorkflowMutationStats 工作流变更统计
type WorkflowMutationStats struct {
	Submitted int `json:"submitted" dc:"提交数量"`
	Created   int `json:"created" dc:"新增数量"`
	Updated   int `json:"updated" dc:"更新数量"`
	Unchanged int `json:"unchanged" dc:"未变化数量"`
}

// WorkflowCreateRes 工作流创建响应
type WorkflowCreateRes struct {
	WorkflowMutationStats
}

// WorkflowUpdateReq 更新工作流
type WorkflowUpdateReq struct {
	g.Meta      `path:"/{id}" method:"put" tags:"工作流" summary:"更新工作流"`
	ID          int64  `json:"id" in:"path" v:"required#id不能为空" dc:"工作流标识"`
	Name        string `json:"name" dc:"工作流名称"`
	Description string `json:"description" dc:"工作流描述"`
	Source      int    `json:"source" d:"0" v:"in:0,1,2#来源只能是0、1、2" dc:"工作流来源"`
	IsProtected bool   `json:"is_protected" d:"false" dc:"是否受保护"`
	Revision    int    `json:"revision" v:"required#版本号不能为空" dc:"版本号"`
}

// WorkflowUpdateRes 工作流更新响应
type WorkflowUpdateRes struct{}

// WorkflowProtectedReq 更新工作流保护状态
type WorkflowProtectedReq struct {
	g.Meta      `path:"/{id}/syncable" method:"put" tags:"工作流" summary:"更新工作流保护状态"`
	ID          int64 `json:"id" in:"path" v:"required#id不能为空" dc:"工作流标识"`
	IsProtected bool  `json:"is_protected" dc:"是否受保护"`
	Revision    int   `json:"revision" v:"required#版本号不能为空" dc:"版本号"`
}

// WorkflowProtectedRes 工作流保护响应
type WorkflowProtectedRes struct{}

// WorkflowBatchDeleteReq 批量删除工作流
type WorkflowBatchDeleteReq struct {
	g.Meta `path:"/" method:"delete" tags:"工作流" summary:"批量删除工作流"`
	IDs    []string `json:"ids" v:"required|min-length:1#工作流id列表不能为空|至少选择一个工作流" dc:"工作流ID列表"`
}

// WorkflowBatchDeleteRes 批量删除响应
type WorkflowBatchDeleteRes struct{}

// WorkflowImportFilesReq 导入工作流压缩包
type WorkflowImportFilesReq struct {
	g.Meta `path:"/import" method:"post" tags:"工作流" summary:"导入工作流压缩包" mime:"multipart/form-data"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required#ZIP压缩文件不能为空" dc:"工作流 ZIP 文件"`
}

// WorkflowImportFilesRes 工作流文件导入响应
type WorkflowImportFilesRes struct {
	WorkflowMutationStats
}

// WorkflowSyncCandidatesReq 获取可同步工作流
type WorkflowSyncCandidatesReq struct {
	g.Meta `path:"/sync-candidates" method:"get" tags:"工作流" summary:"获取可同步工作流"`
	rr.CommonPageReq
	Mode     string `json:"mode" in:"query" d:"client" v:"in:client,workflow#查询维度只能是client或workflow" dc:"查询维度"`
	SourceIP string `json:"source_ip" in:"query" dc:"客户端来源地址"`
	AutomaID string `json:"automa_id" in:"query" dc:"Automa 工作流 ID"`
	Keyword  string `json:"keyword,omitempty" in:"query" dc:"关键字"`
}

// WorkflowSyncCandidatesResModel 同步候选列表项
type WorkflowSyncCandidatesResModel struct {
	Id              string      `json:"id" dc:"客户端工作流标识"`
	AutomaId        string      `json:"automa_id" dc:"Automa 工作流 ID"`
	WorkflowId      string      `json:"workflow_id" dc:"工作流 ID"`
	Name            string      `json:"name" dc:"工作流名称"`
	Description     string      `json:"description" dc:"工作流描述"`
	Source          string      `json:"source" dc:"工作流来源"`
	SourceIp        string      `json:"source_ip" dc:"来源地址"`
	AutomaVersion   string      `json:"automa_version" dc:"Automa 版本"`
	ExtVersion      string      `json:"ext_version" dc:"扩展版本"`
	CreatedAtAutoma int64       `json:"created_at_automa" dc:"Automa 创建时间"`
	UpdatedAtAutoma int64       `json:"updated_at_automa" dc:"Automa 更新时间"`
	IsDisabled      bool        `json:"is_disabled" dc:"是否禁用"`
	IsProtected     bool        `json:"is_protected" dc:"是否受保护"`
	NodeCount       int         `json:"node_count" dc:"节点数量"`
	EdgeCount       int         `json:"edge_count" dc:"连线数量"`
	ContentHash     string      `json:"content_hash" dc:"内容哈希"`
	Synced          bool        `json:"synced" dc:"是否已同步"`
	HasUpdate       bool        `json:"has_update" dc:"是否有更新"`
	SyncStatus      string      `json:"sync_status" dc:"同步状态"`
	ServerId        int64       `json:"server_id" dc:"服务端主键"`
	ServerName      string      `json:"server_name" dc:"服务端名称"`
	ServerDesc      string      `json:"server_description" dc:"服务端描述"`
	ServerRevision  int         `json:"server_revision" dc:"服务端版本"`
	LastSyncedAt    *gtime.Time `json:"last_synced_at" dc:"最近同步时间"`
	ServerUpdatedAt *gtime.Time `json:"server_updated_at" dc:"服务端更新时间"`
	Online          bool        `json:"online" dc:"是否在线"`
}

// WorkflowSyncCandidatesRes 同步候选响应
type WorkflowSyncCandidatesRes struct {
	List  []WorkflowSyncCandidatesResModel `json:"list" dc:"可同步工作流列表"`
	Total int                              `json:"total" dc:"可同步工作流数量"`
}

// WorkflowSyncReq 从客户端同步工作流
type WorkflowSyncReq struct {
	g.Meta               `path:"/sync" method:"post" tags:"工作流" summary:"同步工作流"`
	SourceIP             string          `json:"source_ip" v:"required#客户端ip不能为空" dc:"客户端来源地址"`
	WorkflowIds          []string        `json:"workflow_ids" dc:"需要同步的工作流 ID 列表"`
	WorkflowJsonDataList []model.JSONMap `json:"workflows" dc:"工作流 JSON 数据列表"`
}

// WorkflowSyncRes 同步响应
type WorkflowSyncRes struct{}

// WorkflowCacheReq 获取本地工作流快照
type WorkflowCacheReq struct {
	g.Meta `path:"/cache" method:"get" tags:"工作流" summary:"获取本地工作流快照"`
}

// WorkflowCacheRes 工作流快照响应
type WorkflowCacheRes struct {
	Snapshot *model.AutomaWorkflowSnapshot `json:"snapshot,omitempty" dc:"本地工作流快照"`
}

// WorkflowAgentListReq reads workflows from browser agent 从浏览器执行端读取工作流
type WorkflowAgentListReq struct {
	g.Meta    `path:"/agent/workflows" method:"get" tags:"工作流" summary:"从浏览器执行端读取工作流"`
	BrowserID string `json:"browser_id" in:"query" dc:"浏览器实例ID"`
}

// WorkflowAgentListRes browser agent workflow list 浏览器执行端工作流列表
type WorkflowAgentListRes struct {
	BrowserID string           `json:"browser_id" dc:"浏览器实例ID"`
	Workflows []map[string]any `json:"workflows" dc:"工作流列表"`
	Total     int              `json:"total" dc:"工作流总数"`
}

// WorkflowAgentExportSkillReq exports browser agent workflows as Skill 导出浏览器执行端工作流为 Skill
type WorkflowAgentExportSkillReq struct {
	g.Meta      `path:"/agent/export/skill" method:"post" tags:"工作流" summary:"导出浏览器执行端工作流为 Skill"`
	BrowserID   string   `json:"browser_id" dc:"浏览器实例ID"`
	Scope       string   `json:"scope" d:"filtered" v:"in:selected,filtered,all#导出范围只能是selected、filtered或all" dc:"导出范围"`
	WorkflowIDs []string `json:"workflow_ids" dc:"需要导出的工作流ID列表"`
}

// WorkflowAgentExportSkillRes skill export response Skill 导出响应
type WorkflowAgentExportSkillRes struct{}

// WorkflowRunReq 运行工作流
type WorkflowRunReq struct {
	g.Meta     `path:"/{id}/run" method:"post" tags:"工作流" summary:"运行工作流"`
	ID         string                             `json:"id" in:"path" v:"required#工作流ID不能为空" dc:"工作流ID"`
	BrowserID  string                             `json:"browser_id" dc:"浏览器实例ID"`
	Variables  model.JSONMap                      `json:"variables" dc:"运行变量"`
	WaitResult bool                               `json:"wait_result" d:"false" dc:"是否等待工作流完成"`
	Timeout    int                                `json:"timeout" d:"300" dc:"等待超时秒数"`
	ReturnData *model.WorkflowExecutionReturnData `json:"return_data" dc:"回传数据配置"`
}

// WorkflowRunRes 工作流运行响应
type WorkflowRunRes struct {
	Result    *model.AgentCommandResult `json:"result,omitempty" dc:"运行结果"`
	Execution *model.WorkflowExecution  `json:"execution,omitempty" dc:"执行状态"`
}

// WorkflowExecutionDetailReq 获取工作流执行状态
type WorkflowExecutionDetailReq struct {
	g.Meta      `path:"/executions/{execution_id}" method:"get" tags:"工作流" summary:"获取工作流执行状态"`
	ExecutionID string `json:"execution_id" in:"path" v:"required#执行ID不能为空" dc:"执行ID"`
}

// WorkflowExecutionDetailRes 工作流执行状态响应
type WorkflowExecutionDetailRes struct {
	Execution *model.WorkflowExecution `json:"execution,omitempty" dc:"执行状态"`
}

// WorkflowOpenReq 打开工作流
type WorkflowOpenReq struct {
	g.Meta    `path:"/{id}/open" method:"post" tags:"工作流" summary:"打开工作流"`
	ID        string `json:"id" in:"path" v:"required#工作流ID不能为空" dc:"工作流ID"`
	BrowserID string `json:"browser_id" dc:"浏览器实例ID"`
}

// WorkflowOpenRes 工作流打开响应
type WorkflowOpenRes struct {
	Result *model.AgentCommandResult `json:"result,omitempty" dc:"打开结果"`
}
