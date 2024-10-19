package errs

import "log"

var h = defaultRecoverHandle

type RecoverHandle func(any, *Stack)

func defaultRecoverHandle(err any, stack *Stack) {
	if h != nil {
		h(err, stack)
		return
	}
	log.Printf("panic: %s %+v\r\n", err, stack)
}

// Recover recover 处理，打印堆栈
// 直接defer errs.Recover() 而不能defer func(){errs.Recover()}
func Recover() {
	RecoverFunc(defaultRecoverHandle)
}

func RecoverFunc(fn RecoverHandle) {
	err := recover()
	if err == nil {
		return
	}
	stack := Callers(2, cMaxStackDepth)
	fn(err, stack)
}

// RegisterRecoverHandle 自定义的recover处理函数
func RegisterRecoverHandle(fn RecoverHandle) {
	h = fn
}
