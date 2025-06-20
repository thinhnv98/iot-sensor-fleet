#!/bin/bash

# simulate-dlt.sh - Script to publish broken messages to simulate DLT handling
# This script uses the Kafka console producer to send invalid Avro messages
# to the sensor.raw topic, which should be redirected to the DLT by the anomaly detector

# Default Kafka broker
KAFKA_BROKER=${KAFKA_BROKER:-"localhost:9092"}
TOPIC=${TOPIC:-"sensor.raw"}
NUM_MESSAGES=${NUM_MESSAGES:-10}

echo "Simulating $NUM_MESSAGES broken messages to topic $TOPIC..."

# Function to generate a random UUID
generate_uuid() {
  cat /dev/urandom | tr -dc 'a-f0-9' | fold -w 32 | head -n 1 | 
    sed -r 's/(.{8})(.{4})(.{4})(.{4})(.{12})/\1-\2-\3-\4-\5/'
}

# Generate and send broken messages
for i in $(seq 1 $NUM_MESSAGES); do
  # Generate a random UUID for the message key
  UUID=$(generate_uuid)
  
  # Create an invalid Avro message (not conforming to the schema)
  # This is just random JSON that will fail Avro deserialization
  BROKEN_MSG="{\"id\":\"$UUID\",\"timestamp\":$(date +%s%3N),\"invalid_field\":\"this will break\",\"another_bad_field\":42}"
  
  echo "Sending broken message $i: $BROKEN_MSG"
  
  # Send the message using Kafka console producer
  echo "$BROKEN_MSG" | docker exec -i kafka kafka-console-producer \
    --bootstrap-server $KAFKA_BROKER \
    --topic $TOPIC \
    --property "parse.key=true" \
    --property "key.separator=:" <<< "$UUID:$BROKEN_MSG"
    
  # Small delay between messages
  sleep 0.5
done

echo "Finished sending $NUM_MESSAGES broken messages to topic $TOPIC"
echo "Check the sensor.raw.dlt topic and the anomaly detector logs for DLT handling"