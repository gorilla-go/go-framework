package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-framework/pkg/config"

	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 定义日志级别常量
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
	PanicLevel = "panic"
)

var (
	// Log 全局logger实例
	Log *logrus.Logger
	// ZapLogger Zap日志实例
	ZapLogger *zap.Logger
)

// InitLogger 初始化日志
func InitLogger(cfg *config.LogConfig) error {
	// 创建日志目录
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 初始化logrus
	if err := initLogrus(cfg); err != nil {
		return err
	}

	// 初始化zap
	if err := initZap(cfg); err != nil {
		return err
	}

	return nil
}

// initLogrus 初始化logrus
func initLogrus(cfg *config.LogConfig) error {
	Log = logrus.New()

	// 设置日志格式
	if cfg.Format == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	}

	// 设置输出
	rotator := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups, // 保留旧文件的最大个数
		MaxAge:     cfg.MaxAge,     // 保留旧文件的最大天数
		Compress:   cfg.Compress,   // 是否压缩
	}

	// 同时输出到标准输出和文件
	Log.SetOutput(rotator)

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	Log.SetLevel(level)

	return nil
}

// initZap 初始化zap
func initZap(cfg *config.LogConfig) error {
	// 定义日志级别
	var level zapcore.Level
	switch cfg.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	case PanicLevel:
		level = zapcore.PanicLevel
	default:
		level = zapcore.InfoLevel
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建输出配置
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建日志轮转器
	writer, err := rotateLogs.New(
		cfg.Filename+".%Y%m%d",
		rotateLogs.WithLinkName(cfg.Filename),
		rotateLogs.WithMaxAge(time.Duration(cfg.MaxAge)*24*time.Hour),
		rotateLogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		return fmt.Errorf("failed to create rotate logs: %w", err)
	}

	// 创建Core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(writer)),
		zap.NewAtomicLevelAt(level),
	)

	// 创建Logger
	ZapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Debug 记录debug级别日志
func Debug(args ...interface{}) {
	Log.Debug(args...)
	ZapLogger.Sugar().Debug(args...)
}

// Debugf 记录debug级别日志（格式化）
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
	ZapLogger.Sugar().Debugf(format, args...)
}

// Info 记录info级别日志
func Info(args ...interface{}) {
	Log.Info(args...)
	ZapLogger.Sugar().Info(args...)
}

// Infof 记录info级别日志（格式化）
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
	ZapLogger.Sugar().Infof(format, args...)
}

// Warn 记录warn级别日志
func Warn(args ...interface{}) {
	Log.Warn(args...)
	ZapLogger.Sugar().Warn(args...)
}

// Warnf 记录warn级别日志（格式化）
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
	ZapLogger.Sugar().Warnf(format, args...)
}

// Error 记录error级别日志
func Error(args ...interface{}) {
	Log.Error(args...)
	ZapLogger.Sugar().Error(args...)
}

// Errorf 记录error级别日志（格式化）
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
	ZapLogger.Sugar().Errorf(format, args...)
}

// Fatal 记录fatal级别日志
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
	ZapLogger.Sugar().Fatal(args...)
}

// Fatalf 记录fatal级别日志（格式化）
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
	ZapLogger.Sugar().Fatalf(format, args...)
}

// Panic 记录panic级别日志
func Panic(args ...interface{}) {
	Log.Panic(args...)
	ZapLogger.Sugar().Panic(args...)
}

// Panicf 记录panic级别日志（格式化）
func Panicf(format string, args ...interface{}) {
	Log.Panicf(format, args...)
	ZapLogger.Sugar().Panicf(format, args...)
}
