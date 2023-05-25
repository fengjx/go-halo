package worker

import (
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	worker := New("test-worker", WithCapacity(3), WithSubmitTimeout(time.Second*3))
	for i := 0; i < 10; i++ {
		idx := i
		err := worker.Submit(func() {
			time.Sleep(time.Second * 1)
			t.Logf("task%d end", idx)
		})
		if err != nil {
			t.Log("task", i, "err:", err)
		}
	}
	worker.Release()
}

func TestTimeout(t *testing.T) {
	worker := New("timeout-worker", WithCapacity(3), WithSubmitTimeout(time.Second*1))
	for i := 0; i < 10; i++ {
		idx := i
		err := worker.Submit(func() {
			time.Sleep(time.Second * 3)
			t.Logf("task%d end", idx)
		})
		if err != nil {
			t.Log("task", i, "err:", err)
		}
	}
	t.Log("submit end")
	worker.Release()
}
