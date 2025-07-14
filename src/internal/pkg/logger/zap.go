package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Field is an alias for zap.Field to avoid exposing zap directly
type Field = zap.Field

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Sync() error
	With(fields ...Field) Logger
}

// zapLogger implements Logger interface
type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

var (
	instance Logger
)

// Config holds the configuration for the logger
type Config struct {
	Level    string `mapstructure:"level"`    // debug, info, warn, error
	Format   string `mapstructure:"format"`   // json, console
	FilePath string `mapstructure:"file_path"` // path to log file
}

// Init initializes the logger with the given configuration
func Init(cfg Config) (Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.FilePath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return nil, err
		}

		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    100,    // MB
			MaxBackups: 3,      // number of backups
			MaxAge:     28,     // days
			Compress:   true,   // compress old logs
			LocalTime:  true,   // use local time
		}
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	zapLog := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	instance = &zapLogger{
		logger: zapLog,
		sugar:  zapLog.Sugar(),
	}
	return instance, nil
}

// Get returns the default logger instance
func Get() Logger {
	if instance == nil {
		// Initialize with default configuration if not initialized
		Init(Config{
			Level:  "info",
			Format: "console",
		})
	}
	return instance
}

// Sync flushes any buffered log entries
func Sync() error {
	if instance != nil {
		return instance.Sync()
	}
	return nil
}

// Implement Logger interface methods
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(fields...),
		sugar:  l.logger.With(fields...).Sugar(),
	}
}

// Field constructors
func String(key, val string) Field {
	return zap.String(key, val)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

func ErrorField(err error) Field {
	return zap.Error(err)
}

func NamedError(key string, err error) Field {
	return zap.NamedError(key, err)
}

func Object(key string, val zapcore.ObjectMarshaler) Field {
	return zap.Object(key, val)
}