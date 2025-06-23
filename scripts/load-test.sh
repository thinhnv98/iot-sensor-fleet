#!/bin/bash

# load-test.sh - Script to generate high load for the IoT sensor fleet
# This script increases the number of sensors and decreases the interval
# to generate approximately 5,000 messages per second

# Set environment variables for high load
export SENSOR_COUNT=1000      # 5,000 sensors
export SENSOR_INTERVAL=1s     # 1 second interval

echo "Starting load test with $SENSOR_COUNT sensors at $SENSOR_INTERVAL interval..."
echo "This will generate approximately 5,000 messages per second"
echo "Press Ctrl+C to stop the load test"

# Run the sensor producer with high load settings
cd "$(dirname "$0")/.." || exit
go run cmd/sensor-producer/main.go