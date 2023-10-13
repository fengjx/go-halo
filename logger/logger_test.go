package logger

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConsole(t *testing.T) {
	log := NewConsole()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		log.Info("log1")
		log.Infof("log1f %d", time.Now().Unix())
		log.Infow("log1w", "ts", time.Now().Unix())
		wg.Done()
	}()

	go func() {
		log.Info("log2")
		log.Infof("log2f %d", time.Now().Unix())
		log.Infow("log2w", "ts", time.Now().Unix())
		wg.Done()
	}()
	wg.Wait()
}

func TestFile(t *testing.T) {
	log1 := New(DebugLevel, "./logs/1.log", 100, 3, WithTrace())
	log2 := New(DebugLevel, "./logs/2.log", 100, 3, WithTrace())
	wg := &sync.WaitGroup{}
	count := 1000
	wg.Add(count)
	for i := 0; i < count; i++ {
		id := i
		go func() {
			log1.SetLocalTraceID(fmt.Sprintf("trace-%d", id))
			log2.SetLocalTraceID(fmt.Sprintf("trace-%d", id))

			log1.Info(fmt.Sprintf("log-%d", id))
			log2.Infow(fmt.Sprintf("log-%d", id))
			wg.Done()
		}()
	}
	wg.Wait()
}
