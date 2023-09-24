package hook

import (
	"sort"
	"sync"
	"time"

	"github.com/samber/lo"
)

type Options struct {
	cycle time.Duration
}

type Option func(*Options)

func WithDuration(cycle time.Duration) Option {
	return func(opts *Options) {
		opts.cycle = cycle
	}
}

// order 越小优先级越高
type hookFun struct {
	handler func()
	order   int
	cycle   time.Duration
}

var hookMap map[string][]hookFun
var hookMapLock sync.Mutex

func AddCustomStartHook(name string, handler func(), order int, opts ...Option) {
	hookMapLock.Lock()
	defer hookMapLock.Unlock()
	opt := &Options{}
	for _, item := range opts {
		item(opt)
	}
	hookMap[name] = append(hookMap[name], hookFun{
		handler: handler,
		order:   order,
		cycle:   opt.cycle,
	})
}

func DoCustomHooks(name string) {
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
			if f.cycle > 0 {
				tk := time.NewTicker(f.cycle)
				for range tk.C {
					f.handler()
				}
			}
		}()
	}
}
