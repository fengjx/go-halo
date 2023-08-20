package logger

import (
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
