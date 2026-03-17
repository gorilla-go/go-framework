package validator

import (
	"sync"
)

// Validator 校验接口（参考 Echo 设计）
// 业务层实现后通过 Register() 注入，框架本身不绑定具体实现
// 推荐使用 github.com/go-playground/validator/v10
type Validator interface {
	Validate(i any) error
}

var (
	global Validator
	mu     sync.RWMutex
)

// Register 注册全局校验器（应在应用启动时调用）
func Register(v Validator) {
	mu.Lock()
	defer mu.Unlock()
	global = v
}

// Validate 使用全局校验器校验数据，未注册时直接返回 nil（不强制要求）
func Validate(i any) error {
	mu.RLock()
	v := global
	mu.RUnlock()
	if v == nil {
		return nil
	}
	return v.Validate(i)
}
