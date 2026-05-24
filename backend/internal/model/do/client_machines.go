// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ClientMachines is the golang structure of table client_machines for DAO operations like Where/Data.
type ClientMachines struct {
	g.Meta          `orm:"table:client_machines, do:true"`
	Id              interface{} // 主键 ID，自增
	MachineId       interface{} // 物理机器稳定 ID，同一台 Windows 电脑保持一致
	MachineName     interface{} // Worker 上报的机器名称
	DisplayName     interface{} // 服务端自定义机器显示名称
	ClientIp        interface{} // 机器当前或最后一次连接 IP
	Hostname        interface{} // 系统主机名
	OsName          interface{} // 操作系统名称
	OsVersion       interface{} // 操作系统版本
	WorkerVersion   interface{} // BrowserFlow Worker 版本
	InstallDir      interface{} // Worker 安装目录
	DataDir         interface{} // Worker 数据目录
	Status          interface{} // 机器状态：online 在线，offline 离线
	IsBanned        interface{} // 是否禁止该机器连接
	BanReason       interface{} // 机器被禁止连接的原因
	NodeCount       interface{} // 该机器配置的执行节点数量
	OnlineNodeCount interface{} // 该机器当前在线执行节点数量
	ConfigJson      interface{} // 机器级配置快照
	FirstSeenAt     *gtime.Time // 第一次连接到服务端的时间
	LastSeenAt      *gtime.Time // 最近一次心跳或交互时间
	CreatedAt       *gtime.Time // 记录创建时间
	UpdatedAt       *gtime.Time // 记录更新时间
	DeletedAt       *gtime.Time // 软删除时间
}
