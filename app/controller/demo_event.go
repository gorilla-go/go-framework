package controller

// DemoEventController 演示 EventBus 事件总线
//
// 覆盖的功能：
//   - eventbus.On()    —— 持久监听，每次触发都执行
//   - eventbus.Once()  —— 一次性监听，触发后自动移除
//   - eventbus.Emit()  —— 触发事件，支持多参数
//   - eventbus.Off()   —— 移除监听器
//   - eventbus.Events()         —— 查看已注册的事件列表
//   - eventbus.ListenerCount()  —— 查看某事件的监听器数量
//
// 路由：
//   POST /demo/events/emit        触发一个事件
//   GET  /demo/events/stats       查看事件总线状态
//   GET  /demo/events/log         查看事件日志（内存）
//   POST /demo/events/clear-log   清空事件日志

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/eventbus"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
	"github.com/gorilla-go/go-framework/pkg/router"
	"go.uber.org/fx"
)

// ---- 内存事件日志（供演示查询）----

type eventLogEntry struct {
	Time    string `json:"time"`
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

var (
	eventLog   []eventLogEntry
	eventLogMu sync.Mutex
)

func appendEventLog(event string, payload any) {
	eventLogMu.Lock()
	defer eventLogMu.Unlock()
	eventLog = append(eventLog, eventLogEntry{
		Time:    time.Now().Format("15:04:05.000"),
		Event:   event,
		Payload: payload,
	})
}

// ---- 在包 init 时注册演示监听器 ----
// 这里故意用两种方式注册，演示 On 和 Once 的区别

func init() {
	// On：持久监听 demo.action 事件，每次触发都记录日志
	eventbus.On("demo.action", func(args ...interface{}) {
		payload := args[0]
		appendEventLog("demo.action [On]", payload)
	})

	// On：同一事件可以有多个监听器，独立执行
	eventbus.On("demo.action", func(args ...interface{}) {
		payload := args[0]
		appendEventLog("demo.action [On-2]", fmt.Sprintf("副监听器收到: %v", payload))
	})

	// Once：一次性监听，demo.welcome 事件只处理第一次触发
	eventbus.Once("demo.welcome", func(args ...interface{}) {
		appendEventLog("demo.welcome [Once]", "首次欢迎事件，触发后自动移除监听器")
	})
}

// ---- 控制器 ----

type DemoEventController struct {
	fx.In
}

func (d *DemoEventController) Annotation(rb *router.RouteBuilder) {
	events := rb.Group("/demo/events")
	events.POST("/emit", response.H(d.Emit), "demo@emitEvent")
	events.GET("/stats", d.Stats, "demo@eventStats")
	events.GET("/log", d.Log, "demo@eventLog")
	events.DELETE("/log", d.ClearLog, "demo@clearEventLog")
}

// ---- 请求结构 ----

type emitRequest struct {
	Event   string `json:"event"   binding:"required"`
	Payload any    `json:"payload"`
}

// ---- Handlers ----

// Emit POST /demo/events/emit
// 演示 eventbus.Emit() —— 通过 API 触发一个事件
//
// 示例请求体：
//
//	{ "event": "demo.action", "payload": "hello" }
//	{ "event": "demo.welcome", "payload": null }   // Once 事件，第二次触发无效
func (d *DemoEventController) Emit(c *gin.Context) error {
	var req emitRequest
	if !response.BindJSON(c, &req) {
		return nil
	}

	// 只允许触发 demo. 前缀的事件，防止误触系统事件
	if len(req.Event) < 5 || req.Event[:5] != "demo." {
		return errors.NewBadRequest("只允许触发 demo.* 前缀的事件", nil)
	}

	beforeCount := eventbus.ListenerCount(req.Event)
	eventbus.Emit(req.Event, req.Payload)
	afterCount := eventbus.ListenerCount(req.Event)

	response.SuccessD(c, fmt.Sprintf("事件 %q 已触发", req.Event), gin.H{
		"event":            req.Event,
		"payload":          req.Payload,
		"listeners_before": beforeCount,
		"listeners_after":  afterCount,
		"tip":              "Once 监听器触发后 listeners_after 会比 listeners_before 少",
	})
	return nil
}

// Stats GET /demo/events/stats
// 演示 eventbus.Events() / ListenerCount() —— 查看事件总线当前状态
func (d *DemoEventController) Stats(c *gin.Context) {
	events := eventbus.Events()
	stats := make([]gin.H, 0, len(events))
	for _, name := range events {
		stats = append(stats, gin.H{
			"event":     name,
			"listeners": eventbus.ListenerCount(name),
		})
	}

	response.Success(c, gin.H{
		"total_events": len(events),
		"events":       stats,
	})
}

// Log GET /demo/events/log
// 查看已收集的事件日志
func (d *DemoEventController) Log(c *gin.Context) {
	eventLogMu.Lock()
	snapshot := make([]eventLogEntry, len(eventLog))
	copy(snapshot, eventLog)
	eventLogMu.Unlock()

	response.SuccessD(c, fmt.Sprintf("共 %d 条日志", len(snapshot)), snapshot)
}

// ClearLog DELETE /demo/events/log
// 清空事件日志
func (d *DemoEventController) ClearLog(c *gin.Context) {
	eventLogMu.Lock()
	eventLog = eventLog[:0]
	eventLogMu.Unlock()

	response.SuccessD(c, "日志已清空", nil)
}
