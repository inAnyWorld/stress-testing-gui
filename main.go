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
	"go-stress-testing/server/gohttp"
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
	r.HandleFunc("/v1/testQPS", TestV1ApiQPSHandler).Methods("GET")
	r.HandleFunc("/v1/buildCURL", BuildV1CURLHandler).Methods("POST")
	r.HandleFunc("/v1/sendHttpRequest", SendHttpRequest).Methods("POST")
	err := http.ListenAndServe(":8089", r)
	if err != nil {
		log.Println("http start or listen failed:", err.Error())
		return
	}
}

// TestV1ApiHandler 模板
func TestV1ApiHandler(w http.ResponseWriter, r *http.Request) {
	templateFiles, templateErr := template.ParseFiles("web/testApi.html")
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

// TestV1ApiQPSHandler 模板
func TestV1ApiQPSHandler(w http.ResponseWriter, r *http.Request) {
	templateFiles, templateErr := template.ParseFiles("web/testQPS.html")
	if templateErr != nil {
		log.Println("ParseFiles QPS Error:", templateErr.Error())
		return
	}
	templateVar = templateFiles
	templateExeErr := templateVar.Execute(w, nil)
	if templateExeErr != nil {
		log.Println("Execute QPS Error:", templateExeErr.Error())
		return
	}
}

// SendHttpRequest 发送请求
func SendHttpRequest(w http.ResponseWriter, r *http.Request) {
	var requestParams map[string]interface{}
	rBody, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	jsonErr := json.Unmarshal(rBody, &requestParams)
	if jsonErr != nil {
		log.Println("json Unmarshal Err: ", jsonErr.Error())
		returnJson("fail", fmt.Sprintf("json Unmarshal Err: %v", jsonErr.Error()), w)
		return
	}
	if !helper.Ping(requestParams["uri"].(string)) {
		returnJson("fail", "接口地址不合法", w)
		return
	}
	uuid := helper.Uuid()
	defer delete(global.BufferMap, uuid)
	requestParams["uuid"] = uuid
	helper.OutputResult(
		"<table border='1'>" +
			"<tr>" +
				"<th>QPS</th>" +
				"<th>压测总时长</th>" +
				"<th>请求耗时</th>" +
				"<th>成功数</th>" +
				"<th>失败数</th>" +
			"</tr>", requestParams["uuid"].(string))
	if requestParams["method"].(string) == global.GET {
		gohttp.HttpRequestGet(requestParams)
	}
	if requestParams["method"].(string) == global.POST {
		gohttp.HttpRequestPost(requestParams)
	}
	if buffer, exist := global.BufferMap[uuid]; exist {
		returnJson("success", buffer.String(), w)
		return
	}
	returnJson("fail","system error", w)
	return
}

// BuildV1CURLHandler 构建cURL请求体
func BuildV1CURLHandler(w http.ResponseWriter, r *http.Request) {
	var requestParams map[string]interface{}
	rBody, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	jsonErr := json.Unmarshal(rBody, &requestParams)
	if jsonErr != nil {
		log.Println("json Unmarshal Err: ", jsonErr.Error())
		returnJson("fail", fmt.Sprintf("json Unmarshal Err: %v", jsonErr.Error()), w)
		return
	}

	uuid := helper.Uuid()
	if uuid == "false" {
		returnJson("fail", "uuid 生成失败1", w)
		return
	}

	buildCURLFileRsp := helper.BuildCURLHandlerHelper(requestParams)
	defer delete(global.BufferMap, uuid)

	if _, ok := buildCURLFileRsp["status"]; ok && buildCURLFileRsp["status"] == "success" {
		initFlag()
		if concurrency == 0 || totalNumber == 0 || path == "" {
			returnJson("fail", "缺乏必要参数,并发数|请求数|接口地址", w)
			return
		}
		var exception global.Exception
		doWork(uuid, &exception)

		if !exception.ExeException && exception.CompleteException {
			if buffer, exist := global.BufferMap[uuid]; exist {
				returnJson("success", buffer.String(), w)
				return
			}
			returnJson("fail", "结果读取失败", w)
			return
		}
		returnJson("fail", "system error", w)
		return
	}

	returnJson("fail", buildCURLFileRsp["message"], w)
	return
}

// doWork 执行压测
func doWork(uuid string, exception *global.Exception) {
	//log.Println(uuid)
	runtime.GOMAXPROCS(1)
	if concurrency == 0 || totalNumber == 0 || (requestURL == "" && path == "") {
		log.Printf("示例: go run main.go -c 1 -n 1 -u https://www.baidu.com/ \n")
		log.Printf("压测地址或curl路径必填 \n")
		log.Printf("当前请求参数: -c %d -n %d -d %v -u %s \n", concurrency, totalNumber, debugStr, requestURL)
		helper.SetExeException(exception)
		return
	}
	debug := strings.ToLower(debugStr) == "true"
	request, err := model.NewRequest(requestURL, verify, code, 0, debug, path, headers, body, maxCon, http2, keepalive)
	if err != nil {
		helper.SetExeException(exception)
		log.Printf("参数不合法 %v \n", err)
		return
	}
	if global.Print {
		fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", concurrency, totalNumber)
	}
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
			"</table><br/>", uuid)
	request.Print()
	// 开始处理
	server.Dispose(uuid, concurrency, totalNumber, request, exception)
	return
}

// returnJson 封装返回
func returnJson(status string, message string, w http.ResponseWriter) {
	//返回数据
	response := map[string]interface{}{
		"code":200,
		"status": status,
		"msg": message,
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