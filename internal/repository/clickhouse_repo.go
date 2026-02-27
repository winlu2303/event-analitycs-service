package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/yourusername/event-analytics-service/internal/models"
)

type clickHouseRepo struct {
	conn clickhouse.Conn
}

func NewClickHouseRepository(conn clickhouse.Conn) EventRepository {
	return &clickHouseRepo{
		conn: conn,
	}
}

func (r *clickHouseRepo) InsertEvent(ctx context.Context, event *models.Event) error {
	query := `
        INSERT INTO events (
            id, project_id, user_id, event_type, page_url, 
            metadata, user_agent, ip_address, timestamp
        ) VALUES (
            ?, ?, ?, ?, ?, ?, ?, ?, ?
        )
    `

	return r.conn.Exec(ctx, query,
		event.ID,
		event.ProjectID,
		event.UserID,
		event.EventType,
		event.PageURL,
		event.Metadata,
		event.UserAgent,
		event.IPAddress,
		event.Timestamp,
	)
}

func (r *clickHouseRepo) InsertEventBatch(ctx context.Context, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return err
	}

	for _, event := range events {
		err := batch.Append(
			event.ID,
			event.ProjectID,
			event.UserID,
			event.EventType,
			event.PageURL,
			event.Metadata,
			event.UserAgent,
			event.IPAddress,
			event.Timestamp,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func (r *clickHouseRepo) GetStats(ctx context.Context, filter models.StatsRequest) ([]models.EventStats, error) {
	var timeFormat string
	switch filter.GroupBy {
	case "hour":
		timeFormat = "toStartOfHour(timestamp)"
	case "day":
		timeFormat = "toStartOfDay(timestamp)"
	case "month":
		timeFormat = "toStartOfMonth(timestamp)"
	default:
		timeFormat = "toStartOfDay(timestamp)"
	}

	query := fmt.Sprintf(`
        SELECT 
            %s as time_bucket,
            event_type,
            count() as count
        FROM events
        WHERE event_type = ?
        AND timestamp BETWEEN ? AND ?
        GROUP BY time_bucket, event_type
        ORDER BY time_bucket ASC
    `, timeFormat)

	startDate, _ := time.Parse("2006-01-02", filter.StartDate)
	endDate, _ := time.Parse("2006-01-02", filter.EndDate)

	rows, err := r.conn.Query(ctx, query, filter.EventType, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.EventStats
	for rows.Next() {
		var stat models.EventStats
		if err := rows.Scan(&stat.TimeBucket, &stat.EventType, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *clickHouseRepo) GetStatsByProject(ctx context.Context, projectID string, filter models.StatsRequest) ([]models.EventStats, error) {
	var timeFormat string
	switch filter.GroupBy {
	case "hour":
		timeFormat = "toStartOfHour(timestamp)"
	case "day":
		timeFormat = "toStartOfDay(timestamp)"
	case "month":
		timeFormat = "toStartOfMonth(timestamp)"
	default:
		timeFormat = "toStartOfDay(timestamp)"
	}

	query := fmt.Sprintf(`
        SELECT 
            %s as time_bucket,
            event_type,
            count() as count
        FROM events
        WHERE project_id = ?
        AND event_type = ?
        AND timestamp BETWEEN ? AND ?
        GROUP BY time_bucket, event_type
        ORDER BY time_bucket ASC
    `, timeFormat)

	startDate, _ := time.Parse("2006-01-02", filter.StartDate)
	endDate, _ := time.Parse("2006-01-02", filter.EndDate)

	rows, err := r.conn.Query(ctx, query, projectID, filter.EventType, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.EventStats
	for rows.Next() {
		var stat models.EventStats
		if err := rows.Scan(&stat.TimeBucket, &stat.EventType, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *clickHouseRepo) GetEventsByUser(ctx context.Context, userID string, limit, offset int) ([]models.Event, error) {
	query := `
        SELECT id, project_id, user_id, event_type, page_url, metadata, user_agent, ip_address, timestamp
        FROM events
        WHERE user_id = ?
        ORDER BY timestamp DESC
        LIMIT ? OFFSET ?
    `

	rows, err := r.conn.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.ID,
			&event.ProjectID,
			&event.UserID,
			&event.EventType,
			&event.PageURL,
			&event.Metadata,
			&event.UserAgent,
			&event.IPAddress,
			&event.Timestamp,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *clickHouseRepo) GetEventsByType(ctx context.Context, eventType models.EventType, start, end time.Time) ([]models.Event, error) {
	query := `
        SELECT id, project_id, user_id, event_type, page_url, metadata, user_agent, ip_address, timestamp
        FROM events
        WHERE event_type = ?
        AND timestamp BETWEEN ? AND ?
        ORDER BY timestamp DESC
    `

	rows, err := r.conn.Query(ctx, query, eventType, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.ID,
			&event.ProjectID,
			&event.UserID,
			&event.EventType,
			&event.PageURL,
			&event.Metadata,
			&event.UserAgent,
			&event.IPAddress,
			&event.Timestamp,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *clickHouseRepo) GetTopPages(ctx context.Context, projectID string, limit int) ([]models.PageStat, error) {
	query := `
        SELECT 
            page_url,
            count() as views,
            uniq(user_id) as unique_users
        FROM events
        WHERE project_id = ?
        AND event_type = 'page_view'
        GROUP BY page_url
        ORDER BY views DESC
        LIMIT ?
    `

	rows, err := r.conn.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PageStat
	for rows.Next() {
		var stat models.PageStat
		if err := rows.Scan(&stat.PageURL, &stat.Views, &stat.UniqueUsers); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *clickHouseRepo) GetUserSessions(ctx context.Context, userID string, sessionTimeout time.Duration) ([]models.UserSession, error) {
	query := `
        SELECT 
            groupArray(page_url) as pages,
            min(timestamp) as session_start,
            max(timestamp) as session_end,
            count() as event_count
        FROM events
        WHERE user_id = ?
        GROUP BY toStartOfInterval(timestamp, INTERVAL ? MINUTE)
        ORDER BY session_start
    `

	rows, err := r.conn.Query(ctx, query, userID, int(sessionTimeout.Minutes()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.UserSession
	sessionID := 1
	for rows.Next() {
		var session models.UserSession
		var pages []string

		err := rows.Scan(
			&pages,
			&session.StartTime,
			&session.EndTime,
			&session.EventCount,
		)
		if err != nil {
			return nil, err
		}

		session.SessionID = fmt.Sprintf("session_%d", sessionID)
		session.UserID = userID
		session.Pages = pages
		sessionID++

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *clickHouseRepo) GetFunnelAnalysis(ctx context.Context, projectID string, steps []models.EventType, start, end time.Time) ([]models.FunnelStep, error) {
	var funnelSteps []models.FunnelStep

	for i, step := range steps {
		query := `
            SELECT COUNT(DISTINCT user_id)
            FROM events
            WHERE project_id = ?
            AND event_type = ?
            AND timestamp BETWEEN ? AND ?
        `

		var userCount int64
		err := r.conn.QueryRow(ctx, query, projectID, step, start, end).Scan(&userCount)
		if err != nil {
			return nil, err
		}

		stepName := string(step)
		conversion := 0.0

		if i > 0 && funnelSteps[i-1].UserCount > 0 {
			conversion = float64(userCount) / float64(funnelSteps[i-1].UserCount) * 100
		} else if i == 0 {
			conversion = 100.0
		}

		funnelSteps = append(funnelSteps, models.FunnelStep{
			StepName:   stepName,
			EventType:  string(step),
			UserCount:  userCount,
			Conversion: conversion,
		})
	}

	return funnelSteps, nil
}

func (r *clickHouseRepo) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx)
}

func (r *clickHouseRepo) Close() error {
	return r.conn.Close()
}
