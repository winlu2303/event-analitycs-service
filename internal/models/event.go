package models

import (
	"time"
)

type EventType string

const (
	PageView    EventType = "page_view"
	ButtonClick EventType = "button_click"
	FormSubmit  EventType = "form_submit"
	Purchase    EventType = "purchase"
	Custom      EventType = "custom"
)

type Event struct {
	ID        string                 `json:"id" db:"id"`
	ProjectID string                 `json:"project_id" db:"project_id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	EventType EventType              `json:"event_type" db:"event_type"`
	PageURL   string                 `json:"page_url" db:"page_url"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`

	// Kafka metadata
	KafkaMetadata KafkaMetadata `json:"-" db:"-"`
}

type KafkaMetadata struct {
	Topic      string    `json:"topic"`
	Partition  int       `json:"partition"`
	Offset     int64     `json:"offset"`
	ProducedAt time.Time `json:"produced_at"`
	ConsumedAt time.Time `json:"consumed_at"`
}

type StatsRequest struct {
	EventType EventType `form:"event_type"`
	StartDate string    `form:"start_date"`
	EndDate   string    `form:"end_date"`
	GroupBy   string    `form:"group_by"` // hour, day, month
}

type EventStats struct {
	TimeBucket string `json:"time_bucket" db:"time_bucket"`
	EventType  string `json:"event_type" db:"event_type"`
	Count      int64  `json:"count" db:"count"`
}
