package config

import (
    "log"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    // Server
    Port        string
    Environment string
    
    // ClickHouse
    ClickHouseHost     string
    ClickHousePort     string
    ClickHouseDB       string
    ClickHouseUser     string
    ClickHousePassword string
    
    // Redis
    RedisAddr     string
    RedisPassword string
    RedisDB       int
    
    // Kafka
    KafkaBroker     string
    KafkaTopic      string
    KafkaGroup      string
    ConsumerWorkers int
    
    // JWT
    JWTSecret        string
    JWTTokenExpiry   time.Duration
    JWTRefreshExpiry time.Duration
    
    // PostgreSQL
    PostgresHost     string
    PostgresPort     string
    PostgresDB       string
    PostgresUser     string
    PostgresPassword string
}

func LoadConfig() *Config {
    // Загружаем .env файл если есть
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using environment variables")
    }

    // Парсим числа
    redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
    consumerWorkers, _ := strconv.Atoi(getEnv("CONSUMER_WORKERS", "5"))
    
    // Парсим длительности
    tokenExpiry, _ := time.ParseDuration(getEnv("JWT_TOKEN_EXPIRY", "24h"))
    refreshExpiry, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "720h"))

    return &Config{
        // Server
        Port:        getEnv("PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),
        
        // ClickHouse
        ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "clickhouse"),
        ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
        ClickHouseDB:       getEnv("CLICKHOUSE_DB", "analytics"),
        ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
        ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
        
        // Redis
        RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       redisDB,
        
        // Kafka
        KafkaBroker:     getEnv("KAFKA_BROKER", "kafka:9092"),
        KafkaTopic:      getEnv("KAFKA_TOPIC", "events"),
        KafkaGroup:      getEnv("KAFKA_GROUP", "event-consumers"),
        ConsumerWorkers: consumerWorkers,
        
        // JWT
        JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
        JWTTokenExpiry:   tokenExpiry,
        JWTRefreshExpiry: refreshExpiry,
        
        // PostgreSQL
        PostgresHost:     getEnv("POSTGRES_HOST", "postgres"),
        PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
        PostgresDB:       getEnv("POSTGRES_DB", "analytics"),
        PostgresUser:     getEnv("POSTGRES_USER", "admin"),
        PostgresPassword: getEnv("POSTGRES_PASSWORD", "admin123"),
    }
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}
