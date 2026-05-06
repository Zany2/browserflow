// @Author daixk 2026/4/29 19:47:00
package consts

type AutomaSourceType int

const (
	AutomaTypeImport AutomaSourceType = 1 // 导入
	AutomaTypeSync   AutomaSourceType = 2 // 同步
)

var AutomaSourceTypeMap = map[AutomaSourceType]string{
	AutomaTypeImport: "导入",
	AutomaTypeSync:   "同步",
}
