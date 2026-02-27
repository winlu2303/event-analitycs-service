package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerConfig конфигурация для логгера
type LoggerConfig struct {
	SkipPaths []string
	LogBody   bool
}

// LoggerMiddleware возвращает middleware для логирования запросов
func LoggerMiddleware(logger *logrus.Logger, config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем указанные пути
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Логируем тело запроса если нужно
		var requestBody []byte
		if config.LogBody && (method == "POST" || method == "PUT") {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Создаем запись лога
		entry := logger.WithFields(logrus.Fields{
			"method":     method,
			"path":       path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"request_id": c.GetString("request_id"),
		})

		// Если есть пользователь
		if userID, exists := c.Get("user_id"); exists {
			entry = entry.WithField("user_id", userID)
		}

		entry.Info("request started")

		// Обрабатываем запрос
		c.Next()

		// Логируем после завершения
		latency := time.Since(start)
		status := c.Writer.Status()

		fields := logrus.Fields{
			"status":     status,
			"latency_ms": latency.Milliseconds(),
			"bytes_sent": c.Writer.Size(),
		}

		// Добавляем тело ответа для ошибок
		if status >= 400 {
			if body, exists := c.Get("response_body"); exists {
				fields["response"] = body
			}
		}

		// Выбираем уровень логирования по статусу
		if status >= 500 {
			entry.WithFields(fields).Error("request failed")
		} else if status >= 400 {
			entry.WithFields(fields).Warn("request warning")
		} else {
			entry.WithFields(fields).Info("request completed")
		}
	}
}

// RequestIDMiddleware добавляет request ID в контекст
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// StructuredLogger логирует в структурированном формате
type StructuredLogger struct {
	logger *logrus.Logger
}

func NewStructuredLogger(logger *logrus.Logger) *StructuredLogger {
	return &StructuredLogger{
		logger: logger,
	}
}

func (l *StructuredLogger) Write(p []byte) (n int, err error) {
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err == nil {
		l.logger.WithFields(logrus.Fields(logEntry)).Info()
	} else {
		l.logger.Info(string(p))
	}
	return len(p), nil
}

// Helper functions
func generateRequestID() string {
	return "req_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1)
	}
	return string(result)
}
