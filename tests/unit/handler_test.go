package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/event-analytics-service/internal/handler"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/service"
)

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) ProcessEvent(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) GetEventStatistics(ctx context.Context, req models.StatsRequest) ([]models.EventStats, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]models.EventStats), args.Error(1)
}

func TestEventHandler_TrackEvent(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockEventService := new(MockEventService)
	eventHandler := handler.NewEventHandler(mockEventService)

	router := gin.New()
	router.POST("/events/track", eventHandler.TrackEvent)

	// Test data
	event := models.Event{
		UserID:    "user123",
		EventType: models.PageView,
		PageURL:   "/test",
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	// Mock expectations
	mockEventService.On("ProcessEvent", mock.Anything, mock.AnythingOfType("*models.Event")).Return(nil)

	// Execute request
	jsonData, _ := json.Marshal(event)
	req, _ := http.NewRequest("POST", "/events/track", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "event_id")
	assert.Equal(t, "Event tracked successfully", response["message"])

	mockEventService.AssertExpectations(t)
}

func TestStatsHandler_GetStatistics(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockStatsService := new(MockStatsService)
	statsHandler := handler.NewStatsHandler(mockStatsService)

	router := gin.New()
	router.GET("/stats/events", statsHandler.GetStatistics)

	// Test data
	expectedStats := []models.EventStats{
		{TimeBucket: "2024-01-01", EventType: "page_view", Count: 100},
		{TimeBucket: "2024-01-02", EventType: "page_view", Count: 150},
	}

	// Mock expectations
	mockStatsService.On("GetEventStatistics", mock.Anything, mock.AnythingOfType("models.StatsRequest")).
		Return(expectedStats, nil)

	// Execute request
	req, _ := http.NewRequest("GET", "/stats/events?event_type=page_view&start_date=2024-01-01&end_date=2024-01-02", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	stats, exists := response["statistics"]
	assert.True(t, exists)
	assert.NotNil(t, stats)

	mockStatsService.AssertExpectations(t)
}

func TestStatsHandler_GetStatistics_ValidationError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockStatsService := new(MockStatsService)
	statsHandler := handler.NewStatsHandler(mockStatsService)

	router := gin.New()
	router.GET("/stats/events", statsHandler.GetStatistics)

	// Execute request without required params
	req, _ := http.NewRequest("GET", "/stats/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)
	authHandler := handler.NewAuthHandler(mockAuthService)

	router := gin.New()
	router.POST("/auth/login", authHandler.Login)

	// Test data
	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	expectedTokens := &service.TokenPair{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		ExpiresIn:    3600,
	}

	// Mock expectations
	mockAuthService.On("Login", mock.Anything, "test@example.com", "password123").
		Return(expectedTokens, nil)

	// Execute request
	jsonData, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")

	mockAuthService.AssertExpectations(t)
}
