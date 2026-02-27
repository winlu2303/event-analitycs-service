package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// HTTP метрики
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Event метрики
	eventsReceived      *prometheus.CounterVec
	eventsProcessed     *prometheus.CounterVec
	eventsFailed        *prometheus.CounterVec
	eventProcessingTime *prometheus.HistogramVec

	// Kafka метрики
	kafkaMessagesPublished *prometheus.CounterVec
	kafkaMessagesConsumed  *prometheus.CounterVec
	kafkaConsumerLag       *prometheus.GaugeVec

	// Cache метрики
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec

	// Database метрики
	dbQueryDuration *prometheus.HistogramVec
	dbErrors        *prometheus.CounterVec

	// Business метрики
	activeProjects *prometheus.GaugeVec
	totalEvents    *prometheus.CounterVec
	uniqueUsers    *prometheus.GaugeVec

	// Custom counters for generic metrics
	customCounters   map[string]prometheus.Counter
	customGauges     map[string]prometheus.Gauge
	customHistograms map[string]prometheus.Histogram
}

func NewMetrics(serviceName string) *Metrics {
	m := &Metrics{
		customCounters:   make(map[string]prometheus.Counter),
		customGauges:     make(map[string]prometheus.Gauge),
		customHistograms: make(map[string]prometheus.Histogram),
	}

	// HTTP Requests
	m.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "endpoint", "status"},
	)

	m.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request duration in seconds",
			Buckets:     prometheus.DefBuckets,
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "endpoint"},
	)

	m.httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name:        "http_requests_in_flight",
			Help:        "Number of HTTP requests currently in flight",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
	)

	// Events
	m.eventsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "events_received_total",
			Help:        "Total number of events received",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"event_type", "project_id"},
	)

	m.eventsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "events_processed_total",
			Help:        "Total number of events processed",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"event_type", "project_id"},
	)

	m.eventsFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "events_failed_total",
			Help:        "Total number of events failed",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"event_type", "error_type"},
	)

	m.eventProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "event_processing_seconds",
			Help:        "Time spent processing events",
			Buckets:     []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"event_type"},
	)

	// Kafka
	m.kafkaMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "kafka_messages_published_total",
			Help:        "Total number of Kafka messages published",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"topic"},
	)

	m.kafkaMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "kafka_messages_consumed_total",
			Help:        "Total number of Kafka messages consumed",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"topic", "partition"},
	)

	m.kafkaConsumerLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "kafka_consumer_lag",
			Help:        "Current Kafka consumer lag",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"topic", "partition"},
	)

	// Cache
	m.cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "cache_hits_total",
			Help:        "Total number of cache hits",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"cache_name"},
	)

	m.cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "cache_misses_total",
			Help:        "Total number of cache misses",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"cache_name"},
	)

	// Database
	m.dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "db_query_duration_seconds",
			Help:        "Database query duration in seconds",
			Buckets:     []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"query_type", "table"},
	)

	m.dbErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db_errors_total",
			Help:        "Total number of database errors",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"error_type", "table"},
	)

	// Business metrics
	m.activeProjects = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "active_projects_total",
			Help:        "Total number of active projects",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"plan_type"},
	)

	m.totalEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "total_events_ingested",
			Help:        "Total number of events ingested over time",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"project_id"},
	)

	m.uniqueUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "unique_users_current",
			Help:        "Current number of unique users (last 5 min)",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"project_id"},
	)

	return m
}

// Helper methods
func (m *Metrics) IncrementHTTPRequest(method, endpoint, status string) {
	m.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

func (m *Metrics) ObserveHTTPDuration(method, endpoint string, duration time.Duration) {
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

func (m *Metrics) IncrementEventsReceived(eventType, projectID string) {
	m.eventsReceived.WithLabelValues(eventType, projectID).Inc()
	m.totalEvents.WithLabelValues(projectID).Inc()
}

func (m *Metrics) IncrementEventsProcessed(eventType, projectID string) {
	m.eventsProcessed.WithLabelValues(eventType, projectID).Inc()
}

func (m *Metrics) IncrementEventsFailed(eventType, errorType string) {
	m.eventsFailed.WithLabelValues(eventType, errorType).Inc()
}

func (m *Metrics) ObserveEventProcessing(eventType string, duration time.Duration) {
	m.eventProcessingTime.WithLabelValues(eventType).Observe(duration.Seconds())
}

func (m *Metrics) IncrementCacheHit(cacheName string) {
	m.cacheHits.WithLabelValues(cacheName).Inc()
}

func (m *Metrics) IncrementCacheMiss(cacheName string) {
	m.cacheMisses.WithLabelValues(cacheName).Inc()
}

func (m *Metrics) ObserveDBQuery(queryType, table string, duration time.Duration) {
	m.dbQueryDuration.WithLabelValues(queryType, table).Observe(duration.Seconds())
}

func (m *Metrics) IncrementDBError(errorType, table string) {
	m.dbErrors.WithLabelValues(errorType, table).Inc()
}

func (m *Metrics) IncrementKafkaMessagesPublished(topic string) {
	m.kafkaMessagesPublished.WithLabelValues(topic).Inc()
}

func (m *Metrics) IncrementKafkaMessagesConsumed(topic, partition string) {
	m.kafkaMessagesConsumed.WithLabelValues(topic, partition).Inc()
}

func (m *Metrics) SetKafkaConsumerLag(topic, partition string, lag float64) {
	m.kafkaConsumerLag.WithLabelValues(topic, partition).Set(lag)
}

// Generic metric methods
func (m *Metrics) Increment(name string) {
	counter, exists := m.customCounters[name]
	if !exists {
		counter = promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: "Generic counter: " + name,
		})
		m.customCounters[name] = counter
	}
	counter.Inc()
}

func (m *Metrics) IncrementBy(name string, value int64) {
	counter, exists := m.customCounters[name]
	if !exists {
		counter = promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: "Generic counter: " + name,
		})
		m.customCounters[name] = counter
	}
	counter.Add(float64(value))
}

func (m *Metrics) Timing(name string, duration time.Duration) {
	histogram, exists := m.customHistograms[name]
	if !exists {
		histogram = promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    name,
			Help:    "Generic histogram: " + name,
			Buckets: prometheus.DefBuckets,
		})
		m.customHistograms[name] = histogram
	}
	histogram.Observe(duration.Seconds())
}

func (m *Metrics) Observe(name string, value float64) {
	histogram, exists := m.customHistograms[name]
	if !exists {
		histogram = promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    name,
			Help:    "Generic histogram: " + name,
			Buckets: prometheus.DefBuckets,
		})
		m.customHistograms[name] = histogram
	}
	histogram.Observe(value)
}
