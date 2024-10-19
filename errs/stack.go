package errs

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const cMaxStackDepth = 16 // 默认获取调用栈的深度

var (
	mFramCache sync.Map // uintptr->*Frame
)

type Frame struct {
	Frame    runtime.Frame
	File     string
	Function string
	FuncName string
}

// name returns the name of this function, if known.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.Frame.PC)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Function)
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.File)
		default:
			io.WriteString(s, path.Base(f.File))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Frame.Line))
	case 'n':
		io.WriteString(s, f.FuncName)
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// MarshalText formats a stacktrace Frame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f Frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == "unknown" {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.File, f.Frame.Line)), nil
}

// Stack represents a stack of program counters.
type Stack []uintptr

func (s *Stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := getFrame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

// StackTrace 转换为 StackTrace
func (s *Stack) StackTrace() StackTrace {
	f := make([]*Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = getFrame((*s)[i])
	}
	return f
}

// StackTrace 调用堆栈
type StackTrace []*Frame

func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", st)
		default:
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

// formatSlice will format this StackTrace into the given buffer as a slice of
// Frame, only valid when called with '%s' or '%v'.
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}

// Callers 获取调用站
func Callers(skip, depth int) *Stack {
	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip, pcs[:])
	var st Stack = pcs[0:n]
	return &st
}

// funcname 获取函数名
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// getFrame 根据pc解析调用信息
func getFrame(pc uintptr) *Frame {
	if fr, ok := mFramCache.Load(pc); ok {
		return fr.(*Frame)
	}

	frs := runtime.CallersFrames([]uintptr{pc})
	if nil != frs {
		fr, _ := frs.Next()
		frame := buildFrame(fr)
		mFramCache.Store(pc, frame)
		return frame
	}
	mFramCache.Store(pc, nil)
	return nil
}

func buildFrame(fr runtime.Frame) *Frame {
	frame := &Frame{Frame: fr, File: fr.File, Function: fr.Function, FuncName: funcname(fr.Function)}

	// 目前函数名是带module路径的，文件名是带绝对路径的，这里先尝试把文件名的前缀部分去掉。
	// 比如函数名是gitit.cc/xxxx 文件名是 /x/y/z/gitit.cc/xxxx，则尝试查找文件名里面匹配函数名第一部分的地方，然后截断
	if idx := strings.IndexByte(frame.Function, '/'); idx > 0 {
		if idx = strings.Index(frame.File, frame.Function[:idx]); idx > 0 {
			frame.File = frame.File[idx:]
		}
	}
	return frame
}
