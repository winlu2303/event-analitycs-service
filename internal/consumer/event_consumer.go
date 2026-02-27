package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/event-analytics-service/internal/metrics"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/repository"
)

type EventConsumer struct {
	reader    *kafka.Reader
	eventRepo repository.EventRepository
	redisRepo *repository.RedisRepository
	workers   int
	stopChan  chan struct{}
	wg        sync.WaitGroup
	metrics   *metrics.Metrics
}

func NewEventConsumer(
	brokers []string,
	topic string,
	groupID string,
	eventRepo repository.EventRepository,
	redisRepo *repository.RedisRepository,
	workers int,
	metrics *metrics.Metrics,
) *EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		Topic:           topic,
		GroupID:         groupID,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
		CommitInterval:  time.Second,
		StartOffset:     kafka.LastOffset,
	})

	return &EventConsumer{
		reader:    reader,
		eventRepo: eventRepo,
		redisRepo: redisRepo,
		workers:   workers,
		stopChan:  make(chan struct{}),
		metrics:   metrics,
	}
}

func (c *EventConsumer) Start() {
	log.Printf("Starting Kafka consumer with %d workers", c.workers)

	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}
}

func (c *EventConsumer) worker(id int) {
	defer c.wg.Done()
	log.Printf("Worker %d started", id)

	batchSize := 100
	batch := make([]*models.Event, 0, batchSize)
	batchTicker := time.NewTicker(100 * time.Millisecond)
	defer batchTicker.Stop()

	for {
		select {
		case <-c.stopChan:
			// Финальная обработка батча перед остановкой
			if len(batch) > 0 {
				c.processBatch(batch)
			}
			log.Printf("Worker %d stopped", id)
			return

		case <-batchTicker.C:
			if len(batch) > 0 {
				c.processBatch(batch)
				batch = make([]*models.Event, 0, batchSize)
			}

		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg, err := c.reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err != context.DeadlineExceeded {
					log.Printf("Worker %d error reading message: %v", id, err)
				}
				continue
			}

			var event models.Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Worker %d failed to unmarshal event: %v", id, err)
				c.metrics.Increment("consumer.errors.unmarshal")
				continue
			}

			// Добавляем метаданные из Kafka
			event.KafkaMetadata.Offset = msg.Offset
			event.KafkaMetadata.Partition = msg.Partition
			event.KafkaMetadata.ConsumedAt = time.Now()

			batch = append(batch, &event)
			c.metrics.Increment("consumer.events.received")

			if len(batch) >= batchSize {
				c.processBatch(batch)
				batch = make([]*models.Event, 0, batchSize)
			}
		}
	}
}

func (c *EventConsumer) processBatch(events []*models.Event) {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Пытаемся вставить батч в ClickHouse
	if err := c.eventRepo.InsertEventBatch(ctx, events); err != nil {
		log.Printf("Failed to insert event batch: %v", err)
		c.metrics.Increment("consumer.errors.insert")

		// Сохраняем в Redis для повторной обработки
		c.saveFailedBatch(events)
		return
	}

	// Обновляем кэш в Redis
	c.updateCache(events)

	// Обновляем метрики
	duration := time.Since(startTime)
	c.metrics.Timing("consumer.batch.processing_time", duration)
	c.metrics.IncrementBy("consumer.events.processed", int64(len(events)))

	log.Printf("Processed batch of %d events in %v", len(events), duration)
}

func (c *EventConsumer) saveFailedBatch(events []*models.Event) {
	ctx := context.Background()
	key := "failed_events:" + time.Now().Format("20060102")

	for _, event := range events {
		data, _ := json.Marshal(event)
		c.redisRepo.Client.RPush(ctx, key, data)
	}
	c.redisRepo.Client.Expire(ctx, key, 24*time.Hour)
}

func (c *EventConsumer) updateCache(events []*models.Event) {
	ctx := context.Background()

	for _, event := range events {
		// Обновляем счетчик событий за последний час
		hourKey := "stats:hourly:" + event.Timestamp.Format("2006-01-02-15")
		c.redisRepo.Client.HIncrBy(ctx, hourKey, string(event.EventType), 1)
		c.redisRepo.Client.Expire(ctx, hourKey, 48*time.Hour)

		// Обновляем уникальных пользователей
		userKey := "users:daily:" + event.Timestamp.Format("2006-01-02")
		c.redisRepo.Client.PFAdd(ctx, userKey, event.UserID)
		c.redisRepo.Client.Expire(ctx, userKey, 7*24*time.Hour)
	}
}

func (c *EventConsumer) Stop() {
	log.Println("Stopping Kafka consumer...")
	close(c.stopChan)
	c.wg.Wait()
	c.reader.Close()
	log.Println("Kafka consumer stopped")
}
