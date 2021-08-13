package gohttp

import (
	"fmt"
	"go-stress-testing/global"
	"go-stress-testing/helper"
	"strconv"
	"sync"
	"time"
)

// HttpRequestGet 发起Http Get请求
func HttpRequestGet(requestParams map[string]interface{}) {
	gLimit := NewGLimit(global.RoutineCountTotal)
	wg := &sync.WaitGroup{}
	var lr LimitRate
	qps, _ := strconv.Atoi(requestParams["qps"].(string))
	//rateFloat64, _ := strconv.ParseFloat(requestParams["qps"].(string),64)
	//rateInt := helper.Round(rateFloat64 * 1.2)
	//lr.SetRate(rateInt)
	lr.SetRate(qps) // 单次执行不大于此QPS
	// 记录执行时间
	now := time.Now()
	var successNum,failNum int // 成功数, 失败数
	for i := 0; i < qps; i ++ {
		wg.Add(1)
		gLimit.Run(func() {
			if lr.Limit() {
				res := helper.HttpGet(requestParams)
				if res["status"].(bool) {
					successNum += 1
				} else {
					failNum += 1
				}
			}
			wg.Done()
		})
	}
	wg.Wait()
	// 客户端请求时间
	timestamp, _ := strconv.Atoi(requestParams["timestamp"].(string))
	continuedSecond, _ := strconv.Atoi(requestParams["time"].(string))
	clientTime := time.Unix(int64(timestamp + continuedSecond), 0)
	helper.OutputResult(
		"<tr/>" +
				"<td>"+fmt.Sprintf("%s", requestParams["qps"].(string)) +"</td>" +
				"<td>"+fmt.Sprintf("%s s", requestParams["time"].(string)) +"</td>" +
				"<td>"+fmt.Sprintf("%v",time.Since(now)) +"</td>" +
				"<td>"+fmt.Sprintf("%d", successNum) +"</td>" +
				"<td>"+fmt.Sprintf("%d", failNum) +"</td>" +
			"<tr/>", requestParams["uuid"].(string))
	if clientTime.After(time.Now()) {
		HttpRequestGet(requestParams)
	}
	helper.OutputResult("</table>", requestParams["uuid"].(string))
	return
}

// HttpRequestPost 发起Http Post请求
func HttpRequestPost(requestParams map[string]interface{}) {
	gLimit := NewGLimit(global.RoutineCountTotal)
	wg := &sync.WaitGroup{}
	var lr LimitRate
	qps, _ := strconv.Atoi(requestParams["qps"].(string))
	lr.SetRate(qps) // 单次执行不大于此QPS
	// 记录执行时间
	now := time.Now()
	var successNum,failNum int // 成功数, 失败数
	for i := 0; i < qps; i ++ {
		wg.Add(1)
		gLimit.Run(func() {
			if lr.Limit() {
				res := helper.HttpPost(requestParams)
				if res["status"].(bool) {
					successNum += 1
				} else {
					failNum += 1
				}
			}
			wg.Done()
		})
	}
	wg.Wait()
	// 客户端请求时间
	timestamp, _ := strconv.Atoi(requestParams["timestamp"].(string))
	continuedSecond, _ := strconv.Atoi(requestParams["time"].(string))
	clientTime := time.Unix(int64(timestamp + continuedSecond), 0)
	helper.OutputResult(
		"<tr/>" +
			"<td>"+fmt.Sprintf("%s", requestParams["qps"].(string)) +"</td>" +
			"<td>"+fmt.Sprintf("%s s", requestParams["time"].(string)) +"</td>" +
			"<td>"+fmt.Sprintf("%v",time.Since(now)) +"</td>" +
			"<td>"+fmt.Sprintf("%d", successNum) +"</td>" +
			"<td>"+fmt.Sprintf("%d", failNum) +"</td>" +
			"<tr/>", requestParams["uuid"].(string))
	if clientTime.After(time.Now()) {
		HttpRequestPost(requestParams)
	}
	helper.OutputResult("</table>", requestParams["uuid"].(string))
	return
}