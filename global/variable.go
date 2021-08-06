package global

import (
	"bytes"
)

// RequestParams 客户端参数
var RequestParams map[string]interface{}

// RequestUri cURL请求地址
var RequestUri string

// BufferString 压测报告
var BufferString bytes.Buffer

// CompleteException 执行结果
var CompleteException bool

// ExeException 检查是否执行异常
var ExeException bool

// FrontException 压测前置条件
var FrontException bool
