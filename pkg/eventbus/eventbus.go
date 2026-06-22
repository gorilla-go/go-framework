package eventbus

import (
	"reflect"
	"sync"
)

// EventHandler 事件处理函数类型
type EventHandler func(args ...interface{})

// handlerEntry 内部处理函数条目，区分普通和 once 监听器
type handlerEntry struct {
	handler EventHandler
	once    bool
	called  bool // once 监听器是否已执行
}

// EventBus 事件总线结构体
type EventBus struct {
	mu        sync.RWMutex
	listeners map[string][]*handlerEntry
}

// New 创建新的事件总线实例
func New() *EventBus {
	return &EventBus{
		listeners: make(map[string][]*handlerEntry),
	}
}

// On 注册事件监听器
func (eb *EventBus) On(event string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.listeners[event] = append(eb.listeners[event], &handlerEntry{handler: handler})
}

// Once 注册一次性事件监听器（触发后自动移除）
func (eb *EventBus) Once(event string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.listeners[event] = append(eb.listeners[event], &handlerEntry{handler: handler, once: true})
}

// Emit 触发事件
//
// 在锁内完成两件事：认领待执行的处理函数、移除已认领的 once 监听器；
// 随后在锁外执行处理函数，避免 handler 内部再调用 On/Off/Emit 造成死锁。
// once 监听器通过 called 标志在锁的保护下"认领"，保证并发 Emit 下也只执行一次。
func (eb *EventBus) Emit(event string, args ...interface{}) {
	eb.mu.Lock()
	entries := eb.listeners[event]
	if len(entries) == 0 {
		eb.mu.Unlock()
		return
	}

	toRun := make([]EventHandler, 0, len(entries))
	var remaining []*handlerEntry
	for _, entry := range entries {
		if entry.once {
			// once 监听器只能被认领一次；已被其他 Emit 认领则跳过，且不保留
			if entry.called {
				continue
			}
			entry.called = true
			toRun = append(toRun, entry.handler)
			continue
		}
		toRun = append(toRun, entry.handler)
		remaining = append(remaining, entry)
	}

	// 更新监听器列表：移除已认领的 once 监听器
	if len(remaining) != len(entries) {
		if len(remaining) == 0 {
			delete(eb.listeners, event)
		} else {
			eb.listeners[event] = remaining
		}
	}
	eb.mu.Unlock()

	// 锁外执行处理函数
	for _, handler := range toRun {
		handler(args...)
	}
}

// Off 移除事件监听器
func (eb *EventBus) Off(event string, handler ...EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(handler) == 0 {
		delete(eb.listeners, event)
		return
	}

	entries := eb.listeners[event]
	for _, h := range handler {
		hPtr := reflect.ValueOf(h).Pointer()
		for i := len(entries) - 1; i >= 0; i-- {
			if reflect.ValueOf(entries[i].handler).Pointer() == hPtr {
				entries = append(entries[:i], entries[i+1:]...)
			}
		}
	}
	eb.listeners[event] = entries
}

// ListenerCount 获取指定事件的监听器数量
func (eb *EventBus) ListenerCount(event string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.listeners[event])
}

// Events 获取所有已注册的事件名称
func (eb *EventBus) Events() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	events := make([]string, 0, len(eb.listeners))
	for event := range eb.listeners {
		events = append(events, event)
	}
	return events
}

// Clear 清除所有事件监听器
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.listeners = make(map[string][]*handlerEntry)
}
