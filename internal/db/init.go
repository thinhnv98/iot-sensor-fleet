package db

import (
	"log"

	"github.com/example/iot-sensor-fleet/internal/config"
)

// InitDatabases initializes all database connections and creates necessary tables and indexes
// Returns the PostgreSQL connection that should be closed by the caller when done
func InitDatabases(cfg *config.Config) (*PostgresDB, error) {
	// Initialize PostgreSQL
	log.Println("Initializing PostgreSQL...")
	postgres, err := NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	if err := postgres.InitTables(); err != nil {
		postgres.Close()
		return nil, err
	}

	//// Initialize Elasticsearch
	//log.Println("Initializing Elasticsearch...")
	//elasticsearch := NewElasticsearchDB(cfg)
	//if err := elasticsearch.InitIndex(); err != nil {
	//	postgres.Close()
	//	return nil, err
	//}

	log.Println("All databases initialized successfully")
	return postgres, nil
}
