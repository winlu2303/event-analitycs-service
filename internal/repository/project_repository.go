package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yourusername/event-analytics-service/internal/models"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *models.Project) error
	GetProjectByID(ctx context.Context, id string) (*models.Project, error)
	GetProjectByAPIKey(ctx context.Context, apiKey string) (*models.Project, error)
	GetProjectsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Project, error)
	UpdateProject(ctx context.Context, project *models.Project) error
	DeleteProject(ctx context.Context, id string) error
	GenerateAPIKey(ctx context.Context, projectID string) (string, error)
	ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error)
	GetProjectStats(ctx context.Context, projectID string, start, end time.Time) (*models.ProjectStats, error)
}

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	query := `
        INSERT INTO projects (id, name, api_key, user_id, settings, active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	if project.APIKey == "" {
		project.APIKey = generateAPIKey()
	}
	if project.Settings == nil {
		project.Settings = models.JSON{}
	}
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	project.Active = true

	settingsJSON, err := json.Marshal(project.Settings)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query,
		project.ID,
		project.Name,
		project.APIKey,
		project.UserID,
		settingsJSON,
		project.Active,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique violation
				return models.ErrDuplicateEntry
			}
		}
		return err
	}

	return nil
}

func (r *projectRepository) GetProjectByID(ctx context.Context, id string) (*models.Project, error) {
	query := `
        SELECT id, name, api_key, user_id, settings, active, created_at, updated_at
        FROM projects
        WHERE id = $1 AND active = true
    `

	var project models.Project
	var settings []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.APIKey,
		&project.UserID,
		&settings,
		&project.Active,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}

	// Parse settings JSON
	if len(settings) > 0 {
		if err := json.Unmarshal(settings, &project.Settings); err != nil {
			return nil, err
		}
	}

	return &project, nil
}

func (r *projectRepository) GetProjectByAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	query := `
        SELECT id, name, api_key, user_id, settings, active, created_at, updated_at
        FROM projects
        WHERE api_key = $1 AND active = true
    `

	var project models.Project
	var settings []byte

	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(
		&project.ID,
		&project.Name,
		&project.APIKey,
		&project.UserID,
		&settings,
		&project.Active,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrInvalidAPIKey
	}
	if err != nil {
		return nil, err
	}

	if len(settings) > 0 {
		json.Unmarshal(settings, &project.Settings)
	}

	return &project, nil
}

func (r *projectRepository) GetProjectsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Project, error) {
	query := `
        SELECT id, name, api_key, user_id, settings, active, created_at, updated_at
        FROM projects
        WHERE user_id = $1 AND active = true
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		var project models.Project
		var settings []byte

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.APIKey,
			&project.UserID,
			&settings,
			&project.Active,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(settings) > 0 {
			json.Unmarshal(settings, &project.Settings)
		}

		projects = append(projects, &project)
	}

	return projects, nil
}

func (r *projectRepository) UpdateProject(ctx context.Context, project *models.Project) error {
	query := `
        UPDATE projects
        SET name = $1, settings = $2, updated_at = $3
        WHERE id = $4 AND user_id = $5 AND active = true
    `

	settings, err := json.Marshal(project.Settings)
	if err != nil {
		return err
	}

	project.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		project.Name,
		settings,
		project.UpdatedAt,
		project.ID,
		project.UserID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrProjectNotFound
	}

	return nil
}

func (r *projectRepository) DeleteProject(ctx context.Context, id string) error {
	query := `UPDATE projects SET active = false, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrProjectNotFound
	}

	return nil
}

func (r *projectRepository) GenerateAPIKey(ctx context.Context, projectID string) (string, error) {
	newKey := generateAPIKey()

	query := `UPDATE projects SET api_key = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, newKey, time.Now(), projectID)
	if err != nil {
		return "", err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return "", models.ErrProjectNotFound
	}

	return newKey, nil
}

func (r *projectRepository) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	return r.GetProjectByAPIKey(ctx, apiKey)
}

func (r *projectRepository) GetProjectStats(ctx context.Context, projectID string, start, end time.Time) (*models.ProjectStats, error) {
	// В реальном проекте здесь был бы запрос к ClickHouse
	// Для простоты вернем заглушку
	return &models.ProjectStats{
		UniqueUsers: 0,
		TotalEvents: 0,
		PageViews:   0,
		Purchases:   0,
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
	}, nil
}

// Helper function to generate API key
func generateAPIKey() string {
	return "evnt_" + uuid.New().String() + uuid.New().String()[:8]
}
