package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/riferrei/srclient"
)

// SensorReading represents a reading from an IoT sensor
type SensorReading struct {
	ID          string  `json:"id"`
	Timestamp   int64   `json:"ts"`
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
}

// SensorAlert represents an alert generated from an anomalous sensor reading
type SensorAlert struct {
	SensorID    string  `json:"sensor_id"`
	Timestamp   int64   `json:"ts"`
	Reason      string  `json:"reason"`
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
}

// Schema registry client and schema caches
var (
	schemaRegistryClient *srclient.SchemaRegistryClient
	readingSchemaOnce    sync.Once
	readingSchema        *srclient.Schema
	alertSchemaOnce      sync.Once
	alertSchema          *srclient.Schema
)

// InitSchemaRegistry initializes the schema registry client
func InitSchemaRegistry(url string) {
	schemaRegistryClient = srclient.CreateSchemaRegistryClient(url)
}

// GetSensorReadingSchema returns the schema for sensor readings
func GetSensorReadingSchema() (*srclient.Schema, error) {
	var err error
	readingSchemaOnce.Do(func() {
 	// Load schema from file
 	schemaPath := filepath.Join("internal", "model", "sensor_reading.avsc")
 	schemaBytes, readErr := os.ReadFile(schemaPath)
		if readErr != nil {
			err = fmt.Errorf("failed to read sensor reading schema: %w", readErr)
			return
		}

		// Register schema with Schema Registry
		readingSchema, err = schemaRegistryClient.CreateSchema("sensor.raw", string(schemaBytes), srclient.Avro)
		if err != nil {
			err = fmt.Errorf("failed to register sensor reading schema: %w", err)
			return
		}
	})

	if err != nil {
		return nil, err
	}
	return readingSchema, nil
}

// GetSensorAlertSchema returns the schema for sensor alerts
func GetSensorAlertSchema() (*srclient.Schema, error) {
	var err error
	alertSchemaOnce.Do(func() {
 	// Load schema from file
 	schemaPath := filepath.Join("internal", "model", "sensor_alert.avsc")
 	schemaBytes, readErr := os.ReadFile(schemaPath)
		if readErr != nil {
			err = fmt.Errorf("failed to read sensor alert schema: %w", readErr)
			return
		}

		// Register schema with Schema Registry
		alertSchema, err = schemaRegistryClient.CreateSchema("sensor.alert", string(schemaBytes), srclient.Avro)
		if err != nil {
			err = fmt.Errorf("failed to register sensor alert schema: %w", err)
			return
		}
	})

	if err != nil {
		return nil, err
	}
	return alertSchema, nil
}

// NewSensorReading creates a new sensor reading with a random UUID
func NewSensorReading(timestamp int64, temperature, humidity float32) *SensorReading {
	return &SensorReading{
		ID:          uuid.New().String(),
		Timestamp:   timestamp,
		Temperature: temperature,
		Humidity:    humidity,
	}
}

// NewSensorAlert creates a new sensor alert from a sensor reading
func NewSensorAlert(reading *SensorReading, reason string) *SensorAlert {
	return &SensorAlert{
		SensorID:    reading.ID,
		Timestamp:   reading.Timestamp,
		Reason:      reason,
		Temperature: reading.Temperature,
		Humidity:    reading.Humidity,
	}
}

// SerializeSensorReading serializes a sensor reading to Avro format
func SerializeSensorReading(reading *SensorReading) ([]byte, error) {
	schema, err := GetSensorReadingSchema()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(reading)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sensor reading to JSON: %w", err)
	}

	native, _, err := schema.Codec().NativeFromTextual(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to native: %w", err)
	}

	binary, err := schema.Codec().BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("failed to convert native to binary: %w", err)
	}

	return binary, nil
}

// DeserializeSensorReading deserializes Avro data to a sensor reading
func DeserializeSensorReading(data []byte) (*SensorReading, error) {
	schema, err := GetSensorReadingSchema()
	if err != nil {
		return nil, err
	}

	native, _, err := schema.Codec().NativeFromBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert binary to native: %w", err)
	}

	jsonData, err := schema.Codec().TextualFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("failed to convert native to JSON: %w", err)
	}

	var reading SensorReading
	if err := json.Unmarshal(jsonData, &reading); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to sensor reading: %w", err)
	}

	return &reading, nil
}

// SerializeSensorAlert serializes a sensor alert to Avro format
func SerializeSensorAlert(alert *SensorAlert) ([]byte, error) {
	schema, err := GetSensorAlertSchema()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(alert)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sensor alert to JSON: %w", err)
	}

	native, _, err := schema.Codec().NativeFromTextual(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to native: %w", err)
	}

	binary, err := schema.Codec().BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("failed to convert native to binary: %w", err)
	}

	return binary, nil
}

// DeserializeSensorAlert deserializes Avro data to a sensor alert
func DeserializeSensorAlert(data []byte) (*SensorAlert, error) {
	schema, err := GetSensorAlertSchema()
	if err != nil {
		return nil, err
	}

	native, _, err := schema.Codec().NativeFromBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert binary to native: %w", err)
	}

	jsonData, err := schema.Codec().TextualFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("failed to convert native to JSON: %w", err)
	}

	var alert SensorAlert
	if err := json.Unmarshal(jsonData, &alert); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to sensor alert: %w", err)
	}

	return &alert, nil
}

// ValidateSensorReading checks if a sensor reading is within valid ranges
// Returns true if valid, false if invalid
func ValidateSensorReading(reading *SensorReading) (bool, string) {
	if reading.Temperature > 50.0 {
		return false, "Temperature exceeds 50°C"
	}
	if reading.Humidity < 10.0 {
		return false, "Humidity below 10%"
	}
	return true, ""
}
