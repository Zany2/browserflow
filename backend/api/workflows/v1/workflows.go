package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
)

type WorkflowListReq struct {
	g.Meta `path:"/" method:"get" tags:"宸ヤ綔娴? summary:"鑾峰彇宸ヤ綔娴佸垪琛?`
	rr.CommonPageReq
	rr.CommonTimeReq
	Keyword       string `json:"keyword" in:"query" dc:"鍏抽敭瀛?`
	CustomKeyword string `json:"custom_keyword" in:"query" dc:"鑷畾涔夊伐浣滄祦鍚嶇О銆佹弿杩板叧閿瓧"`
	Source        int    `json:"source" in:"query" d:"0" v:"in:0,1,2#鏉ユ簮鍙兘鏄?銆?銆?" dc:"宸ヤ綔娴佹潵婧?`
	SourceIP      string `json:"source_ip" in:"query" dc:"瀹㈡埛绔潵婧愬湴鍧€"`
	SourceNodeID  string `json:"source_node_id" in:"query" dc:"鏉ユ簮鎵ц鑺傜偣 ID"`
	Syncable      int    `json:"syncable" in:"query" d:"0" v:"in:0,1,2#鍚屾绛涢€夊彧鑳芥槸0銆?銆?" dc:"鍚屾绛涢€?`
}

type WorkflowListResModel struct {
	Id                int64       `json:"id" dc:"鏈嶅姟绔富閿?`
	AutomaId          string      `json:"automa_id" dc:"Automa 鍘熷宸ヤ綔娴佹爣璇?`
	Name              string      `json:"name" dc:"宸ヤ綔娴佸悕绉?`
	Description       string      `json:"description" dc:"宸ヤ綔娴佹弿杩?`
	AutomaName        string      `json:"automa_name" dc:"Automa 宸ヤ綔娴佸師濮嬪悕绉?`
	AutomaDescription string      `json:"automa_description" dc:"Automa 宸ヤ綔娴佸師濮嬫弿杩?`
	Source            string      `json:"source" dc:"宸ヤ綔娴佹潵婧?`
	SourceIp          string      `json:"source_ip" dc:"鏉ユ簮瀹㈡埛绔湴鍧€"`
	SourceNodeId      string      `json:"source_node_id" dc:"鏉ユ簮鎵ц鑺傜偣 ID"`
	CreatedAtAutoma   int64       `json:"created_at_automa" dc:"Automa 鍘熷鍒涘缓鏃堕棿"`
	UpdatedAtAutoma   int64       `json:"updated_at_automa" dc:"Automa 鍘熷鏇存柊鏃堕棿"`
	IsDisabled        bool        `json:"is_disabled" dc:"鏄惁绂佺敤"`
	IsProtected       bool        `json:"is_protected" dc:"鏄惁鍙椾繚鎶?`
	NodeCount         int         `json:"node_count" dc:"鑺傜偣鏁伴噺"`
	EdgeCount         int         `json:"edge_count" dc:"杩炵嚎鏁伴噺"`
	ContentHash       string      `json:"content_hash" dc:"鍐呭鍝堝笇"`
	Revision          int         `json:"revision" dc:"鐗堟湰鍙?`
	CreatedAt         *gtime.Time `json:"created_at" dc:"鏈嶅姟绔垱寤烘椂闂?`
	UpdatedAt         *gtime.Time `json:"updated_at" dc:"鏈嶅姟绔洿鏂版椂闂?`
}

type WorkflowListRes struct {
	List  []WorkflowListResModel `json:"list" dc:"宸ヤ綔娴佸垪琛?`
	Total int                    `json:"total" dc:"宸ヤ綔娴佹€绘暟"`
}

type WorkflowDetailReq struct {
	g.Meta `path:"/{id}" method:"get" tags:"宸ヤ綔娴? summary:"鑾峰彇宸ヤ綔娴佽鎯?`
	ID     string `json:"id" in:"path" v:"required#ID涓嶈兘涓虹┖" dc:"宸ヤ綔娴佹爣璇?`
}

type WorkflowDetailRes struct {
	Id                int64       `json:"id" dc:"鏈嶅姟绔富閿?`
	AutomaId          string      `json:"automa_id" dc:"Automa 鍘熷宸ヤ綔娴佹爣璇?`
	Name              string      `json:"name" dc:"宸ヤ綔娴佸悕绉?`
	Description       string      `json:"description" dc:"宸ヤ綔娴佹弿杩?`
	AutomaName        string      `json:"automa_name" dc:"Automa 宸ヤ綔娴佸師濮嬪悕绉?`
	AutomaDescription string      `json:"automa_description" dc:"Automa 宸ヤ綔娴佸師濮嬫弿杩?`
	Source            string      `json:"source" dc:"宸ヤ綔娴佹潵婧?`
	SourceIp          string      `json:"source_ip" dc:"鏉ユ簮瀹㈡埛绔湴鍧€"`
	SourceNodeId      string      `json:"source_node_id" dc:"鏉ユ簮鎵ц鑺傜偣 ID"`
	SourceUserAgent   string      `json:"source_user_agent" dc:"鏉ユ簮鐢ㄦ埛浠ｇ悊"`
	AutomaVersion     string      `json:"automa_version" dc:"Automa 鐗堟湰"`
	ExtVersion        string      `json:"ext_version" dc:"鎵╁睍鐗堟湰"`
	CreatedAtAutoma   int64       `json:"created_at_automa" dc:"Automa 鍘熷鍒涘缓鏃堕棿"`
	UpdatedAtAutoma   int64       `json:"updated_at_automa" dc:"Automa 鍘熷鏇存柊鏃堕棿"`
	IsDisabled        bool        `json:"is_disabled" dc:"鏄惁绂佺敤"`
	IsProtected       bool        `json:"is_protected" dc:"鏄惁鍙椾繚鎶?`
	NodeCount         int         `json:"node_count" dc:"鑺傜偣鏁伴噺"`
	EdgeCount         int         `json:"edge_count" dc:"杩炵嚎鏁伴噺"`
	RawJson           string      `json:"raw_json" dc:"鍘熷 JSON"`
	NormalizedJson    string      `json:"normalized_json" dc:"瑙勮寖鍖?JSON"`
	ContentHash       string      `json:"content_hash" dc:"鍐呭鍝堝笇"`
	Revision          int         `json:"revision" dc:"鐗堟湰鍙?`
	FirstSyncedAt     *gtime.Time `json:"first_synced_at" dc:"棣栨鍚屾鏃堕棿"`
	LastSyncedAt      *gtime.Time `json:"last_synced_at" dc:"鏈€杩戝悓姝ユ椂闂?`
	CreatedAt         *gtime.Time `json:"created_at" dc:"鏈嶅姟绔垱寤烘椂闂?`
	UpdatedAt         *gtime.Time `json:"updated_at" dc:"鏈嶅姟绔洿鏂版椂闂?`
}

type WorkflowCreateMeta struct {
	Name        string `json:"name" dc:"宸ヤ綔娴佸悕绉?`
	Description string `json:"description" dc:"宸ヤ綔娴佹弿杩?`
	Source      int    `json:"source" d:"1" v:"in:1,2#鏉ユ簮蹇呴』鏄?鎴?" dc:"宸ヤ綔娴佹潵婧?`
	IsProtected bool   `json:"is_protected" d:"false" dc:"鏄惁鍙椾繚鎶?`
}

type WorkflowCreateReq struct {
	g.Meta        `path:"/" method:"post" tags:"宸ヤ綔娴? summary:"鍒涘缓宸ヤ綔娴? mime:"multipart/form-data"`
	WorkflowFiles ghttp.UploadFiles `json:"workflow_files" type:"file" dc:"宸ヤ綔娴?JSON 鏂囦欢鍒楄〃"`
	WorkflowMetas string            `json:"workflow_metas" dc:"宸ヤ綔娴佸厓鏁版嵁 JSON"`
}

type WorkflowMutationStats struct {
	Submitted int `json:"submitted" dc:"鎻愪氦鏁伴噺"`
	Created   int `json:"created" dc:"鏂板鏁伴噺"`
	Updated   int `json:"updated" dc:"鏇存柊鏁伴噺"`
	Unchanged int `json:"unchanged" dc:"鏈彉鍖栨暟閲?`
}

type WorkflowCreateRes struct {
	WorkflowMutationStats
}

type WorkflowUpdateReq struct {
	g.Meta      `path:"/{id}" method:"put" tags:"宸ヤ綔娴? summary:"鏇存柊宸ヤ綔娴?`
	ID          int64  `json:"id" in:"path" v:"required#id涓嶈兘涓虹┖" dc:"宸ヤ綔娴佹爣璇?`
	Name        string `json:"name" dc:"宸ヤ綔娴佸悕绉?`
	Description string `json:"description" dc:"宸ヤ綔娴佹弿杩?`
	Source      int    `json:"source" d:"0" v:"in:0,1,2#鏉ユ簮鍙兘鏄?銆?銆?" dc:"宸ヤ綔娴佹潵婧?`
	IsProtected bool   `json:"is_protected" d:"false" dc:"鏄惁鍙椾繚鎶?`
	Revision    int    `json:"revision" v:"required#鐗堟湰鍙蜂笉鑳戒负绌? dc:"鐗堟湰鍙?`
}

type WorkflowUpdateRes struct{}

type WorkflowProtectedReq struct {
	g.Meta      `path:"/{id}/syncable" method:"put" tags:"宸ヤ綔娴? summary:"鏇存柊宸ヤ綔娴佷繚鎶ょ姸鎬?`
	ID          int64 `json:"id" in:"path" v:"required#id涓嶈兘涓虹┖" dc:"宸ヤ綔娴佹爣璇?`
	IsProtected bool  `json:"is_protected" dc:"鏄惁鍙椾繚鎶?`
	Revision    int   `json:"revision" v:"required#鐗堟湰鍙蜂笉鑳戒负绌? dc:"鐗堟湰鍙?`
}

type WorkflowProtectedRes struct{}

type WorkflowBatchDeleteReq struct {
	g.Meta `path:"/" method:"delete" tags:"宸ヤ綔娴? summary:"鎵归噺鍒犻櫎宸ヤ綔娴?`
	IDs    []string `json:"ids" v:"required|min-length:1#宸ヤ綔娴乮d鍒楄〃涓嶈兘涓虹┖|鑷冲皯閫夋嫨涓€涓伐浣滄祦" dc:"宸ヤ綔娴?ID 鍒楄〃"`
}

type WorkflowBatchDeleteRes struct{}

type WorkflowImportFilesReq struct {
	g.Meta `path:"/import" method:"post" tags:"宸ヤ綔娴? summary:"瀵煎叆宸ヤ綔娴佸帇缂╁寘" mime:"multipart/form-data"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required#ZIP鍘嬬缉鏂囦欢涓嶈兘涓虹┖" dc:"宸ヤ綔娴?ZIP 鏂囦欢"`
}

type WorkflowImportFilesRes struct {
	WorkflowMutationStats
}

type WorkflowSyncCandidatesReq struct {
	g.Meta `path:"/sync-candidates" method:"get" tags:"宸ヤ綔娴? summary:"鑾峰彇鍙悓姝ュ伐浣滄祦"`
	rr.CommonPageReq
	Mode           string `json:"mode" in:"query" d:"client" v:"in:client,workflow#鏌ヨ缁村害鍙兘鏄痗lient鎴杦orkflow" dc:"鏌ヨ缁村害"`
	SourceIP       string `json:"source_ip" in:"query" dc:"瀹㈡埛绔潵婧愬湴鍧€"`
	SourceNodeID   string `json:"source_node_id" in:"query" dc:"鏉ユ簮鎵ц鑺傜偣 ID"`
	AutomaID       string `json:"automa_id" in:"query" dc:"Automa 宸ヤ綔娴?ID"`
	Keyword        string `json:"keyword,omitempty" in:"query" dc:"鍏抽敭瀛?`
	SyncStatus     string `json:"sync_status,omitempty" in:"query" dc:"同步状态"`
	WorkflowStatus string `json:"workflow_status,omitempty" in:"query" dc:"工作流状态"`
	Refresh        bool   `json:"refresh" in:"query" d:"false" dc:"鏄惁鍏堝埛鏂板鎴风宸ヤ綔娴佹竻鍗?`
}

type WorkflowSyncCandidatesResModel struct {
	Id                string      `json:"id" dc:"瀹㈡埛绔伐浣滄祦鏍囪瘑"`
	AutomaId          string      `json:"automa_id" dc:"Automa 宸ヤ綔娴?ID"`
	WorkflowId        string      `json:"workflow_id" dc:"宸ヤ綔娴?ID"`
	Name              string      `json:"name" dc:"宸ヤ綔娴佸悕绉?`
	Description       string      `json:"description" dc:"宸ヤ綔娴佹弿杩?`
	AutomaName        string      `json:"automa_name" dc:"Automa 宸ヤ綔娴佸師濮嬪悕绉?`
	AutomaDescription string      `json:"automa_description" dc:"Automa 宸ヤ綔娴佸師濮嬫弿杩?`
	Source            string      `json:"source" dc:"宸ヤ綔娴佹潵婧?`
	SourceIp          string      `json:"source_ip" dc:"鏉ユ簮鍦板潃"`
	MachineId         string      `json:"machine_id" dc:"鏈哄櫒 ID"`
	NodeId            string      `json:"node_id" dc:"鎵ц鑺傜偣 ID"`
	NodeName          string      `json:"node_name" dc:"鎵ц鑺傜偣鍚嶇О"`
	AutomaVersion     string      `json:"automa_version" dc:"Automa 鐗堟湰"`
	ExtVersion        string      `json:"ext_version" dc:"鎵╁睍鐗堟湰"`
	CreatedAtAutoma   int64       `json:"created_at_automa" dc:"Automa 鍒涘缓鏃堕棿"`
	UpdatedAtAutoma   int64       `json:"updated_at_automa" dc:"Automa 鏇存柊鏃堕棿"`
	IsDisabled        bool        `json:"is_disabled" dc:"鏄惁绂佺敤"`
	IsProtected       bool        `json:"is_protected" dc:"鏄惁鍙椾繚鎶?`
	NodeCount         int         `json:"node_count" dc:"鑺傜偣鏁伴噺"`
	EdgeCount         int         `json:"edge_count" dc:"杩炵嚎鏁伴噺"`
	ContentHash       string      `json:"content_hash" dc:"鍐呭鍝堝笇"`
	Synced            bool        `json:"synced" dc:"鏄惁宸插悓姝?`
	HasUpdate         bool        `json:"has_update" dc:"鏄惁鏈夋洿鏂?`
	SyncStatus        string      `json:"sync_status" dc:"鍚屾鐘舵€?`
	ServerId          int64       `json:"server_id" dc:"鏈嶅姟绔富閿?`
	ServerName        string      `json:"server_name" dc:"鏈嶅姟绔悕绉?`
	ServerDesc        string      `json:"server_description" dc:"鏈嶅姟绔弿杩?`
	ServerAutomaName  string      `json:"server_automa_name" dc:"鏈嶅姟绔?Automa 鍘熷鍚嶇О"`
	ServerAutomaDesc  string      `json:"server_automa_description" dc:"鏈嶅姟绔?Automa 鍘熷鎻忚堪"`
	ServerRevision    int         `json:"server_revision" dc:"鏈嶅姟绔増鏈?`
	LastSyncedAt      *gtime.Time `json:"last_synced_at" dc:"鏈€杩戝悓姝ユ椂闂?`
	ServerUpdatedAt   *gtime.Time `json:"server_updated_at" dc:"鏈嶅姟绔洿鏂版椂闂?`
	Online            bool        `json:"online" dc:"鏄惁鍦ㄧ嚎"`
}

type WorkflowSyncCandidatesRes struct {
	List  []WorkflowSyncCandidatesResModel `json:"list" dc:"鍙悓姝ュ伐浣滄祦鍒楄〃"`
	Total int                              `json:"total" dc:"鍙悓姝ュ伐浣滄祦鏁伴噺"`
}

type WorkflowSyncReq struct {
	g.Meta               `path:"/sync" method:"post" tags:"宸ヤ綔娴? summary:"鍚屾宸ヤ綔娴?`
	SourceIP             string          `json:"source_ip" v:"required#瀹㈡埛绔痠p涓嶈兘涓虹┖" dc:"瀹㈡埛绔潵婧愬湴鍧€"`
	SourceNodeID         string          `json:"source_node_id" dc:"鏉ユ簮鎵ц鑺傜偣 ID"`
	WorkflowIds          []string        `json:"workflow_ids" dc:"闇€瑕佸悓姝ョ殑宸ヤ綔娴?ID 鍒楄〃"`
	WorkflowJsonDataList []model.JSONMap `json:"workflows" dc:"宸ヤ綔娴?JSON 鏁版嵁鍒楄〃"`
}

type WorkflowSyncRes struct{}

type WorkflowClientMaintenanceReq struct {
	g.Meta        `path:"/client-maintenance" method:"post" tags:"宸ヤ綔娴? summary:"缁存姢瀹㈡埛绔墽琛岃妭鐐瑰伐浣滄祦"`
	SourceIP      string   `json:"source_ip" v:"required#瀹㈡埛绔?IP 涓嶈兘涓虹┖" dc:"瀹㈡埛绔潵婧愬湴鍧€"`
	SourceNodeID  string   `json:"source_node_id" dc:"鎵ц鑺傜偣 ID"`
	SourceNodeIDs []string `json:"source_node_ids" dc:"鎵ц鑺傜偣 ID 鍒楄〃"`
	Action        string   `json:"action" v:"required|in:install,update,delete#鎿嶄綔涓嶈兘涓虹┖|鎿嶄綔鍙兘鏄?install銆乽pdate銆乨elete" dc:"缁存姢鍔ㄤ綔"`
	WorkflowIds   []string `json:"workflow_ids" v:"required|min-length:1#宸ヤ綔娴佷笉鑳戒负绌簗鑷冲皯閫夋嫨涓€涓伐浣滄祦" dc:"宸ヤ綔娴?ID 鍒楄〃"`
	Timeout       int      `json:"timeout" d:"20" dc:"绛夊緟瓒呮椂绉掓暟"`
	Refresh       bool     `json:"refresh" d:"true" dc:"瀹屾垚鍚庢槸鍚﹀埛鏂拌妭鐐规竻鍗?`
}

type WorkflowClientMaintenanceRes struct {
	Submitted   int      `json:"submitted" dc:"鎻愪氦鏁伴噺"`
	Succeeded   int      `json:"succeeded" dc:"鎴愬姛鏁伴噺"`
	Failed      int      `json:"failed" dc:"澶辫触鏁伴噺"`
	WorkflowIds []string `json:"workflow_ids" dc:"宸ヤ綔娴?ID 鍒楄〃"`
	Message     string   `json:"message" dc:"缁撴灉璇存槑"`
}

type WorkflowCacheReq struct {
	g.Meta `path:"/cache" method:"get" tags:"宸ヤ綔娴? summary:"鑾峰彇鏈湴宸ヤ綔娴佸揩鐓?`
}

type WorkflowCacheRes struct {
	Snapshot *model.AutomaWorkflowSnapshot `json:"snapshot,omitempty" dc:"鏈湴宸ヤ綔娴佸揩鐓?`
}

type WorkflowAgentListReq struct {
	g.Meta    `path:"/agent/workflows" method:"get" tags:"宸ヤ綔娴? summary:"浠庢祻瑙堝櫒鎵ц绔鍙栧伐浣滄祦"`
	BrowserID string `json:"browser_id" in:"query" dc:"娴忚鍣ㄥ疄渚?ID"`
}

type WorkflowAgentListRes struct {
	BrowserID string           `json:"browser_id" dc:"娴忚鍣ㄥ疄渚?ID"`
	Workflows []map[string]any `json:"workflows" dc:"宸ヤ綔娴佸垪琛?`
	Total     int              `json:"total" dc:"宸ヤ綔娴佹€绘暟"`
}

type WorkflowAgentExportSkillReq struct {
	g.Meta      `path:"/agent/export/skill" method:"post" tags:"宸ヤ綔娴? summary:"瀵煎嚭娴忚鍣ㄦ墽琛岀宸ヤ綔娴佷负 Skill"`
	BrowserID   string   `json:"browser_id" dc:"娴忚鍣ㄥ疄渚?ID"`
	Scope       string   `json:"scope" d:"filtered" v:"in:selected,filtered,all#瀵煎嚭鑼冨洿鍙兘鏄痵elected銆乫iltered鎴朼ll" dc:"瀵煎嚭鑼冨洿"`
	WorkflowIDs []string `json:"workflow_ids" dc:"闇€瑕佸鍑虹殑宸ヤ綔娴?ID 鍒楄〃"`
}

type WorkflowAgentExportSkillRes struct{}

type WorkflowExportSkillReq struct {
	g.Meta      `path:"/export/skill" method:"post" tags:"宸ヤ綔娴? summary:"瀵煎嚭鏈嶅姟绔伐浣滄祦涓?Skill"`
	Scope       string   `json:"scope" d:"filtered" v:"in:selected,filtered,all#瀵煎嚭鑼冨洿鍙兘鏄痵elected銆乫iltered鎴朼ll" dc:"瀵煎嚭鑼冨洿"`
	WorkflowIDs []string `json:"workflow_ids" dc:"闇€瑕佸鍑虹殑宸ヤ綔娴?ID 鍒楄〃"`
}

type WorkflowExportSkillRes struct{}

type WorkflowRunReq struct {
	g.Meta     `path:"/{id}/run" method:"post" tags:"宸ヤ綔娴? summary:"杩愯宸ヤ綔娴?`
	ID         string                             `json:"id" in:"path" v:"required#宸ヤ綔娴両D涓嶈兘涓虹┖" dc:"宸ヤ綔娴?ID"`
	BrowserID  string                             `json:"browser_id" dc:"娴忚鍣ㄥ疄渚?ID"`
	Variables  model.JSONMap                      `json:"variables" dc:"杩愯鍙橀噺"`
	WaitResult bool                               `json:"wait_result" d:"false" dc:"鏄惁绛夊緟宸ヤ綔娴佸畬鎴?`
	Timeout    int                                `json:"timeout" d:"300" dc:"绛夊緟瓒呮椂绉掓暟"`
	ReturnData *model.WorkflowExecutionReturnData `json:"return_data" dc:"鍥炰紶鏁版嵁閰嶇疆"`
}

type WorkflowRunRes struct {
	Result    *model.AgentCommandResult `json:"result,omitempty" dc:"杩愯缁撴灉"`
	Execution *model.WorkflowExecution  `json:"execution,omitempty" dc:"鎵ц鐘舵€?`
}

type WorkflowExecutionDetailReq struct {
	g.Meta      `path:"/executions/{execution_id}" method:"get" tags:"宸ヤ綔娴? summary:"鑾峰彇宸ヤ綔娴佹墽琛岀姸鎬?`
	ExecutionID string `json:"execution_id" in:"path" v:"required#鎵цID涓嶈兘涓虹┖" dc:"鎵ц ID"`
}

type WorkflowExecutionDetailRes struct {
	Execution *model.WorkflowExecution `json:"execution,omitempty" dc:"鎵ц鐘舵€?`
}

type WorkflowOpenReq struct {
	g.Meta    `path:"/{id}/open" method:"post" tags:"宸ヤ綔娴? summary:"鎵撳紑宸ヤ綔娴?`
	ID        string `json:"id" in:"path" v:"required#宸ヤ綔娴両D涓嶈兘涓虹┖" dc:"宸ヤ綔娴?ID"`
	BrowserID string `json:"browser_id" dc:"娴忚鍣ㄥ疄渚?ID"`
}

type WorkflowOpenRes struct {
	Result *model.AgentCommandResult `json:"result,omitempty" dc:"鎵撳紑缁撴灉"`
}
