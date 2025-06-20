package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/example/iot-sensor-fleet/internal/config"
)

// ElasticsearchDB represents an Elasticsearch connection
type ElasticsearchDB struct {
	url   string
	index string
}

// NewElasticsearchDB creates a new Elasticsearch connection
func NewElasticsearchDB(cfg *config.Config) *ElasticsearchDB {
	return &ElasticsearchDB{
		url:   cfg.ElasticsearchURL,
		index: cfg.ElasticsearchIndex,
	}
}

// InitIndex creates the necessary index if it doesn't exist
func (e *ElasticsearchDB) InitIndex() error {
	// Check if index exists
	resp, err := http.Head(fmt.Sprintf("%s/%s", e.url, e.index))
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	// If index exists, return
	if resp.StatusCode == http.StatusOK {
		log.Printf("Elasticsearch index '%s' already exists", e.index)
		return nil
	}

	// Create index with mapping
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"ts": map[string]interface{}{
					"type": "long",
				},
				"temperature": map[string]interface{}{
					"type": "float",
				},
				"humidity": map[string]interface{}{
					"type": "float",
				},
			},
		},
	}

	// Convert mapping to JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping to JSON: %w", err)
	}

	// Create index
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/%s", e.url, e.index),
		bytes.NewBuffer(mappingJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create index, status code: %d", resp.StatusCode)
	}

	log.Printf("Elasticsearch index '%s' created successfully", e.index)
	return nil
}
