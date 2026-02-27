package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/event-analytics-service/internal/models"
)

type RedisRepository struct {
	Client *redis.Client
}

func NewRedisRepository(addr, password string, db int) *RedisRepository {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &RedisRepository{
		Client: client,
	}
}

// Кэширование статистики
func (r *RedisRepository) CacheStats(ctx context.Context, key string, stats []models.EventStats, ttl time.Duration) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, "stats:"+key, data, ttl).Err()
}

func (r *RedisRepository) GetCachedStats(ctx context.Context, key string) ([]models.EventStats, error) {
	data, err := r.Client.Get(ctx, "stats:"+key).Bytes()
	if err == redis.Nil {
		return nil, nil // Кэш пуст
	}
	if err != nil {
		return nil, err
	}

	var stats []models.EventStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Очередь для Retry
func (r *RedisRepository) PushToRetryQueue(ctx context.Context, event *models.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.Client.LPush(ctx, "retry_queue", data).Err()
}

func (r *RedisRepository) PopFromRetryQueue(ctx context.Context) (*models.Event, error) {
	data, err := r.Client.RPop(ctx, "retry_queue").Bytes()
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// Rate limiting
func (r *RedisRepository) CheckRateLimit(ctx context.Context, projectID string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", projectID)

	pipe := r.Client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return incr.Val() <= int64(limit), nil
}

// Real-time агрегации
func (r *RedisRepository) IncrementRealtimeMetric(ctx context.Context, metric string, value int64) error {
	key := "realtime:" + metric + ":" + time.Now().Format("2006-01-02-15-04")
	return r.Client.IncrBy(ctx, key, value).Err()
}

func (r *RedisRepository) GetRealtimeMetrics(ctx context.Context, metric string, minutes int) (map[string]int64, error) {
	pattern := "realtime:" + metric + ":*"
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, key := range keys {
		val, err := r.Client.Get(ctx, key).Int64()
		if err != nil {
			continue
		}
		result[key] = val
	}

	return result, nil
}

// Проверка соединения
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisRepository) Close() error {
	return r.Client.Close()
}
