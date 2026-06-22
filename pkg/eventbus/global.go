package eventbus

// 全局事件总线实例
var defaultEventBus = New()

// Default 返回全局事件总线实例。
// 用于让依赖注入得到的 *EventBus 与全局 On/Emit/Off 等函数共享同一份监听器，
// 避免出现"注入的总线"和"全局总线"两套互不相通的状态。
func Default() *EventBus {
	return defaultEventBus
}

// On 在全局事件总线上注册事件监听器
func On(event string, handler EventHandler) {
	defaultEventBus.On(event, handler)
}

// Once 在全局事件总线上注册一次性事件监听器
func Once(event string, handler EventHandler) {
	defaultEventBus.Once(event, handler)
}

// Emit 在全局事件总线上触发事件
func Emit(event string, args ...interface{}) {
	defaultEventBus.Emit(event, args...)
}

// Off 在全局事件总线上移除事件监听器
func Off(event string, handler ...EventHandler) {
	defaultEventBus.Off(event, handler...)
}

// ListenerCount 获取全局事件总线上指定事件的监听器数量
func ListenerCount(event string) int {
	return defaultEventBus.ListenerCount(event)
}

// Events 获取全局事件总线上所有已注册的事件名称
func Events() []string {
	return defaultEventBus.Events()
}

// Clear 清除全局事件总线上所有事件监听器
func Clear() {
	defaultEventBus.Clear()
}
