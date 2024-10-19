package errs

import "log"

// Recover recover 处理，打印堆栈
// 直接defer errs.Recover() 而不能defer func(){errs.Recover()}
func Recover() {
	RecoverFunc(func(err any, stack *Stack) {
		log.Printf("panic: %s %+v\r\n", err, stack)
	})
}

func RecoverFunc(fn func(any, *Stack)) {
	err := recover()
	if err == nil {
		return
	}
	stack := Callers(2, cMaxStackDepth)
	fn(err, stack)
}
