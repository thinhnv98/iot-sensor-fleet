package model

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
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

// InitSchemaRegistry is kept for backward compatibility but does nothing
func InitSchemaRegistry(url string) {
	// No-op in JSON implementation
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

// SerializeSensorReading serializes a sensor reading to JSON format
func SerializeSensorReading(reading *SensorReading) ([]byte, error) {
	jsonData, err := json.Marshal(reading)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sensor reading to JSON: %w", err)
	}
	return jsonData, nil
}

// DeserializeSensorReading deserializes JSON data to a sensor reading
func DeserializeSensorReading(data []byte) (*SensorReading, error) {
	var reading SensorReading
	if err := json.Unmarshal(data, &reading); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to sensor reading: %w", err)
	}
	return &reading, nil
}

// SerializeSensorAlert serializes a sensor alert to JSON format
func SerializeSensorAlert(alert *SensorAlert) ([]byte, error) {
	jsonData, err := json.Marshal(alert)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sensor alert to JSON: %w", err)
	}
	return jsonData, nil
}

// DeserializeSensorAlert deserializes JSON data to a sensor alert
func DeserializeSensorAlert(data []byte) (*SensorAlert, error) {
	var alert SensorAlert
	if err := json.Unmarshal(data, &alert); err != nil {
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
