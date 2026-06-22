package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newCtx 构造一个带指定 query 的 gin.Context
func newCtx(rawQuery string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)
	return c
}

func TestInputString(t *testing.T) {
	c := newCtx("name=alice")
	if got := Input(c, "name", "def"); got != "alice" {
		t.Errorf("name: 期望 alice, 得到 %q", got)
	}
	// 缺失返回默认值
	if got := Input(c, "missing", "def"); got != "def" {
		t.Errorf("missing: 期望默认 def, 得到 %q", got)
	}
}

func TestInputInt(t *testing.T) {
	c := newCtx("age=30&bad=xx")
	if got := Input(c, "age", 18); got != 30 {
		t.Errorf("age: 期望 30, 得到 %d", got)
	}
	// 解析失败回退默认值
	if got := Input(c, "bad", 18); got != 18 {
		t.Errorf("bad: 解析失败应回退 18, 得到 %d", got)
	}
	// 缺失回退默认值
	if got := Input(c, "none", 7); got != 7 {
		t.Errorf("none: 期望默认 7, 得到 %d", got)
	}
}

func TestInputBool(t *testing.T) {
	c := newCtx("active=true&off=false")
	if got := Input(c, "active", false); got != true {
		t.Errorf("active: 期望 true, 得到 %v", got)
	}
	if got := Input(c, "off", true); got != false {
		t.Errorf("off: 期望 false, 得到 %v", got)
	}
}

func TestInputFloat(t *testing.T) {
	c := newCtx("price=9.99")
	if got := Input(c, "price", 0.0); got != 9.99 {
		t.Errorf("price: 期望 9.99, 得到 %v", got)
	}
}

func TestInputStringSlice(t *testing.T) {
	// 多值形式
	c := newCtx("tags=a&tags=b")
	got := Input(c, "tags", []string{})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("tags(多值): 期望 [a b], 得到 %v", got)
	}
	// 逗号分隔兜底
	c2 := newCtx("tags=x,y,z")
	got2 := Input(c2, "tags", []string{})
	if len(got2) != 3 {
		t.Errorf("tags(逗号): 期望 3 个, 得到 %v", got2)
	}
}

func TestInputIntSlice(t *testing.T) {
	c := newCtx("ids=1,2,3")
	got := Input(c, "ids", []int{})
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("ids: 期望 [1 2 3], 得到 %v", got)
	}
}

func TestFileAbsent(t *testing.T) {
	c := newCtx("")
	if f := File(c, "avatar"); f != nil {
		t.Errorf("无文件时 File 应返回 nil, 得到 %v", f)
	}
	if fs := Files(c, "images"); fs != nil {
		t.Errorf("无文件时 Files 应返回 nil, 得到 %v", fs)
	}
}
