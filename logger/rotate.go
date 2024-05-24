package logger

import (
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

type rotateWriter struct {
	sync.Mutex
	*lumberjack.Logger
	date string
}

// Write 重写 Write，支持每日切割
func (r *rotateWriter) Write(p []byte) (n int, err error) {
	d := currentTime().Format(backupDayFormat)
	if r.date != "" && r.date != d {
		r.rotate(d)
	}
	return r.Logger.Write(p)
}

func (r *rotateWriter) rotate(date string) {
	r.Lock()
	defer r.Unlock()
	// double check
	if r.date == date {
		return
	}
	r.date = date
	_ = r.Logger.Rotate()
}
