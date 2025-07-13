package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/imraushankr/brevity/server/src/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log   *zap.Logger
	sugar *zap.SugaredLogger
)

// InitBasic initializes a basic console logger for startup
func InitBasic() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	var err error
	log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to initialize basic logger: " + err.Error())
	}
	sugar = log.Sugar()
}

// Init initializes logger with full configuration
func Init(cfg *configs.LoggerConfig) {
	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.FilePath != "" { // If file path is specified, log to file
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0744); err != nil {
			log.Fatal("Failed to create log directory", zap.Error(err))
		}

		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		}
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	} else { // Otherwise log to console
		writeSyncer = zapcore.AddSync(os.Stdout)
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

	core := zapcore.NewCore(encoder, writeSyncer, level)
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = log.Sugar()
}

func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}

func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Sync() {
	_ = log.Sync()
	_ = sugar.Sync()
}

func NewLogger() *zap.Logger {
	return log
}