package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/example/iot-sensor-fleet/internal/config"
	"github.com/example/iot-sensor-fleet/internal/db"
	"github.com/example/iot-sensor-fleet/internal/kafka"
	"github.com/example/iot-sensor-fleet/internal/metrics"
	"github.com/example/iot-sensor-fleet/internal/model"
)

// AnomalyDetector processes sensor readings and detects anomalies
type AnomalyDetector struct {
	consumer       *kafka.Consumer
	producer       *kafka.Producer
	dltProducer    *kafka.Producer
	metrics        *metrics.AnomalyDetectorMetrics
	maxTemperature float32
	minHumidity    float32
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	dltProducer *kafka.Producer,
	metrics *metrics.AnomalyDetectorMetrics,
	maxTemperature float32,
	minHumidity float32,
) *AnomalyDetector {
	return &AnomalyDetector{
		consumer:       consumer,
		producer:       producer,
		dltProducer:    dltProducer,
		metrics:        metrics,
		maxTemperature: maxTemperature,
		minHumidity:    minHumidity,
	}
}

// Start starts the anomaly detector
func (a *AnomalyDetector) Start() error {
	return a.consumer.Start()
}

// Stop stops the anomaly detector
func (a *AnomalyDetector) Stop() {
	a.consumer.Stop()
}

// handleMessage processes a message from Kafka
func (a *AnomalyDetector) handleMessage(message *sarama.ConsumerMessage) error {
	startTime := time.Now()

	// Update metrics
	if a.metrics != nil {
		a.metrics.MessagesProcessedTotal.Inc()
	}

	// Deserialize the message
	reading, err := model.DeserializeSensorReading(message.Value)
	if err != nil {
		log.Printf("Error deserializing message: %v", err)

		// Send to DLT
		if a.dltProducer != nil {
			a.dltProducer.SendMessage(message.Key, message.Value)
			if a.metrics != nil {
				a.metrics.DLTMessagesTotal.Inc()
			}
		}

		return err
	}

	// Validate the reading
	valid, reason := model.ValidateSensorReading(reading)
	if !valid {
		log.Printf("Anomaly detected: %s, sensor: %s, temp: %.1f°C, humidity: %.1f%%",
			reason, reading.ID, reading.Temperature, reading.Humidity)

		// Create alert
		alert := model.NewSensorAlert(reading, reason)

		// Serialize alert
		alertData, err := model.SerializeSensorAlert(alert)
		if err != nil {
			log.Printf("Error serializing alert: %v", err)
			return err
		}

		// Send alert to Kafka
		a.producer.SendMessageWithKey(alert.SensorID, alertData)

		// Update metrics
		if a.metrics != nil {
			a.metrics.AlertsGeneratedTotal.Inc()
		}
	}

	// Update processing latency metric
	if a.metrics != nil {
		a.metrics.ProcessingLatency.Observe(time.Since(startTime).Seconds())
	}

	return nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Schema Registry client
	model.InitSchemaRegistry(cfg.SchemaRegistryURL)

	// Initialize databases (PostgreSQL and Elasticsearch)
	log.Println("Initializing databases...")
	if _, err := db.InitDatabases(cfg); err != nil {
		log.Printf("Warning: Failed to initialize databases: %v", err)
		// Continue execution even if database initialization fails
	}

	// Create metrics server (on a different port than the producer)
	metricsPort := cfg.MetricsPort + 1 // Use port 2113 by default
	metricsServer := metrics.NewMetricsServer(metricsPort)
	metricsServer.Start()
	defer metricsServer.Stop()

	// Create anomaly detector metrics
	anomalyMetrics := metrics.NewAnomalyDetectorMetrics(metricsServer.Registry())

	// Create Kafka producer metrics for the alert producer
	alertProducerMetrics := kafka.NewProducerMetrics("iot", "alert_producer", metricsServer.Registry())

	// Create Kafka producer metrics for the DLT producer
	dltProducerMetrics := kafka.NewProducerMetrics("iot", "dlt_producer", metricsServer.Registry())

	// Create Kafka consumer metrics
	consumerMetrics := kafka.NewConsumerMetrics("iot", "sensor_consumer", metricsServer.Registry())

	// Create Kafka alert producer
	alertProducer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:         cfg.KafkaBrokers,
		Topic:           cfg.TopicSensorAlert,
		RequiredAcks:    sarama.RequiredAcks(cfg.ProducerRequiredAcks),
		ReturnSuccesses: cfg.ProducerReturnSuccess,
		ReturnErrors:    cfg.ProducerReturnErrors,
		Metrics:         alertProducerMetrics,
		Version:         cfg.KafkaVersion,
	})
	if err != nil {
		log.Fatalf("Failed to create alert producer: %v", err)
	}
	defer alertProducer.Close()

	// Create Kafka DLT producer
	dltProducer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:         cfg.KafkaBrokers,
		Topic:           cfg.TopicSensorRawDLT,
		RequiredAcks:    sarama.RequiredAcks(cfg.ProducerRequiredAcks),
		ReturnSuccesses: cfg.ProducerReturnSuccess,
		ReturnErrors:    cfg.ProducerReturnErrors,
		Metrics:         dltProducerMetrics,
		Version:         cfg.KafkaVersion,
	})
	if err != nil {
		log.Fatalf("Failed to create DLT producer: %v", err)
	}
	defer dltProducer.Close()

	// Create anomaly detector instance
	detector := NewAnomalyDetector(
		nil, // Will be set after consumer creation
		alertProducer,
		dltProducer,
		anomalyMetrics,
		cfg.MaxTemperature,
		cfg.MinHumidity,
	)

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(
		kafka.ConsumerConfig{
			Brokers:         cfg.KafkaBrokers,
			GroupID:         cfg.ConsumerGroupID,
			Topics:          []string{cfg.TopicSensorRaw},
			OffsetInitial:   cfg.ConsumerOffsetInitial,
			ReturnErrors:    cfg.ConsumerReturnErrors,
			Metrics:         consumerMetrics,
			Version:         cfg.KafkaVersion,
			BalanceStrategy: cfg.ConsumerBalanceStrategy,
		},
		detector.handleMessage,
	)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// Set the consumer in the detector
	detector.consumer = consumer

	// Start the anomaly detector
	if err := detector.Start(); err != nil {
		log.Fatalf("Failed to start anomaly detector: %v", err)
	}

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan
	log.Println("Received termination signal, shutting down...")

	// Stop the anomaly detector
	detector.Stop()

	log.Println("Anomaly detector shutdown complete")
}
