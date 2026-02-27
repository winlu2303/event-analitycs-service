package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/event-analytics-service/internal/metrics"
)

func MetricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Для отслеживания количества одновременных запросов
		m.Increment("http_requests_in_flight")
		defer m.IncrementBy("http_requests_in_flight", -1)

		// Начало замера времени
		start := time.Now()

		// Обработка запроса
		c.Next()

		// Длительность запроса
		duration := time.Since(start)

		// Статус ответа
		status := strconv.Itoa(c.Writer.Status())

		// Метод и путь
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Сохраняем метрики
		m.IncrementHTTPRequest(method, path, status)
		m.ObserveHTTPDuration(method, path, duration)

		// Дополнительные метрики для ошибок
		if c.Writer.Status() >= 400 {
			m.Increment("http_errors_total")
		}
	}
}

// RequestSizeMiddleware отслеживает размер запросов
func RequestSizeMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Размер ответа
		size := c.Writer.Size()
		if size > 0 {
			m.Observe("http_response_size", float64(size))
		}
	}
}

// PanicRecoveryMiddleware восстанавливается после паник и логирует их
func PanicRecoveryMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		m.Increment("http_panics_total")
		c.AbortWithStatusJSON(500, gin.H{
			"error": "internal server error",
		})
	})
}
