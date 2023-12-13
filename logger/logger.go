package logger

import (
	"fmt"
	"log"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

	Flush()

	SetLevel(level Level)
}

type logger struct {
	level Level
	log   *zap.Logger
}

func New(logLevel Level, logFile string, maxSizeMB int, maxDays int) Logger {
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
	return newWithZap(l)
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
	return newWithZap(l)
}

func newWithZap(l *zap.Logger) Logger {
	l = l.WithOptions(zap.AddCallerSkip(1))
	return &logger{
		level: Level(l.Level()),
		log:   l,
	}
}

func (l *logger) With(fields ...zap.Field) Logger {
	if len(fields) > 0 {
		return &logger{
			level: l.level,
			log:   l.log.With(fields...),
		}
	}
	return l
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.log.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.log.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.log.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.log.Error(msg, fields...)
}

func (l *logger) Panic(msg string, fields ...zap.Field) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.log.Panic(msg, fields...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.checkLevel(DebugLevel) {
		return
	}
	l.log.Debug(getMessage(format, args))
}

func (l *logger) Infof(format string, args ...interface{}) {
	if !l.checkLevel(InfoLevel) {
		return
	}
	l.log.Info(getMessage(format, args))
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if !l.checkLevel(WarnLevel) {
		return
	}
	l.log.Warn(getMessage(format, args))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if !l.checkLevel(ErrorLevel) {
		return
	}
	l.log.Error(getMessage(format, args))
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.log.Panic(getMessage(format, args))
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

func (l *logger) checkLevel(lv Level) bool {
	return Level(l.log.Level()) <= lv
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

func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}
