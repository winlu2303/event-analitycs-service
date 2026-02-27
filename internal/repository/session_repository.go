package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/event-analytics-service/internal/models"
)

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID string) error
	CleanupExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
        INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
	)

	return err
}

func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	query := `
        SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
        FROM sessions
        WHERE refresh_token = $1 AND expires_at > $2
    `

	var session models.Session
	err := r.db.QueryRowContext(ctx, query, refreshToken, time.Now()).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	query := `
        SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
        FROM sessions
        WHERE user_id = $1 AND expires_at > $2
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.UserAgent,
			&session.IPAddress,
			&session.ExpiresAt,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sessionRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *sessionRepository) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
