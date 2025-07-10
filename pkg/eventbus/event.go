package eventbus

import (
	"time"
)

// Event 表示系统中的一个事件
type Event struct {
	// Type 是事件的类型，用于标识不同种类的事件
	Type string

	// Data 是事件携带的数据
	Data interface{}

	// Timestamp 是事件发生的时间戳
	Timestamp time.Time

	// Source 是事件的来源
	Source string
}

// NewEvent 创建一个新的事件
func NewEvent(eventType string, data interface{}) *Event {
	return &Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// WithSource 设置事件的来源
func (e *Event) WithSource(source string) *Event {
	e.Source = source
	return e
}
