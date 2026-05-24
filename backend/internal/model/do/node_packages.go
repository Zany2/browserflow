// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NodePackages is the golang structure of table node_packages for DAO operations like Where/Data.
type NodePackages struct {
	g.Meta          `orm:"table:node_packages, do:true"`
	Id              interface{} // 自增 ID
	NodeId          interface{} // 执行节点 ID
	MachineId       interface{} // 机器 ID
	PackageId       interface{} // 关联的 Automa 插件安装包 ID
	ExpectedVersion interface{} // 服务端期望安装的插件版本
	ActualVersion   interface{} // 节点实际上报的插件版本
	ExpectedSha256  interface{} // 服务端期望安装包 sha256
	ActualSha256    interface{} // 节点实际上报的安装内容 sha256
	DesiredStatus   interface{} // 期望状态：installed 安装，removed 移除
	InstallStatus   interface{} // 安装状态：unknown、pending、installing、installed、failed、removed
	LastInstallAt   *gtime.Time // 最近一次安装或更新完成时间
	LastCheckAt     *gtime.Time // 最近一次节点上报检查时间
	LastError       interface{} // 最近一次安装或检查错误
	CreatedAt       *gtime.Time // 记录创建时间
	UpdatedAt       *gtime.Time // 记录更新时间
	DeletedAt       *gtime.Time // 软删除时间
}
