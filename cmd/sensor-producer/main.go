package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/example/iot-sensor-fleet/internal/config"
	"github.com/example/iot-sensor-fleet/internal/kafka"
	"github.com/example/iot-sensor-fleet/internal/metrics"
	"github.com/example/iot-sensor-fleet/internal/model"
)

// Sensor represents a virtual IoT sensor
type Sensor struct {
	ID       string
	Producer *kafka.Producer
	Interval time.Duration
	Metrics  *metrics.SensorProducerMetrics
	stopCh   chan struct{}
}

// NewSensor creates a new virtual sensor
func NewSensor(id string, producer *kafka.Producer, interval time.Duration, metrics *metrics.SensorProducerMetrics) *Sensor {
	return &Sensor{
		ID:       id,
		Producer: producer,
		Interval: interval,
		Metrics:  metrics,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the sensor simulation
func (s *Sensor) Start() {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Generate random sensor reading
			reading := s.generateReading()

			// Serialize the reading
			data, err := model.SerializeSensorReading(reading)
			if err != nil {
				log.Printf("Error serializing sensor reading: %v", err)
				if s.Metrics != nil {
					s.Metrics.SensorReadingErrors.Inc()
				}
				continue
			}

			// Send the reading to Kafka
			startTime := time.Now()
			s.Producer.SendMessageWithKey(reading.ID, data)

			// Update metrics
			if s.Metrics != nil {
				s.Metrics.SensorReadingsTotal.Inc()
				s.Metrics.SensorReadingBytes.Add(float64(len(data)))
				s.Metrics.SensorReadingLatency.Observe(time.Since(startTime).Seconds())
			}

		case <-s.stopCh:
			return
		}
	}
}

// Stop stops the sensor simulation
func (s *Sensor) Stop() {
	close(s.stopCh)
}

// generateReading generates a random sensor reading
func (s *Sensor) generateReading() *model.SensorReading {
	// Generate random temperature between 10°C and 60°C
	// This will occasionally generate anomalies (>50°C)
	temperature := 10.0 + rand.Float32()*50.0

	// Generate random humidity between 5% and 95%
	// This will occasionally generate anomalies (<10%)
	humidity := 5.0 + rand.Float32()*90.0

	return model.NewSensorReading(
		time.Now().UnixMilli(),
		temperature,
		humidity,
	)
}

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Schema Registry client
	model.InitSchemaRegistry(cfg.SchemaRegistryURL)

	// Create metrics server
	metricsServer := metrics.NewMetricsServer(cfg.MetricsPort)
	metricsServer.Start()
	defer metricsServer.Stop()

	// Create sensor producer metrics
	sensorMetrics := metrics.NewSensorProducerMetrics(metricsServer.Registry())

	// Create Kafka producer metrics
	producerMetrics := kafka.NewProducerMetrics("iot", "kafka_producer", metricsServer.Registry())

	// Create Kafka producer
	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:         cfg.KafkaBrokers,
		Topic:           cfg.TopicSensorRaw,
		RequiredAcks:    sarama.RequiredAcks(cfg.ProducerRequiredAcks),
		ReturnSuccesses: cfg.ProducerReturnSuccess,
		ReturnErrors:    cfg.ProducerReturnErrors,
		Metrics:         producerMetrics,
		Version:         cfg.KafkaVersion,
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// Create context with cancellation for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create wait group for sensors
	var wg sync.WaitGroup

	// Create and start sensors
	log.Printf("Starting %d sensors...", cfg.SensorCount)
	sensorMetrics.ActiveSensors.Set(float64(cfg.SensorCount))

	for i := 0; i < cfg.SensorCount; i++ {
		sensor := NewSensor(
			fmt.Sprintf("sensor-%d", i),
			producer,
			cfg.SensorInterval,
			sensorMetrics,
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			sensor.Start()
		}()
	}

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan
	log.Println("Received termination signal, shutting down...")

	// Cancel context to stop all sensors
	cancel()

	// Wait for all sensors to stop
	wg.Wait()

	// Close the producer
	if err := producer.GracefulShutdown(context.Background()); err != nil {
		log.Printf("Error during producer shutdown: %v", err)
	}

	log.Println("Sensor producer shutdown complete")
}
