package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/service"
)

// Mock репозитория
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) InsertEvent(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) InsertEventBatch(ctx context.Context, events []*models.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockEventRepository) GetStats(ctx context.Context, filter models.StatsRequest) ([]models.EventStats, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.EventStats), args.Error(1)
}

// Остальные методы интерфейса...

func TestEventService_ProcessEvent(t *testing.T) {
	// Arrange
	mockRepo := new(MockEventRepository)
	service := service.NewEventService(mockRepo)

	event := &models.Event{
		UserID:    "user123",
		EventType: models.PageView,
		PageURL:   "/home",
		Timestamp: time.Now(),
	}

	mockRepo.On("InsertEvent", mock.Anything, mock.AnythingOfType("*models.Event")).Return(nil)

	// Act
	err := service.ProcessEvent(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID) // Проверяем, что ID сгенерировался
	mockRepo.AssertExpectations(t)
}

func TestStatsService_CalculateConversionRate(t *testing.T) {
	// Arrange
	mockRepo := new(MockEventRepository)
	service := service.NewStatsService(mockRepo)

	viewStats := []models.EventStats{
		{TimeBucket: "2024-01-01", EventType: "page_view", Count: 100},
		{TimeBucket: "2024-01-02", EventType: "page_view", Count: 150},
	}

	purchaseStats := []models.EventStats{
		{TimeBucket: "2024-01-01", EventType: "purchase", Count: 10},
		{TimeBucket: "2024-01-02", EventType: "purchase", Count: 15},
	}

	mockRepo.On("GetStats", mock.Anything, mock.Anything).Return(viewStats, nil).Once()
	mockRepo.On("GetStats", mock.Anything, mock.Anything).Return(purchaseStats, nil).Once()

	// Act
	rate, err := service.CalculateConversionRate(context.Background(), "2024-01-01", "2024-01-02")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 10.0, rate) // (10+15)/(100+150) * 100 = 10%
	mockRepo.AssertExpectations(t)
}
