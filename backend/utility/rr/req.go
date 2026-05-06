// @Author daixk 2024/8/29 0:09:00
package rr

// CommonPageReq 公共分页请求参数
type CommonPageReq struct {
	PageNum  int `json:"page_num" in:"query" d:"1" v:"min:1#页码必须大于等于1" dc:"页码，从 1 开始"`
	PageSize int `json:"page_size" in:"query" d:"30" v:"in:10,30,60#页大小只能是10/30/60" dc:"每页数量，只能是 10、30 或 60"`
}

// CommonTimeReq 公共时间请求参数
type CommonTimeReq struct {
	StartTime string `json:"start_time" in:"query" v:"datetime#开始时间格式错误" dc:"开始时间，格式为日期时间字符串"`
	EndTime   string `json:"end_time" in:"query" v:"datetime#结束时间格式错误" dc:"结束时间，格式为日期时间字符串"`
}
