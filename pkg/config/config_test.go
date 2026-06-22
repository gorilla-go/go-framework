package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTempConfig 写入一个临时配置文件并返回路径
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时配置失败: %v", err)
	}
	return path
}

// TestLoadDefaults 配置文件缺字段时应回退到默认值，而非零值
func TestLoadDefaults(t *testing.T) {
	// 仅提供 server.port，其余字段缺失
	path := writeTempConfig(t, "server:\n  port: 8081\n")

	cfg, err := load(path)
	if err != nil {
		t.Fatalf("load 失败: %v", err)
	}

	if cfg.Server.Port != 8081 {
		t.Errorf("Port: 期望文件值 8081，得到 %d", cfg.Server.Port)
	}
	// 缺失字段应取默认值而非 0 / ""
	if cfg.Server.ReadTimeout != 60 {
		t.Errorf("ReadTimeout: 期望默认 60，得到 %d", cfg.Server.ReadTimeout)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Mode: 期望默认 release，得到 %q", cfg.Server.Mode)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level: 期望默认 info，得到 %q", cfg.Log.Level)
	}
	if cfg.Database.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns: 期望默认 100，得到 %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Session.Store != "cookie" {
		t.Errorf("Session.Store: 期望默认 cookie，得到 %q", cfg.Session.Store)
	}
	// 可信代理默认仅本机回环
	if len(cfg.Server.TrustedProxies) != 2 || cfg.Server.TrustedProxies[0] != "127.0.0.1" {
		t.Errorf("TrustedProxies: 期望默认 [127.0.0.1 ::1]，得到 %v", cfg.Server.TrustedProxies)
	}
}

// TestLoadFileOverridesDefault 文件值应覆盖默认值
func TestLoadFileOverridesDefault(t *testing.T) {
	path := writeTempConfig(t, "server:\n  mode: debug\n  read_timeout: 30\n")

	cfg, err := load(path)
	if err != nil {
		t.Fatalf("load 失败: %v", err)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("Mode: 期望文件值 debug，得到 %q", cfg.Server.Mode)
	}
	if cfg.Server.ReadTimeout != 30 {
		t.Errorf("ReadTimeout: 期望文件值 30，得到 %d", cfg.Server.ReadTimeout)
	}
}

// TestLoadEnvOverrides 环境变量应覆盖文件与默认值（含嵌套 key）
func TestLoadEnvOverrides(t *testing.T) {
	path := writeTempConfig(t, "server:\n  port: 8081\n  mode: debug\ndatabase:\n  host: localhost\n")

	t.Setenv("SERVER_PORT", "9999")       // 覆盖文件中的 server.port
	t.Setenv("SERVER_MODE", "release")    // 覆盖文件中的 server.mode
	t.Setenv("DATABASE_HOST", "10.0.0.1") // 覆盖文件中的 database.host
	t.Setenv("REDIS_PORT", "6400")        // 文件未设置，覆盖默认值

	cfg, err := load(path)
	if err != nil {
		t.Fatalf("load 失败: %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("Port: 期望 env 覆盖为 9999，得到 %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Mode: 期望 env 覆盖为 release，得到 %q", cfg.Server.Mode)
	}
	if cfg.Database.Host != "10.0.0.1" {
		t.Errorf("Database.Host: 期望 env 覆盖为 10.0.0.1，得到 %q", cfg.Database.Host)
	}
	if cfg.Redis.Port != 6400 {
		t.Errorf("Redis.Port: 期望 env 覆盖默认值为 6400，得到 %d", cfg.Redis.Port)
	}
}

// TestLoadMissingFile 配置文件不存在时应返回错误
func TestLoadMissingFile(t *testing.T) {
	if _, err := load(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("配置文件不存在时应返回错误")
	}
}
