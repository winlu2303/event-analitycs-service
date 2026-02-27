package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/producer"
)

func TestKafkaProducerConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Kafka
	topic := "test-events-" + time.Now().Format("20060102150405")
	brokers := []string{"localhost:9092"}

	// Create topic
	err := createTopic(topic, brokers)
	require.NoError(t, err)

	// Create producer
	prod := producer.NewEventProducer(brokers, topic)
	defer prod.Close()

	// Create consumer
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "test-group",
		MaxWait: 1 * time.Second,
	})
	defer reader.Close()

	// Test event
	event := &models.Event{
		ID:        "test-event-1",
		UserID:    "test-user",
		EventType: models.PageView,
		PageURL:   "/test",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"test":  true,
			"value": 123,
		},
	}

	// Send event
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = prod.SendEvent(ctx, event)
	cancel()
	require.NoError(t, err)

	// Receive event
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	msg, err := reader.ReadMessage(ctx)
	cancel()
	require.NoError(t, err)

	// Verify
	var receivedEvent models.Event
	err = json.Unmarshal(msg.Value, &receivedEvent)
	require.NoError(t, err)

	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.UserID, receivedEvent.UserID)
	assert.Equal(t, event.EventType, receivedEvent.EventType)
	assert.Equal(t, event.PageURL, receivedEvent.PageURL)
}

func TestKafkaBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	topic := "test-batch-" + time.Now().Format("20060102150405")
	brokers := []string{"localhost:9092"}

	err := createTopic(topic, brokers)
	require.NoError(t, err)

	prod := producer.NewEventProducer(brokers, topic)
	defer prod.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "test-batch-group",
		MaxWait:  2 * time.Second,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// Send batch of events
	events := make([]*models.Event, 10)
	for i := 0; i < 10; i++ {
		events[i] = &models.Event{
			ID:        "test-event-" + string(rune(i)),
			UserID:    "test-user",
			EventType: models.PageView,
			PageURL:   "/test",
			Timestamp: time.Now(),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = prod.SendBatch(ctx, events)
	cancel()
	require.NoError(t, err)

	// Receive and verify all events
	receivedCount := 0
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for receivedCount < 10 {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			break
		}

		var event models.Event
		json.Unmarshal(msg.Value, &event)
		receivedCount++
	}

	assert.Equal(t, 10, receivedCount)
}

func TestKafkaConsumerGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	topic := "test-group-" + time.Now().Format("20060102150405")
	brokers := []string{"localhost:9092"}

	err := createTopic(topic, brokers)
	require.NoError(t, err)

	prod := producer.NewEventProducer(brokers, topic)
	defer prod.Close()

	// Create two consumers in same group
	reader1 := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "test-group-1",
	})
	defer reader1.Close()

	reader2 := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "test-group-1",
	})
	defer reader2.Close()

	// Send 20 events
	for i := 0; i < 20; i++ {
		event := &models.Event{
			ID:     "test-event-" + string(rune(i)),
			UserID: "test-user",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		prod.SendEvent(ctx, event)
		cancel()
	}

	// Read events from both consumers
	received := make(chan int, 2)

	go func() {
		count := 0
		for i := 0; i < 10; i++ {
			_, err := reader1.ReadMessage(context.Background())
			if err == nil {
				count++
			}
		}
		received <- count
	}()

	go func() {
		count := 0
		for i := 0; i < 10; i++ {
			_, err := reader2.ReadMessage(context.Background())
			if err == nil {
				count++
			}
		}
		received <- count
	}()

	// Verify total events processed
	total1 := <-received
	total2 := <-received

	assert.Equal(t, 20, total1+total2)
}

// Helper function to create Kafka topic
func createTopic(topic string, brokers []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", controller.Host)
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}

	return controllerConn.CreateTopics(topicConfigs...)
}
