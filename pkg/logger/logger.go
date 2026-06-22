package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gorilla-go/go-framework/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 初始化zap
	if err := initZap(cfg); err != nil {
		panic(err)
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

	atomicLevel := zap.NewAtomicLevelAt(level)

	// 文件编码器：始终使用 JSON 格式，便于日志平台采集
	fileEncoderCfg := zapcore.EncoderConfig{
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
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)

	// 日志文件写入器：使用 lumberjack 实现按大小切割、保留份数、按天清理与压缩
	logWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,    // 单个文件最大体积（MB）
		MaxBackups: cfg.MaxBackups, // 保留的旧文件最大份数
		MaxAge:     cfg.MaxAge,     // 旧文件最长保留天数
		Compress:   cfg.Compress,   // 是否 gzip 压缩旧文件
		LocalTime:  true,           // 切割文件名使用本地时间
	}

	// 文件 Core
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logWriter), atomicLevel)

	// 根据配置决定是否同时输出到控制台
	var core zapcore.Core
	if cfg.Stdout {
		consoleEncoderCfg := fileEncoderCfg
		consoleEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色大写级别
		consoleEncoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderCfg)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel)
		core = zapcore.NewTee(fileCore, consoleCore)
	} else {
		core = fileCore
	}

	// 创建Logger
	ZapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	// 创建SugarLogger，提供更便捷的API
	SugarLogger = ZapLogger.Sugar()

	return nil
}

// Debug 记录debug级别日志
func Debug(args ...any) {
	SugarLogger.Debug(args...)
}

// Debugf 记录debug级别日志（格式化）
func Debugf(format string, args ...any) {
	SugarLogger.Debugf(format, args...)
}

// Info 记录info级别日志
func Info(args ...any) {
	SugarLogger.Info(args...)
}

// Infof 记录info级别日志（格式化）
func Infof(format string, args ...any) {
	SugarLogger.Infof(format, args...)
}

// Warn 记录warn级别日志
func Warn(args ...any) {
	SugarLogger.Warn(args...)
}

// Warnf 记录warn级别日志（格式化）
func Warnf(format string, args ...any) {
	SugarLogger.Warnf(format, args...)
}

// Error 记录error级别日志
func Error(args ...any) {
	SugarLogger.Error(args...)
}

// Errorf 记录error级别日志（格式化）
func Errorf(format string, args ...any) {
	SugarLogger.Errorf(format, args...)
}

// Fatal 记录fatal级别日志
func Fatal(args ...any) {
	SugarLogger.Fatal(args...)
}

// Fatalf 记录fatal级别日志（格式化）
func Fatalf(format string, args ...any) {
	SugarLogger.Fatalf(format, args...)
}

// Panic 记录panic级别日志
func Panic(args ...any) {
	SugarLogger.Panic(args...)
}

// Panicf 记录panic级别日志（格式化）
func Panicf(format string, args ...any) {
	SugarLogger.Panicf(format, args...)
}

// GetLogger 获取底层的zap.Logger实例
func GetLogger() *zap.Logger {
	return ZapLogger
}
