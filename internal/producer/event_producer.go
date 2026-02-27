package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/event-analytics-service/internal/models"
)

type EventProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewEventProducer(brokers []string, topic string) *EventProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        true,
	})

	return &EventProducer{
		writer: writer,
		topic:  topic,
	}
}

func (p *EventProducer) SendEvent(ctx context.Context, event *models.Event) error {
	// Добавляем метаданные для Kafka
	event.KafkaMetadata = models.KafkaMetadata{
		ProducedAt: time.Now(),
		Partition:  0, // будет установлено Kafka
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.UserID), // Партицирование по UserID
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "project_id", Value: []byte(event.ProjectID)},
		},
		Time: time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *EventProducer) SendBatch(ctx context.Context, events []*models.Event) error {
	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(event.UserID),
			Value: data,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(event.EventType)},
			},
		})
	}

	return p.writer.WriteMessages(ctx, messages...)
}

func (p *EventProducer) Close() error {
	return p.writer.Close()
}
