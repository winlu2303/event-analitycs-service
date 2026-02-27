package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yourusername/event-analytics-service/internal/config"
	"github.com/yourusername/event-analytics-service/internal/handler"
	"github.com/yourusername/event-analytics-service/internal/metrics"
	"github.com/yourusername/event-analytics-service/internal/middleware"
	"github.com/yourusername/event-analytics-service/internal/producer"
	"github.com/yourusername/event-analytics-service/internal/repository"
	"github.com/yourusername/event-analytics-service/internal/service"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Инициализируем метрики
	appMetrics := metrics.NewMetrics("event-analytics-api")

	// Подключаемся к ClickHouse
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

	// Подключаемся к PostgreSQL
	psqlConnStr := "host=" + cfg.PostgresHost +
		" port=" + cfg.PostgresPort +
		" user=" + cfg.PostgresUser +
		" password=" + cfg.PostgresPassword +
		" dbname=" + cfg.PostgresDB +
		" sslmode=disable"

	psqlDB, err := sql.Open("postgres", psqlConnStr)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	if err := psqlDB.Ping(); err != nil {
		log.Fatal("PostgreSQL ping failed:", err)
	}

	// Подключаемся к Redis
	redisRepo := repository.NewRedisRepository(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redisRepo.Ping(context.Background()); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Инициализируем Kafka producer
	kafkaProducer := producer.NewEventProducer(
		[]string{cfg.KafkaBroker},
		cfg.KafkaTopic,
	)
	defer kafkaProducer.Close()

	// Инициализируем репозитории
	eventRepo := repository.NewClickHouseRepository(conn)
	userRepo := repository.NewUserRepository(psqlDB)
	sessionRepo := repository.NewSessionRepository(psqlDB)
	projectRepo := repository.NewProjectRepository(psqlDB)

	// Инициализируем сервисы
	eventService := service.NewEventService(eventRepo, kafkaProducer, appMetrics)
	statsService := service.NewStatsService(eventRepo, redisRepo, appMetrics)
	exportService := service.NewExportService(eventRepo)
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.JWTSecret,
		cfg.JWTTokenExpiry,
		cfg.JWTRefreshExpiry,
	)
	projectService := service.NewProjectService(projectRepo, eventRepo, redisRepo)

	// Инициализируем хендлеры
	eventHandler := handler.NewEventHandler(eventService)
	statsHandler := handler.NewStatsHandler(statsService)
	exportHandler := handler.NewExportHandler(exportService)
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)

	// Инициализируем middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, cfg.JWTTokenExpiry)

	// Настраиваем роутер
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.Default())
	router.Use(middleware.MetricsMiddleware(appMetrics))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// Публичные Auth endpoints
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)

		// Endpoint для трекинга событий (с API ключом)
		api.POST("/events/track", authMiddleware.ValidateAPIKey(), eventHandler.TrackEvent)

		// Protected endpoints (требуют JWT токен)
		protected := api.Group("/")
		protected.Use(authMiddleware.ValidateToken())
		{
			// User endpoints
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/auth/change-password", authHandler.ChangePassword)

			// Project endpoints
			protected.POST("/projects", projectHandler.CreateProject)
			protected.GET("/projects", projectHandler.ListProjects)
			protected.GET("/projects/:id", projectHandler.GetProject)
			protected.PUT("/projects/:id", projectHandler.UpdateProject)
			protected.DELETE("/projects/:id", projectHandler.DeleteProject)
			protected.POST("/projects/:id/regenerate-key", projectHandler.RegenerateAPIKey)
			protected.GET("/projects/:id/stats", projectHandler.GetProjectStats)

			// Stats endpoints
			protected.GET("/stats/events", statsHandler.GetStatistics)
			protected.GET("/stats/conversion", statsHandler.GetConversionRate)

			// Export endpoints
			protected.GET("/export/csv", exportHandler.ExportCSV)
			protected.GET("/export/json", exportHandler.ExportJSON)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
