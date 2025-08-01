package eventbus

import (
	"reflect"
	"sync"
)

// EventHandler 事件处理函数类型
type EventHandler func(args ...interface{})

// EventBus 事件总线结构体
type EventBus struct {
	mu        sync.RWMutex
	listeners map[string][]EventHandler
	onceMap   map[string]map[int]bool // 记录once监听器的索引
}

// New 创建新的事件总线实例
func New() *EventBus {
	return &EventBus{
		listeners: make(map[string][]EventHandler),
		onceMap:   make(map[string]map[int]bool),
	}
}

// On 注册事件监听器
func (eb *EventBus) On(event string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.listeners[event] = append(eb.listeners[event], handler)
}

// Once 注册一次性事件监听器
func (eb *EventBus) Once(event string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.onceMap[event] == nil {
		eb.onceMap[event] = make(map[int]bool)
	}

	index := len(eb.listeners[event])
	eb.listeners[event] = append(eb.listeners[event], handler)
	eb.onceMap[event][index] = true
}

// Emit 触发事件
func (eb *EventBus) Emit(event string, args ...interface{}) {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.listeners[event]))
	copy(handlers, eb.listeners[event])
	onceIndexes := make([]int, 0)

	// 收集需要删除的once监听器索引
	if eb.onceMap[event] != nil {
		for index := range eb.onceMap[event] {
			onceIndexes = append(onceIndexes, index)
		}
	}
	eb.mu.RUnlock()

	// 执行所有处理函数
	for _, handler := range handlers {
		handler(args...)
	}

	// 删除once监听器
	if len(onceIndexes) > 0 {
		eb.removeOnceListeners(event, onceIndexes)
	}
}

// Off 移除事件监听器
func (eb *EventBus) Off(event string, handler ...EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(handler) == 0 {
		// 如果没有指定处理函数，移除所有监听器
		delete(eb.listeners, event)
		delete(eb.onceMap, event)
		return
	}

	// 移除指定的处理函数
	handlers := eb.listeners[event]
	for _, h := range handler {
		for i := len(handlers) - 1; i >= 0; i-- {
			if reflect.ValueOf(handlers[i]).Pointer() == reflect.ValueOf(h).Pointer() {
				// 删除监听器
				handlers = append(handlers[:i], handlers[i+1:]...)
				// 删除对应的once记录
				if eb.onceMap[event] != nil {
					delete(eb.onceMap[event], i)
				}
			}
		}
	}
	eb.listeners[event] = handlers
}

// removeOnceListeners 移除once监听器
func (eb *EventBus) removeOnceListeners(event string, indexes []int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// 从大到小排序索引，避免删除时索引错位
	for i := 0; i < len(indexes); i++ {
		for j := i + 1; j < len(indexes); j++ {
			if indexes[i] < indexes[j] {
				indexes[i], indexes[j] = indexes[j], indexes[i]
			}
		}
	}

	handlers := eb.listeners[event]
	for _, index := range indexes {
		if index < len(handlers) {
			handlers = append(handlers[:index], handlers[index+1:]...)
		}
	}
	eb.listeners[event] = handlers

	// 清理once映射
	delete(eb.onceMap, event)
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

	eb.listeners = make(map[string][]EventHandler)
	eb.onceMap = make(map[string]map[int]bool)
}
