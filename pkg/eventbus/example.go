package eventbus

import (
	"fmt"
)

// 这个文件提供了事件总线的使用示例

// 定义一些事件类型常量
const (
	UserRegisteredEvent = "user.registered"
	UserLoginEvent      = "user.login"
	OrderCreatedEvent   = "order.created"
)

// ExampleUsage 展示事件总线的基本用法
func ExampleUsage() {
	// 获取默认的事件总线
	bus := Default()

	// 启用异步处理
	bus.EnableAsync()

	// 注册一个处理用户注册事件的处理器
	bus.RegisterFunc(UserRegisteredEvent, func(e *Event) error {
		userData := e.Data.(map[string]interface{})
		fmt.Printf("新用户注册: %s (%s)\n", userData["username"], userData["email"])
		return nil
	})

	// 注册一个多事件类型处理器
	loggingHandler := &LoggingHandler{}
	bus.Register(loggingHandler)

	// 发布一个用户注册事件
	bus.PublishType(UserRegisteredEvent, map[string]interface{}{
		"username": "张三",
		"email":    "zhangsan@example.com",
		"id":       1001,
	})

	// 发布一个订单创建事件
	bus.PublishType(OrderCreatedEvent, map[string]interface{}{
		"orderID":   "ORD-2023-001",
		"userID":    1001,
		"amount":    199.99,
		"timestamp": "2023-05-15T14:30:00Z",
	})
}

// LoggingHandler 是一个记录所有感兴趣事件的示例处理器
type LoggingHandler struct{}

// Handle 实现Handler接口的Handle方法，记录事件信息
func (h *LoggingHandler) Handle(e *Event) error {
	fmt.Printf("[%s] 事件类型: %s, 来源: %s\n", e.Timestamp.Format("2006-01-02 15:04:05"), e.Type, e.Source)
	return nil
}

// InterestedIn 实现Handler接口的InterestedIn方法
// 这个处理器对所有用户和订单相关事件感兴趣
func (h *LoggingHandler) InterestedIn() []string {
	return []string{
		UserRegisteredEvent,
		UserLoginEvent,
		OrderCreatedEvent,
	}
}
