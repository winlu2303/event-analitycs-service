-- Создаем базу данных
CREATE DATABASE IF NOT EXISTS analytics;

-- Используем базу данных
USE analytics;

-- Таблица для событий
CREATE TABLE IF NOT EXISTS events (
    -- Основные поля
    id String,
    project_id String,
    user_id String,
    event_type Enum8(
        'page_view' = 1,
        'button_click' = 2,
        'form_submit' = 3,
        'purchase' = 4,
        'custom' = 5
    ),
    
    -- Детали события
    page_url String,
    referrer String,
    
    -- Данные
    metadata JSON,
    properties Map(String, String),
    
    -- Техническая информация
    user_agent String,
    ip_address String,
    country_code FixedString(2),
    device_type String,
    browser String,
    os String,
    
    -- Временные метки
    timestamp DateTime64(3),
    processed_at DateTime64(3) DEFAULT now64(),
    
    -- Kafka метаданные (для отладки)
    kafka_offset UInt64,
    kafka_partition UInt16
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp, event_type)
TTL timestamp + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

-- Материализованное представление для ежедневной статистики по проектам
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_project_stats
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (project_id, day, event_type)
AS SELECT
    project_id,
    toDate(timestamp) as day,
    event_type,
    count() as events_count,
    uniq(user_id) as unique_users
FROM events
GROUP BY project_id, day, event_type;

-- Материализованное представление для почасовой статистики
CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_project_stats
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (project_id, hour, event_type)
AS SELECT
    project_id,
    toStartOfHour(timestamp) as hour,
    event_type,
    count() as events_count,
    uniq(user_id) as unique_users
FROM events
GROUP BY project_id, hour, event_type;

-- Таблица для уникальных пользователей (приблизительные значения для быстрых подсчетов)
CREATE TABLE IF NOT EXISTS daily_unique_users (
    project_id String,
    date Date,
    users AggregateFunction(uniq, String)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date);

-- Материализованное представление для заполнения уникальных пользователей
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_unique_users_mv
TO daily_unique_users
AS SELECT
    project_id,
    toDate(timestamp) as date,
    uniqState(user_id) as users
FROM events
GROUP BY project_id, date;

-- Таблица для топ страниц
CREATE TABLE IF NOT EXISTS top_pages (
    project_id String,
    date Date,
    page_url String,
    views AggregateFunction(count)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date, page_url);

-- Материализованное представление для топ страниц
CREATE MATERIALIZED VIEW IF NOT EXISTS top_pages_mv
TO top_pages
AS SELECT
    project_id,
    toDate(timestamp) as date,
    page_url,
    countState() as views
FROM events
WHERE event_type = 'page_view'
GROUP BY project_id, date, page_url;

-- Таблица для воронок конверсии
CREATE TABLE IF NOT EXISTS conversion_funnels (
    project_id String,
    date Date,
    step_name String,
    step_order UInt8,
    user_count AggregateFunction(uniq, String)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date, step_order);

-- Системная таблица для мониторинга
CREATE TABLE IF NOT EXISTS system_metrics (
    metric_name String,
    metric_value Float64,
    tags Map(String, String),
    timestamp DateTime
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (metric_name, timestamp);

-- Создаем пользователя для приложения
CREATE USER IF NOT EXISTS analytics_app IDENTIFIED WITH plaintext_password BY 'secure_password';
GRANT SELECT, INSERT ON analytics.* TO analytics_app;
