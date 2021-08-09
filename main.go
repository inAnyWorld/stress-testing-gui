// Package main go 实现的压测工具
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go-stress-testing/global"
	"go-stress-testing/helper"
	"go-stress-testing/model"
	"go-stress-testing/server"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

// array 自定义数组参数
type array []string

// String string
func (a *array) String() string {
	return fmt.Sprint(*a)
}

// Set set
func (a *array) Set(s string) error {
	*a = append(*a, s)

	return nil
}

var (
	concurrency uint64  = 1       // 并发数
	totalNumber uint64  = 1       // 请求数(单个并发/协程)
	debugStr            = "false" // 是否是debug
	requestURL          = ""      // 压测的url 目前支持，http/https ws/wss
	path                = ""      // curl文件路径 http接口压测，自定义参数设置
	verify              = ""      // verify 验证方法 在server/verify中 http 支持:statusCode、json webSocket支持:json
	headers     array             // 自定义头信息传递给服务器
	body        = ""              // HTTP POST方式传送数据
	maxCon      = 1               // 单个连接最大请求数
	code        = 200             //成功状态码
	http2       = false           // 是否开http2.0
	keepalive   = false           // 是否开启长连接
)
var templateVar *template.Template

// initFlag 设置压测必要参数
func initFlag() {
	if _, ok := global.RequestParams["concurrent"]; ok {
		int64c, _ := strconv.ParseInt(global.RequestParams["concurrent"].(string), 10, 64)
		concurrency = uint64(int64c)
	}
	if _, ok := global.RequestParams["request"]; ok {
		int64n, _ := strconv.ParseInt(global.RequestParams["request"].(string), 10, 64)
		totalNumber = uint64(int64n)
	}
	path = global.RequestUri
}

// main go 实现的压测工具
// 编译可执行文件

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/testApi", TestV1ApiHandler).Methods("GET")
	r.HandleFunc("/v1/buildCURL", BuildV1CURLHandler).Methods("POST")
	err := http.ListenAndServe(":8181", r)
	if err != nil {
		log.Println("http start or listen failed:", err.Error())
		return
	}
}

// TestV1ApiHandler 模板
func TestV1ApiHandler(w http.ResponseWriter, r *http.Request) {
	templateFiles, templateErr := template.ParseFiles("web/test.html")
	if templateErr != nil {
		log.Println("ParseFiles Error:", templateErr.Error())
		return
	}
	templateVar = templateFiles
	templateExeErr := templateVar.Execute(w, nil)
	if templateExeErr != nil {
		log.Println("Execute Error:", templateExeErr.Error())
		return
	}
}

// BuildV1CURLHandler 构建cURL请求体
func BuildV1CURLHandler(w http.ResponseWriter, r *http.Request) {
	var requestMap map[string]interface{}
	var returnMessage, returnStatus string
	rBody, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	err := json.Unmarshal(rBody, &requestMap)
	if err != nil {
		returnMessage = fmt.Sprintf("json Unmarshal Err: %v", err.Error())
		returnStatus = "fail"
		log.Println(returnMessage)
		global.FrontException = true
	} else {
		buildCURLFileRsp := helper.BuildCURLHandlerHelper(requestMap)
		if _, ok := buildCURLFileRsp["status"]; ok && buildCURLFileRsp["status"] == "success" {
			initFlag()
			doWork()
			if !global.ExeException && global.CompleteException{
				returnMessage = global.BufferString.String()
				returnStatus = "success"
			} else {
				returnMessage = "system error"
			}
		}
	}
	defer global.BufferString.Reset()
	//返回数据
	response := map[string]interface{}{
		"code":200,
		"status": returnStatus,
		"msg": returnMessage,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-Type","application/json")
	indent, indentErr := json.MarshalIndent(response, "", "\t")
	if indentErr != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found data"))
	}
	w.Write(indent)
}

// doWork 执行压测
func doWork()  {
	runtime.GOMAXPROCS(1)
	if concurrency == 0 || totalNumber == 0 || (requestURL == "" && path == "") {
		log.Printf("示例: go run main.go -c 1 -n 1 -u https://www.baidu.com/ \n")
		log.Printf("压测地址或curl路径必填 \n")
		log.Printf("当前请求参数: -c %d -n %d -d %v -u %s \n", concurrency, totalNumber, debugStr, requestURL)
		global.ExeException = true
		return
	}
	debug := strings.ToLower(debugStr) == "true"
	request, err := model.NewRequest(requestURL, verify, code, 0, debug, path, headers, body, maxCon, http2, keepalive)
	if err != nil {
		global.ExeException = true
		log.Printf("参数不合法 %v \n", err)
		return
	}
	fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", concurrency, totalNumber)
	// 输出结果
	helper.OutputResult(
		"<h5>执行信息</h5><table border='1'>" +
				"<tr>" +
					"<th>执行信息</th>" +
					"<th>并发数</th>" +
					"<th>请求数</th>" +
				"</tr>" +
				"<tr>" +
					"<td>数量</td>" +
					"<td>"+fmt.Sprintf("%d",concurrency) +"</td>" +
					"<td>"+fmt.Sprintf("%d",totalNumber) +"</td>" +
				"</tr>" +
			"</table><br/>")
	request.Print()
	// 开始处理
	server.Dispose(concurrency, totalNumber, request)
	return
}