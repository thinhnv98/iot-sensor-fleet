package kafka

import (
	"github.com/IBM/sarama"
	"time"
)

// OptionFunc defines a function type for configuring Kafka
type OptionFunc func(*sarama.Config)

// Producer options

// WithProducerRetryMax sets the maximum number of times to retry sending a message
func WithProducerRetryMax(max int) OptionFunc {
	return func(config *sarama.Config) {
		config.Producer.Retry.Max = max
	}
}

// WithProducerRetryBackoff sets the backoff time between retries in milliseconds
func WithProducerRetryBackoff(backoff int) OptionFunc {
	return func(config *sarama.Config) {
		config.Producer.Retry.Backoff = time.Duration(backoff) * time.Millisecond
	}
}

// WithProducerReturnSuccesses configures the producer to return successes
func WithProducerReturnSuccesses(isReturnSuccesses bool) OptionFunc {
	return func(config *sarama.Config) {
		config.Producer.Return.Successes = isReturnSuccesses
	}
}

// WithProducerRequiredAcks sets the required acknowledgments for producer
func WithProducerRequiredAcks(requiredAcks int) OptionFunc {
	return func(config *sarama.Config) {
		config.Producer.RequiredAcks = sarama.RequiredAcks(requiredAcks)
	}
}

// Consumer options

// WithConsumerReturnErrors configures the consumer to return errors
func WithConsumerReturnErrors(isReturnErrors bool) OptionFunc {
	return func(config *sarama.Config) {
		config.Consumer.Return.Errors = isReturnErrors
	}
}

// WithConsumerGroupRebalanceStrategy sets the consumer group rebalance strategy
func WithConsumerGroupRebalanceStrategy(strategy sarama.BalanceStrategy) OptionFunc {
	return func(config *sarama.Config) {
		config.Consumer.Group.Rebalance.Strategy = strategy
	}
}

// WithConsumerOffsetsInitial sets the initial offset for consumer
func WithConsumerOffsetsInitial(offset int64) OptionFunc {
	return func(config *sarama.Config) {
		config.Consumer.Offsets.Initial = offset
	}
}

// General options

// WithKafkaVersion sets the Kafka version
func WithKafkaVersion(version string) OptionFunc {
	return func(config *sarama.Config) {
		kafkaVersion, err := sarama.ParseKafkaVersion(version)
		if err != nil {
			// Default to a safe version if parsing fails
			// Use a default version string instead of a constant
			kafkaVersion, _ = sarama.ParseKafkaVersion("3.7.0")
		}
		config.Version = kafkaVersion
	}
}

// GetBalanceStrategy returns the appropriate balance strategy based on the string name
func GetBalanceStrategy(strategyName string) sarama.BalanceStrategy {
	switch strategyName {
	case "range":
		return sarama.BalanceStrategyRange
	case "roundrobin":
		return sarama.BalanceStrategyRoundRobin
	case "sticky":
		return sarama.BalanceStrategySticky
	default:
		return sarama.BalanceStrategyRange // Default to range strategy
	}
}
