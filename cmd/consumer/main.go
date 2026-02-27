package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yourusername/event-analytics-service/internal/config"
	"github.com/yourusername/event-analytics-service/internal/consumer"
	"github.com/yourusername/event-analytics-service/internal/metrics"
	"github.com/yourusername/event-analytics-service/internal/repository"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Инициализируем метрики
	appMetrics := metrics.NewMetrics("event-analytics-consumer")

	// Подключаемся к ClickHouse (исправлено: добавляем порт)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouseHost + ":" + cfg.ClickHousePort},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDB,
			Username: cfg.ClickHouseUser,
			Password: cfg.ClickHousePassword,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:     5 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse:", err)
	}

	// Подключаемся к Redis
	redisRepo := repository.NewRedisRepository(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redisRepo.Ping(context.Background()); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Инициализируем репозиторий событий
	eventRepo := repository.NewClickHouseRepository(conn)

	// Создаем Kafka consumer
	kafkaConsumer := consumer.NewEventConsumer(
		[]string{cfg.KafkaBroker},
		cfg.KafkaTopic,
		cfg.KafkaGroup,
		eventRepo,
		redisRepo,
		cfg.ConsumerWorkers,
		appMetrics,
	)

	// Запускаем HTTP сервер для метрик и health check
	go func() {
		// Создаем мультиплексор для маршрутов
		mux := http.NewServeMux()

		// Метрики Prometheus
		mux.Handle("/metrics", promhttp.Handler())

		// Health check для Docker
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		log.Printf("Consumer metrics server starting on :8081")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Запускаем consumer
	kafkaConsumer.Start()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down consumer...")
	kafkaConsumer.Stop()
	redisRepo.Close()
	conn.Close()
	log.Println("Consumer stopped")
}
