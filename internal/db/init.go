package db

import (
	"log"

	"github.com/example/iot-sensor-fleet/internal/config"
)

// InitDatabases initializes all database connections and creates necessary tables and indexes
func InitDatabases(cfg *config.Config) error {
	// Initialize PostgreSQL
	log.Println("Initializing PostgreSQL...")
	postgres, err := NewPostgresDB(cfg)
	if err != nil {
		return err
	}
	defer postgres.Close()

	if err := postgres.InitTables(); err != nil {
		return err
	}

	// Initialize Elasticsearch
	log.Println("Initializing Elasticsearch...")
	elasticsearch := NewElasticsearchDB(cfg)
	if err := elasticsearch.InitIndex(); err != nil {
		return err
	}

	log.Println("All databases initialized successfully")
	return nil
}
