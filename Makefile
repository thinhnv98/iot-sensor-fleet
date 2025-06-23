# Makefile for IoT Sensor Fleet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
PRODUCER_BIN=sensor-producer
DETECTOR_BIN=anomaly-detector

# Source directories
PRODUCER_SRC=./cmd/sensor-producer
DETECTOR_SRC=./cmd/anomaly-detector

# Build directory
BUILD_DIR=./bin

# Docker compose file
DOCKER_COMPOSE=docker/docker-compose.yml

.PHONY: all build clean test run-producer run-detector docker-build docker-up docker-down docker-logs load simulate-dlt tidy docker

all: build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(PRODUCER_BIN) $(PRODUCER_SRC)
	$(GOBUILD) -o $(BUILD_DIR)/$(DETECTOR_BIN) $(DETECTOR_SRC)

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GOTEST) -v ./...

run-producer:
	$(GORUN) $(PRODUCER_SRC)/main.go

run-detector:
	$(GORUN) $(DETECTOR_SRC)/main.go

up:
	@if command -v docker compose >/dev/null 2>&1; then \
		echo "Using docker compose..."; \
		docker compose -f $(DOCKER_COMPOSE) up -d; \
	else \
		echo "Using docker-compose..."; \
		docker-compose -f $(DOCKER_COMPOSE) up -d; \
	fi

down:
	@if command -v docker compose >/dev/null 2>&1; then \
		echo "Using docker compose..."; \
		docker compose -f $(DOCKER_COMPOSE) down; \
	else \
		echo "Using docker-compose..."; \
		docker-compose -f $(DOCKER_COMPOSE) down; \
	fi

docker-build:
	@if command -v docker compose >/dev/null 2>&1; then \
		echo "Using docker compose..."; \
		docker compose -f $(DOCKER_COMPOSE) build sensor-producer anomaly-detector; \
	else \
		echo "Using docker-compose..."; \
		docker-compose -f $(DOCKER_COMPOSE) build sensor-producer anomaly-detector; \
	fi

docker-logs:
	@if command -v docker compose >/dev/null 2>&1; then \
		echo "Using docker compose..."; \
		docker compose -f $(DOCKER_COMPOSE) logs -f sensor-producer anomaly-detector; \
	else \
		echo "Using docker-compose..."; \
		docker-compose -f $(DOCKER_COMPOSE) logs -f sensor-producer anomaly-detector; \
	fi

# Alias for docker-up
docker: docker-up

load:
	chmod +x ./scripts/load-test.sh
	./scripts/load-test.sh

simulate-dlt:
	chmod +x ./scripts/simulate-dlt.sh
	./scripts/simulate-dlt.sh

tidy:
	$(GOMOD) tidy

# Ensure scripts are executable
scripts-executable:
	chmod +x ./scripts/*.sh
