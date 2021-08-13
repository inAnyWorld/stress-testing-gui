// Package helper 帮助函数，时间、数组的通用处理
package helper

import (
	"bytes"
	"encoding/json"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/valyala/fasthttp"
	"go-stress-testing/global"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

// DiffNano 时间差，纳秒
func DiffNano(startTime time.Time) (diff int64) {
	diff = int64(time.Since(startTime))
	return
}

// InArrayStr 判断字符串是否在数组内
func InArrayStr(str string, arr []string) (inArray bool) {
	for _, s := range arr {
		if s == str {
			inArray = true
			break
		}
	}
	return
}

// Round 向下取整
func Round(x float64) int {
	return int(math.Floor(x + 0/5))
}

// BuildCURLHandlerHelper 助手函数, 构建cURL请求体
func BuildCURLHandlerHelper(requestMap map[string]interface{}) map[string]string {
	if _, ok := requestMap["uri"]; !ok {
		return map[string]string{
			"status": "fail",
			"message": "接口地址缺失",
		}
	}
	if !Ping(requestMap["uri"].(string)) {
		return map[string]string{
			"status": "fail",
			"message": "接口地址不合法",
		}
	}
	var write2File string

	// paramsString url参数
	var paramsString, contentTypeString string
	for p, v := range requestMap["params"].(map[string]interface{}) {
		paramsString += p + "=" + v.(string) + "&"
	}

	requestMode := "GET"
	if requestMap["method"].(string) != "" {
		requestMode = requestMap["method"].(string)
	}
	write2File = `curl -X ` + requestMode + ` \`
	write2File += "\n"

	// url
	uri := requestMap["uri"].(string)
	if paramsString != "" {
		paramsString = strings.Trim(paramsString, "&")
		if strings.Contains(requestMap["uri"].(string), "?") {
			uri = requestMap["uri"].(string) + `&` + paramsString
		} else {
			uri = requestMap["uri"].(string) + `?` + paramsString
		}
	}

	write2File += `  '` + uri + `' \`
	write2File += "\n"

	// headers body为raw的text无header
	if requestMap["raw"].(string) != "text" {
		var headerStr string
		for h, v := range requestMap["headers"].(map[string]interface{}) {
			hv := h + ":" + v.(string)
			// 避免POST设置错误的content-type, 统一不设置
			if strings.Contains(write2File, "content-type") {
				contentTypeString = `  -H '` + hv + `' \`
				continue
			}
			headerStr += `  -H '` + hv + `' \`
			headerStr += "\n"
		}
		write2File += headerStr
	}
	if requestMap["method"].(string) == "GET" {
		write2File += contentTypeString
		write2File += "\n"
	}
	// 构造POST
	if requestMap["method"].(string) == "POST" {
		// form-data
		if requestMap["body"].(string) == "form-data" {
			var formData string
			for fd, v := range requestMap["form-data"].(map[string]interface{}) {
				fdv := fd +":"+ v.(string)
				formData += `  -F '`+fdv+`' \`
				formData += "\n"
			}
			write2File += formData
		}

		// x-www-form-urlencoded
		if requestMap["body"].(string) == "x-www-form-urlencoded" {
			write2File += `  -H 'Content-Type: application/x-www-form-urlencoded' \`
			write2File += "\n"
			var xWwwFormUrlencodedString string
			for x, v := range requestMap["x-www-form-urlencoded"].(map[string]interface{}) {
				xWwwFormUrlencodedString += x + "=" + v.(string) + "&"
			}
			xWwwFormUrlencodedString = strings.Trim(xWwwFormUrlencodedString, "&")
			write2File += `  -d '`+xWwwFormUrlencodedString+`'`
			write2File += "\n"
		}

		// raw
		if requestMap["body"].(string) == "raw" {
			var raw string
			if requestMap["raw-type"].(string) == "app-json" {
				write2File += `  -H 'Content-Type: application/json' \`
				write2File += "\n"
				raw += `  -d '` + requestMap["raw"].(string) + `'`
			} else {
				if requestMap["raw-type"].(string) == "text-plain" {
					write2File += `  -H 'Content-Type: application/text-plain' \`
				}
				if requestMap["raw-type"].(string) == "text-xml" {
					write2File += `  -H 'Content-Type: text/xml' \`
				}
				if requestMap["raw-type"].(string) == "text-html" {
					write2File += `  -H 'Content-Type: text/html' \`
				}
				if requestMap["raw-type"].(string) == "app-js" {
					write2File += `  -H 'Content-Type: application/javascript' \`
				}
				if requestMap["raw-type"].(string) == "app-xml" {
					write2File += `  -H 'Content-Type: application/xml' \`
				}
				if requestMap["raw-type"].(string) != "text" {
					write2File += "\n"
				}
				raw += `  -d ` + requestMap["raw"].(string) + ``
			}
			write2File += raw
		}
	}
	// GET 请求会直接来到这来, 写入文件
	writeRsp := Write2File(write2File)
	if _, ok := writeRsp["status"]; ok && writeRsp["status"] == "success" {
		global.RequestParams = requestMap
		global.RequestUri = writeRsp["message"]
		return map[string]string{
			"status": "success",
			"message": global.RequestUri,
		}
	}

	return map[string]string{
		"status": "fail",
		"message": "system error",
	}
}

// Ping URL
func Ping(uri string) bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(uri)
	if err != nil {
		log.Println("client ping error: ", err.Error())
		return false
	}
	defer resp.Body.Close()
	return true
}

// Write2File 写入文件
func Write2File(writeString string) map[string]string {
	content := []byte(writeString)
	fileId := Uuid()
	if fileId == "false" {
		return map[string]string{
			"status": "fail",
			"message": "uuid 生成失败",
		}
	}
	fileName := "curl/" + fileId + ".txt"
	writeErr := ioutil.WriteFile(fileName, content, 0644)
	if writeErr != nil {
		log.Println("cURL 写入Err:", writeErr)
		return map[string]string{
			"status": "fail",
			"message": writeErr.Error(),
		}
	}
	//log.Println("fileName ", fileName)
	return map[string]string{
		"status": "success",
		"message": fileName,
	}
}

// Uuid 生成uuid
func Uuid() string {
	fileId, uuidErr := uuid.NewV4()
	if uuidErr != nil {
		log.Println("uuid 生成Err:", uuidErr)
		return "false"
	}
	return fileId.String()
}

// OutputResult 输出到页面的信息
func OutputResult(print string, uuid string) {
	if write, exist := global.BufferMap[uuid]; exist {
		write.WriteString(print)
		global.BufferMap[uuid] = write
		return
	}
	var writeBuffer bytes.Buffer
	writeBuffer.WriteString(print)
	global.BufferMap[uuid] = writeBuffer
}

// HttpGet 发送get请求
func HttpGet(requestParams map[string]interface{}) map[string]interface{} {
	requestGetTime := time.Now()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req) // 用完需要释放资源
	req.Header.SetMethod("GET")
	req.SetRequestURI(SetUri(requestParams))
	SetHeader(req, requestParams)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源

	if err := fasthttp.Do(req, resp); err != nil {
		log.Println("请求失败:", err.Error())
		return map[string]interface{}{
			"status":false,
			"time": time.Since(requestGetTime),
			"message": "fail",
		}
	}

	return map[string]interface{}{
		"status":true,
		"time":time.Since(requestGetTime),
		"message": "success",
	}
}

// HttpPost 发送Post请求
func HttpPost(requestParams map[string]interface{}) map[string]interface{} {
	requestGetTime := time.Now()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req) // 用完需要释放资源

	// 默认是application/x-www-form-urlencoded
	if SetContentType(requestParams) != "" {
		req.Header.SetContentType(SetContentType(requestParams))
	}

	req.Header.SetMethod("POST")
	req.SetRequestURI(SetUri(requestParams))
	SetHeader(req, requestParams)
	var requestBody []byte
	if requestParams["body"].(string) == global.FormData {
		SetFormData(req, requestParams)
		requestBody = []byte(MapToJson(requestParams["form-data"].(map[string]interface{})))
	}

	if requestParams["body"].(string) == global.XWwwFormUrlencoded {
		requestBody = []byte(MapToJson(requestParams["x-www-form-urlencoded"].(map[string]interface{})))
	}
	log.Println(requestParams["raw"].(map[string]interface{}))
	if requestParams["body"].(string) == global.Raw {
		requestBody = []byte(MapToJson(requestParams["raw"].(map[string]interface{})))
	}
	req.SetBody(requestBody)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	if err := fasthttp.Do(req, resp); err != nil {
		log.Println("请求失败:", err.Error())
		return map[string]interface{}{
			"status":false,
			"time": time.Since(requestGetTime),
			"message": "fail",
		}
	}

	return map[string]interface{}{
		"status":true,
		"time":time.Since(requestGetTime),
		"message": "success",
	}
}

func SetUri(requestParams map[string]interface{}) string {
	var paramsString string
	uri := requestParams["uri"].(string)
	for p, v := range requestParams["params"].(map[string]interface{}) {
		paramsString += p + "=" + v.(string) + "&"
	}
	if paramsString != "" {
		paramsString = strings.Trim(paramsString, "&")
		if strings.Contains(requestParams["uri"].(string), "?") {
			uri = requestParams["uri"].(string) + `&` + paramsString
		} else {
			uri = requestParams["uri"].(string) + `?` + paramsString
		}
	}
	return uri
}

// SetHeader 设置header
func SetHeader(req *fasthttp.Request, requestParams map[string]interface{}) *fasthttp.Request{
	for h, v := range requestParams["headers"].(map[string]interface{}) {
		req.Header.Set(h, v.(string))
	}
	return req
}

// SetFormData 设置form-data参数
func SetFormData(req *fasthttp.Request, requestParams map[string]interface{}) *fasthttp.Request{
	for fd, v := range requestParams["form-data"].(map[string]interface{}) {
		req.PostArgs().Add(fd, v.(string))
	}
	return req
}

// SetContentType 设置请求头
func SetContentType(requestParams map[string]interface{})  string {
	if requestParams["body"].(string) == global.FormData {
		return global.AppJson
	}

	if requestParams["body"].(string) == global.XWwwFormUrlencoded {
		return global.AppXWwwFormUrlencoded
	}

	if requestParams["body"].(string) == global.Raw {
		if requestParams["raw-type"].(string) == global.Json {
			return global.AppJson
		}

		if requestParams["raw-type"].(string) == global.Plain {
			return global.AppTextPlain
		}

		if requestParams["raw-type"].(string) == global.XmlTX {
			return global.TextXml
		}
		if requestParams["raw-type"].(string) == global.XmlAX {
			return global.AppTextXml
		}
		if requestParams["raw-type"].(string) == global.Javascript {
			return global.AppJs
		}
		if requestParams["raw-type"].(string) == global.Html {
			return global.TextHtml
		}

		if requestParams["raw-type"].(string) == global.Text {
			return global.AppJson
		}
	}
	return global.AppJson
}

// MapToJson map转为json字符串
func MapToJson(m map[string]interface{}) string {
	m2Json , _ := json.Marshal(m)
	map2String := string(m2Json)
	return map2String
}

// SetCompleteException 是否完成
func SetCompleteException(e *global.Exception) {
	e.CompleteException = true
}

// SetExeException 执行异常
func SetExeException(e *global.Exception) {
	e.ExeException = true
}
