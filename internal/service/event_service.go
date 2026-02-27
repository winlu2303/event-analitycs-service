package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/event-analytics-service/internal/metrics"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/producer"
	"github.com/yourusername/event-analytics-service/internal/repository"
)

type EventService struct {
	repo     repository.EventRepository
	producer *producer.EventProducer
	metrics  *metrics.Metrics
}

func NewEventService(repo repository.EventRepository, producer *producer.EventProducer, metrics *metrics.Metrics) *EventService {
	return &EventService{
		repo:     repo,
		producer: producer,
		metrics:  metrics,
	}
}

func (s *EventService) ProcessEvent(ctx context.Context, event *models.Event) error {
	// Обогащаем событие
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Базовая валидация
	if event.EventType == "" {
		event.EventType = models.PageView
	}

	// Отправляем в Kafka асинхронно
	if err := s.producer.SendEvent(ctx, event); err != nil {
		s.metrics.Increment("event.producer.errors")
		return err
	}

	s.metrics.IncrementEventsReceived(string(event.EventType), event.ProjectID)

	return nil
}
