// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AutomaPackages is the golang structure for table automa_packages.
type AutomaPackages struct {
	Id           int64       `json:"id"            orm:"id"            ` // 自增 ID
	PackageName  string      `json:"package_name"  orm:"package_name"  ` // 安装包名称，例如 Automa Chrome Extension
	Version      string      `json:"version"       orm:"version"       ` // 安装包版本号
	FileName     string      `json:"file_name"     orm:"file_name"     ` // 原始文件名
	FilePath     string      `json:"file_path"     orm:"file_path"     ` // 服务端存储路径
	FileSize     int64       `json:"file_size"     orm:"file_size"     ` // 文件大小，单位字节
	Sha256       string      `json:"sha_256"       orm:"sha256"        ` // 安装包 sha256 校验值
	ManifestJson string      `json:"manifest_json" orm:"manifest_json" ` // 插件 manifest 内容快照
	Status       string      `json:"status"        orm:"status"        ` // 安装包状态：active 可用，disabled 禁用
	IsDefault    bool        `json:"is_default"    orm:"is_default"    ` // 是否为默认下发版本
	CreatedAt    *gtime.Time `json:"created_at"    orm:"created_at"    ` // 记录创建时间
	UpdatedAt    *gtime.Time `json:"updated_at"    orm:"updated_at"    ` // 记录更新时间
	DeletedAt    *gtime.Time `json:"deleted_at"    orm:"deleted_at"    ` // 软删除时间
}
