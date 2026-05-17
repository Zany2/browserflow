package browserexecutor

import "github.com/gogf/gf/v2/net/ghttp"

// RequestBaseURL builds the API base URL from an HTTP request. 根据请求生成 API 基础地址。
func RequestBaseURL(request *ghttp.Request) string {
	scheme := "http"
	if request != nil && request.TLS != nil {
		scheme = "https"
	}
	host := ""
	if request != nil {
		host = request.Host
	}
	return scheme + "://" + host + "/api/v1"
}
