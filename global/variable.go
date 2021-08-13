package global

import (
	"bytes"
)

// RequestParams 客户端参数
var RequestParams map[string]interface{}

// RequestUri cURL请求地址
var RequestUri string

// BufferMap 压测报告
var BufferMap = map[string]bytes.Buffer{}

// Print 是否打印
var Print bool
