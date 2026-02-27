package service

import (
    "context"
    "fmt"
    "time"
    
    "github.com/yourusername/event-analytics-service/internal/models"
    "github.com/yourusername/event-analytics-service/internal/repository"
    "github.com/yourusername/event-analytics-service/internal/metrics"
)

type StatsService struct {
    repo    repository.EventRepository
    cache   *repository.RedisRepository
    metrics *metrics.Metrics
}

func NewStatsService(repo repository.EventRepository, cache *repository.RedisRepository, metrics *metrics.Metrics) *StatsService {
    return &StatsService{
        repo:    repo,
        cache:   cache,
        metrics: metrics,
    }
}

func (s *StatsService) GetEventStatistics(ctx context.Context, req models.StatsRequest) ([]models.EventStats, error) {
    // Пробуем получить из кэша
    cacheKey := fmt.Sprintf("stats:%s:%s:%s", req.EventType, req.StartDate, req.EndDate)
    cached, err := s.cache.GetCachedStats(ctx, cacheKey)
    if err == nil && cached != nil {
        s.metrics.IncrementCacheHit("stats")
        return cached, nil
    }
    s.metrics.IncrementCacheMiss("stats")

    // Получаем из репозитория
    stats, err := s.repo.GetStats(ctx, req)
    if err != nil {
        return nil, err
    }

    // Сохраняем в кэш на 5 минут
    s.cache.CacheStats(ctx, cacheKey, stats, 5*time.Minute)

    return stats, nil
}

func (s *StatsService) GetTopPages(ctx context.Context, projectID string, limit int) ([]models.PageStat, error) {
    // Можно добавить кэширование позже
    return s.repo.GetTopPages(ctx, projectID, limit)
}

func (s *StatsService) CalculateConversionRate(ctx context.Context, startDate, endDate string) (float64, error) {
    // Получаем статистику просмотров
    viewStats, err := s.repo.GetStats(ctx, models.StatsRequest{
        EventType: models.PageView,
        StartDate: startDate,
        EndDate:   endDate,
    })
    if err != nil {
        return 0, err
    }

    // Получаем статистику покупок
    purchaseStats, err := s.repo.GetStats(ctx, models.StatsRequest{
        EventType: models.Purchase,
        StartDate: startDate,
        EndDate:   endDate,
    })
    if err != nil {
        return 0, err
    }

    var totalViews, totalPurchases int64
    for _, stat := range viewStats {
        totalViews += stat.Count
    }
    for _, stat := range purchaseStats {
        totalPurchases += stat.Count
    }

    if totalViews == 0 {
        return 0, nil
    }

    return float64(totalPurchases) / float64(totalViews) * 100, nil
}
