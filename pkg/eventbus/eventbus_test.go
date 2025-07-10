package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	// 创建一个新的事件总线
	bus := New()

	// 测试事件类型
	const testEventType = "test.event"

	// 测试事件数据
	testEventData := map[string]interface{}{
		"message": "测试消息",
		"code":    200,
	}

	// 测试变量用于验证处理器被调用
	handlerCalled := false
	handlerWg := sync.WaitGroup{}
	handlerWg.Add(1)

	// 注册处理器
	bus.RegisterFunc(testEventType, func(e *Event) error {
		defer handlerWg.Done()
		handlerCalled = true

		// 验证事件类型
		if e.Type != testEventType {
			t.Errorf("预期事件类型 %s，实际得到 %s", testEventType, e.Type)
		}

		// 验证事件数据
		data, ok := e.Data.(map[string]interface{})
		if !ok {
			t.Errorf("事件数据类型错误")
			return nil
		}

		if data["message"] != testEventData["message"] || data["code"] != testEventData["code"] {
			t.Errorf("事件数据不匹配")
		}

		return nil
	})

	// 发布事件
	bus.PublishType(testEventType, testEventData)

	// 等待处理器执行完成
	handlerWg.Wait()

	// 验证处理器被调用
	if !handlerCalled {
		t.Error("处理器未被调用")
	}
}

func TestAsyncEventHandling(t *testing.T) {
	bus := New()
	bus.EnableAsync()

	const testEventType = "test.async"
	handlerCalled := false
	handlerWg := sync.WaitGroup{}
	handlerWg.Add(1)

	// 注册处理器
	bus.RegisterFunc(testEventType, func(e *Event) error {
		// 模拟处理延迟
		time.Sleep(100 * time.Millisecond)
		handlerCalled = true
		handlerWg.Done()
		return nil
	})

	// 记录开始时间
	startTime := time.Now()

	// 发布事件
	bus.PublishType(testEventType, "异步测试")

	// 验证发布立即返回（异步处理）
	elapsed := time.Since(startTime)
	if elapsed >= 100*time.Millisecond {
		t.Errorf("异步处理应该立即返回，但耗时 %v", elapsed)
	}

	// 等待处理器执行完成
	handlerWg.Wait()

	// 验证处理器被调用
	if !handlerCalled {
		t.Error("处理器未被调用")
	}
}

func TestMultipleHandlers(t *testing.T) {
	bus := New()

	const testEventType = "test.multiple"

	// 跟踪处理器调用
	handlersCalledCount := 0
	handlerWg := sync.WaitGroup{}
	handlerWg.Add(3) // 期望3个处理器

	// 注册多个处理器
	for i := 0; i < 3; i++ {
		bus.RegisterFunc(testEventType, func(e *Event) error {
			handlersCalledCount++
			handlerWg.Done()
			return nil
		})
	}

	// 发布事件
	bus.PublishType(testEventType, "测试多处理器")

	// 等待所有处理器执行完成
	handlerWg.Wait()

	// 验证所有处理器都被调用
	if handlersCalledCount != 3 {
		t.Errorf("期望调用3个处理器，实际调用了 %d 个", handlersCalledCount)
	}
}

func TestUnregister(t *testing.T) {
	bus := New()

	const testEventType = "test.unregister"

	// 创建一个可以被注销的处理器
	handler := NewSingleTypeHandler(testEventType, func(e *Event) error {
		t.Error("这个处理器应该已被注销，不应被调用")
		return nil
	})

	// 注册处理器
	err := bus.Register(handler)
	if err != nil {
		t.Fatalf("注册处理器失败：%v", err)
	}

	// 验证处理器被正确注册
	if !bus.HasHandlersFor(testEventType) {
		t.Fatal("处理器注册失败")
	}

	// 注销处理器
	err = bus.Unregister(handler)
	if err != nil {
		t.Fatalf("注销处理器失败：%v", err)
	}

	// 验证处理器已被注销
	if bus.HasHandlersFor(testEventType) {
		t.Fatal("处理器注销失败")
	}

	// 发布事件，此时不应有处理器被调用
	bus.PublishType(testEventType, "测试注销")

	// 尝试注销一个不存在的处理器
	err = bus.Unregister(handler)
	if err != ErrHandlerNotFound {
		t.Errorf("期望错误 %v，实际得到 %v", ErrHandlerNotFound, err)
	}
}
