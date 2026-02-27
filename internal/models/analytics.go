package models

import "time"

// PageStat represents page view statistics
type PageStat struct {
	PageURL     string `json:"page_url" db:"page_url"`
	Views       int64  `json:"views" db:"views"`
	UniqueUsers int64  `json:"unique_users" db:"unique_users"`
}

// UserSession represents a user session for analytics
type UserSession struct {
	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	EventCount int       `json:"event_count"`
	Pages      []string  `json:"pages"`
}

// FunnelStep represents a step in conversion funnel
type FunnelStep struct {
	StepName   string  `json:"step_name"`
	EventType  string  `json:"event_type"`
	UserCount  int64   `json:"user_count"`
	Conversion float64 `json:"conversion_rate"`
}

// ProjectStats represents project statistics
type ProjectStats struct {
	UniqueUsers int64     `json:"unique_users"`
	TotalEvents int64     `json:"total_events"`
	PageViews   int64     `json:"page_views"`
	Purchases   int64     `json:"purchases"`
	FirstEvent  time.Time `json:"first_event"`
	LastEvent   time.Time `json:"last_event"`
}
