package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserInstancePayload 浏览器实例载荷
type BrowserInstancePayload struct {
	Name        string   `json:"name" dc:"浏览器实例名称"`
	Description string   `json:"description" dc:"浏览器实例描述"`
	IsDefault   bool     `json:"is_default" dc:"是否默认实例"`
	Type        string   `json:"type" dc:"浏览器类型"`
	BinPath     string   `json:"bin_path" dc:"浏览器可执行文件路径"`
	UserDataDir string   `json:"user_data_dir" dc:"用户数据目录"`
	ControlURL  string   `json:"control_url" dc:"浏览器控制地址"`
	UserAgent   string   `json:"user_agent,omitempty" dc:"用户代理"`
	Headless    *bool    `json:"headless,omitempty" dc:"是否无头模式"`
	NoSandbox   *bool    `json:"no_sandbox,omitempty" dc:"是否禁用沙箱"`
	LaunchArgs  []string `json:"launch_args,omitempty" dc:"启动参数列表"`
	Proxy       string   `json:"proxy,omitempty" dc:"代理地址"`
}

// BrowserStatusResModel 浏览器状态响应模型
type BrowserStatusResModel = model.BrowserStatus

// BrowserInstanceListResModel 浏览器实例列表项
type BrowserInstanceListResModel = model.BrowserInstance

// BrowserStatusReq 获取浏览器状态
type BrowserStatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"浏览器" summary:"获取浏览器状态"`
}

// BrowserStatusRes 浏览器状态响应
type BrowserStatusRes struct {
	Status *BrowserStatusResModel `json:"status,omitempty" dc:"浏览器状态"`
}

// BrowserStartReq 启动默认浏览器
type BrowserStartReq struct {
	g.Meta `path:"/start" method:"post" tags:"浏览器" summary:"启动默认浏览器"`
}

// BrowserStartRes 浏览器启动响应
type BrowserStartRes struct {
	Status *BrowserStatusResModel `json:"status,omitempty" dc:"浏览器状态"`
}

// BrowserStopReq 停止当前浏览器
type BrowserStopReq struct {
	g.Meta `path:"/stop" method:"post" tags:"浏览器" summary:"停止当前浏览器"`
}

// BrowserStopRes 浏览器停止响应
type BrowserStopRes struct {
	Status *BrowserStatusResModel `json:"status,omitempty" dc:"浏览器状态"`
}

// BrowserInstanceListReq 获取浏览器实例列表
type BrowserInstanceListReq struct {
	g.Meta `path:"/instances" method:"get" tags:"浏览器" summary:"获取浏览器实例列表"`
}

// BrowserInstanceListRes 浏览器实例列表响应
type BrowserInstanceListRes struct {
	Instances []*BrowserInstanceListResModel `json:"instances,omitempty" dc:"浏览器实例列表"`
}

// BrowserInstanceCreateReq 创建浏览器实例
type BrowserInstanceCreateReq struct {
	g.Meta `path:"/instances" method:"post" tags:"浏览器" summary:"创建浏览器实例"`
	BrowserInstancePayload
}

// BrowserInstanceCreateRes 浏览器实例创建响应
type BrowserInstanceCreateRes struct {
	Instance *BrowserInstanceListResModel `json:"instance,omitempty" dc:"浏览器实例"`
}

// BrowserInstanceCurrentReq 获取当前浏览器实例
type BrowserInstanceCurrentReq struct {
	g.Meta `path:"/instances/current" method:"get" tags:"浏览器" summary:"获取当前浏览器实例"`
}

// BrowserInstanceCurrentRes 当前浏览器实例响应
type BrowserInstanceCurrentRes struct {
	Status   *BrowserStatusResModel       `json:"status,omitempty" dc:"浏览器状态"`
	Instance *BrowserInstanceListResModel `json:"instance,omitempty" dc:"当前浏览器实例"`
}

// BrowserInstanceDetailReq 获取浏览器实例详情
type BrowserInstanceDetailReq struct {
	g.Meta `path:"/instances/{id}" method:"get" tags:"浏览器" summary:"获取浏览器实例详情"`
	ID     string `json:"id" in:"path" v:"required#浏览器实例ID不能为空" dc:"浏览器实例ID"`
}

// BrowserInstanceDetailRes 浏览器实例详情响应
type BrowserInstanceDetailRes struct {
	Instance *BrowserInstanceListResModel `json:"instance,omitempty" dc:"浏览器实例"`
}

// BrowserInstanceUpdateReq 更新浏览器实例
type BrowserInstanceUpdateReq struct {
	g.Meta `path:"/instances/{id}" method:"put" tags:"浏览器" summary:"更新浏览器实例"`
	ID     string `json:"id" in:"path" v:"required#浏览器实例ID不能为空" dc:"浏览器实例ID"`
	BrowserInstancePayload
}

// BrowserInstanceUpdateRes 浏览器实例更新响应
type BrowserInstanceUpdateRes struct {
	Instance *BrowserInstanceListResModel `json:"instance,omitempty" dc:"浏览器实例"`
}

// BrowserInstanceDeleteReq 删除浏览器实例
type BrowserInstanceDeleteReq struct {
	g.Meta `path:"/instances/{id}" method:"delete" tags:"浏览器" summary:"删除浏览器实例"`
	ID     string `json:"id" in:"path" v:"required#浏览器实例ID不能为空" dc:"浏览器实例ID"`
}

// BrowserInstanceDeleteRes 删除响应
type BrowserInstanceDeleteRes struct {
	Message string `json:"message,omitempty" dc:"删除结果消息"`
}

// BrowserInstanceStartReq 启动浏览器实例
type BrowserInstanceStartReq struct {
	g.Meta `path:"/instances/{id}/start" method:"post" tags:"浏览器" summary:"启动浏览器实例"`
	ID     string `json:"id" in:"path" v:"required#浏览器实例ID不能为空" dc:"浏览器实例ID"`
}

// BrowserInstanceStartRes 浏览器实例启动响应
type BrowserInstanceStartRes struct {
	Status *BrowserStatusResModel `json:"status,omitempty" dc:"浏览器状态"`
}

// BrowserInstanceStopReq 停止浏览器实例
type BrowserInstanceStopReq struct {
	g.Meta `path:"/instances/{id}/stop" method:"post" tags:"浏览器" summary:"停止浏览器实例"`
	ID     string `json:"id" in:"path" v:"required#浏览器实例ID不能为空" dc:"浏览器实例ID"`
}

// BrowserInstanceStopRes 浏览器实例停止响应
type BrowserInstanceStopRes struct {
	Status *BrowserStatusResModel `json:"status,omitempty" dc:"浏览器状态"`
}

// BrowserInstanceSwitchReq 切换当前浏览器实例
type BrowserInstanceSwitchReq struct {
	g.Meta `path:"/instances/{id}/switch" method:"post" tags:"浏览器" summary:"切换浏览器实例"`
	ID     string `json:"id" in:"path" v:"required#浏览器实例ID不能为空" dc:"浏览器实例ID"`
}

// BrowserInstanceSwitchRes 浏览器实例切换响应
type BrowserInstanceSwitchRes struct {
	Status *BrowserStatusResModel `json:"status,omitempty" dc:"浏览器状态"`
}
