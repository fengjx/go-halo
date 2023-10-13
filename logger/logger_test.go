package logger

import (
	"fmt"
	"sync"
	"testing"
)

func TestConsole(t *testing.T) {
	log := NewConsole()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		log.Info("log1")
		wg.Done()
	}()

	go func() {
		log.Info("log2")
		wg.Done()
	}()
	wg.Wait()
}

func TestFile(t *testing.T) {
	log := New(DebugLevel, "./logs/tes.log", 100, 3, WithTrace())
	wg := &sync.WaitGroup{}
	count := 1000
	wg.Add(count)
	for i := 0; i < count; i++ {
		id := i
		go func() {
			log.SetLocalTraceID(fmt.Sprintf("trace-%d", id))
			log.Info(fmt.Sprintf("log-%d", id))
			wg.Done()
		}()
	}
	wg.Wait()
}
