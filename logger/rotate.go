package logger

import (
	"gopkg.in/natefinch/lumberjack.v2"
)

type rotateWriter struct {
	*lumberjack.Logger
	date string
}

// Write 重写 Write，支持每日切割
func (r *rotateWriter) Write(p []byte) (n int, err error) {
	d := currentTime().Format(backupDayFormat)
	if r.date != "" && r.date != d {
		_ = r.Logger.Rotate()
	}
	r.date = d
	return r.Logger.Write(p)
}
