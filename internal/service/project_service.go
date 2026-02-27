package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/repository"
)

type ProjectService struct {
	projectRepo repository.ProjectRepository
	eventRepo   repository.EventRepository
	cacheRepo   *repository.RedisRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	eventRepo repository.EventRepository,
	cacheRepo *repository.RedisRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		eventRepo:   eventRepo,
		cacheRepo:   cacheRepo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, userID string, name string, settings models.JSON) (*models.Project, error) {
	// Check project limit for user
	projects, err := s.projectRepo.GetProjectsByUserID(ctx, userID, 100, 0)
	if err != nil {
		return nil, err
	}

	if len(projects) >= 10 { // Max 10 projects per user
		return nil, models.ErrProjectLimitReached
	}

	project := &models.Project{
		Name:     name,
		UserID:   userID,
		Settings: settings,
		Active:   true,
	}

	if err := s.projectRepo.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, projectID, userID string) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Check access
	if project.UserID != userID {
		return nil, models.ErrProjectAccessDenied
	}

	return project, nil
}

func (s *ProjectService) GetUserProjects(ctx context.Context, userID string, limit, offset int) ([]*models.Project, error) {
	return s.projectRepo.GetProjectsByUserID(ctx, userID, limit, offset)
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID, userID string, updates map[string]interface{}) (*models.Project, error) {
	// Get existing project
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Check access
	if project.UserID != userID {
		return nil, models.ErrProjectAccessDenied
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		project.Name = name
	}
	if settings, ok := updates["settings"].(map[string]interface{}); ok {
		project.Settings = settings
	}

	// Save
	if err := s.projectRepo.UpdateProject(ctx, project); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cacheRepo.Client.Del(ctx, "project:"+projectID)

	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	// Check access
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.UserID != userID {
		return models.ErrProjectAccessDenied
	}

	// Delete project
	if err := s.projectRepo.DeleteProject(ctx, projectID); err != nil {
		return err
	}

	// Invalidate cache
	s.cacheRepo.Client.Del(ctx, "project:"+projectID)

	return nil
}

func (s *ProjectService) RegenerateAPIKey(ctx context.Context, projectID, userID string) (string, error) {
	// Check access
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return "", err
	}

	if project.UserID != userID {
		return "", models.ErrProjectAccessDenied
	}

	// Generate new key
	newKey, err := s.projectRepo.GenerateAPIKey(ctx, projectID)
	if err != nil {
		return "", err
	}

	return newKey, nil
}

func (s *ProjectService) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	// Try cache first
	cacheKey := "apikey:" + apiKey
	cached, err := s.cacheRepo.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		var project models.Project
		if json.Unmarshal([]byte(cached), &project) == nil {
			return &project, nil
		}
	}

	// Validate in database
	project, err := s.projectRepo.ValidateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	if data, err := json.Marshal(project); err == nil {
		s.cacheRepo.Client.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return project, nil
}

func (s *ProjectService) GetProjectStats(ctx context.Context, projectID, userID string, start, end time.Time) (*models.ProjectStats, error) {
	// Check access
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.UserID != userID {
		return nil, models.ErrProjectAccessDenied
	}

	// Try cache
	cacheKey := "stats:" + projectID + ":" + start.Format("2006-01-02") + ":" + end.Format("2006-01-02")
	cached, err := s.cacheRepo.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats models.ProjectStats
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return &stats, nil
		}
	}

	// Get from repository
	stats, err := s.projectRepo.GetProjectStats(ctx, projectID, start, end)
	if err != nil {
		return nil, err
	}

	// Cache for 1 hour
	if data, err := json.Marshal(stats); err == nil {
		s.cacheRepo.Client.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return stats, nil
}
