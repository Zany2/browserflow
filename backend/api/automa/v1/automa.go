package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
)

// AutomaWorkflowListReq 获取 Automa 工作流列表
type AutomaWorkflowListReq struct {
	g.Meta `path:"/list" method:"get" tags:"Automa工作流" summary:"获取 Automa 工作流列表" description:"分页查询已保存到服务端的 Automa 工作流，支持按关键字、来源、客户端地址和创建时间筛选"`
	rr.CommonPageReq
	rr.CommonTimeReq
	Keyword  string `json:"keyword" in:"query" dc:"关键字，匹配工作流名称、描述或 Automa 原始工作流标识"`
	Source   int    `json:"source" in:"query" d:"0" v:"in:0,1,2#来源只能是0(全部)、1(页面导入)、2(客户端同步)" dc:"工作流来源筛选，0 表示全部，1 表示页面新增或导入，2 表示客户端同步"`
	SourceIP string `json:"source_ip" in:"query" dc:"客户端来源地址筛选"`
	Syncable int    `json:"syncable" in:"query" d:"0" v:"in:0,1,2#同步筛选只能是0(全部)、1(可同步)、2(不可同步)" dc:"客户端拉取服务端工作流时使用，1 表示 is_protected=false"`
}

// AutomaWorkflowListResModel Automa 工作流列表返回结构
type AutomaWorkflowListResModel struct {
	Id              int64       `json:"id" dc:"服务端自增主键"`
	AutomaId        string      `json:"automa_id" dc:"Automa 原始工作流标识，用于判断同一个工作流"`
	Name            string      `json:"name" dc:"服务端展示的工作流名称，默认从工作流数据中解析"`
	Description     string      `json:"description" dc:"服务端展示的工作流描述，默认从工作流数据中解析"`
	Source          string      `json:"source" dc:"工作流来源，返回中文展示值"`
	SourceIp        string      `json:"source_ip" dc:"最后同步来源客户端地址"`
	CreatedAtAutoma int64       `json:"created_at_automa" dc:"Automa 原始创建时间，通常为毫秒时间戳"`
	UpdatedAtAutoma int64       `json:"updated_at_automa" dc:"Automa 原始更新时间，通常为毫秒时间戳"`
	IsDisabled      bool        `json:"is_disabled" dc:"是否在 Automa 中被禁用"`
	IsProtected     bool        `json:"is_protected" dc:"是否不可被客户端拉取同步，false 表示可同步到客户端"`
	NodeCount       int         `json:"node_count" dc:"工作流节点数量"`
	EdgeCount       int         `json:"edge_count" dc:"工作流连线数量"`
	ContentHash     string      `json:"content_hash" dc:"核心内容哈希，用于判断工作流内容是否变化"`
	Revision        int         `json:"revision" dc:"服务端版本号，内容变化时递增"`
	CreatedAt       *gtime.Time `json:"created_at" dc:"服务端记录创建时间"`
	UpdatedAt       *gtime.Time `json:"updated_at" dc:"服务端记录更新时间"`
}

// AutomaWorkflowListRes Automa 工作流列表响应
type AutomaWorkflowListRes struct {
	List  []AutomaWorkflowListResModel `json:"list" dc:"工作流列表"`
	Total int                          `json:"total" dc:"工作流总数"`
}

// AutomaWorkflowDetailReq 获取 Automa 工作流详情
type AutomaWorkflowDetailReq struct {
	g.Meta `path:"/{id}" method:"get" tags:"Automa工作流" summary:"获取 Automa 工作流详情" description:"按服务端工作流标识获取 Automa 工作流详情，返回原始数据、规范化数据、内容哈希和同步信息"`
	ID     string `json:"id" in:"path" v:"required#ID不能为空" dc:"服务端工作流标识"`
}

// AutomaWorkflowDetailRes Automa 工作流详情响应
type AutomaWorkflowDetailRes struct {
	Id              int64       `json:"id" dc:"服务端自增主键"`
	AutomaId        string      `json:"automa_id" dc:"Automa 原始工作流标识，用于判断同一个工作流"`
	Name            string      `json:"name" dc:"服务端展示的工作流名称，默认从工作流数据中解析"`
	Description     string      `json:"description" dc:"服务端展示的工作流描述，默认从工作流数据中解析"`
	Source          string      `json:"source" dc:"工作流来源，返回中文展示值"`
	SourceIp        string      `json:"source_ip" dc:"最后同步来源客户端地址"`
	SourceUserAgent string      `json:"source_user_agent" dc:"最后同步来源浏览器用户代理"`
	AutomaVersion   string      `json:"automa_version" dc:"工作流保存时的 Automa 插件版本"`
	ExtVersion      string      `json:"ext_version" dc:"Automa 导出文件中的扩展版本"`
	CreatedAtAutoma int64       `json:"created_at_automa" dc:"Automa 原始创建时间，通常为毫秒时间戳"`
	UpdatedAtAutoma int64       `json:"updated_at_automa" dc:"Automa 原始更新时间，通常为毫秒时间戳"`
	IsDisabled      bool        `json:"is_disabled" dc:"是否在 Automa 中被禁用"`
	IsProtected     bool        `json:"is_protected" dc:"是否受保护，受保护时客户端同步不能覆盖"`
	NodeCount       int         `json:"node_count" dc:"工作流节点数量"`
	EdgeCount       int         `json:"edge_count" dc:"工作流连线数量"`
	RawJson         string      `json:"raw_json" dc:"Automa 导出的原始完整工作流数据"`
	NormalizedJson  string      `json:"normalized_json" dc:"规范化后的工作流数据，用于计算内容哈希或内容对比"`
	ContentHash     string      `json:"content_hash" dc:"核心内容哈希，用于判断工作流内容是否变化"`
	Revision        int         `json:"revision" dc:"服务端版本号，内容变化时递增"`
	FirstSyncedAt   *gtime.Time `json:"first_synced_at" dc:"第一次同步到服务端的时间"`
	LastSyncedAt    *gtime.Time `json:"last_synced_at" dc:"最近一次同步到服务端的时间"`
	CreatedAt       *gtime.Time `json:"created_at" dc:"服务端记录创建时间"`
	UpdatedAt       *gtime.Time `json:"updated_at" dc:"服务端记录更新时间"`
}

// AutomaWorkflowCreateMeta 创建 Automa 工作流元数据
type AutomaWorkflowCreateMeta struct {
	Name        string `json:"name" dc:"工作流名称，未填写时可从工作流数据中解析"`                                            // 工作流名称
	Description string `json:"description" dc:"工作流描述，未填写时可从工作流数据中解析"`                                     // 工作流描述
	Source      int    `json:"source" d:"1" v:"in:1,2#来源必须为1(导入)或2(同步)" dc:"工作流来源，1 表示页面新增或导入，2 表示客户端同步"` // 默认1，只能是1或2
	IsProtected bool   `json:"is_protected" d:"false" dc:"是否受保护，是表示客户端不能同步覆盖，否表示客户端可以同步"`                 // 默认false
}

// AutomaWorkflowCreateReq 创建 Automa 工作流
type AutomaWorkflowCreateReq struct {
	g.Meta        `path:"/add" method:"post" tags:"Automa工作流" summary:"批量创建 Automa 工作流" description:"批量上传 Automa 工作流文件，并根据每个条目的元信息创建或更新服务端工作流记录" mime:"multipart/form-data"`
	WorkflowFiles ghttp.UploadFiles `json:"workflow_files" type:"file" dc:"工作流 JSON 文件列表"` // WorkflowFiles 上传文件列表
	WorkflowMetas string            `json:"workflow_metas" dc:"工作流文件对应的 JSON 元数据列表"`       // WorkflowMetas 文件元数据 JSON
}

// AutomaWorkflowMutationStats Automa 工作流变更统计
type AutomaWorkflowMutationStats struct {
	Submitted int `json:"submitted" dc:"提交的工作流文件数量"`  // Submitted 提交数量
	Created   int `json:"created" dc:"新增的工作流数量"`      // Created 新增数量
	Updated   int `json:"updated" dc:"更新的工作流数量"`      // Updated 更新数量
	Unchanged int `json:"unchanged" dc:"重复且内容未变化的数量"` // Unchanged 重复未变化数量
}

// AutomaWorkflowCreateRes Automa 工作流新增响应
type AutomaWorkflowCreateRes struct {
	AutomaWorkflowMutationStats
}

// AutomaWorkflowUpdateReq 修改 Automa 工作流
type AutomaWorkflowUpdateReq struct {
	g.Meta      `path:"/update" method:"put" tags:"Automa工作流" summary:"修改 Automa 工作流" description:"按服务端工作流标识修改 Automa 工作流记录"`
	ID          int64  `json:"id" v:"required#id不能为空" dc:"服务端工作流标识"`                                                         // ID 后端自增主键
	Name        string `json:"name" dc:"工作流名称，未填写时可从工作流数据中解析"`                                                               // Name 工作流名称
	Description string `json:"description" dc:"工作流描述，未填写时可从工作流数据中解析"`                                                        // Description 工作流描述
	Source      int    `json:"source" d:"0" v:"in:0,1,2#来源只能是0(保持不变)、1(导入)或2(同步)" dc:"工作流来源，0 表示保持不变，1 表示页面新增或导入，2 表示客户端同步"` // Source 工作流来源
	IsProtected bool   `json:"is_protected" d:"false" dc:"是否受保护，是表示客户端不能同步覆盖，否表示客户端可以同步"`                                    // IsProtected 是否受保护
	Revision    int    `json:"revision" v:"required#版本号不能为空" dc:"工作流版本号，用于防止覆盖较新的数据"`                                        // Revision 乐观锁版本号
}

// AutomaWorkflowUpdateRes Automa 工作流修改响应
type AutomaWorkflowUpdateRes struct{}

// AutomaWorkflowProtectedReq 修改 Automa 工作流是否受保护
type AutomaWorkflowProtectedReq struct {
	g.Meta      `path:"/syncable" method:"put" tags:"Automa工作流" summary:"修改 Automa 工作流是否受保护" description:"按服务端工作流标识修改是否保护服务端工作流，受保护时客户端不能同步覆盖"`
	ID          int64 `json:"id" v:"required#id不能为空" dc:"服务端工作流标识"`                    // ID 后端自增主键
	IsProtected bool  `json:"is_protected" dc:"是否受保护，true 表示受保护不可覆盖，false 表示客户端可同步覆盖"` // IsProtected 是否受保护
	Revision    int   `json:"revision" v:"required#版本号不能为空" dc:"工作流版本号，用于防止覆盖较新的数据"`   // Revision 乐观锁版本号
}

// AutomaWorkflowProtectedRes Automa 工作流是否受保护响应
type AutomaWorkflowProtectedRes struct{}

// AutomaWorkflowBatchDeleteReq 批量删除 Automa 工作流
type AutomaWorkflowBatchDeleteReq struct {
	g.Meta `path:"/delete" method:"delete" tags:"Automa工作流" summary:"批量删除 Automa 工作流" description:"一次请求批量删除多个 Automa 工作流"`
	IDs    []string `json:"ids" v:"required|min-length:1#工作流id列表不能为空|至少选择一个工作流" dc:"工作流标识列表"` // IDs 工作流 ID 列表
}

// AutomaWorkflowBatchDeleteRes 批量删除 Automa 工作流响应
type AutomaWorkflowBatchDeleteRes struct{}

// AutomaWorkflowImportFilesReq 导入 Automa 工作流压缩文件
type AutomaWorkflowImportFilesReq struct {
	g.Meta `path:"/import" method:"post" tags:"Automa工作流" summary:"导入 Automa 工作流压缩包" description:"上传一个压缩文件，压缩包内必须全部为 Automa 工作流文件" mime:"multipart/form-data"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required#ZIP压缩文件不能为空" dc:"Automa 工作流压缩文件"` // File 工作流 ZIP 压缩文件
}

// AutomaWorkflowImportFilesRes Automa 工作流文件导入响应
type AutomaWorkflowImportFilesRes struct {
	AutomaWorkflowMutationStats
}

// AutomaWorkflowSyncCandidatesReq 获取客户端可同步工作流
type AutomaWorkflowSyncCandidatesReq struct {
	g.Meta `path:"/sync-client-workflow" method:"get" tags:"Automa工作流" summary:"获取客户端可同步工作流" description:"获取某个客户端可以同步到服务端的 Automa 工作流列表"`
	rr.CommonPageReq
	Mode     string `json:"mode" in:"query" d:"client" v:"in:client,workflow#查询维度只能是client或workflow" dc:"查询维度，client 表示按客户端 IP，workflow 表示按工作流"`
	SourceIP string `json:"source_ip" in:"query" dc:"客户端来源地址"`
	AutomaID string `json:"automa_id" in:"query" dc:"Automa 工作流 ID"`
	Keyword  string `json:"keyword,omitempty" in:"query" dc:"关键字，匹配工作流相关字段"`
}

// AutomaWorkflowSyncCandidatesResModel 客户端可同步工作流
type AutomaWorkflowSyncCandidatesResModel struct {
	Id              string      `json:"id" dc:"客户端工作流标识，默认等于 Automa 原始工作流 ID"`
	AutomaId        string      `json:"automa_id" dc:"Automa 原始工作流 ID"`
	WorkflowId      string      `json:"workflow_id" dc:"兼容前端选择逻辑的工作流 ID"`
	Name            string      `json:"name" dc:"工作流名称"`
	Description     string      `json:"description" dc:"工作流描述"`
	Source          string      `json:"source" dc:"工作流来源展示值"`
	SourceIp        string      `json:"source_ip" dc:"客户端来源地址"`
	AutomaVersion   string      `json:"automa_version" dc:"客户端 Automa 工作流版本"`
	ExtVersion      string      `json:"ext_version" dc:"Automa 导出文件中的 extVersion"`
	CreatedAtAutoma int64       `json:"created_at_automa" dc:"Automa 原始创建时间"`
	UpdatedAtAutoma int64       `json:"updated_at_automa" dc:"Automa 原始更新时间"`
	IsDisabled      bool        `json:"is_disabled" dc:"是否在 Automa 中被禁用"`
	IsProtected     bool        `json:"is_protected" dc:"服务端工作流是否不可被客户端拉取同步"`
	NodeCount       int         `json:"node_count" dc:"工作流节点数量"`
	EdgeCount       int         `json:"edge_count" dc:"工作流连线数量"`
	ContentHash     string      `json:"content_hash" dc:"客户端工作流内容哈希"`
	Synced          bool        `json:"synced" dc:"是否已同步到服务端且内容一致"`
	HasUpdate       bool        `json:"has_update" dc:"客户端是否有可同步更新"`
	SyncStatus      string      `json:"sync_status" dc:"同步状态：not_synced、synced、has_update、client_newer、server_newer、protected"`
	ServerId        int64       `json:"server_id" dc:"服务端已有工作流主键"`
	ServerName      string      `json:"server_name" dc:"服务端工作流名称"`
	ServerDesc      string      `json:"server_description" dc:"服务端工作流描述"`
	ServerRevision  int         `json:"server_revision" dc:"服务端已有工作流版本号"`
	LastSyncedAt    *gtime.Time `json:"last_synced_at" dc:"服务端最近同步时间"`
	ServerUpdatedAt *gtime.Time `json:"server_updated_at" dc:"服务端记录更新时间"`
	Online          bool        `json:"online" dc:"客户端是否在线"`
}

// AutomaWorkflowSyncCandidatesRes 客户端可同步工作流响应
type AutomaWorkflowSyncCandidatesRes struct {
	List  []AutomaWorkflowSyncCandidatesResModel `json:"list" dc:"可同步工作流列表"`
	Total int                                    `json:"total" dc:"可同步工作流数量"`
}

// AutomaWorkflowSyncReq 从客户端同步 Automa 工作流
type AutomaWorkflowSyncReq struct {
	g.Meta               `path:"/sync" method:"post" tags:"Automa工作流" summary:"同步 Automa 工作流" description:"从客户端同步 Automa 工作流到服务端"`
	SourceIP             string          `json:"source_ip" v:"required#客户端ip不能为空" dc:"客户端来源地址"`
	WorkflowIds          []string        `json:"workflow_ids" dc:"需要同步的客户端工作流 ID 列表，为空时同步 workflows 中全部工作流"`
	WorkflowJsonDataList []model.JSONMap `json:"workflows" dc:"客户端工作流 JSON 数据列表"`
}

// AutomaWorkflowSyncRes 从客户端同步 Automa 工作流响应
type AutomaWorkflowSyncRes struct{}

// AutomaWorkflowCacheReq 获取本地工作流快照
type AutomaWorkflowCacheReq struct {
	g.Meta `path:"/cache" method:"get" tags:"Automa工作流" summary:"获取本地工作流快照"`
}

// AutomaWorkflowCacheRes 工作流快照响应
type AutomaWorkflowCacheRes struct {
	Snapshot *model.AutomaWorkflowSnapshot `json:"snapshot,omitempty" dc:"本地工作流快照"`
}

// AutomaWorkflowRunReq 通过浏览器执行端运行工作流
type AutomaWorkflowRunReq struct {
	g.Meta    `path:"/{id}/run" method:"post" tags:"Automa工作流" summary:"运行 Automa 工作流"`
	ID        string        `json:"id" in:"path"`
	BrowserID string        `json:"browser_id"`
	Variables model.JSONMap `json:"variables"`
}

// AutomaWorkflowRunRes 工作流运行响应
type AutomaWorkflowRunRes struct {
	Result *model.AgentCommandResult `json:"result,omitempty" dc:"运行结果"`
}

// AutomaWorkflowOpenReq 通过浏览器执行端打开工作流
type AutomaWorkflowOpenReq struct {
	g.Meta    `path:"/{id}/open" method:"post" tags:"Automa工作流" summary:"打开 Automa 工作流"`
	ID        string `json:"id" in:"path"`
	BrowserID string `json:"browser_id"`
}

// AutomaWorkflowOpenRes 工作流打开响应
type AutomaWorkflowOpenRes struct {
	Result *model.AgentCommandResult `json:"result,omitempty" dc:"打开结果"`
}
