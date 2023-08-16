package logger

import (
	"context"
	"log"

	"github.com/petermattis/goid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	TraceID = "X-Trace-ID"
)

type Logger interface {
	With(ctx context.Context, args ...interface{}) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Flush()
}

type logger struct {
	*zap.SugaredLogger
}

func New(logLevel string, logFile string, maxSizeMB int, maxDays int) Logger {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSizeMB,
		MaxBackups: 3,
		MaxAge:     maxDays,
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		GetLogLevel(logLevel),
	)
	l := zap.New(core)
	return NewWithZap(l)
}

func NewConsole() Logger {
	l, _ := zap.NewDevelopment()
	return NewWithZap(l)
}

func NewWithZap(l *zap.Logger) Logger {
	return &logger{l.Sugar()}
}

func (l *logger) With(ctx context.Context, args ...interface{}) Logger {
	gid := goid.Get()
	args = append(args, zap.Int64("goid", gid))
	if ctx != nil {
		if id, ok := ctx.Value(TraceID).(string); ok {
			args = append(args, zap.String("traceId", id))
		}
	}
	if len(args) > 0 {
		return &logger{l.SugaredLogger.With(args...)}
	}
	return l
}

func (l *logger) Flush() {
	err := l.Sync()
	if err != nil {
		log.Printf("log flush err - %v \n", err)
	}
}

func GetLogLevel(logLevel string) zapcore.Level {
	var level zapcore.Level
	switch logLevel {
	case "panic":
		level = zapcore.PanicLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "error":
		level = zapcore.ErrorLevel
	case "warn":
		level = zapcore.WarnLevel
	case "info":
		level = zapcore.InfoLevel
	case "debug":
		level = zapcore.DebugLevel
	default:
		level = zapcore.InfoLevel
	}
	return level
}
