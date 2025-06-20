package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"time"
)

// Producer is a wrapper around IPublisher that provides the same API as internal/kafka.Producer
type Producer struct {
	publisher IPublisher
	topic     string
	metrics   *ProducerMetrics
}

// ProducerMetrics holds Prometheus metrics for the producer
type ProducerMetrics struct {
	MessagesSent   prometheus.Counter
	BytesSent      prometheus.Counter
	ErrorsTotal    prometheus.Counter
	MessageLatency prometheus.Histogram
	registry       prometheus.Registerer
}

// NewProducerMetrics creates a new set of producer metrics
func NewProducerMetrics(namespace, subsystem string, registry prometheus.Registerer) *ProducerMetrics {
	metrics := &ProducerMetrics{
		MessagesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "messages_sent_total",
			Help:      "Total number of messages sent",
		}),
		BytesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "bytes_sent_total",
			Help:      "Total number of bytes sent",
		}),
		ErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total number of errors",
		}),
		MessageLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "message_latency_seconds",
			Help:      "Latency of message production in seconds",
			Buckets:   prometheus.DefBuckets,
		}),
		registry: registry,
	}

	registry.MustRegister(
		metrics.MessagesSent,
		metrics.BytesSent,
		metrics.ErrorsTotal,
		metrics.MessageLatency,
	)

	return metrics
}

// ProducerConfig holds configuration for the producer
type ProducerConfig struct {
	Brokers         []string
	Topic           string
	RequiredAcks    sarama.RequiredAcks
	ReturnSuccesses bool
	ReturnErrors    bool
	Metrics         *ProducerMetrics
	Version         string
}

// NewProducer creates a new Kafka producer
func NewProducer(config ProducerConfig) (*Producer, error) {
	// Create options for the publisher
	opts := []OptionFunc{
		WithProducerRequiredAcks(int(config.RequiredAcks)),
		WithProducerReturnSuccesses(config.ReturnSuccesses),
	}

	// Set Kafka version if provided
	if config.Version != "" {
		opts = append(opts, WithKafkaVersion(config.Version))
	}

	// Create the publisher
	publisher, err := NewKafkaPublisher(config.Brokers, config.Topic, opts...)
	if err != nil {
		return nil, err
	}

	return &Producer{
		publisher: publisher,
		topic:     config.Topic,
		metrics:   config.Metrics,
	}, nil
}

// SendMessage sends a message to the configured topic
func (p *Producer) SendMessage(key, value []byte) {
	startTime := time.Now()

	// Publish the message
	ctx := context.Background()
	err := p.publisher.Publish(ctx, key, value)

	// Update metrics
	if p.metrics != nil {
		if err == nil {
			p.metrics.MessagesSent.Inc()
			p.metrics.BytesSent.Add(float64(len(value)))
			p.metrics.MessageLatency.Observe(time.Since(startTime).Seconds())
		} else {
			p.metrics.ErrorsTotal.Inc()
		}
	}
}

// SendMessageWithKey sends a message with the specified key to the configured topic
func (p *Producer) SendMessageWithKey(key string, value []byte) {
	startTime := time.Now()

	// Convert string key to []byte
	keyBytes := []byte(key)

	// Publish the message
	ctx := context.Background()
	err := p.publisher.Publish(ctx, keyBytes, value)

	// Update metrics
	if p.metrics != nil {
		if err == nil {
			p.metrics.MessagesSent.Inc()
			p.metrics.BytesSent.Add(float64(len(value)))
			p.metrics.MessageLatency.Observe(time.Since(startTime).Seconds())
		} else {
			p.metrics.ErrorsTotal.Inc()
		}
	}
}

// SendMessageToTopic sends a message to the specified topic
func (p *Producer) SendMessageToTopic(topic string, key, value []byte) {
	// For this adapter, we'll just use the configured topic
	// since the underlying publisher doesn't support changing topics
	log.Printf("Warning: SendMessageToTopic called with topic %s, but using configured topic %s", topic, p.topic)
	p.SendMessage(key, value)
}

// Close closes the producer
func (p *Producer) Close() error {
	p.publisher.Stop()
	return nil
}

// GracefulShutdown performs a graceful shutdown of the producer
func (p *Producer) GracefulShutdown(ctx context.Context) error {
	p.publisher.Stop()
	return nil
}

// Consumer is a wrapper around IConsumer that provides the same API as internal/kafka.Consumer
type Consumer struct {
	consumer IConsumer
	metrics  *ConsumerMetrics
}

// ConsumerMetrics holds Prometheus metrics for the consumer
type ConsumerMetrics struct {
	MessagesReceived prometheus.Counter
	BytesReceived    prometheus.Counter
	ErrorsTotal      prometheus.Counter
	ProcessingTime   prometheus.Histogram
	LagGauge         prometheus.Gauge
	registry         prometheus.Registerer
}

// NewConsumerMetrics creates a new set of consumer metrics
func NewConsumerMetrics(namespace, subsystem string, registry prometheus.Registerer) *ConsumerMetrics {
	metrics := &ConsumerMetrics{
		MessagesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "messages_received_total",
			Help:      "Total number of messages received",
		}),
		BytesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "bytes_received_total",
			Help:      "Total number of bytes received",
		}),
		ErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total number of errors",
		}),
		ProcessingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "processing_time_seconds",
			Help:      "Time taken to process messages in seconds",
			Buckets:   prometheus.DefBuckets,
		}),
		LagGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "consumer_lag",
			Help:      "Current consumer lag (messages behind)",
		}),
		registry: registry,
	}

	registry.MustRegister(
		metrics.MessagesReceived,
		metrics.BytesReceived,
		metrics.ErrorsTotal,
		metrics.ProcessingTime,
		metrics.LagGauge,
	)

	return metrics
}

// ConsumerConfig holds configuration for the consumer
type ConsumerConfig struct {
	Brokers         []string
	GroupID         string
	Topics          []string
	OffsetInitial   int64
	ReturnErrors    bool
	Metrics         *ConsumerMetrics
	Version         string
	BalanceStrategy string
}

// MessageHandler is a function that processes a Kafka message
type MessageHandler func(message *sarama.ConsumerMessage) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(config ConsumerConfig, handler MessageHandler) (*Consumer, error) {
	// We need to adapt the handler function to match the expected signature
	adaptedHandler := func(ctx context.Context, message *sarama.ConsumerMessage) error {
		startTime := time.Now()

		// Update metrics before processing
		if config.Metrics != nil {
			config.Metrics.MessagesReceived.Inc()
			config.Metrics.BytesReceived.Add(float64(len(message.Value)))
		}

		// Call the original handler
		err := handler(message)

		// Update metrics after processing
		if config.Metrics != nil {
			config.Metrics.ProcessingTime.Observe(time.Since(startTime).Seconds())
			if err != nil {
				config.Metrics.ErrorsTotal.Inc()
			}
		}

		return err
	}

	// Create options for the consumer
	opts := []OptionFunc{
		WithConsumerReturnErrors(config.ReturnErrors),
		WithConsumerOffsetsInitial(config.OffsetInitial),
	}

	// Set Kafka version if provided
	if config.Version != "" {
		opts = append(opts, WithKafkaVersion(config.Version))
	}

	// Set balance strategy if provided
	if config.BalanceStrategy != "" {
		strategy := GetBalanceStrategy(config.BalanceStrategy)
		opts = append(opts, WithConsumerGroupRebalanceStrategy(strategy))
	}

	// Since the original consumer only supports a single topic, we'll use the first one
	topic := ""
	if len(config.Topics) > 0 {
		topic = config.Topics[0]
		if len(config.Topics) > 1 {
			log.Printf("Warning: Multiple topics provided, but only using the first one: %s", topic)
		}
	}

	// Create the consumer
	consumer, err := NewKafkaConsumer(
		config.Brokers,
		topic,
		config.GroupID,
		adaptedHandler,
		DefaultWorkerPoolSize,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		metrics:  config.Metrics,
	}, nil
}

// Start starts consuming messages
func (c *Consumer) Start() error {
	return c.consumer.Start()
}

// StartWithSignalHandler starts consuming messages and sets up a signal handler for graceful shutdown
func (c *Consumer) StartWithSignalHandler() error {
	return c.Start()
}

// Stop stops the consumer
func (c *Consumer) Stop() {
	c.consumer.Stop()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// This is not used in the adapter, as the underlying consumer handles this
	return nil
}
