package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/example/iot-sensor-fleet/internal/config"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// InitTables creates the necessary tables if they don't exist
func (p *PostgresDB) InitTables() error {
	// Create sensor_readings table
	_, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS sensor_readings (
			id VARCHAR(36) PRIMARY KEY,
			ts BIGINT NOT NULL,
			temperature REAL NOT NULL,
			humidity REAL NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sensor_readings table: %w", err)
	}

	// Create sensor_alerts table
	_, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS sensor_alerts (
			sensor_id VARCHAR(36) NOT NULL,
			ts BIGINT NOT NULL,
			reason TEXT NOT NULL,
			temperature REAL NOT NULL,
			humidity REAL NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (sensor_id, ts)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sensor_alerts table: %w", err)
	}

	// Create indexes for better query performance
	_, err = p.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sensor_readings_ts ON sensor_readings (ts);
		CREATE INDEX IF NOT EXISTS idx_sensor_alerts_ts ON sensor_alerts (ts);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("PostgreSQL tables initialized successfully")
	return nil
}
