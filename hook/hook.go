package hook

import (
	"sort"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/fengjx/go-halo/halo"
)

type Options struct {
	interval time.Duration
}

type Option func(*Options)

// WithInterval 定时执行
func WithInterval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.interval = interval
	}
}

// order 越小优先级越高
type hookFun struct {
	handler  func()
	order    int
	interval time.Duration
}

var hookMap map[string][]hookFun
var hookMapLock sync.Mutex

func init() {
	hookMap = make(map[string][]hookFun)
}

// AddHook 注册回调
func AddHook(name string, order int, handler func(), opts ...Option) {
	hookMapLock.Lock()
	defer hookMapLock.Unlock()
	opt := &Options{}
	for _, item := range opts {
		item(opt)
	}
	hookMap[name] = append(hookMap[name], hookFun{
		handler:  handler,
		order:    order,
		interval: opt.interval,
	})
}

// DoHooks 执行回调
func DoHooks(name string) {
	doHooks(hookMap[name])
}

func doHooks(hookFns []hookFun) {
	hookGroup := make(map[int][]hookFun)
	for _, hook := range hookFns {
		fnList := hookGroup[hook.order]
		hookGroup[hook.order] = append(fnList, hook)
	}
	keys := lo.Keys[int, []hookFun](hookGroup)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, order := range keys {
		hooks := hookGroup[order]
		group := &sync.WaitGroup{}
		group.Add(len(hooks))
		execHooks(hooks, group)
		group.Wait()
	}
}

func execHooks(hooks []hookFun, wg *sync.WaitGroup) {
	for _, fn := range hooks {
		f := fn
		go func() {
			defer wg.Done()
			f.handler()
			if f.interval > 0 {
				go func() {
					defer halo.Recover()
					tk := time.NewTicker(f.interval)
					for range tk.C {
						f.handler()
					}
				}()
			}
		}()
	}
}
