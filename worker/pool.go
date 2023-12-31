package worker

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fengjx/go-halo/halo"
)

var (
	ErrSubmitTimeout  = errors.New("exceeded maximum capacity abd submit timeout")
	ErrWorkerReleased = errors.New("worker has released")
)

type Pool struct {
	name          string         // worker 名称，用来输出日志，便于排查问题
	capacity      int            // 最大协程数量
	submitTimeout time.Duration  // 任务提交超时时间，避免长时间阻塞，占用资源导致宕机
	active        chan struct{}  // 用来计数，控制并发
	tasks         chan Task      // 用来提交任务
	wg            sync.WaitGroup // 保证任务优雅停止
	quit          chan struct{}  // worker 停止信号
	log           Logger
}

type Task func()

// Logger is used for logging formatted messages.
type Logger interface {
	// Printf must have the same semantics as log.Printf.
	Printf(format string, args ...interface{})
}

const (
	defaultCapacity      = 100
	maxCapacity          = 10000
	defaultSubmitTimeout = time.Millisecond * 500
)

func New(name string, opts ...Option) *Pool {
	p := &Pool{
		name:  name,
		tasks: make(chan Task),
		quit:  make(chan struct{}),
	}

	for _, opt := range opts {
		opt(p)
	}

	capacity := p.capacity
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}
	p.capacity = capacity
	p.active = make(chan struct{}, capacity)

	timeout := p.submitTimeout
	if timeout == 0 {
		timeout = defaultSubmitTimeout
	}
	p.submitTimeout = timeout

	if p.log == nil {
		p.log = Logger(log.New(os.Stderr, fmt.Sprintf("[worker-%s]: ", p.name), log.LstdFlags|log.Lmsgprefix|log.Lmicroseconds))
	}

	go p.run()
	return p
}

func (p *Pool) doTask(t Task) {
	p.wg.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := halo.Stack(10)
				p.log.Printf("recover panic[%s] and exit - %s\n", err, stack)
			}
			p.wg.Done()
			<-p.active
		}()
		t()
	}()
}

func (p *Pool) run() {
	for {
		select {
		case <-p.quit:
			p.log.Printf("exit")
			<-p.active
			return
		case t := <-p.tasks:
			p.doTask(t)
		}
	}
}

func (p *Pool) Submit(t Task) error {
	select {
	case <-p.quit:
		return ErrWorkerReleased
	case p.active <- struct{}{}:
		p.tasks <- t
		return nil
	case <-time.After(p.submitTimeout):
		p.log.Printf("submit worker task timeout")
		return ErrSubmitTimeout
	}
}

func (p *Pool) Release() {
	close(p.quit)
	// 等待所有任务执行完成
	p.wg.Wait()
	p.log.Printf("release")
}
