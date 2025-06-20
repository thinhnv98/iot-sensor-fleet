# IoT Sensor Fleet

A complete, runnable Go project that simulates an IoT environment using Apache Kafka 3.7.x. This project showcases Kafka producers, consumers, Kafka Connect sinks, Schema Registry, dead-letter handling, Prometheus metrics, and alert stream processing – all within Docker Compose.

## Architecture

```
                                                                 ┌───────────────┐
                                                                 │               │
                                                                 │  Prometheus   │
                                                                 │               │
                                                                 └───────┬───────┘
                                                                         │
                                                                         │ scrape
                                                                         │ metrics
                                                                         ▼
┌───────────────┐    produce     ┌───────────────┐    consume    ┌───────────────┐
│               │                │               │                │               │
│     Sensor    ├───────────────►│  sensor.raw   ├───────────────►│    Anomaly    │
│    Producer   │   Avro msgs    │    (topic)    │                │    Detector   │
│               │                │               │                │               │
└───────────────┘                └───────────────┘                └───────┬───────┘
                                         │                                │
                                         │                                │ produce
                                         │                                │ alerts
                                         │                                ▼
                                         │                        ┌───────────────┐
                                         │                        │               │
                                         │                        │  sensor.alert │
                                         │                        │    (topic)    │
                                         │                        │               │
                                         │                        └───────────────┘
                                         │
                                         │ sink
                                         ▼
                 ┌───────────────┬───────────────┬───────────────┐
                 │               │               │               │
                 │  PostgreSQL   │ Elasticsearch │   S3 (MinIO)  │
                 │               │               │               │
                 └───────────────┘               └───────────────┘
                                 └───────────────┘
```

## Features

- 1,000 virtual sensor devices (temperature & humidity) publish Avro-encoded messages every 2s to topic **sensor.raw**
- Kafka Streams service **anomaly-detector** reads **sensor.raw**, flags out-of-range values (> 50°C or RH < 10%), and writes alert events to **sensor.alert**
- Kafka Connect sinks:
  - **PostgreSQL** (table `sensor_readings`) for raw data
  - **Elasticsearch** (index `sensor_readings`) for search
  - **S3** (local MinIO) for cold storage
- Dead-letter topic **sensor.raw.dlt** for deserialization or processing errors
- Observability: Prometheus metrics from producers/consumers + Grafana dashboards
- Everything runs via `docker compose up -d`; zero external dependencies

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- Make (optional, for using the Makefile)

## Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/example/iot-sensor-fleet.git
   cd iot-sensor-fleet
   ```

2. Start the entire system with Docker Compose:
   ```bash
   docker-compose -f docker/docker-compose.yml up -d
   ```

   This will start:
   - Kafka ecosystem (Kafka, Zookeeper, Schema Registry, Kafka Connect)
   - Data storage (PostgreSQL, Elasticsearch, MinIO)
   - Monitoring (Prometheus, Grafana)
   - Sensor Producer (simulating IoT devices)
   - Anomaly Detector (processing sensor data)

3. Access the dashboards:
   - Grafana: http://localhost:3000 (admin/admin)
   - Kafka UI: http://localhost:8080
   - Prometheus: http://localhost:9090

4. Monitor the logs:
   ```bash
   # View logs from the sensor producer
   docker logs -f sensor-producer

   # View logs from the anomaly detector
   docker logs -f anomaly-detector
   ```

5. To run the services locally instead of in Docker:
   ```bash
   # Run the sensor producer locally
   go run cmd/sensor-producer/main.go

   # Run the anomaly detector locally
   go run cmd/anomaly-detector/main.go
   ```

## Using the Makefile

The project includes a Makefile for common operations:

```bash
# Build the project
make build

# Run the sensor producer locally
make run-producer

# Run the anomaly detector locally
make run-detector

# Run tests
make test

# Docker operations
make docker-build    # Build Docker images for sensor-producer and anomaly-detector
make docker-up       # Start all services with Docker Compose
make docker-down     # Stop all services
make docker-logs     # View logs from sensor-producer and anomaly-detector

# Generate high load (5,000 msgs/sec)
make load

# Simulate DLT messages
make simulate-dlt
```

## Configuration

The application is configured via environment variables (12-factor app):

| Variable | Description | Default |
|----------|-------------|---------|
| KAFKA_BROKERS | Comma-separated list of Kafka brokers | localhost:9092 |
| SCHEMA_REGISTRY_URL | URL of the Schema Registry | http://localhost:8081 |
| SENSOR_COUNT | Number of virtual sensors to simulate | 1000 |
| SENSOR_INTERVAL | Interval between sensor readings | 2s |
| MAX_TEMPERATURE | Maximum temperature threshold | 50.0 |
| MIN_HUMIDITY | Minimum humidity threshold | 10.0 |
| METRICS_PORT | Port for Prometheus metrics (producer) | 2112 |

## Sample Queries

### Kafka UI

Browse topics and messages at http://localhost:8080

### PostgreSQL

```bash
docker exec -it postgres psql -U postgres -d sensordb -c "SELECT * FROM sensor_readings LIMIT 10;"
```

### Elasticsearch

```bash
curl -X GET "http://localhost:9200/sensor_readings/_search?pretty" -H "Content-Type: application/json" -d'
{
  "query": {
    "range": {
      "temperature": {
        "gt": 45
      }
    }
  },
  "size": 5
}'
```

### MinIO

Access the MinIO console at http://localhost:9001 (minioadmin/minioadmin) to browse the sensor-cold bucket.

## Project Structure

```
.
├── cmd/
│   ├── sensor-producer/       # generates mock data
│   └── anomaly-detector/      # Kafka Streams app
├── internal/
│   ├── model/                 # Avro schemas (*.avsc) + Go structs
│   ├── kafka/                 # common producer/consumer helpers
│   ├── metrics/               # Prometheus collectors
│   └── config/                # env config loader
├── docker/
│   ├── docker-compose.yml
│   └── grafana/               # pre-baked dashboards JSON
├── scripts/
│   ├── load-test.sh           # spin 5k msg/s for stress
│   └── simulate-dlt.sh        # publish broken messages
├── README.md                  # architecture diagram & how-to-run
└── Makefile                   # make run-producer / make test / make load
```

## Testing

Run the unit tests:

```bash
go test ./...
```

### Load Testing

To simulate high load (approximately 5,000 messages per second):

```bash
./scripts/load-test.sh
```

### DLT Simulation

To simulate messages that will be sent to the dead-letter topic:

```bash
./scripts/simulate-dlt.sh
```

## Monitoring

The project includes comprehensive monitoring:

- **Prometheus** scrapes metrics from:
  - Sensor producer (port 2112)
  - Anomaly detector (port 2113)
  - Kafka brokers
  - Kafka Connect

- **Grafana Dashboards**:
  - IoT Sensor Overview: General metrics about the system
  - End-to-End Latency: Detailed latency metrics from production to alert

## License

MIT
