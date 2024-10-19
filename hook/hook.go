package hook

import (
	"sort"
	"sync"

	"github.com/samber/lo"
)

// order 越小优先级越高
type hookFun struct {
	handler func()
	order   int
}

var hookMap map[string][]hookFun
var hookMapLock sync.Mutex

func init() {
	hookMap = make(map[string][]hookFun)
}

// AddHook 注册回调
func AddHook(name string, order int, handler func()) {
	hookMapLock.Lock()
	defer hookMapLock.Unlock()
	hookMap[name] = append(hookMap[name], hookFun{
		handler: handler,
		order:   order,
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
		}()
	}
}
