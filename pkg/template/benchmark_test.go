package template

import (
	"bytes"
	"testing"
	"time"

	"github.com/gorilla-go/go-framework/pkg/config"
)

// 设置测试环境
func setupBenchmark() Manager {
	cfg := config.TemplateConfig{
		Path:      "../../templates",
		Layouts:   "../../templates/layouts",
		Extension: ".html",
	}
	return NewTemplateManager(cfg, false)
}

// 基准测试数据
var benchmarkData = map[string]any{
	"title":   "测试页面",
	"content": "这是一个测试内容",
	"items":   []string{"item1", "item2", "item3", "item4", "item5"},
	"count":   100,
	"price":   99.99,
	"active":  true,
	"user": map[string]any{
		"name":  "测试用户",
		"email": "test@example.com",
		"age":   25,
	},
}

// 数学函数基准测试
func BenchmarkMathFunctions(b *testing.B) {
	b.Run("Add_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastAdd(10, 20)
		}
	})

	b.Run("Add_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Add(10, 20)
		}
	})

	b.Run("Subtract_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastSubtract(100, 30)
		}
	})

	b.Run("Subtract_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Subtract(100, 30)
		}
	})

	b.Run("Multiply_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastMultiply(5, 10)
		}
	})

	b.Run("Multiply_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Multiply(5, 10)
		}
	})

	b.Run("Divide_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastDivide(100, 5)
		}
	})

	b.Run("Divide_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Divide(100, 5)
		}
	})
}

// 集合函数基准测试
func BenchmarkCollectionFunctions(b *testing.B) {
	testArray := []string{"a", "b", "c", "d", "e"}
	testMap := map[string]any{"key1": "value1", "key2": 42, "key3": true}

	b.Run("Empty_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastEmpty("")
			fastEmpty(0)
			fastEmpty(testArray)
			fastEmpty(testMap)
		}
	})

	b.Run("Empty_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Empty("")
			Empty(0)
			Empty(testArray)
			Empty(testMap)
		}
	})

	b.Run("Length_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastLength("hello world")
			fastLength(testArray)
			fastLength(testMap)
		}
	})

	b.Run("Length_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Length("hello world")
			Length(testArray)
			Length(testMap)
		}
	})

	b.Run("InArray_Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastInArray("c", testArray)
			fastInArray("z", testArray)
		}
	})

	b.Run("InArray_Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InArray("c", testArray)
			InArray("z", testArray)
		}
	})
}

// 字符串函数基准测试
func BenchmarkStringFunctions(b *testing.B) {
	testHTML := "<p>这是一段<strong>HTML</strong>内容，包含<a href='#'>链接</a>和其他标签</p>"

	b.Run("StripTags_Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StripTags(testHTML)
		}
	})

	testText := "这是第一行\n这是第二行\n这是第三行"
	b.Run("Nl2br", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Nl2br(testText)
		}
	})

	longText := "这是一段非常长的文本，用于测试截断功能的性能表现，包含了中文字符"
	b.Run("Truncate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Truncate(longText, 10)
		}
	})
}

// 模板渲染基准测试
func BenchmarkTemplateRendering(b *testing.B) {
	// 注意：这些测试需要实际的模板文件，如果没有文件会失败
	// 在实际环境中运行时，请确保模板文件存在

	b.Skip("跳过模板渲染测试 - 需要实际的模板文件")

	manager := setupBenchmark()
	var buf bytes.Buffer

	b.Run("Simple_Template", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			_ = manager.RenderPartial(&buf, "simple", benchmarkData)
		}
	})

	b.Run("Complex_Template_With_Layout", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			_ = manager.Render(&buf, "complex", benchmarkData, "main")
		}
	})
}

// 函数映射构建基准测试
func BenchmarkFuncMapCreation(b *testing.B) {
	b.Run("FuncMap_Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FuncMap()
		}
	})
}

// 时间函数基准测试
func BenchmarkTimeFunctions(b *testing.B) {
	testTime := time.Now()

	b.Run("FormatDateTime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FormatDateTime(testTime)
		}
	})

	b.Run("DateFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DateFormat(testTime, "Y-m-d H:i:s")
		}
	})

	b.Run("HumanizeTime", func(b *testing.B) {
		pastTime := testTime.Add(-2 * time.Hour)
		for i := 0; i < b.N; i++ {
			HumanizeTime(pastTime)
		}
	})
}

// 比较函数基准测试
func BenchmarkComparisonFunctions(b *testing.B) {
	b.Run("Eq", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Eq(10, 10)
			Eq("hello", "hello")
			Eq(true, true)
		}
	})

	b.Run("Compare", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			compare(10, 20)
			compare(3.14, 2.71)
			compare("apple", "banana")
		}
	})
}

// 内存分配基准测试
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("NewMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewMap("key1", "value1", "key2", 42, "key3", true)
		}
	})

	testMap := map[string]any{"key1": "value1", "key2": 42}
	b.Run("MapGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MapGet(testMap, "key1")
			MapGet(testMap, "key2")
			MapGet(testMap, "nonexistent")
		}
	})
}

// 错误处理基准测试
func BenchmarkErrorHandling(b *testing.B) {
	b.Run("ValidateTemplateName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ValidateTemplateName("valid_template_name")
			ValidateTemplateName("another/valid/name")
		}
	})

	b.Run("NewTemplateError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewTemplateError("TEST_ERROR", "测试错误", "test_template", nil)
		}
	})
}