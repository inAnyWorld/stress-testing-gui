package global

// Exception 执行异常
type Exception struct {
	CompleteException bool // CompleteException 执行结果
	ExeException bool // ExeException 检查是否执行异常
	FrontException bool // FrontException 前置条件
}

