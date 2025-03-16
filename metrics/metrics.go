package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector – интерфейс для сбора метрик.
type MetricsCollector interface {
	IncMessageSent()
	ObserveMessageLatency(latency float64)
	IncErrorCount()
}

// PrometheusMetrics – реализация MetricsCollector с помощью Prometheus.
type PrometheusMetrics struct {
	messageSentCounter   prometheus.Counter
	messageLatencyHist   prometheus.Histogram
	errorCounter         prometheus.Counter
}

// NewPrometheusMetrics создаёт новый экземпляр PrometheusMetrics.
func NewPrometheusMetrics() MetricsCollector {
	pm := &PrometheusMetrics{
		messageSentCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "bot_message_sent_total",
			Help: "Общее количество отправленных сообщений ботом",
		}),
		messageLatencyHist: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "bot_message_latency_seconds",
			Help:    "Время отправки сообщения в секундах",
			Buckets: prometheus.DefBuckets,
		}),
		errorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "bot_error_total",
			Help: "Общее количество ошибок",
		}),
	}
	prometheus.MustRegister(pm.messageSentCounter, pm.messageLatencyHist, pm.errorCounter)
	return pm
}

func (pm *PrometheusMetrics) IncMessageSent() {
	pm.messageSentCounter.Inc()
}

func (pm *PrometheusMetrics) ObserveMessageLatency(latency float64) {
	pm.messageLatencyHist.Observe(latency)
}

func (pm *PrometheusMetrics) IncErrorCount() {
	pm.errorCounter.Inc()
}

// ExposeMetricsHandler возвращает HTTP-обработчик для экспонирования метрик.
func ExposeMetricsHandler(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}
