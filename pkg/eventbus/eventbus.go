package eventbus

import (
	"reflect"
	"sort"
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
func (eb *EventBus) Emit(event string, args ...interface{}) {
	eb.mu.RLock()
	entries := make([]*handlerEntry, len(eb.listeners[event]))
	copy(entries, eb.listeners[event])
	eb.mu.RUnlock()

	// 执行所有处理函数，标记 once 条目
	var onceIndexes []int
	for i, entry := range entries {
		entry.handler(args...)
		if entry.once {
			onceIndexes = append(onceIndexes, i)
		}
	}

	// 移除已执行的 once 监听器
	if len(onceIndexes) > 0 {
		eb.removeOnceListeners(event, onceIndexes)
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

// removeOnceListeners 从大到小删除指定索引的条目，避免索引错位
func (eb *EventBus) removeOnceListeners(event string, indexes []int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// 从大到小排序，保证删除时不影响前面的索引
	sort.Sort(sort.Reverse(sort.IntSlice(indexes)))

	entries := eb.listeners[event]
	for _, idx := range indexes {
		if idx < len(entries) {
			entries = append(entries[:idx], entries[idx+1:]...)
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
