package kafka

// Default configuration values
const (
	// Default retry configuration
	DefaultRetryMax     = 3
	DefaultRetryBackoff = 100 // milliseconds

	// Default producer configuration
	DefaultRequiredAcks       = 1 // WaitForLocal
	DefaultProducerReturnSucc = true

	// Default consumer configuration
	DefaultConsumerReturnErrors  = true
	DefaultConsumerOffsetInitial = -1      // OffsetNewest
	DefaultConsumerGroupStrategy = "range" // Range strategy

	// Default worker pool size
	DefaultWorkerPoolSize = 10

	// Default Kafka version
	DefaultKafkaVersion = "3.7.0" // Updated to match iot-sensor-fleet version
)

// RebalanceStrategyMap maps string names to sarama BalanceStrategy implementations
var RebalanceStrategyMap = map[string]string{
	"range":      "Range",
	"roundrobin": "RoundRobin",
	"sticky":     "Sticky",
}
