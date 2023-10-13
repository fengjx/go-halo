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
	With(fields ...zap.Field) Logger

	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})

	Flush()

	SetLevel(level Level)
	SetLocalTraceID(traceID string)
	SetLocalContext(ctx context.Context)
	GetLocalContext() context.Context
	RemoveLocalContext()
}

type logger struct {
	level     Level
	log       *zap.Logger
	sugar     *zap.SugaredLogger
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

func New(logLevel Level, logFile string, maxSizeMB int, maxDays int, opts ...Option) Logger {
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
	encoderConfig.FunctionKey = "fn"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zapcore.Level(logLevel),
	)
	l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel))
	return NewWithZap(l, ops.openTrace)
}

func NewConsole() Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.FunctionKey = "fn"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.Level.SetLevel(zapcore.DebugLevel)

	l, _ := config.Build()
	return NewWithZap(l, false)
}

func NewWithZap(l *zap.Logger, openTrace bool) Logger {
	l = l.WithOptions(zap.AddCallerSkip(1))
	return &logger{
		level:     Level(l.Level()),
		log:       l,
		sugar:     l.Sugar(),
		openTrace: openTrace,
	}
}

func (l *logger) With(fields ...zap.Field) Logger {
	if len(fields) > 0 {
		return &logger{
			log:       l.log.With(fields...),
			sugar:     l.sugar,
			openTrace: l.openTrace,
		}
	}
	return l
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.log.Debug(msg, l.addFields(fields)...)
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.log.Info(msg, l.addFields(fields)...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.log.Warn(msg, l.addFields(fields)...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.log.Error(msg, l.addFields(fields)...)
}

func (l *logger) Panic(msg string, fields ...zap.Field) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.log.Panic(msg, l.addFields(fields)...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.sugar.With().Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.sugar.With().Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.sugar.With().Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.sugar.With().Errorf(format, args...)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.sugar.With().Panicf(format, args...)
}

func (l *logger) Debugw(format string, keysAndValues ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.sugar.With().Debugw(format, l.addArgs(keysAndValues)...)
}

func (l *logger) Infow(format string, keysAndValues ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.sugar.With().Infow(format, l.addArgs(keysAndValues)...)
}

func (l *logger) Warnw(format string, keysAndValues ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.sugar.With().Warnw(format, l.addArgs(keysAndValues)...)
}

func (l *logger) Errorw(format string, keysAndValues ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.sugar.With().Errorw(format, l.addArgs(keysAndValues)...)
}

func (l *logger) Panicw(format string, keysAndValues ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.sugar.With().Panicw(format, l.addArgs(keysAndValues)...)
}

func (l *logger) SetLevel(level Level) {
	l.level = level
}

func (l *logger) Flush() {
	err := l.log.Sync()
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
	return Level(l.log.Level()) <= lv
}

func (l *logger) addFields(fields []zap.Field) (fs []zap.Field) {
	fs = []zap.Field{zap.Int64("goid", goid.Get())}
	if l.openTrace {
		if id, ok := l.GetLocalContext().Value(TraceIDKey).(string); ok {
			fields = append(fields, zap.String("traceId", id))
		}
	}
	fs = append(fs, fields...)
	return
}

func (l *logger) addArgs(args []interface{}) (retArgs []interface{}) {
	retArgs = []interface{}{"goid", goid.Get()}
	if l.openTrace {
		if traceId, ok := l.GetLocalContext().Value(TraceIDKey).(string); ok {
			retArgs = append(retArgs, "traceId", traceId)
		}
	}
	retArgs = append(retArgs, args...)
	return
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
