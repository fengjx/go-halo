package logger

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/fengjx/go-halo/halo"
)

var fakeCurrentTime = time.Now()

func fakeTime() time.Time {
	return fakeCurrentTime
}

// makeFakeTime 日期便宜
func makeFakeTime(d time.Duration) {
	fakeCurrentTime = fakeCurrentTime.Add(d)
}

func TestLogLevel(t *testing.T) {
	log := NewConsole()
	log.Debug("debug msg")
	log.Info("info msg")
	log.Warn("warn msg")
	log.Error("error msg")
}

func TestConsole(t *testing.T) {
	log := NewConsole()
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
	log1 := New(&Options{
		LogFile: "./logs/1.log",
	})
	log2 := New(&Options{
		LogFile: "./logs/2.log",
	})
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
	log := NewConsole()
	log.Info("before with")
	log = log.With(zap.Int64("goid", halo.GetGoID()))
	log.Info("after with goid")
	log = log.With(zap.String("uid", "1000"))
	log.Info("after with uid")
}

func TestRotate(t *testing.T) {
	currentTime = fakeTime

	logFilepath := "./logs/rotate1.log"
	log := New(&Options{
		LogFile: logFilepath,
	})
	log.Info("test log")

	makeFakeTime(time.Hour * 24)

	log.Info("test log2")
	log.Info("test log3")
	log.Info("test log4")
	log.Flush()
}

func TestThin(t *testing.T) {
	log := New(&Options{
		Thin:    true,
		LogFile: "./logs/thin.log",
	})
	log.Info("", zap.String("foo", "bar"))
}
