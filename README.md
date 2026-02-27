
# **Event Analytics Service** 

Сервис для сбора и анализа событий в реальном времени на Go. Проект построен на микросервисной архитектуре с использованием современных инструментов для обработки больших данных.

##  **Архитектура**
```
┌─────────────┐     ┌────────┐     ┌─────────────┐
│   Клиенты   │────▶│  API   │────▶│    Kafka    │
└─────────────┘     └────────┘     └──────┬──────┘
                                          │
                                          ▼
┌─────────────┐     ┌────────┐     ┌─────────────┐
│   Grafana   │◀────│Prometeus│◀────│  Consumer   │
└─────────────┘     └────────┘     └──────┬──────┘
                                          │
                                          ▼
                                   ┌─────────────┐
                                   │  ClickHouse │
                                   └─────────────┘
```

##  **Технологический стек**

### **Backend**
- **Go 1.21+** - основной язык разработки
- **Gin** - HTTP фреймворк
- **ClickHouse** - колоночная БД для аналитики
- **PostgreSQL** - реляционная БД для пользователей и проектов
- **Redis** - кэширование и real-time данные
- **Kafka** - очередь сообщений для асинхронной обработки

### **Мониторинг**
- **Prometheus** - сбор метрик
- **Grafana** - визуализация дашбордов

### **Инфраструктура**
- **Docker** & **Docker Compose** - контейнеризация
- **Make** - автоматизация команд

##  **Функциональность**

###  **Реализовано**
-  Аутентификация (JWT + Refresh токены)
-  Управление проектами и API ключами
-  Сбор событий (page_view, click, purchase и др.)
-  Асинхронная обработка через Kafka
-  Кэширование статистики в Redis
-  Агрегация данных в ClickHouse
-  Статистика по событиям с группировкой
-  Экспорт данных в CSV/JSON
-  Метрики Prometheus
-  Дашборды Grafana

##  **Быстрый старт**

### **Предварительные требования**
- Docker & Docker Compose
- Make
- Go 1.21+ (для локальной разработки)

### **Установка и запуск**

# 1. Клонируйте репозиторий
git clone https://github.com/yourusername/event-analytics-service.git
cd event-analytics-service

# 2. Соберите и запустите все сервисы
make docker-up

# 3. Примените миграции
make migrate-all

# 4. Проверьте, что всё работает
curl http://localhost:8080/health


##  **API Документация**

### **Аутентификация**

#### Регистрация

curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "full_name": "John Doe",
    "company": "Acme Inc"
  }'

#### Логин

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'


### **События**

#### Отправка события

curl -X POST http://localhost:8080/api/v1/events/track \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "user_id": "user123",
    "event_type": "page_view",
    "page_url": "/home",
    "metadata": {"browser": "chrome"}
  }'


### **Статистика**

#### Получение статистики

curl "http://localhost:8080/api/v1/stats/events?event_type=page_view&start_date=2024-01-01&end_date=2024-12-31&group_by=day" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"


#### Экспорт в CSV

curl "http://localhost:8080/api/v1/export/csv?event_type=page_view&start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --output events.csv


##  **Мониторинг**

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Метрики API**: http://localhost:8080/metrics
- **Метрики Consumer**: http://localhost:8081/metrics

##  **Архитектура проекта Структура**

```
event-analytics-service/
├── cmd/
│   ├── api/                        # HTTP API сервер
│   │   └── main.go     
│   └── consumer/                   # Kafka consumer
│       └── main.go           
├── internal/
│   ├── handler/                    # HTTP обработчики
│   │   ├── event_handler.go
│   │   ├── stats_handler.go
│   │   ├── auth_handler.go       
│   │   └── export_handler.go  
│   ├── models/                     # Модели данных
│   │   ├── event.go
│   │   ├── project.go            
│   │   └── errors.go             
│   ├── repository/                 # Работа с БД
│   │   ├── event_repository.go      
│   │   ├── clickhouse_repo.go       
│   │   ├── project_repository.go    
│   │   └── redis_repository.go       
│   ├── service/                    # Бизнес-логика
│   │   ├── event_service.go
│   │   ├── stats_service.go
│   │   ├── auth_service.go         
│   │   ├── project_service.go      
│   │   └── export_service.go       
│   ├── consumer/                   # Kafka consumer logic
│   │   └── event_consumer.go      
│   ├── producer/                   # Kafka producer
│   │   └── event_producer.go     
│   ├── middleware/                 # HTTP middleware
│   │   ├── auth.go          
│   │   ├── metrics.go             
│   │   └── logger.go              
│   ├── metrics/                    # Метрики Prometheus
│   │   └── prometheus.go         
│   └── config/                     # Конфигурация
│       └── config.go
├── migrations/                     # SQL миграции
│   ├── clickhouse/
│   │   └── 001_init.sql
│   └── postgres/
│       └── 001_init.sql           
├── tests/                          # Тесты
│   ├── unit/
│   │   ├── service_test.go
│   │   └── handler_test.go
│   └── integration/
│       ├── api_test.go
│       └── kafka_test.go
├── docker-compose.yml              # Docker композиция
├── Dockerfile
├── Dockerfile.consumer                  
├── prometheus.yml                          
├── grafana-dashboard.json                   
└── Makefile                        # Автоматизация
```

##  **Команды Makefile**

make docker-up        # Запустить все сервисы
make docker-down      # Остановить все сервисы
make migrate-all      # Применить миграции
make test            # Запустить тесты
make load-test       # Отправить тестовые события
make logs            # Просмотр логов


##  **Тестирование**

# Unit тесты
go test ./tests/unit/...

# Интеграционные тесты
go test ./tests/integration/...

# Все тесты с покрытием
make test-cover


##  **Производительность**

- **События**: до 10,000 событий/сек
- **Задержка**: < 100ms (p99)
- **Доступность**: 99.9%
- **Масштабирование**: горизонтальное через Kafka

##  **Вклад в проект**

1. Форкните репозиторий
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Add amazing feature'`)
4. Запушьте в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request
