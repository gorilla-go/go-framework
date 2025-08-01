package eventbus

import (
	"fmt"
	"time"
)

// ExampleUsage 展示事件总线的基本用法
func ExampleUsage() {
	// 创建事件总线实例
	eb := New()

	// 注册普通事件监听器
	eb.On("user.login", func(args ...interface{}) {
		if len(args) > 0 {
			fmt.Printf("用户登录: %s\n", args[0])
		}
	})

	// 注册一次性事件监听器
	eb.Once("app.start", func(args ...interface{}) {
		fmt.Println("应用启动完成")
	})

	// 注册多个监听器到同一个事件
	eb.On("user.login", func(args ...interface{}) {
		fmt.Println("记录登录日志")
	})

	eb.On("user.login", func(args ...interface{}) {
		fmt.Println("更新用户状态")
	})

	// 触发事件
	eb.Emit("app.start")
	eb.Emit("user.login", "张三")
	eb.Emit("user.login", "李四")

	// 再次触发app.start事件，由于是once监听器，不会被执行
	eb.Emit("app.start")

	// 查看监听器数量
	fmt.Printf("user.login事件监听器数量: %d\n", eb.ListenerCount("user.login"))

	// 移除特定监听器
	handler := func(args ...interface{}) {
		fmt.Println("这个监听器会被移除")
	}
	eb.On("test.event", handler)
	eb.Off("test.event", handler)

	// 移除事件的所有监听器
	eb.Off("user.login")
	fmt.Printf("移除后user.login事件监听器数量: %d\n", eb.ListenerCount("user.login"))
}

// ExampleGlobalUsage 展示全局事件总线的用法
func ExampleGlobalUsage() {
	// 使用全局函数
	On("message.send", func(args ...interface{}) {
		if len(args) >= 2 {
			fmt.Printf("发送消息: %s -> %s\n", args[0], args[1])
		}
	})

	Once("system.shutdown", func(args ...interface{}) {
		fmt.Println("系统正在关闭...")
	})

	// 触发全局事件
	Emit("message.send", "用户A", "Hello World")
	Emit("system.shutdown")

	// 清理所有全局监听器
	Clear()
}

// ExampleAsyncUsage 展示异步事件处理
func ExampleAsyncUsage() {
	eb := New()

	// 注册异步处理的事件监听器
	eb.On("task.process", func(args ...interface{}) {
		go func() {
			if len(args) > 0 {
				taskID := args[0]
				fmt.Printf("开始处理任务: %v\n", taskID)
				// 模拟耗时操作
				time.Sleep(100 * time.Millisecond)
				fmt.Printf("任务处理完成: %v\n", taskID)

				// 处理完成后触发另一个事件
				eb.Emit("task.completed", taskID)
			}
		}()
	})

	eb.On("task.completed", func(args ...interface{}) {
		if len(args) > 0 {
			fmt.Printf("任务完成通知: %v\n", args[0])
		}
	})

	// 触发多个任务
	for i := 1; i <= 3; i++ {
		eb.Emit("task.process", fmt.Sprintf("task-%d", i))
	}

	// 等待异步任务完成
	time.Sleep(200 * time.Millisecond)
}
