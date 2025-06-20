package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Kafka configuration
	KafkaBrokers      []string
	KafkaVersion      string
	SchemaRegistryURL string

	// Topics
	TopicSensorRaw    string
	TopicSensorAlert  string
	TopicSensorRawDLT string

	// Producer configuration
	ProducerRequiredAcks  int
	ProducerReturnSuccess bool
	ProducerReturnErrors  bool

	// Consumer configuration
	ConsumerGroupID         string
	ConsumerOffsetInitial   int64
	ConsumerReturnErrors    bool
	ConsumerBalanceStrategy string

	// Sensor simulation configuration
	SensorCount    int
	SensorInterval time.Duration

	// HTTP server configuration
	MetricsPort int

	// Anomaly detector configuration
	MaxTemperature float32
	MinHumidity    float32

	// PostgreSQL configuration
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// Elasticsearch configuration
	ElasticsearchURL   string
	ElasticsearchIndex string

	// MinIO configuration
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		// Default values
		KafkaBrokers:      []string{"localhost:9092"},
		KafkaVersion:      "3.7.0",
		SchemaRegistryURL: "http://localhost:8081",

		TopicSensorRaw:    "sensor.raw",
		TopicSensorAlert:  "sensor.alert",
		TopicSensorRawDLT: "sensor.raw.dlt",

		ProducerRequiredAcks:  1, // WaitForLocal
		ProducerReturnSuccess: true,
		ProducerReturnErrors:  true,

		ConsumerGroupID:         "iot-sensor-group",
		ConsumerOffsetInitial:   -1, // OffsetNewest
		ConsumerReturnErrors:    true,
		ConsumerBalanceStrategy: "range",

		SensorCount:    1000,
		SensorInterval: 2 * time.Second,

		MetricsPort: 2112,

		MaxTemperature: 50.0,
		MinHumidity:    10.0,

		// PostgreSQL defaults
		PostgresHost:     "localhost",
		PostgresPort:     5432,
		PostgresUser:     "postgres",
		PostgresPassword: "postgres",
		PostgresDB:       "sensordb",

		// Elasticsearch defaults
		ElasticsearchURL:   "http://localhost:9200",
		ElasticsearchIndex: "sensor_readings",

		// MinIO defaults
		MinioEndpoint:  "localhost:9000",
		MinioAccessKey: "minioadmin",
		MinioSecretKey: "minioadmin",
		MinioBucket:    "sensor-cold",
	}

	// Override defaults with environment variables
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.KafkaBrokers = strings.Split(brokers, ",")
	}

	if version := os.Getenv("KAFKA_VERSION"); version != "" {
		config.KafkaVersion = version
	}

	if url := os.Getenv("SCHEMA_REGISTRY_URL"); url != "" {
		config.SchemaRegistryURL = url
	}

	if topic := os.Getenv("TOPIC_SENSOR_RAW"); topic != "" {
		config.TopicSensorRaw = topic
	}

	if topic := os.Getenv("TOPIC_SENSOR_ALERT"); topic != "" {
		config.TopicSensorAlert = topic
	}

	if topic := os.Getenv("TOPIC_SENSOR_RAW_DLT"); topic != "" {
		config.TopicSensorRawDLT = topic
	}

	if acks := os.Getenv("PRODUCER_REQUIRED_ACKS"); acks != "" {
		acksInt, err := strconv.Atoi(acks)
		if err != nil {
			return nil, fmt.Errorf("invalid PRODUCER_REQUIRED_ACKS: %w", err)
		}
		config.ProducerRequiredAcks = acksInt
	}

	if returnSuccess := os.Getenv("PRODUCER_RETURN_SUCCESS"); returnSuccess != "" {
		returnSuccessBool, err := strconv.ParseBool(returnSuccess)
		if err != nil {
			return nil, fmt.Errorf("invalid PRODUCER_RETURN_SUCCESS: %w", err)
		}
		config.ProducerReturnSuccess = returnSuccessBool
	}

	if returnErrors := os.Getenv("PRODUCER_RETURN_ERRORS"); returnErrors != "" {
		returnErrorsBool, err := strconv.ParseBool(returnErrors)
		if err != nil {
			return nil, fmt.Errorf("invalid PRODUCER_RETURN_ERRORS: %w", err)
		}
		config.ProducerReturnErrors = returnErrorsBool
	}

	if groupID := os.Getenv("CONSUMER_GROUP_ID"); groupID != "" {
		config.ConsumerGroupID = groupID
	}

	if offsetInitial := os.Getenv("CONSUMER_OFFSET_INITIAL"); offsetInitial != "" {
		offsetInitialInt, err := strconv.ParseInt(offsetInitial, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CONSUMER_OFFSET_INITIAL: %w", err)
		}
		config.ConsumerOffsetInitial = offsetInitialInt
	}

	if returnErrors := os.Getenv("CONSUMER_RETURN_ERRORS"); returnErrors != "" {
		returnErrorsBool, err := strconv.ParseBool(returnErrors)
		if err != nil {
			return nil, fmt.Errorf("invalid CONSUMER_RETURN_ERRORS: %w", err)
		}
		config.ConsumerReturnErrors = returnErrorsBool
	}

	if balanceStrategy := os.Getenv("CONSUMER_BALANCE_STRATEGY"); balanceStrategy != "" {
		config.ConsumerBalanceStrategy = strings.ToLower(balanceStrategy)
	}

	if sensorCount := os.Getenv("SENSOR_COUNT"); sensorCount != "" {
		sensorCountInt, err := strconv.Atoi(sensorCount)
		if err != nil {
			return nil, fmt.Errorf("invalid SENSOR_COUNT: %w", err)
		}
		config.SensorCount = sensorCountInt
	}

	if sensorInterval := os.Getenv("SENSOR_INTERVAL"); sensorInterval != "" {
		sensorIntervalDuration, err := time.ParseDuration(sensorInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid SENSOR_INTERVAL: %w", err)
		}
		config.SensorInterval = sensorIntervalDuration
	}

	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		metricsPortInt, err := strconv.Atoi(metricsPort)
		if err != nil {
			return nil, fmt.Errorf("invalid METRICS_PORT: %w", err)
		}
		config.MetricsPort = metricsPortInt
	}

	if maxTemperature := os.Getenv("MAX_TEMPERATURE"); maxTemperature != "" {
		maxTemperatureFloat, err := strconv.ParseFloat(maxTemperature, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_TEMPERATURE: %w", err)
		}
		config.MaxTemperature = float32(maxTemperatureFloat)
	}

	if minHumidity := os.Getenv("MIN_HUMIDITY"); minHumidity != "" {
		minHumidityFloat, err := strconv.ParseFloat(minHumidity, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid MIN_HUMIDITY: %w", err)
		}
		config.MinHumidity = float32(minHumidityFloat)
	}

	// PostgreSQL configuration
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		config.PostgresHost = host
	}

	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
		}
		config.PostgresPort = portInt
	}

	if user := os.Getenv("POSTGRES_USER"); user != "" {
		config.PostgresUser = user
	}

	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		config.PostgresPassword = password
	}

	if db := os.Getenv("POSTGRES_DB"); db != "" {
		config.PostgresDB = db
	}

	// Elasticsearch configuration
	if url := os.Getenv("ELASTICSEARCH_URL"); url != "" {
		config.ElasticsearchURL = url
	}

	if index := os.Getenv("ELASTICSEARCH_INDEX"); index != "" {
		config.ElasticsearchIndex = index
	}

	// MinIO configuration
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		config.MinioEndpoint = endpoint
	}

	if accessKey := os.Getenv("MINIO_ACCESS_KEY"); accessKey != "" {
		config.MinioAccessKey = accessKey
	}

	if secretKey := os.Getenv("MINIO_SECRET_KEY"); secretKey != "" {
		config.MinioSecretKey = secretKey
	}

	if bucket := os.Getenv("MINIO_BUCKET"); bucket != "" {
		config.MinioBucket = bucket
	}

	return config, nil
}
