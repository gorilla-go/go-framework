package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	apperrors "github.com/gorilla-go/go-framework/pkg/errors"
)

// TestMain 切换到仓库根目录，使 config.MustFetch() 能定位到 config/config.yaml
// （wrapH 的 HTML 错误页分支依赖全局配置）。
func TestMain(m *testing.M) {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "config", "config.yaml")); err == nil {
			_ = os.Chdir(dir)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // 到达文件系统根仍未找到，按原目录运行
		}
		dir = parent
	}
	os.Exit(m.Run())
}

// newWrapEngine 注册一个返回指定错误的处理器
func newWrapEngine(err error) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", wrapH(func(c *gin.Context) error { return err }))
	return r
}

// TestWrapHPageRequestRendersHTML 页面请求遇到非预期错误时应渲染 HTML 错误页，而非 JSON
func TestWrapHPageRequestRendersHTML(t *testing.T) {
	r := newWrapEngine(fmt.Errorf("boom"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html") // 模拟浏览器页面请求
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望 500，得到 %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("页面请求应返回 HTML，得到 Content-Type=%q", ct)
	}
}

// TestWrapHAjaxRequestReturnsJSON AJAX/JSON 请求遇到非预期错误时应返回 JSON
func TestWrapHAjaxRequestReturnsJSON(t *testing.T) {
	r := newWrapEngine(fmt.Errorf("boom"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest") // 标记为 AJAX
	r.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("AJAX 请求应返回 JSON，得到 Content-Type=%q", ct)
	}
}

// TestWrapHAppErrorAlwaysJSON 业务 AppError 无论页面还是 API 请求都走统一 JSON 响应
func TestWrapHAppErrorAlwaysJSON(t *testing.T) {
	var appErr error = apperrors.NewBadRequest("参数不合法", errors.New("invalid"))
	r := newWrapEngine(appErr)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("AppError 应映射为对应状态码 400，得到 %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("AppError 应返回 JSON，得到 Content-Type=%q", ct)
	}
}
