package repository

import (
	"context"
	"time"

	"github.com/yourusername/event-analytics-service/internal/models"
)

// EventRepository defines the interface for event storage operations
type EventRepository interface {
	// Основные операции
	InsertEvent(ctx context.Context, event *models.Event) error
	InsertEventBatch(ctx context.Context, events []*models.Event) error

	// Статистика
	GetStats(ctx context.Context, filter models.StatsRequest) ([]models.EventStats, error)
	GetStatsByProject(ctx context.Context, projectID string, filter models.StatsRequest) ([]models.EventStats, error)

	// Детальные запросы
	GetEventsByUser(ctx context.Context, userID string, limit, offset int) ([]models.Event, error)
	GetEventsByType(ctx context.Context, eventType models.EventType, start, end time.Time) ([]models.Event, error)

	// Аналитические запросы
	GetTopPages(ctx context.Context, projectID string, limit int) ([]models.PageStat, error)
	GetUserSessions(ctx context.Context, userID string, sessionTimeout time.Duration) ([]models.UserSession, error)
	GetFunnelAnalysis(ctx context.Context, projectID string, steps []models.EventType, start, end time.Time) ([]models.FunnelStep, error)

	// Вспомогательные
	Ping(ctx context.Context) error
	Close() error
}
