package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"math/rand"
	"sync"
	"time"
)

// MessageHandlerFunc defines the function for handling messages
type MessageHandlerFunc func(ctx context.Context, message *sarama.ConsumerMessage) error

// IConsumer defines the interface for a Kafka consumer
type IConsumer interface {
	Start() error
	Stop()
}

// kafkaConsumer implements both IConsumer and sarama.ConsumerGroupHandler
type kafkaConsumer struct {
	brokers       []string
	topic         string
	groupID       string
	consumerGroup sarama.ConsumerGroup
	handler       MessageHandlerFunc
	config        *sarama.Config
	workerPool    chan struct{}
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers []string, topic, groupID string, handler MessageHandlerFunc, workerPoolSize int, opts ...OptionFunc) (IConsumer, error) {
	config := sarama.NewConfig()

	// Set default values
	config.Consumer.Return.Errors = DefaultConsumerReturnErrors
	config.Consumer.Offsets.Initial = DefaultConsumerOffsetInitial

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &kafkaConsumer{
		brokers:       brokers,
		topic:         topic,
		groupID:       groupID,
		consumerGroup: consumerGroup,
		handler:       handler,
		config:        config,
		workerPool:    make(chan struct{}, workerPoolSize),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start begins consuming messages
func (c *kafkaConsumer) Start() error {
	c.wg.Add(1)
	go c.consume()
	return nil
}

// Stop stops consuming messages and closes the consumer group
func (c *kafkaConsumer) Stop() {
	c.cancel()
	c.wg.Wait()
	if err := c.consumerGroup.Close(); err != nil {
		log.Printf("Failed to close Kafka consumer group: %v", err)
	}
}

// consume runs the consumer loop
func (c *kafkaConsumer) consume() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if err := c.consumerGroup.Consume(c.ctx, []string{c.topic}, c); err != nil {
				log.Printf("Error from consumer: %v", err)
				time.Sleep(time.Second) // Wait before retrying
			}
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (c *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		select {
		case <-c.ctx.Done():
			return nil
		case c.workerPool <- struct{}{}: // Acquire worker
			c.wg.Add(1)
			go func(msg *sarama.ConsumerMessage) {
				defer c.wg.Done()
				defer func() { <-c.workerPool }() // Release worker

				c.processMessage(session, msg)
			}(message)
		}
	}
	return nil
}

// processMessage processes a single message with retry logic
func (c *kafkaConsumer) processMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	// Simple retry mechanism with exponential backoff
	var err error
	maxRetries := 3
	maxWait := 2 * time.Minute
	deadline := time.Now().Add(maxWait)

	for i := 0; i < maxRetries; i++ {
		// Check if context is done
		select {
		case <-c.ctx.Done():
			log.Printf("Context canceled while processing message")
			return
		default:
			// Try to process the message
			err = c.handler(c.ctx, msg)
			if err == nil {
				break // Success, exit the loop
			}

			// Check if we've exceeded the deadline
			if time.Now().After(deadline) {
				log.Printf("Exceeded retry deadline for message")
				break
			}

			// Calculate backoff time (exponential with jitter)
			backoffTime := time.Duration(100*(1<<i)) * time.Millisecond
			// Add some jitter (±20%)
			jitter := time.Duration(float64(backoffTime) * (0.8 + 0.4*rand.Float64()))

			log.Printf("Retrying message after %v (attempt %d/%d): %v", jitter, i+1, maxRetries, err)

			// Wait before retrying
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(jitter):
				// Continue with next retry
			}
		}
	}

	if err != nil {
		log.Printf("Failed to process message after retries: %v", err)
		// Here you could implement a Dead Letter Queue (DLQ) for failed messages
	}

	// Mark message as processed
	session.MarkMessage(msg, "")
}
