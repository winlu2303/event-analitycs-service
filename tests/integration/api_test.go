package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/event-analytics-service/internal/handler"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/repository"
	"github.com/yourusername/event-analytics-service/internal/service"
)

func setupTestServer(t *testing.T) *gin.Engine {
	// Подключаемся к тестовой БД
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "analytics_test",
			Username: "default",
			Password: "",
		},
	})
	assert.NoError(t, err)

	// Инициализируем репозитории
	eventRepo := repository.NewClickHouseRepository(conn)

	// Инициализируем сервисы
	eventService := service.NewEventService(eventRepo)
	statsService := service.NewStatsService(eventRepo)

	// Инициализируем хендлеры
	eventHandler := handler.NewEventHandler(eventService)
	statsHandler := handler.NewStatsHandler(statsService)

	// Настраиваем роутер
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/events/track", eventHandler.TrackEvent)
		api.GET("/stats/events", statsHandler.GetStatistics)
	}

	return router
}

func TestTrackEventIntegration(t *testing.T) {
	router := setupTestServer(t)

	event := models.Event{
		UserID:    "test_user",
		EventType: models.PageView,
		PageURL:   "/test",
		Metadata: map[string]interface{}{
			"test": "data",
		},
	}

	jsonData, _ := json.Marshal(event)
	req, _ := http.NewRequest("POST", "/api/v1/events/track", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "event_id")
}

func TestGetStatsIntegration(t *testing.T) {
	router := setupTestServer(t)

	// Сначала отправляем тестовые события
	for i := 0; i < 5; i++ {
		event := models.Event{
			UserID:    "test_user",
			EventType: models.PageView,
			PageURL:   "/test",
		}
		jsonData, _ := json.Marshal(event)
		req, _ := http.NewRequest("POST", "/api/v1/events/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Получаем статистику
	req, _ := http.NewRequest("GET", "/api/v1/stats/events?event_type=page_view&start_date=2024-01-01&end_date=2024-12-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "statistics")
}
