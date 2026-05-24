// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AutomaPackages is the golang structure of table automa_packages for DAO operations like Where/Data.
type AutomaPackages struct {
	g.Meta       `orm:"table:automa_packages, do:true"`
	Id           interface{} // 自增 ID
	PackageName  interface{} // 安装包名称，例如 Automa Chrome Extension
	Version      interface{} // 安装包版本号
	FileName     interface{} // 原始文件名
	FilePath     interface{} // 服务端存储路径
	FileSize     interface{} // 文件大小，单位字节
	Sha256       interface{} // 安装包 sha256 校验值
	ManifestJson interface{} // 插件 manifest 内容快照
	Status       interface{} // 安装包状态：active 可用，disabled 禁用
	IsDefault    interface{} // 是否为默认下发版本
	CreatedAt    *gtime.Time // 记录创建时间
	UpdatedAt    *gtime.Time // 记录更新时间
	DeletedAt    *gtime.Time // 软删除时间
}
