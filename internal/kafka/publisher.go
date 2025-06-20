package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"math/rand"
	"time"
)

// IPublisher defines the interface for a Kafka publisher
type IPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
	Stop()
}

// kafkaPublisher implements the IPublisher interface
type kafkaPublisher struct {
	brokers  []string
	topic    string
	producer sarama.SyncProducer
	config   *sarama.Config
}

// NewKafkaPublisher creates a new Kafka publisher
func NewKafkaPublisher(brokers []string, topic string, opts ...OptionFunc) (IPublisher, error) {
	config := sarama.NewConfig()

	// Set default values
	config.Producer.RequiredAcks = sarama.RequiredAcks(DefaultRequiredAcks)
	config.Producer.Return.Successes = DefaultProducerReturnSucc
	config.Producer.Retry.Max = DefaultRetryMax
	config.Producer.Retry.Backoff = time.Duration(DefaultRetryBackoff) * time.Millisecond

	// Apply options
	for _, o := range opts {
		o(config)
	}

	// Create producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &kafkaPublisher{
		brokers:  brokers,
		topic:    topic,
		producer: producer,
		config:   config,
	}, nil
}

// Publish sends a message to Kafka with retry logic
func (p *kafkaPublisher) Publish(ctx context.Context, key, value []byte) error {
	// Create producer message with default topic and provided key and value
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	// Simple retry mechanism with exponential backoff
	maxRetries := 3
	maxWait := 2 * time.Minute
	deadline := time.Now().Add(maxWait)

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// Check if context is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Try to send the message
			_, _, err := p.producer.SendMessage(msg)
			if err == nil {
				return nil // Success
			}

			lastErr = err

			// Check if we've exceeded the deadline
			if time.Now().After(deadline) {
				break
			}

			// Calculate backoff time (exponential with jitter)
			backoffTime := time.Duration(100*(1<<i)) * time.Millisecond
			// Add some jitter (±20%)
			jitter := time.Duration(float64(backoffTime) * (0.8 + 0.4*rand.Float64()))

			// Wait before retrying
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(jitter):
				// Continue with next retry
			}
		}
	}

	return fmt.Errorf("failed to publish message after retries: %w", lastErr)
}

// Stop closes the producer
func (p *kafkaPublisher) Stop() {
	if err := p.producer.Close(); err != nil {
		log.Printf("Failed to close Kafka producer: %v", err)
	}
}
