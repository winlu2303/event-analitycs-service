package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/repository"
)

type ExportService struct {
	eventRepo repository.EventRepository
}

func NewExportService(eventRepo repository.EventRepository) *ExportService {
	return &ExportService{
		eventRepo: eventRepo,
	}
}

func (s *ExportService) ExportToCSV(ctx context.Context, filter models.StatsRequest) ([]byte, error) {
	// Получаем события
	events, err := s.eventRepo.GetEventsByType(ctx, filter.EventType,
		parseDate(filter.StartDate), parseDate(filter.EndDate))
	if err != nil {
		return nil, err
	}

	// Создаем буфер для CSV
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Записываем заголовки
	headers := []string{"ID", "UserID", "EventType", "PageURL", "Timestamp", "UserAgent", "IPAddress"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Записываем данные
	for _, event := range events {
		record := []string{
			event.ID,
			event.UserID,
			string(event.EventType),
			event.PageURL,
			event.Timestamp.Format(time.RFC3339),
			event.UserAgent,
			event.IPAddress,
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func (s *ExportService) ExportToJSON(ctx context.Context, filter models.StatsRequest) ([]byte, error) {
	events, err := s.eventRepo.GetEventsByType(ctx, filter.EventType,
		parseDate(filter.StartDate), parseDate(filter.EndDate))
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(events, "", "  ")
}

func (s *ExportService) ExportAggregatedCSV(ctx context.Context, filter models.StatsRequest) ([]byte, error) {
	stats, err := s.eventRepo.GetStats(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Заголовки для агрегированных данных
	headers := []string{"TimeBucket", "EventType", "Count"}
	writer.Write(headers)

	for _, stat := range stats {
		record := []string{
			stat.TimeBucket,
			stat.EventType,
			fmt.Sprintf("%d", stat.Count),
		}
		writer.Write(record)
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
