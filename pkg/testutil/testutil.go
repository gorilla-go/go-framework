// Package testutil 提供框架集成测试工具（参考 Fiber App.Test() 设计）
// 无需启动真实 HTTP 服务器即可测试 controller 路由逻辑
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// Request 向 gin.Engine 发送一个测试请求，返回 ResponseRecorder
// 不启动真实服务器，直接调用 router.ServeHTTP
//
// 示例：
//
//	w := testutil.Request(router, "GET", "/users/1", nil)
//	assert.Equal(t, 200, w.Code)
func Request(router *gin.Engine, method, path string, body io.Reader, headers ...map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, body)
	for _, h := range headers {
		for k, v := range h {
			req.Header.Set(k, v)
		}
	}
	router.ServeHTTP(w, req)
	return w
}

// RequestJSON 发送 JSON 请求，自动设置 Content-Type: application/json
func RequestJSON(router *gin.Engine, method, path string, payload any, headers ...map[string]string) *httptest.ResponseRecorder {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewBuffer(b)
	}
	merged := map[string]string{"Content-Type": "application/json"}
	for _, h := range headers {
		for k, v := range h {
			merged[k] = v
		}
	}
	return Request(router, method, path, body, merged)
}

// DecodeJSON 从 ResponseRecorder 中解码 JSON 响应体
//
// 示例：
//
//	var resp response.Response
//	testutil.DecodeJSON(w, &resp)
func DecodeJSON(w *httptest.ResponseRecorder, v any) error {
	return json.NewDecoder(w.Body).Decode(v)
}
