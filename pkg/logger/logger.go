package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go-framework/pkg/config"

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
	// ZapLogger Zap日志实例
	ZapLogger *zap.Logger
	// SugarLogger 提供更便捷的API
	SugarLogger *zap.SugaredLogger
)

// InitLogger 初始化日志
func InitLogger(cfg *config.LogConfig) error {
	// 创建日志目录
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 初始化zap
	if err := initZap(cfg); err != nil {
		return err
	}

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

	// 创建日志文件
	logFile, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// 创建Core，只输出到文件
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(logFile),
		zap.NewAtomicLevelAt(level),
	)

	// 创建Logger
	ZapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	// 创建SugarLogger，提供更便捷的API
	SugarLogger = ZapLogger.Sugar()

	return nil
}

// Debug 记录debug级别日志
func Debug(args ...interface{}) {
	SugarLogger.Debug(args...)
}

// Debugf 记录debug级别日志（格式化）
func Debugf(format string, args ...interface{}) {
	SugarLogger.Debugf(format, args...)
}

// Info 记录info级别日志
func Info(args ...interface{}) {
	SugarLogger.Info(args...)
}

// Infof 记录info级别日志（格式化）
func Infof(format string, args ...interface{}) {
	SugarLogger.Infof(format, args...)
}

// Warn 记录warn级别日志
func Warn(args ...interface{}) {
	SugarLogger.Warn(args...)
}

// Warnf 记录warn级别日志（格式化）
func Warnf(format string, args ...interface{}) {
	SugarLogger.Warnf(format, args...)
}

// Error 记录error级别日志
func Error(args ...interface{}) {
	SugarLogger.Error(args...)
}

// Errorf 记录error级别日志（格式化）
func Errorf(format string, args ...interface{}) {
	SugarLogger.Errorf(format, args...)
}

// Fatal 记录fatal级别日志
func Fatal(args ...interface{}) {
	SugarLogger.Fatal(args...)
}

// Fatalf 记录fatal级别日志（格式化）
func Fatalf(format string, args ...interface{}) {
	SugarLogger.Fatalf(format, args...)
}

// Panic 记录panic级别日志
func Panic(args ...interface{}) {
	SugarLogger.Panic(args...)
}

// Panicf 记录panic级别日志（格式化）
func Panicf(format string, args ...interface{}) {
	SugarLogger.Panicf(format, args...)
}

// GetLogger 获取底层的zap.Logger实例
func GetLogger() *zap.Logger {
	return ZapLogger
}
