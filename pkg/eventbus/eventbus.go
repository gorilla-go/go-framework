package eventbus

import (
	"errors"
	"sync"
)

var (
	// ErrHandlerAlreadyRegistered 表示处理器已经注册
	ErrHandlerAlreadyRegistered = errors.New("handler already registered")

	// ErrHandlerNotFound 表示处理器未找到
	ErrHandlerNotFound = errors.New("handler not found")

	// 默认事件总线实例
	defaultBus *EventBus
	once       sync.Once
)

// EventBus 是事件总线的实现
type EventBus struct {
	handlers     map[string][]Handler // 按事件类型映射处理器
	handlersLock sync.RWMutex
	asyncEnabled bool
}

// New 创建一个新的事件总线
func New() *EventBus {
	return &EventBus{
		handlers: make(map[string][]Handler),
	}
}

// Default 返回默认的事件总线实例（单例模式）
func Default() *EventBus {
	once.Do(func() {
		defaultBus = New()
	})
	return defaultBus
}

// EnableAsync 启用异步事件处理
func (b *EventBus) EnableAsync() {
	b.asyncEnabled = true
}

// DisableAsync 禁用异步事件处理
func (b *EventBus) DisableAsync() {
	b.asyncEnabled = false
}

// Register 注册一个事件处理器
func (b *EventBus) Register(h Handler) error {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()

	// 获取处理器关心的事件类型
	eventTypes := h.InterestedIn()

	for _, eventType := range eventTypes {
		// 检查处理器是否已注册
		for _, existingHandler := range b.handlers[eventType] {
			if existingHandler == h {
				return ErrHandlerAlreadyRegistered
			}
		}

		// 添加处理器到对应的事件类型
		b.handlers[eventType] = append(b.handlers[eventType], h)
	}

	return nil
}

// RegisterFunc 注册一个函数作为指定事件类型的处理器
func (b *EventBus) RegisterFunc(eventType string, handlerFunc func(*Event) error) {
	handler := NewSingleTypeHandler(eventType, handlerFunc)
	_ = b.Register(handler)
}

// Unregister 注销一个事件处理器
func (b *EventBus) Unregister(h Handler) error {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()

	eventTypes := h.InterestedIn()
	found := false

	for _, eventType := range eventTypes {
		handlers := b.handlers[eventType]
		for i, existingHandler := range handlers {
			if existingHandler == h {
				// 从处理器列表中移除
				b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				found = true
				break
			}
		}
	}

	if !found {
		return ErrHandlerNotFound
	}

	return nil
}

// Publish 发布一个事件
func (b *EventBus) Publish(event *Event) {
	b.handlersLock.RLock()
	handlers := b.handlers[event.Type]
	b.handlersLock.RUnlock()

	for _, handler := range handlers {
		if b.asyncEnabled {
			// 异步处理
			go func(h Handler, e *Event) {
				_ = h.Handle(e)
			}(handler, event)
		} else {
			// 同步处理
			_ = handler.Handle(event)
		}
	}
}

// PublishType 以指定类型和数据发布事件
func (b *EventBus) PublishType(eventType string, data interface{}) {
	event := NewEvent(eventType, data)
	b.Publish(event)
}

// HasHandlersFor 检查是否有处理器注册了指定事件类型
func (b *EventBus) HasHandlersFor(eventType string) bool {
	b.handlersLock.RLock()
	defer b.handlersLock.RUnlock()

	return len(b.handlers[eventType]) > 0
}

// Reset 重置事件总线，清除所有处理器
func (b *EventBus) Reset() {
	b.handlersLock.Lock()
	defer b.handlersLock.Unlock()

	b.handlers = make(map[string][]Handler)
}
