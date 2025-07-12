package errs

import (
	"errors"
	"fmt"
	"io"
)

// withStack 错误包装结构体,支持携带堆栈信息，
// 只包装根因异常即可，因为堆栈信息只需要携带一次
type withStack struct {
	error
	*Stack
}

// Cause 返回错误根因
func (w *withStack) Cause() error {
	return w.error
}

// Unwrap 适配 Go 1.13 标准库接口
func (w *withStack) Unwrap() error {
	return w.error
}

func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", w.Cause())
			w.Stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

// WithStack 给错误添加堆栈信息
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &withStack{
		err,
		callers(),
	}
}

// Wrap 包装错误信息，并且保存堆栈信息
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	// %w 可以自动生成一个可以 Unwrap 的 error
	err = fmt.Errorf("%s: %w", msg, err)
	var s *Stack
	var w withStack
	if errors.As(err, &w) {
		// 调用栈只需要获取一次，如果本身 err 就是 withStack 了，则不重复获取
		s = w.Stack
	} else {
		s = callers()
	}
	return &withStack{
		err,
		s,
	}
}

// Wrapf 包装错误信息，并且保存堆栈信息，错误信息支持格式化
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	// %w 可以自动生成一个可以 Unwrap 的 error
	err = fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
	var s *Stack
	var w withStack
	if errors.As(err, &w) {
		// 调用栈只需要获取一次，如果本身 err 就是 withStack 了，则不重复获取
		s = w.Stack
	} else {
		s = callers()
	}
	return &withStack{
		err,
		s,
	}
}

// Cause 返回根因 error
// 查找第一个没有实现 causer 的 error，认为是根因
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func callers() *Stack {
	return Callers(4, cMaxStackDepth)
}
