package template

import (
	"math"
	"reflect"
	"strconv"
)

// 优化的数学运算函数，减少反射使用

// fastAdd 快速加法运算，优先处理常见类型
func fastAdd(a, b any) any {
	// 优先处理最常见的类型，避免反射
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			return aVal + bVal
		}
		if bVal, ok := b.(float64); ok {
			return float64(aVal) + bVal
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal + bVal
		}
		if bVal, ok := b.(int); ok {
			return aVal + float64(bVal)
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			return aVal + bVal
		}
		if bVal, ok := b.(float64); ok {
			return float64(aVal) + bVal
		}
	case string:
		// 字符串数值转换
		if aFloat, err := strconv.ParseFloat(aVal, 64); err == nil {
			if bVal, ok := b.(string); ok {
				if bFloat, err := strconv.ParseFloat(bVal, 64); err == nil {
					return aFloat + bFloat
				}
			}
		}
	}

	// 回退到原有的反射方式
	return Add(a, b)
}

// fastSubtract 快速减法运算
func fastSubtract(a, b any) any {
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			return aVal - bVal
		}
		if bVal, ok := b.(float64); ok {
			return float64(aVal) - bVal
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal - bVal
		}
		if bVal, ok := b.(int); ok {
			return aVal - float64(bVal)
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			return aVal - bVal
		}
		if bVal, ok := b.(float64); ok {
			return float64(aVal) - bVal
		}
	}

	return Subtract(a, b)
}

// fastMultiply 快速乘法运算
func fastMultiply(a, b any) any {
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			return aVal * bVal
		}
		if bVal, ok := b.(float64); ok {
			return float64(aVal) * bVal
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal * bVal
		}
		if bVal, ok := b.(int); ok {
			return aVal * float64(bVal)
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			return aVal * bVal
		}
		if bVal, ok := b.(float64); ok {
			return float64(aVal) * bVal
		}
	}

	return Multiply(a, b)
}

// fastDivide 快速除法运算
func fastDivide(a, b any) any {
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / float64(bVal)
		}
		if bVal, ok := b.(float64); ok {
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / bVal
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			if bVal == 0 {
				return "除数不能为零"
			}
			return aVal / bVal
		}
		if bVal, ok := b.(int); ok {
			if bVal == 0 {
				return "除数不能为零"
			}
			return aVal / float64(bVal)
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / float64(bVal)
		}
		if bVal, ok := b.(float64); ok {
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / bVal
		}
	}

	return Divide(a, b)
}

// fastEmpty 优化的空值检查
func fastEmpty(a any) bool {
	if a == nil {
		return true
	}

	// 优先处理常见类型
	switch v := a.(type) {
	case string:
		return v == ""
	case int:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return v == 0
	case bool:
		return !v
	case []string:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	}

	// 回退到原有的反射方式
	return Empty(a)
}

// fastLength 优化的长度计算
func fastLength(a any) int {
	switch v := a.(type) {
	case string:
		return len([]rune(v)) // 正确处理中文字符
	case []string:
		return len(v)
	case []int:
		return len(v)
	case []any:
		return len(v)
	case map[string]any:
		return len(v)
	case map[string]string:
		return len(v)
	case map[string]int:
		return len(v)
	}

	// 回退到原有的反射方式
	return Length(a)
}

// fastInArray 优化的数组包含检查
func fastInArray(needle any, haystack any) bool {
	switch arr := haystack.(type) {
	case []string:
		if needleStr, ok := needle.(string); ok {
			for _, item := range arr {
				if item == needleStr {
					return true
				}
			}
			return false
		}
	case []int:
		if needleInt, ok := needle.(int); ok {
			for _, item := range arr {
				if item == needleInt {
					return true
				}
			}
			return false
		}
	case []int64:
		if needleInt64, ok := needle.(int64); ok {
			for _, item := range arr {
				if item == needleInt64 {
					return true
				}
			}
			return false
		}
	case []any:
		for _, item := range arr {
			if reflect.DeepEqual(needle, item) {
				return true
			}
		}
		return false
	}

	// 回退到原有的反射方式
	return InArray(needle, haystack)
}

// 数值类型转换优化
func toFloat64Fast(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return toFloat64(v)
	}
}

// 优化的四舍五入函数
func fastRound(a any, precision int) float64 {
	var f float64
	var err error

	if f, err = toFloat64Fast(a); err != nil {
		return 0
	}

	p := math.Pow10(precision)
	return math.Round(f*p) / p
}