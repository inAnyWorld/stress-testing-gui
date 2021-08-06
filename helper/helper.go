// Package helper 帮助函数，时间、数组的通用处理
package helper

import (
	uuid "github.com/nu7hatch/gouuid"
	"go-stress-testing/global"
	"io/ioutil"
	"log"
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

// BuildCURLHandlerHelper 助手函数, 构建cURL请求体
func BuildCURLHandlerHelper(requestMap map[string]interface{}) map[string]string {
	if _, ok := requestMap["uri"]; !ok {
		return map[string]string{
			"status": "fail",
			"message": "uri empty",
		}
	}
	if !Ping(requestMap["uri"].(string)) {
		return map[string]string{
			"status": "fail",
			"message": "uri ping error",
		}
	}
	var write2File string

	// paramsString url参数
	var paramsString, contentTypeString string
	for p, v := range requestMap["params"].(map[string]interface{}) {
		paramsString += p + "=" + v.(string) + "&"
	}
	paramsString = strings.Trim(paramsString, "&")
	requestMode := "GET"
	if requestMap["method"].(string) != "" {
		requestMode = requestMap["method"].(string)
	}
	write2File = `curl -X ` + requestMode + ` \`
	write2File += "\n"

	// url
	uri := requestMap["uri"].(string)
	if paramsString != "" {
		uri = requestMap["uri"].(string) + `?` + paramsString
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
					write2File += `  -H 'Content-Type: application/text-xml' \`
				}
				if requestMap["raw-type"].(string) == "app-js" {
					write2File += `  -H 'Content-Type: application/javascript' \`
				}
				if requestMap["raw-type"].(string) == "app-xml" {
					write2File += `  -H 'Content-Type: text/xml' \`
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
		log.Println("client.Get error: ", err)
		return false
	}
	defer resp.Body.Close()
	return true
}

// Write2File 写入文件
func Write2File(writeString string) map[string]string {
	content := []byte(writeString)
	fileId, uuidErr := uuid.NewV4()
	if uuidErr!= nil {
		log.Println("uuid 生成Err:", uuidErr)
		return map[string]string{
			"status": "fail",
			"message": uuidErr.Error(),
		}
	}
	fileName := "curl/" + fileId.String() + ".txt"
	writeErr := ioutil.WriteFile(fileName, content, 0644)
	if writeErr != nil {
		log.Println("cURL 写入Err:", writeErr)
		return map[string]string{
			"status": "fail",
			"message": writeErr.Error(),
		}
	}
	log.Println("fileName ", fileName)
	return map[string]string{
		"status": "success",
		"message": fileName,
	}
}

func OutputResult(print string)  {
	global.BufferString.WriteString(print)
}