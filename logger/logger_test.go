package logger_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/fengjx/go-halo/halo"
	"github.com/fengjx/go-halo/logger"
)

func TestConsole(t *testing.T) {
	log := logger.NewConsole()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		log.Info("log1")
		log.Infof("log1f %d", time.Now().Unix())
		wg.Done()
	}()

	go func() {
		log.Info("log2")
		log.Infof("log2f %d", time.Now().Unix())
		wg.Done()
	}()
	wg.Wait()
}

func TestFile(t *testing.T) {
	log1 := logger.New(logger.DebugLevel, "./logs/1.log", 100, 3)
	log2 := logger.New(logger.DebugLevel, "./logs/2.log", 100, 3)
	wg := &sync.WaitGroup{}
	count := 1000
	wg.Add(count)
	for i := 0; i < count; i++ {
		id := i
		go func() {
			log1.Info(fmt.Sprintf("log1-%d", id))
			log2.Info(fmt.Sprintf("log2-%d", id))
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestWith(t *testing.T) {
	log := logger.NewConsole()
	log.Info("before with")
	log = log.With(zap.Int64("goid", halo.GetGoID()))
	log.Info("after with goid")
	log = log.With(zap.String("uid", "1000"))
	log.Info("after with uid")
}
