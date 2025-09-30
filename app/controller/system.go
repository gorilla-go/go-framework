package controller

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/router"
	"go.uber.org/fx"
)

var startTime = time.Now()

type SystemController struct {
	fx.In
}

func (s *SystemController) Annotation(rb *router.RouteBuilder) {
	rb.GET("/system/stats", s.Stats, "system@stats")
}

func (s *SystemController) Stats(ctx *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"memory": gin.H{
			"alloc":          formatBytes(memStats.Alloc),
			"total_alloc":    formatBytes(memStats.TotalAlloc),
			"sys":            formatBytes(memStats.Sys),
			"heap_alloc":     formatBytes(memStats.HeapAlloc),
			"heap_sys":       formatBytes(memStats.HeapSys),
			"heap_idle":      formatBytes(memStats.HeapIdle),
			"heap_in_use":    formatBytes(memStats.HeapInuse),
			"heap_released":  formatBytes(memStats.HeapReleased),
			"heap_objects":   memStats.HeapObjects,
			"stack_in_use":   formatBytes(memStats.StackInuse),
			"stack_sys":      formatBytes(memStats.StackSys),
			"gc_sys":         formatBytes(memStats.GCSys),
			"num_gc":         memStats.NumGC,
			"last_gc_time":   time.Unix(0, int64(memStats.LastGC)).Format(time.RFC3339),
			"gc_pause_total": time.Duration(memStats.PauseTotalNs).String(),
		},
		"runtime": gin.H{
			"go_version":   runtime.Version(),
			"num_cpu":      runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
			"gomaxprocs":   runtime.GOMAXPROCS(0),
		},
	}

	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

func formatBytes(bytes uint64) gin.H {
	const unit = 1024
	if bytes < unit {
		return gin.H{
			"value": bytes,
			"unit":  "B",
		}
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(bytes) / float64(div)

	return gin.H{
		"value": value,
		"unit":  units[exp],
	}
}