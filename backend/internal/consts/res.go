// @Author daixk 2025/3/10 14:12:00
package consts

// CodeMessageMap code message map 状态码消息映射
var CodeMessageMap map[int]string

// ResCode response code 响应状态码
const (
	CodeOK          = 20000 // 请求成功
	CodeBadRequest  = 40000 // 错误请求
	CodeParamError  = 40002 // 参数错误
	CodeServerError = 50000 // 服务异常
)

func init() {
	CodeMessageMap = make(map[int]string)

	CodeMessageMap[CodeOK] = "请求成功"
	CodeMessageMap[CodeBadRequest] = "请求错误"
	CodeMessageMap[CodeParamError] = "请求参数有误"
	CodeMessageMap[CodeServerError] = "服务异常，请稍后再试或联系管理员"
}

// ResMessage common response message 通用返回消息
var (
	DoSuccessMessage          = "操作成功"
	InsertSuccessMessage      = "新增成功"
	UpdateSuccessMessage      = "修改成功"
	DeleteSuccessMessage      = "删除成功"
	DeleteBatchSuccessMessage = "成功删除%d条数据"

	DataDeletedMessage = "该数据已删除"
	DataRefreshMessage = "数据状态已更新，请刷新后查看"
)
