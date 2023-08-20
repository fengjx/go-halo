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
		zapcore.Level(GetLogLevel(logLevel)),
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

func (l *logger) Debug(args ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.With(nil).Debug(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.With(nil).Debugf(format, args...)
}

func (l *logger) Info(args ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.With(nil).Info(args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.With(nil).Infof(format, args...)
}

func (l *logger) Warn(args ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.With(nil).Warn(args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.With(nil).Warnf(format, args...)
}

func (l *logger) Error(args ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.With(nil).Error(args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.With(nil).Errorf(format, args...)
}

func (l *logger) Panic(args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.With(nil).Panic(args...)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.With(nil).Panicf(format, args...)
}

func (l *logger) Flush() {
	err := l.Sync()
	if err != nil {
		log.Printf("log flush err - %v \n", err)
	}
}

func (l *logger) checkLevel(lv Level) bool {
	return Level(l.Level()) >= lv
}

type Level zapcore.Level

var (
	DebugLevel = Level(zapcore.DebugLevel)
	// InfoLevel is the default logging priority.
	InfoLevel = Level(zapcore.InfoLevel)
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = Level(zapcore.WarnLevel)
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = Level(zapcore.ErrorLevel)
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = Level(zapcore.DPanicLevel)
	// PanicLevel logs a message, then panics.
	PanicLevel = Level(zapcore.PanicLevel)
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = Level(zapcore.FatalLevel)
)

func GetLogLevel(logLevel string) Level {
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
	return Level(level)
}
