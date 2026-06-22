package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCORSEngine(h gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h)
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

// TestCORSAllowAll 通配模式：返回 * 且不携带凭证（符合规范）
func TestCORSAllowAll(t *testing.T) {
	r := newCORSEngine(CORSMiddleware())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("通配模式 Allow-Origin 期望 *，得到 %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("通配模式不应携带 Allow-Credentials，得到 %q", got)
	}
}

// TestCORSWhitelist 白名单模式：命中来源回显并带凭证；未命中来源不回显
func TestCORSWhitelist(t *testing.T) {
	r := newCORSEngine(CORSMiddleware("https://a.com"))

	// 命中白名单
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://a.com")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://a.com" {
		t.Errorf("白名单命中应回显来源，得到 %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("白名单模式应允许凭证，得到 %q", got)
	}

	// 未命中白名单
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Origin", "https://evil.com")
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("白名单未命中不应设置 Allow-Origin，得到 %q", got)
	}
}

// TestCORSPreflight 预检请求应返回 204 并带方法/头部信息
func TestCORSPreflight(t *testing.T) {
	r := newCORSEngine(CORSMiddleware())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://a.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("预检应返回 204，得到 %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("预检应设置 Access-Control-Allow-Methods")
	}
}
