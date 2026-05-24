// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// NodePackages is the golang structure for table node_packages.
type NodePackages struct {
	Id              int64       `json:"id"               orm:"id"               ` // 自增 ID
	NodeId          string      `json:"node_id"          orm:"node_id"          ` // 执行节点 ID
	MachineId       string      `json:"machine_id"       orm:"machine_id"       ` // 机器 ID
	PackageId       int64       `json:"package_id"       orm:"package_id"       ` // 关联的 Automa 插件安装包 ID
	ExpectedVersion string      `json:"expected_version" orm:"expected_version" ` // 服务端期望安装的插件版本
	ActualVersion   string      `json:"actual_version"   orm:"actual_version"   ` // 节点实际上报的插件版本
	ExpectedSha256  string      `json:"expected_sha_256" orm:"expected_sha256"  ` // 服务端期望安装包 sha256
	ActualSha256    string      `json:"actual_sha_256"   orm:"actual_sha256"    ` // 节点实际上报的安装内容 sha256
	DesiredStatus   string      `json:"desired_status"   orm:"desired_status"   ` // 期望状态：installed 安装，removed 移除
	InstallStatus   string      `json:"install_status"   orm:"install_status"   ` // 安装状态：unknown、pending、installing、installed、failed、removed
	LastInstallAt   *gtime.Time `json:"last_install_at"  orm:"last_install_at"  ` // 最近一次安装或更新完成时间
	LastCheckAt     *gtime.Time `json:"last_check_at"    orm:"last_check_at"    ` // 最近一次节点上报检查时间
	LastError       string      `json:"last_error"       orm:"last_error"       ` // 最近一次安装或检查错误
	CreatedAt       *gtime.Time `json:"created_at"       orm:"created_at"       ` // 记录创建时间
	UpdatedAt       *gtime.Time `json:"updated_at"       orm:"updated_at"       ` // 记录更新时间
	DeletedAt       *gtime.Time `json:"deleted_at"       orm:"deleted_at"       ` // 软删除时间
}
