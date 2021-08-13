package global

const (
	RoutineCountTotal = 50 // 限制线程数

	GET				  = "GET" // 请求类型
	POST			  = "POST"

	AppJs	      	  = "application/javascript" // content-type
	AppJson			  = "application/json"
	AppTextXml	      = "application/xml"
	AppTextPlain	  = "application/text-plain"
	TextXml	          = "text/xml"
	TextHtml	      = "text/html"
	AppXWwwFormUrlencoded= "application/x-www-form-urlencoded"

	Raw  			  = "raw" // 表单类型
	FormData		  = "form-data"
	XWwwFormUrlencoded= "x-www-form-urlencoded"

	Plain			  = "text-plain" // raw类型
	Json			  = "app-json"
	Javascript		  = "app-js"
	XmlAX			  = "app-xml"
	XmlTX			  = "text-xml"
	Html			  = "text-html"
	Text		  	  = "text"
)
