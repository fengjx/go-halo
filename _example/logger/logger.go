package main

import (
	"sync"

	"github.com/fengjx/go-halo/logger"
)

func main() {
	log := logger.NewConsole()
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
