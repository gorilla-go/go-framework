package eventbus

// Handler 是事件处理器接口，负责处理特定类型的事件
type Handler interface {
	// Handle 处理一个事件
	// 返回错误表示处理过程中发生的问题
	Handle(*Event) error

	// InterestedIn 返回处理器感兴趣的事件类型列表
	// 当事件总线收到事件时，会检查每个处理器是否对该事件类型感兴趣
	InterestedIn() []string
}

// HandlerFunc 是一个函数类型，实现了Handler接口
type HandlerFunc func(*Event) error

// Handle 实现Handler接口的Handle方法
func (f HandlerFunc) Handle(e *Event) error {
	return f(e)
}

// SingleTypeHandler 是一个处理单一事件类型的处理器
type SingleTypeHandler struct {
	EventType  string
	HandleFunc HandlerFunc
}

// Handle 实现Handler接口的Handle方法
func (h *SingleTypeHandler) Handle(e *Event) error {
	return h.HandleFunc(e)
}

// InterestedIn 实现Handler接口的InterestedIn方法
func (h *SingleTypeHandler) InterestedIn() []string {
	return []string{h.EventType}
}

// NewSingleTypeHandler 创建一个处理单一事件类型的处理器
func NewSingleTypeHandler(eventType string, handleFunc HandlerFunc) *SingleTypeHandler {
	return &SingleTypeHandler{
		EventType:  eventType,
		HandleFunc: handleFunc,
	}
}
