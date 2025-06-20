#!/bin/bash
set -e

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c '\q'; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done
echo "PostgreSQL is up and running!"

# Create tables in PostgreSQL
echo "Creating tables in PostgreSQL..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB << EOF
-- Create sensor_readings table if it doesn't exist
CREATE TABLE IF NOT EXISTS sensor_readings (
  id VARCHAR(36) PRIMARY KEY,
  ts BIGINT NOT NULL,
  temperature REAL NOT NULL,
  humidity REAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create sensor_alerts table if it doesn't exist
CREATE TABLE IF NOT EXISTS sensor_alerts (
  sensor_id VARCHAR(36) NOT NULL,
  ts BIGINT NOT NULL,
  reason TEXT NOT NULL,
  temperature REAL NOT NULL,
  humidity REAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (sensor_id, ts)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_sensor_readings_ts ON sensor_readings (ts);
CREATE INDEX IF NOT EXISTS idx_sensor_alerts_ts ON sensor_alerts (ts);

echo "Database initialization completed successfully!"