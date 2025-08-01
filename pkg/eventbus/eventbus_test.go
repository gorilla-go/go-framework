package eventbus

import (
	"sync"
	"testing"
)

func TestEventBus_On(t *testing.T) {
	eb := New()
	called := false

	eb.On("test", func(args ...interface{}) {
		called = true
	})

	eb.Emit("test")

	if !called {
		t.Error("Event handler was not called")
	}
}

func TestEventBus_Once(t *testing.T) {
	eb := New()
	callCount := 0

	eb.Once("test", func(args ...interface{}) {
		callCount++
	})

	// 触发两次事件
	eb.Emit("test")
	eb.Emit("test")

	if callCount != 1 {
		t.Errorf("Expected call count to be 1, got %d", callCount)
	}
}

func TestEventBus_Emit(t *testing.T) {
	eb := New()
	var receivedArgs []interface{}

	eb.On("test", func(args ...interface{}) {
		receivedArgs = args
	})

	expectedArgs := []interface{}{"hello", 123, true}
	eb.Emit("test", expectedArgs...)

	if len(receivedArgs) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(receivedArgs))
	}

	for i, arg := range expectedArgs {
		if receivedArgs[i] != arg {
			t.Errorf("Expected arg %d to be %v, got %v", i, arg, receivedArgs[i])
		}
	}
}

func TestEventBus_Off(t *testing.T) {
	eb := New()
	called := false

	handler := func(args ...interface{}) {
		called = true
	}

	eb.On("test", handler)
	eb.Off("test", handler)
	eb.Emit("test")

	if called {
		t.Error("Event handler should not have been called after removal")
	}
}

func TestEventBus_OffAll(t *testing.T) {
	eb := New()
	callCount := 0

	eb.On("test", func(args ...interface{}) {
		callCount++
	})
	eb.On("test", func(args ...interface{}) {
		callCount++
	})

	// 移除所有监听器
	eb.Off("test")
	eb.Emit("test")

	if callCount != 0 {
		t.Errorf("Expected call count to be 0, got %d", callCount)
	}
}

func TestEventBus_ListenerCount(t *testing.T) {
	eb := New()

	if eb.ListenerCount("test") != 0 {
		t.Error("Expected listener count to be 0 for non-existent event")
	}

	eb.On("test", func(args ...interface{}) {})
	eb.On("test", func(args ...interface{}) {})

	if eb.ListenerCount("test") != 2 {
		t.Errorf("Expected listener count to be 2, got %d", eb.ListenerCount("test"))
	}
}

func TestEventBus_Events(t *testing.T) {
	eb := New()

	eb.On("event1", func(args ...interface{}) {})
	eb.On("event2", func(args ...interface{}) {})

	events := eb.Events()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	// 检查事件名称是否存在
	eventMap := make(map[string]bool)
	for _, event := range events {
		eventMap[event] = true
	}

	if !eventMap["event1"] || !eventMap["event2"] {
		t.Error("Expected events 'event1' and 'event2' to be present")
	}
}

func TestEventBus_Clear(t *testing.T) {
	eb := New()

	eb.On("event1", func(args ...interface{}) {})
	eb.On("event2", func(args ...interface{}) {})

	eb.Clear()

	if len(eb.Events()) != 0 {
		t.Error("Expected no events after clear")
	}
}

func TestEventBus_Concurrent(t *testing.T) {
	eb := New()
	var wg sync.WaitGroup
	var mu sync.Mutex
	callCount := 0

	// 注册多个监听器
	for i := 0; i < 10; i++ {
		eb.On("test", func(args ...interface{}) {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
	}

	// 并发触发事件
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Emit("test")
		}()
	}

	wg.Wait()

	mu.Lock()
	expected := 10 * 100 // 10个监听器 * 100次触发
	if callCount != expected {
		t.Errorf("Expected call count to be %d, got %d", expected, callCount)
	}
	mu.Unlock()
}

func TestEventBus_MultipleOnce(t *testing.T) {
	eb := New()
	callCount := 0

	// 注册多个once监听器
	for i := 0; i < 5; i++ {
		eb.Once("test", func(args ...interface{}) {
			callCount++
		})
	}

	// 触发一次事件
	eb.Emit("test")

	if callCount != 5 {
		t.Errorf("Expected call count to be 5, got %d", callCount)
	}

	// 再次触发，应该没有监听器被调用
	eb.Emit("test")

	if callCount != 5 {
		t.Errorf("Expected call count to remain 5, got %d", callCount)
	}
}

// 测试全局函数
func TestGlobalFunctions(t *testing.T) {
	// 清理全局状态
	Clear()

	called := false
	On("global_test", func(args ...interface{}) {
		called = true
	})

	Emit("global_test")

	if !called {
		t.Error("Global event handler was not called")
	}

	if ListenerCount("global_test") != 1 {
		t.Errorf("Expected global listener count to be 1, got %d", ListenerCount("global_test"))
	}

	Off("global_test")

	if ListenerCount("global_test") != 0 {
		t.Errorf("Expected global listener count to be 0 after removal, got %d", ListenerCount("global_test"))
	}
}

func TestGlobalOnce(t *testing.T) {
	Clear()

	callCount := 0
	Once("global_once_test", func(args ...interface{}) {
		callCount++
	})

	Emit("global_once_test")
	Emit("global_once_test")

	if callCount != 1 {
		t.Errorf("Expected global once call count to be 1, got %d", callCount)
	}
}

// 基准测试
func BenchmarkEventBus_Emit(b *testing.B) {
	eb := New()
	eb.On("benchmark", func(args ...interface{}) {
		// 空处理函数
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Emit("benchmark", "test", 123)
	}
}

func BenchmarkEventBus_On(b *testing.B) {
	eb := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.On("benchmark", func(args ...interface{}) {})
	}
}
