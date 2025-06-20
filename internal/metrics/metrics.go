package metrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer represents a server that exposes Prometheus metrics
type MetricsServer struct {
	registry *prometheus.Registry
	server   *http.Server
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	registry := prometheus.NewRegistry()
	
	// Register the Go collector (collects runtime metrics about the Go process)
	registry.MustRegister(prometheus.NewGoCollector())
	
	// Register the process collector (collects metrics about the process)
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	
	return &MetricsServer{
		registry: registry,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}
}

// Registry returns the Prometheus registry
func (m *MetricsServer) Registry() *prometheus.Registry {
	return m.registry
}

// Start starts the metrics server
func (m *MetricsServer) Start() {
	mux := http.NewServeMux()
	
	// Register the metrics handler
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	
	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	m.server.Handler = mux
	
	go func() {
		log.Printf("Starting metrics server on %s", m.server.Addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting metrics server: %v", err)
		}
	}()
}

// Stop stops the metrics server
func (m *MetricsServer) Stop() error {
	log.Println("Stopping metrics server")
	return m.server.Close()
}

// SensorProducerMetrics holds metrics for the sensor producer
type SensorProducerMetrics struct {
	SensorReadingsTotal    prometheus.Counter
	SensorReadingErrors    prometheus.Counter
	SensorReadingBytes     prometheus.Counter
	SensorReadingLatency   prometheus.Histogram
	ActiveSensors          prometheus.Gauge
}

// NewSensorProducerMetrics creates a new set of sensor producer metrics
func NewSensorProducerMetrics(registry prometheus.Registerer) *SensorProducerMetrics {
	metrics := &SensorProducerMetrics{
		SensorReadingsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "sensor_producer",
			Name:      "readings_total",
			Help:      "Total number of sensor readings produced",
		}),
		SensorReadingErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "sensor_producer",
			Name:      "reading_errors_total",
			Help:      "Total number of sensor reading errors",
		}),
		SensorReadingBytes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "sensor_producer",
			Name:      "reading_bytes_total",
			Help:      "Total number of bytes produced",
		}),
		SensorReadingLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "iot",
			Subsystem: "sensor_producer",
			Name:      "reading_latency_seconds",
			Help:      "Latency of sensor reading production in seconds",
			Buckets:   prometheus.DefBuckets,
		}),
		ActiveSensors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iot",
			Subsystem: "sensor_producer",
			Name:      "active_sensors",
			Help:      "Number of active sensors",
		}),
	}
	
	registry.MustRegister(
		metrics.SensorReadingsTotal,
		metrics.SensorReadingErrors,
		metrics.SensorReadingBytes,
		metrics.SensorReadingLatency,
		metrics.ActiveSensors,
	)
	
	return metrics
}

// AnomalyDetectorMetrics holds metrics for the anomaly detector
type AnomalyDetectorMetrics struct {
	MessagesProcessedTotal prometheus.Counter
	AlertsGeneratedTotal   prometheus.Counter
	DLTMessagesTotal       prometheus.Counter
	ProcessingLatency      prometheus.Histogram
	ConsumerLag            prometheus.Gauge
}

// NewAnomalyDetectorMetrics creates a new set of anomaly detector metrics
func NewAnomalyDetectorMetrics(registry prometheus.Registerer) *AnomalyDetectorMetrics {
	metrics := &AnomalyDetectorMetrics{
		MessagesProcessedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "anomaly_detector",
			Name:      "messages_processed_total",
			Help:      "Total number of messages processed",
		}),
		AlertsGeneratedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "anomaly_detector",
			Name:      "alerts_generated_total",
			Help:      "Total number of alerts generated",
		}),
		DLTMessagesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "anomaly_detector",
			Name:      "dlt_messages_total",
			Help:      "Total number of messages sent to DLT",
		}),
		ProcessingLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "iot",
			Subsystem: "anomaly_detector",
			Name:      "processing_latency_seconds",
			Help:      "Latency of message processing in seconds",
			Buckets:   prometheus.DefBuckets,
		}),
		ConsumerLag: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "iot",
			Subsystem: "anomaly_detector",
			Name:      "consumer_lag",
			Help:      "Current consumer lag (messages behind)",
		}),
	}
	
	registry.MustRegister(
		metrics.MessagesProcessedTotal,
		metrics.AlertsGeneratedTotal,
		metrics.DLTMessagesTotal,
		metrics.ProcessingLatency,
		metrics.ConsumerLag,
	)
	
	return metrics
}