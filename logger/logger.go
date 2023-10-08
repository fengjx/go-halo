package logger

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/petermattis/goid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	TraceIDKey = "X-Trace-ID"
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

	SetLocalTraceID(traceID string)
	SetLocalContext(ctx context.Context)
	GetLocalContext() context.Context
	RemoveLocalContext()
}

type logger struct {
	adapter   *zap.SugaredLogger
	traceMap  sync.Map
	openTrace bool
}

type Options struct {
	openTrace bool
}

type Option func(*Options)

func WithTrace() Option {
	return func(ops *Options) {
		ops.openTrace = true
	}
}

func New(logLevel string, logFile string, maxSizeMB int, maxDays int, opts ...Option) Logger {
	ops := &Options{}
	for _, opt := range opts {
		opt(ops)
	}
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSizeMB,
		MaxBackups: 3,
		MaxAge:     maxDays,
	})
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zapcore.Level(GetLogLevel(logLevel)),
	)
	l := zap.New(core)
	return NewWithZap(l, ops.openTrace)
}

func NewConsole() Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.Level.SetLevel(zapcore.DebugLevel)

	l, _ := config.Build()
	return NewWithZap(l, true)
}

func NewWithZap(l *zap.Logger, openTrace bool) Logger {
	return &logger{
		adapter:   l.Sugar(),
		openTrace: openTrace,
	}
}

func (l *logger) With(ctx context.Context, args ...interface{}) Logger {
	gid := goid.Get()
	args = append(args, zap.Int64("goid", gid))
	if l.openTrace && ctx == nil {
		ctx = l.GetLocalContext()
	}
	if ctx != nil {
		if id, ok := ctx.Value(TraceIDKey).(string); ok {
			args = append(args, zap.String("traceId", id))
		}
	}
	if len(args) > 0 {
		return &logger{adapter: l.adapter.With(args...)}
	}
	return l
}

func (l *logger) Debug(args ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Debug(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Debugf(format, args...)
}

func (l *logger) Info(args ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Info(args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Infof(format, args...)
}

func (l *logger) Warn(args ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Warn(args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Warnf(format, args...)
}

func (l *logger) Error(args ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Error(args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Errorf(format, args...)
}

func (l *logger) Panic(args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Panic(args...)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.With(nil).(*logger).adapter.Panicf(format, args...)
}

func (l *logger) Flush() {
	err := l.adapter.Sync()
	if err != nil {
		log.Printf("log flush err - %v \n", err)
	}
}

func (l *logger) SetLocalTraceID(traceID string) {
	ctx := l.GetLocalContext()
	if ctx == nil {
		ctx = context.Background()
	}
	l.SetLocalContext(context.WithValue(ctx, TraceIDKey, traceID))
}

func (l *logger) SetLocalContext(ctx context.Context) {
	gid := goid.Get()
	l.traceMap.Store(gid, ctx)
}

func (l *logger) GetLocalContext() context.Context {
	gid := goid.Get()
	if value, ok := l.traceMap.Load(gid); ok {
		return value.(context.Context)
	}
	return nil
}

func (l *logger) RemoveLocalContext() {
	gid := goid.Get()
	l.traceMap.Delete(gid)
}

func (l *logger) checkLevel(lv Level) bool {
	return Level(l.adapter.Level()) <= lv
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
	switch strings.ToLower(logLevel) {
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
