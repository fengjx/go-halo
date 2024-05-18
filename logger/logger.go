package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	backupDayFormat = "20060102"
)

// currentTime 时间获取
// 参考 lumberjack 可以替换这个实现来做一些 go test，例如按日期分割日志
var currentTime = time.Now

type Logger interface {
	With(fields ...zap.Field) Logger

	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	DPanic(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	DPanicf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Flush()

	SetLevel(level Level)
}

type logger struct {
	level Level
	log   *zap.Logger
}

// New 创建 Logger
// maxSizeMB 如果不希望按日志大小切割，则传 0
func New(logLevel Level, logFile string, maxSizeMB int, maxDays int, opts ...zap.Option) Logger {
	jl := &lumberjack.Logger{
		Filename:  logFile,
		MaxSize:   maxSizeMB,
		MaxAge:    maxDays,
		LocalTime: true,
	}
	rw := &rotateWriter{
		Logger: jl,
	}
	fstat, err := os.Stat(logFile)
	if err == nil && fstat.Size() > 0 {
		// 记录文件最后修改位置
		rw.date = fstat.ModTime().Format(backupDayFormat)
	}
	w := zapcore.AddSync(rw)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.FunctionKey = "fn"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zap.NewAtomicLevelAt(zapcore.Level(logLevel)),
	)
	l := zap.New(core, zap.AddCaller())
	return newWithZap(l, opts...)
}

func NewConsole(opts ...zap.Option) Logger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.FunctionKey = "fn"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.Level.SetLevel(zapcore.DebugLevel)

	l, _ := config.Build()
	return newWithZap(l, opts...)
}

func newWithZap(l *zap.Logger, opts ...zap.Option) Logger {
	options := []zap.Option{
		zap.AddStacktrace(zap.PanicLevel),
	}
	options = append(options, opts...)
	l = l.WithOptions(options...)
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

func (l *logger) DPanic(format string, fields ...zap.Field) {
	if !l.checkLevel(DPanicLevel) {
		return
	}
	l.log.DPanic(format, fields...)
}

func (l *logger) Panic(msg string, fields ...zap.Field) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.log.Panic(msg, fields...)
}

func (l *logger) Fatal(format string, fields ...zap.Field) {
	if !l.checkLevel(FatalLevel) {
		return
	}
	l.log.Fatal(format, fields...)
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

func (l *logger) DPanicf(format string, args ...interface{}) {
	if !l.checkLevel(DPanicLevel) {
		return
	}
	l.log.DPanic(getMessage(format, args))
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if !l.checkLevel(PanicLevel) {
		return
	}
	l.log.Panic(getMessage(format, args))
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	if !l.checkLevel(FatalLevel) {
		return
	}
	l.log.Fatal(getMessage(format, args))
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
	var level Level
	switch strings.ToLower(logLevel) {
	case "fatal":
		level = FatalLevel
	case "panic":
		level = PanicLevel
	case "dpanic":
		level = DPanicLevel
	case "error":
		level = ErrorLevel
	case "warn":
		level = WarnLevel
	case "info":
		level = InfoLevel
	case "debug":
		level = DebugLevel
	default:
		level = InfoLevel
	}
	return level
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

func backupDayName(name string, local bool) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := currentTime()
	if !local {
		t = t.UTC()
	}

	date := t.Format(backupDayFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, date, ext))
}
