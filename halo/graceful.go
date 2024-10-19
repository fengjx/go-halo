package halo

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fengjx/go-halo/errs"
)

var (
	ErrGracefulShutdownTimeout = errors.New("graceful shutdown timeout")

	s = NewSingleton[graceful](func() *graceful {
		return newGraceful()
	})
)

type graceful struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	onStop []func() // 注册的停止回调函数
}

func newGraceful() *graceful {
	ctx, cancel := context.WithCancel(context.Background())
	g := &graceful{
		ctx:    ctx,
		cancel: cancel,
	}
	g.listenSignal()
	return g
}

func (g *graceful) gracefulRun(f func(ctx context.Context)) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer errs.Recover()
		f(g.ctx) // 将全局 context 传递给 Goroutine
	}()
}

func (g *graceful) gracefulRunSync(f func(ctx context.Context)) {
	g.wg.Add(1)
	defer g.wg.Done()
	defer errs.Recover()
	f(g.ctx) // 将全局 context 传递给 Goroutine
}

func (g *graceful) addShutdownCallback(f func()) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onStop = append(g.onStop, f)
}

func (g *graceful) shutdown() {
	g.mu.Lock()
	defer g.mu.Unlock()
	// 执行注册的停止回调
	for _, f := range g.onStop {
		f()
	}
	// 取消所有 context
	g.cancel()
}

func (g *graceful) wait(timeout time.Duration) error {
	select {
	case <-g.ctx.Done():
	}

	waitDone := make(chan struct{})

	go func() {
		g.wg.Wait() // 等待所有 Goroutine 退出
		close(waitDone)
	}()

	select {
	case <-waitDone:
		return nil
	case <-time.After(timeout):
		return ErrGracefulShutdownTimeout
	}
}

// listenSignal 启动信号监听器，捕获 SIGINT 和 SIGTERM 信号
func (g *graceful) listenSignal() {
	// 监听操作系统的信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		g.shutdown() // 收到信号后停止所有 Goroutine
	}()
}

// SetInterval 定时执行
func SetInterval(fn func(), d time.Duration) {
	g := s.Get()
	g.gracefulRun(func(ctx context.Context) {
		ticker := time.NewTicker(d)
		for {
			select {
			case <-ticker.C:
				errs.Recover()
				fn()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	})
}

// SafeGo 启动一个 goroutine，并捕获 panic
func SafeGo(f func()) {
	errs.Recover()
	f()
}

// GracefulRun 启动一个优雅停机 goroutine
func GracefulRun(f func(ctx context.Context)) {
	g := s.Get()
	g.gracefulRun(f)
}

// GracefulRunSync 同步执行一个优雅停机函数
func GracefulRunSync(f func(ctx context.Context)) {
	g := s.Get()
	g.gracefulRunSync(f)
}

// AddShutdownCallback 注册一个进程停止时的回调函数
func AddShutdownCallback(f func()) {
	g := s.Get()
	g.addShutdownCallback(f)
}

// Wait 等待所有 goroutine 退出
// 当收到 kill 信号时，最多等待 timeout 时长，超时则返回 ErrGracefulShutdownTimeout
func Wait(timeout time.Duration) error {
	g := s.Get()
	return g.wait(timeout)
}
