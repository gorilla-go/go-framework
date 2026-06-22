package logger

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla-go/go-framework/pkg/config"
)

// TestLogRotation 验证日志按 MaxSize 切割：写入超过 1MB 后应产生多个日志文件
func TestLogRotation(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.LogConfig{
		Level:      "info",
		Filename:   filepath.Join(dir, "app.log"),
		MaxSize:    1, // 1MB，lumberjack 的最小有效切割阈值
		MaxBackups: 3,
		MaxAge:     1,
		Compress:   false,
		Format:     "json",
		Stdout:     false,
	}

	if err := InitLogger(cfg); err != nil {
		t.Fatalf("InitLogger 失败: %v", err)
	}

	// 写入约 1.5MB 日志触发至少一次切割
	line := strings.Repeat("x", 500)
	for i := 0; i < 3000; i++ {
		Info(line)
	}
	_ = ZapLogger.Sync()

	files, err := filepath.Glob(filepath.Join(dir, "app*"))
	if err != nil {
		t.Fatalf("读取日志目录失败: %v", err)
	}
	if len(files) < 2 {
		t.Fatalf("期望发生日志轮转产生多个文件，实际仅 %d 个: %v", len(files), files)
	}
}
