package event

import (
	"github.com/fengjx/go-halo/worker"
	"sync"
	"time"
)

type eventHandle func(msg interface{})

var eventHandles = make(map[Topic][]eventHandle)

type Topic string

var lock sync.Mutex
var workerPool = worker.New("event-pool", worker.WithCapacity(5000), worker.WithSubmitTimeout(time.Millisecond*500))

// Subscribe 订阅事件处理
func Subscribe(topic Topic, handle eventHandle) {
	lock.Lock()
	handles := eventHandles[topic]
	handles = append(handles, handle)
	eventHandles[topic] = handles
	lock.Unlock()
}

func Publish(topic Topic, msg interface{}) {
	handles := eventHandles[topic]
	for _, fun := range handles {
		if fun == nil {
			continue
		}
		// 这一步很重要，不要直接使用fun变量，闭包会持有外部变量引用，下一个循环fun会指向其他 handle 导致执行错误
		// 后续如果重构需要注意别改错了
		task := fun
		err := workerPool.Submit(func() {
			task(msg)
		})
		if err != nil {
			// todo print log
		}
	}
}

func Quit() {
	workerPool.Release()
}
